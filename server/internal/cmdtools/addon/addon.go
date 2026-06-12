// +----------------------------------------------------------------------
// | XYGo Admin 扩展安装/卸载逻辑 (V2 - 隔离式)
// +----------------------------------------------------------------------
// | 用途：从 addons/ 目录读取 ZIP 扩展包，执行安装或卸载
// |
// | V2 扩展包结构（新）：
// |   addons/{name}.zip 解压后包含：
// |     addon.yaml              -- 扩展元信息
// |     server/                 -- 后端代码 → 安装到 server/addons/{name}/
// |     web/                    -- 前端组件 → 安装到 web/src/addons/{name}/
// |     install/pgsql.sql       -- PgSQL 安装 SQL
// |     install/mysql.sql       -- MySQL 安装 SQL
// |     uninstall/pgsql.sql     -- PgSQL 卸载 SQL
// |     uninstall/mysql.sql     -- MySQL 卸载 SQL
// |     upgrade/pgsql.sql       -- PgSQL 增量升级 SQL（可选）
// |     upgrade/mysql.sql       -- MySQL 增量升级 SQL（可选）
// |
// | 安装目标路径映射：
// |   ZIP server/    → server/addons/{name}/
// |   ZIP web/       → web/src/addons/{name}/
// |   ZIP web-{n}/   → web-{name}/（独立前端，可选）
// |   addon.yaml     → server/addons/{name}/addon.yaml
// +----------------------------------------------------------------------

package addon

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"gopkg.in/yaml.v3"
)

const addonTable = "xy_addon"

var parsedLink string

type AddonMeta struct {
	Name           string     `yaml:"name"`
	Version        string     `yaml:"version"`
	Title          string     `yaml:"title"`
	Description    string     `yaml:"description"`
	Author         string     `yaml:"author"`
	MinVersion     string     `yaml:"min_version"`
	MinUpgradeFrom string     `yaml:"min_upgrade_from"`
	Changelog      []string   `yaml:"changelog"`
	Menus          AddonMenus `yaml:"menus"`
}

func Install(ctx context.Context, name string) error {
	if name == "" {
		name = gcmd.Scan("  请输入扩展包名称（如 tenant 或 tenant-1.0.0）: ")
		name = strings.TrimSpace(name)
		if name == "" {
			fmt.Println("  取消安装")
			return nil
		}
	}

	zipPath, err := resolveAddonZipPath("addons", name)
	if err != nil {
		return err
	}
	zipStem := strings.TrimSuffix(filepath.Base(zipPath), ".zip")

	fmt.Println()
	fmt.Printf("  正在安装扩展包: %s\n", zipStem)
	fmt.Println("  ────────────────────────────────")

	// [1/9] 解压到临时目录
	tmpDir := filepath.Join("addons", ".tmp", zipStem)
	os.RemoveAll(tmpDir)
	fmt.Print("  [1/9] 解压扩展包 ... ")
	if err := unzip(zipPath, tmpDir); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("解压失败: %v", err)
	}
	fmt.Println("OK")
	defer os.RemoveAll(filepath.Join("addons", ".tmp"))

	// 确定解压根目录（可能有一层包装目录）
	extractRoot := resolveExtractRoot(tmpDir, zipStem)

	// [2/9] 读取 addon.yaml
	fmt.Print("  [2/9] 读取扩展信息 ... ")
	metaPath := filepath.Join(extractRoot, "addon.yaml")
	if !gfile.Exists(metaPath) {
		fmt.Println("FAILED")
		return fmt.Errorf("addon.yaml 不存在")
	}
	metaData, _ := os.ReadFile(metaPath)
	var meta AddonMeta
	if err := yaml.Unmarshal(metaData, &meta); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("addon.yaml 解析失败: %v", err)
	}

	// 扩展身份以 addon.yaml 的 name 为准，避免 qrcode-0.1.0 被当成扩展名
	addonName := strings.TrimSpace(meta.Name)
	if addonName == "" {
		addonName = stripVersionZipStem(zipStem)
	}
	if addonName == "" {
		addonName = name
	}
	meta.Name = addonName

	if zipStem != addonName {
		fmt.Printf("OK (%s v%s, 扩展标识: %s)\n", meta.Title, meta.Version, addonName)
	} else {
		fmt.Printf("OK (%s v%s)\n", meta.Title, meta.Version)
	}

	// [3/9] 检查安装状态
	db := initDB(ctx)
	ensureAddonTable(ctx, db)
	installed, _ := db.GetOne(ctx, "SELECT * FROM "+addonTable+" WHERE name=$1 AND status=1", addonName)

	isUpgrade := false
	oldVersion := ""
	if installed != nil && !installed.IsEmpty() {
		installedVer := installed["version"].String()
		oldVersion = installedVer
		cmp := compareVersion(meta.Version, installedVer)
		if cmp > 0 {
			if meta.MinUpgradeFrom != "" && compareVersion(installedVer, meta.MinUpgradeFrom) < 0 {
				fmt.Printf("  [3/9] 当前版本 v%s 低于最低升级要求 v%s\n", installedVer, meta.MinUpgradeFrom)
				return fmt.Errorf("请先升级到 v%s 再安装此包", meta.MinUpgradeFrom)
			}
			fmt.Printf("  [3/9] 检测到已安装 v%s，新版本 v%s\n", installedVer, meta.Version)
			confirm := gcmd.Scan(fmt.Sprintf("  确认升级 v%s → v%s？[Y/n] ", installedVer, meta.Version))
			if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
				fmt.Println("  取消升级")
				return nil
			}
			isUpgrade = true
		} else if cmp == 0 {
			confirm := gcmd.Scan(fmt.Sprintf("  [3/9] 扩展 %s 已是 v%s，是否覆盖重装？[y/N] ", addonName, installedVer))
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Println("  取消安装")
				return nil
			}
		} else {
			fmt.Printf("  [3/9] 当前已安装 v%s，包版本 v%s 更低，不支持降级\n", installedVer, meta.Version)
			return nil
		}
	} else {
		fmt.Println("  [3/9] 全新安装")
	}

	// [4/9] 执行数据库变更
	dbType := detectDBType(ctx)
	var sqlDir string
	if isUpgrade {
		fmt.Print("  [4/9] 执行增量升级 SQL ... ")
		sqlDir = filepath.Join(extractRoot, "upgrade")
	} else {
		fmt.Print("  [4/9] 执行数据库变更 ... ")
		sqlDir = filepath.Join(extractRoot, "install")
	}
	sqlFile := filepath.Join(sqlDir, dbType+".sql")
	if gfile.Exists(sqlFile) {
		if err := execSQLFile(ctx, db, sqlFile); err != nil {
			fmt.Println("FAILED")
			return err
		}
		fmt.Println("OK")
	} else {
		if isUpgrade {
			fmt.Println("SKIP (无升级 SQL)")
		} else {
			fmt.Println("SKIP (无安装 SQL)")
		}
	}

	// [5/9] 安装菜单（从 addon.yaml 声明式写入）
	adminMenuCount := len(collectAllNodes(meta.Menus.Admin))
	tenantMenuCount := len(collectAllNodes(meta.Menus.Tenant))
	totalMenuCount := adminMenuCount + tenantMenuCount
	if totalMenuCount > 0 {
		fmt.Printf("  [5/9] 安装扩展菜单 (%d 项) ... ", totalMenuCount)
		if err := installMenus(ctx, db, addonName, meta.Menus); err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("菜单安装失败: %v", err)
		}
		fmt.Println("OK")
	} else {
		fmt.Println("  [5/9] 安装扩展菜单 ... SKIP (无菜单声明)")
	}

	// [6/9] 升级前备份旧文件
	projectRoot := getProjectRoot()
	if isUpgrade && oldVersion != "" {
		fmt.Print("  [6/9] 备份旧版本文件 ... ")
		backupBase := filepath.Join(projectRoot, "server", "addons", ".backup", fmt.Sprintf("%s-%s", addonName, oldVersion))
		os.MkdirAll(backupBase, 0755)
		backupCount := 0

		serverOld := filepath.Join(projectRoot, "server", "addons", addonName)
		if gfile.Exists(serverOld) {
			serverBackup := filepath.Join(backupBase, "server")
			os.MkdirAll(serverBackup, 0755)
			n, _ := copyDir(serverOld, serverBackup)
			backupCount += n
		}

		webOld := filepath.Join(projectRoot, "web", "src", "addons", addonName)
		if gfile.Exists(webOld) {
			webBackup := filepath.Join(backupBase, "web")
			os.MkdirAll(webBackup, 0755)
			n, _ := copyDir(webOld, webBackup)
			backupCount += n
		}
		fmt.Printf("OK (%d 个文件 → %s)\n", backupCount, filepath.ToSlash(backupBase))
	} else {
		fmt.Println("  [6/9] 备份旧版本文件 ... SKIP (全新安装)")
	}

	// [7/9] 复制文件到隔离目录
	fmt.Print("  [7/9] 复制扩展文件 ... ")
	copiedCount := 0

	// server/ → server/addons/{name}/
	serverSrc := filepath.Join(extractRoot, "server")
	if gfile.Exists(serverSrc) {
		serverDest := filepath.Join(projectRoot, "server", "addons", addonName)
		os.MkdirAll(serverDest, 0755)
		n, err := copyDir(serverSrc, serverDest)
		if err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("复制 server/ 失败: %v", err)
		}
		copiedCount += n
	}

	// 复制 addon.yaml 到 server/addons/{name}/addon.yaml
	addonYamlDest := filepath.Join(projectRoot, "server", "addons", addonName, "addon.yaml")
	os.MkdirAll(filepath.Dir(addonYamlDest), 0755)
	_ = copyFile(metaPath, addonYamlDest)

	// web/ → web/src/addons/{name}/
	webSrc := filepath.Join(extractRoot, "web")
	if gfile.Exists(webSrc) {
		webDest := filepath.Join(projectRoot, "web", "src", "addons", addonName)
		os.MkdirAll(webDest, 0755)
		n, err := copyDir(webSrc, webDest)
		if err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("复制 web/ 失败: %v", err)
		}
		copiedCount += n
	}

	// web-{name}/ → web-{name}/（独立前端，可选）
	webIndepSrc := filepath.Join(extractRoot, "web-"+addonName)
	if gfile.Exists(webIndepSrc) {
		webIndepDest := filepath.Join(projectRoot, "web-"+addonName)
		os.MkdirAll(webIndepDest, 0755)
		n, err := copyDir(webIndepSrc, webIndepDest)
		if err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("复制 web-%s/ 失败: %v", addonName, err)
		}
		copiedCount += n
	}

	fmt.Printf("OK (%d 个文件)\n", copiedCount)

	// [8/9] 更新 addons/addons.go
	fmt.Print("  [8/9] 更新扩展导入文件 ... ")
	addonsDir := filepath.Join(projectRoot, "server", "addons")
	if err := updateAddonsImport(addonsDir); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("更新 addons.go 失败: %v", err)
	}
	fmt.Println("OK")

	// [9/9] 记录安装信息
	fmt.Print("  [9/9] 记录安装信息 ... ")
	_, _ = db.Exec(ctx, "DELETE FROM "+addonTable+" WHERE name=$1", addonName)
	_, err = db.Exec(ctx,
		"INSERT INTO "+addonTable+" (name, version, title, status, installed_at) VALUES ($1, $2, $3, 1, $4)",
		addonName, meta.Version, meta.Title, time.Now().Unix(),
	)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("OK")

	// 打印后续操作
	fmt.Println()
	fmt.Println("  ════════════════════════════════════════")
	action := "安装"
	if isUpgrade {
		action = "升级"
	}
	fmt.Printf("  扩展 %s v%s %s完成！\n", meta.Title, meta.Version, action)
	fmt.Println("  ════════════════════════════════════════")

	if isUpgrade && len(meta.Changelog) > 0 {
		fmt.Println()
		fmt.Printf("  更新日志 (v%s):\n", meta.Version)
		for _, entry := range meta.Changelog {
			fmt.Printf("    - %s\n", entry)
		}
	}

	fmt.Println()
	fmt.Println("  扩展文件已安装到隔离目录：")
	fmt.Printf("    后端: server/addons/%s/\n", addonName)
	fmt.Printf("    前端: web/src/addons/%s/\n", addonName)
	fmt.Println()
	fmt.Println("  请依次执行以下操作：")
	fmt.Println("    1. gf gen dao        (重新生成数据模型)")
	fmt.Println("    2. gf gen service    (重新生成服务接口)")
	fmt.Println("    3. 重启后端服务")
	fmt.Println("    4. 前端: cd web && pnpm install && pnpm run build (如有前端变更)")
	fmt.Println()

	return nil
}

func Uninstall(ctx context.Context, name string) error {
	if name == "" {
		name = gcmd.Scan("  请输入要卸载的扩展名称（如 tenant）: ")
		name = strings.TrimSpace(name)
		if name == "" {
			fmt.Println("  取消卸载")
			return nil
		}
	}

	db := initDB(ctx)
	ensureAddonTable(ctx, db)

	// [1/5] 确认
	record, _ := db.GetOne(ctx, "SELECT * FROM "+addonTable+" WHERE name=$1 AND status=1", name)
	if record == nil || record.IsEmpty() {
		return fmt.Errorf("扩展 %s 未安装", name)
	}

	fmt.Println()
	fmt.Printf("  准备卸载扩展: %s (v%s)\n", record["title"].String(), record["version"].String())
	fmt.Println("  ────────────────────────────────")
	confirm := gcmd.Scan("  警告：卸载将删除扩展代码和数据表，此操作不可恢复！确认卸载？[y/N] ")
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("  取消卸载")
		return nil
	}

	// [2/6] 执行卸载 SQL
	fmt.Print("  [2/6] 执行数据库回滚 ... ")
	dbType := detectDBType(ctx)
	projectRoot := getProjectRoot()

	// 从已安装的 addon.yaml 同级 uninstall/ 读取，或从 ZIP 包读取
	uninstallSQL := ""
	addonDir := filepath.Join(projectRoot, "server", "addons", name)
	sqlFile := filepath.Join(addonDir, "uninstall", dbType+".sql")
	if gfile.Exists(sqlFile) {
		data, _ := os.ReadFile(sqlFile)
		uninstallSQL = string(data)
	} else {
		// 回退：从 ZIP 包中提取（支持 name.zip 与 name-version.zip）
		if zipPath, err := resolveAddonZipPath("addons", name); err == nil {
			uninstallSQL = extractSQLFromZip(zipPath, name, "uninstall/"+dbType+".sql")
		}
	}

	if uninstallSQL != "" {
		stmts := splitStatements(uninstallSQL)
		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || isCommentOnly(stmt) {
				continue
			}
			if _, err := db.Exec(ctx, stmt); err != nil {
				fmt.Printf("WARNING: %v\n", err)
			}
		}
		fmt.Println("OK")
	} else {
		fmt.Println("SKIP (无卸载 SQL)")
	}

	// [3/6] 卸载扩展菜单
	fmt.Print("  [3/6] 卸载扩展菜单 ... ")
	menuCount, menuErr := uninstallMenus(ctx, db, name)
	if menuErr != nil {
		fmt.Printf("WARNING: %v\n", menuErr)
	} else if menuCount > 0 {
		fmt.Printf("OK (删除 %d 条菜单)\n", menuCount)
	} else {
		fmt.Println("SKIP (无扩展菜单)")
	}

	// [4/6] 删除扩展目录
	fmt.Print("  [4/6] 删除扩展目录 ... ")
	removedDirs := []string{}

	serverDir := filepath.Join(projectRoot, "server", "addons", name)
	if gfile.Exists(serverDir) {
		os.RemoveAll(serverDir)
		removedDirs = append(removedDirs, "server/addons/"+name)
	}

	webDir := filepath.Join(projectRoot, "web", "src", "addons", name)
	if gfile.Exists(webDir) {
		os.RemoveAll(webDir)
		removedDirs = append(removedDirs, "web/src/addons/"+name)
	}

	webIndepDir := filepath.Join(projectRoot, "web-"+name)
	if gfile.Exists(webIndepDir) {
		os.RemoveAll(webIndepDir)
		removedDirs = append(removedDirs, "web-"+name)
	}

	fmt.Printf("OK (已删除 %d 个目录)\n", len(removedDirs))

	// [5/6] 更新 addons/addons.go
	fmt.Print("  [5/6] 更新扩展导入文件 ... ")
	addonsDir := filepath.Join(projectRoot, "server", "addons")
	if err := updateAddonsImport(addonsDir); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("更新 addons.go 失败: %v", err)
	}
	fmt.Println("OK")

	// [6/6] 更新安装记录
	fmt.Print("  [6/6] 更新安装记录 ... ")
	_, _ = db.Exec(ctx,
		"UPDATE "+addonTable+" SET status=0, uninstalled_at=$1 WHERE name=$2",
		time.Now().Unix(), name,
	)
	fmt.Println("OK")

	fmt.Println()
	fmt.Println("  ════════════════════════════════════════")
	fmt.Printf("  扩展 %s 已卸载\n", name)
	fmt.Println("  ════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  请依次执行以下操作：")
	fmt.Println("    1. gf gen dao        (重新生成数据模型)")
	fmt.Println("    2. gf gen service    (重新生成服务接口)")
	fmt.Println("    3. 重启后端服务")
	fmt.Println()

	return nil
}

// ==================== addons.go 自动维护 ====================

// updateAddonsImport 扫描 addonsDir 下所有子目录，
// 如果子目录中包含 .go 文件则视为有效扩展模块，重新生成聚合导入文件。
func updateAddonsImport(addonsDir string) error {
	entries, err := os.ReadDir(addonsDir)
	if err != nil {
		return err
	}

	var addonNames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		if dirName == ".tmp" || dirName == ".backup" || strings.HasPrefix(dirName, ".") {
			continue
		}
		// 检查目录中是否有 .go 文件
		subDir := filepath.Join(addonsDir, dirName)
		hasGo := false
		subEntries, _ := os.ReadDir(subDir)
		for _, f := range subEntries {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
				hasGo = true
				break
			}
		}
		if hasGo {
			addonNames = append(addonNames, dirName)
		}
	}

	sort.Strings(addonNames)

	var content string
	if len(addonNames) == 0 {
		content = "// Code generated and maintained by addon installer. DO NOT EDIT.\n\npackage addons\n"
	} else {
		imports := make([]string, len(addonNames))
		for i, n := range addonNames {
			imports[i] = fmt.Sprintf("\t_ \"xygo/addons/%s\"", n)
		}
		content = fmt.Sprintf(
			"// Code generated and maintained by addon installer. DO NOT EDIT.\n\npackage addons\n\nimport (\n%s\n)\n",
			strings.Join(imports, "\n"),
		)
	}

	return os.WriteFile(filepath.Join(addonsDir, "addons.go"), []byte(content), 0644)
}

// ==================== 工具函数 ====================

// resolveExtractRoot 确定解压后的实际根目录
// ZIP 可能在顶层有一个与 name 同名的包装目录
func resolveExtractRoot(tmpDir, name string) string {
	// 优先检查是否有 addon.yaml 直接在 tmpDir
	if gfile.Exists(filepath.Join(tmpDir, "addon.yaml")) {
		return tmpDir
	}
	// 检查 tmpDir/{name}/addon.yaml
	nested := filepath.Join(tmpDir, name)
	if gfile.Exists(filepath.Join(nested, "addon.yaml")) {
		return nested
	}
	// 尝试找第一个含 addon.yaml 的子目录
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if e.IsDir() {
			candidate := filepath.Join(tmpDir, e.Name())
			if gfile.Exists(filepath.Join(candidate, "addon.yaml")) {
				return candidate
			}
		}
	}
	return tmpDir
}

// extractSQLFromZip 从 ZIP 包中直接读取指定 SQL 文件内容
func extractSQLFromZip(zipPath, addonName, relPath string) string {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return ""
	}
	defer r.Close()

	candidates := []string{
		relPath,
		addonName + "/" + relPath,
	}

	for _, f := range r.File {
		fName := strings.ReplaceAll(f.Name, "\\", "/")
		for _, c := range candidates {
			if fName == c || strings.HasSuffix(fName, "/"+c) {
				rc, err := f.Open()
				if err != nil {
					continue
				}
				data, _ := io.ReadAll(rc)
				rc.Close()
				return string(data)
			}
		}
	}
	return ""
}

// copyDir 递归复制目录，返回复制的文件数量
func copyDir(src, dst string) (int, error) {
	count := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		os.MkdirAll(filepath.Dir(destPath), 0755)
		if err := copyFile(path, destPath); err != nil {
			return fmt.Errorf("复制 %s 失败: %v", relPath, err)
		}
		count++
		return nil
	})
	return count, err
}

func initDB(ctx context.Context) gdb.DB {
	configPath := findConfigPath()
	if configPath == "" {
		panic("配置文件未找到 (搜索: manifest/config/config.yaml)")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败: %v", err))
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(content, &raw); err != nil {
		panic(fmt.Sprintf("解析 YAML 失败: %v\n  文件: %s", err, configPath))
	}
	dbSection, _ := raw["database"].(map[string]interface{})
	if dbSection == nil {
		panic(fmt.Sprintf("配置中缺少 database 节点\n  文件: %s\n  顶层 key: %v", configPath, mapKeys(raw)))
	}
	defaultSection, _ := dbSection["default"].(map[string]interface{})
	if defaultSection == nil {
		panic(fmt.Sprintf("配置中缺少 database.default 节点\n  database 下的 key: %v", mapKeys(dbSection)))
	}
	link, _ := defaultSection["link"].(string)
	if link == "" {
		panic(fmt.Sprintf("database.default.link 为空\n  default 下的 key: %v", mapKeys(defaultSection)))
	}
	prefix, _ := defaultSection["Prefix"].(string)
	if prefix == "" {
		prefix, _ = defaultSection["prefix"].(string)
	}
	debug, _ := defaultSection["debug"].(bool)

	parsedLink = link
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			{Link: link, Prefix: prefix, Debug: debug},
		},
	})
	db, err := gdb.Instance()
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v\n  link: %s", err, link))
	}
	return db
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func findConfigPath() string {
	for _, c := range []string{"manifest/config/config.yaml", "../manifest/config/config.yaml"} {
		if gfile.Exists(c) {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return ""
}

func detectDBType(ctx context.Context) string {
	link := parsedLink
	if link == "" {
		return "pgsql"
	}
	if strings.HasPrefix(strings.ToLower(link), "pgsql:") || strings.HasPrefix(strings.ToLower(link), "postgres:") {
		return "pgsql"
	}
	return "mysql"
}

func getProjectRoot() string {
	if gfile.Exists("main.go") {
		abs, _ := filepath.Abs("..")
		return abs
	}
	abs, _ := filepath.Abs(".")
	return abs
}

func ensureAddonTable(ctx context.Context, db gdb.DB) {
	dbType := detectDBType(ctx)
	var sql string
	if dbType == "pgsql" {
		sql = `CREATE TABLE IF NOT EXISTS xy_addon (
			id bigserial PRIMARY KEY,
			name varchar(64) NOT NULL UNIQUE,
			version varchar(32) NOT NULL DEFAULT '',
			title varchar(128) NOT NULL DEFAULT '',
			status smallint NOT NULL DEFAULT 1,
			installed_at bigint NOT NULL DEFAULT 0,
			uninstalled_at bigint NOT NULL DEFAULT 0
		)`
	} else {
		sql = "CREATE TABLE IF NOT EXISTS `xy_addon` (" +
			"`id` bigint UNSIGNED NOT NULL AUTO_INCREMENT," +
			"`name` varchar(64) NOT NULL," +
			"`version` varchar(32) NOT NULL DEFAULT ''," +
			"`title` varchar(128) NOT NULL DEFAULT ''," +
			"`status` tinyint NOT NULL DEFAULT 1," +
			"`installed_at` bigint NOT NULL DEFAULT 0," +
			"`uninstalled_at` bigint NOT NULL DEFAULT 0," +
			"PRIMARY KEY (`id`)," +
			"UNIQUE KEY `uk_name` (`name`)" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
	}
	db.Exec(ctx, sql)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法路径: %s", f.Name)
		}
		if f.FileInfo().IsDir() || f.UncompressedSize64 == 0 && strings.HasSuffix(f.Name, "/") {
			os.MkdirAll(fpath, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(fpath), 0755)
		if f.UncompressedSize64 == 0 {
			os.MkdirAll(fpath, 0755)
			continue
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func splitStatements(sql string) []string {
	var stmts []string
	var current strings.Builder
	inDelimiter := false
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)
		if upper == "DELIMITER $$" {
			inDelimiter = true
			continue
		}
		if upper == "DELIMITER ;" {
			inDelimiter = false
			if s := strings.TrimSpace(current.String()); s != "" {
				stmts = append(stmts, s)
				current.Reset()
			}
			continue
		}
		if inDelimiter {
			if strings.HasSuffix(trimmed, "$$") {
				current.WriteString(strings.TrimSuffix(trimmed, "$$"))
				stmts = append(stmts, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteString(line + "\n")
			}
		} else {
			if trimmed == "" {
				if s := strings.TrimSpace(current.String()); s != "" && isCommentOnly(s) {
					current.Reset()
				}
				current.WriteString(line + "\n")
			} else if strings.HasSuffix(trimmed, ";") && !strings.HasPrefix(upper, "--") {
				current.WriteString(line + "\n")
				s := strings.TrimSpace(current.String())
				s = strings.TrimSuffix(s, ";")
				if s != "" {
					stmts = append(stmts, s)
				}
				current.Reset()
			} else {
				current.WriteString(line + "\n")
			}
		}
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		s = strings.TrimSuffix(s, ";")
		if s != "" {
			stmts = append(stmts, s)
		}
	}
	return stmts
}

func isCommentOnly(stmt string) bool {
	for _, line := range strings.Split(stmt, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			return false
		}
	}
	return true
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func execSQLFile(ctx context.Context, db gdb.DB, sqlFile string) error {
	sqlContent, _ := os.ReadFile(sqlFile)
	stmts := splitStatements(string(sqlContent))
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || isCommentOnly(stmt) {
			continue
		}
		if _, err := db.Exec(ctx, stmt); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "already exists") ||
				strings.Contains(errMsg, "已存在") ||
				strings.Contains(errMsg, "多个主键") ||
				strings.Contains(errMsg, "duplicate key") ||
				strings.Contains(errMsg, "重复键") {
				continue
			}
			return fmt.Errorf("SQL 执行失败: %v\n  语句: %s", err, truncate(stmt, 200))
		}
	}
	return nil
}

// resolveAddonZipPath 解析扩展 ZIP 路径，支持 name.zip 与 name-version.zip。
func resolveAddonZipPath(addonsDir, input string) (string, error) {
	direct := filepath.Join(addonsDir, input+".zip")
	if gfile.Exists(direct) {
		return direct, nil
	}

	pattern := filepath.Join(addonsDir, input+"-*.zip")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("查找扩展包失败: %v", err)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, filepath.Base(m))
		}
		sort.Strings(names)
		return "", fmt.Errorf("扩展包 %q 存在多个版本: %s，请指定完整包名（如 %s）",
			input, strings.Join(names, ", "), strings.TrimSuffix(names[len(names)-1], ".zip"))
	}

	return "", fmt.Errorf("扩展包不存在: 已查找 %s 和 %s",
		filepath.Join(addonsDir, input+".zip"),
		filepath.Join(addonsDir, input+"-*.zip"))
}

// stripVersionZipStem 从 ZIP 文件名推断扩展名，如 qrcode-0.1.0 -> qrcode。
func stripVersionZipStem(zipStem string) string {
	if i := strings.LastIndex(zipStem, "-"); i > 0 {
		suffix := zipStem[i+1:]
		if isVersionLike(suffix) {
			return zipStem[:i]
		}
	}
	return zipStem
}

func isVersionLike(s string) bool {
	if s == "" {
		return false
	}
	for _, p := range strings.Split(s, ".") {
		if p == "" {
			return false
		}
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

func compareVersion(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	maxLen := len(ap)
	if len(bp) > maxLen {
		maxLen = len(bp)
	}
	for i := 0; i < maxLen; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			fmt.Sscanf(ap[i], "%d", &av)
		}
		if i < len(bp) {
			fmt.Sscanf(bp[i], "%d", &bv)
		}
		if av != bv {
			return av - bv
		}
	}
	return 0
}
