// +----------------------------------------------------------------------
// | XYGo Admin 在线更新逻辑
// +----------------------------------------------------------------------
// | 用途：从远程检查更新、下载更新包、对比冲突、应用文件变更、执行数据库迁移
// +----------------------------------------------------------------------

package updater

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"xygo/internal/cmdtools/migrate"

	"gopkg.in/yaml.v3"
)

var addonSkipPrefixes = []string{
	"server/addons/",
	"web/src/addons/",
}

func isAddonPath(p string) bool {
	normalized := strings.ReplaceAll(p, "\\", "/")
	for _, prefix := range addonSkipPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			rest := normalized[len(prefix):]
			if strings.Contains(rest, "/") {
				return true
			}
		}
	}
	return false
}

const (
	defaultIndexURL = "https://xygoupload.xingyunwangluo.com/updates/open/update-index.json"
	httpTimeout     = 30 * time.Second
)

// ==================== 数据结构 ====================

type VersionInfo struct {
	Version   string `json:"version"   yaml:"version"`
	Edition   string `json:"edition"   yaml:"edition"`
	UpdatedAt string `json:"updated_at" yaml:"updated_at"`
}

type UpdateIndex struct {
	Latest   string                      `json:"latest"`
	Editions map[string]EditionUpdateList `json:"editions"`
}

type EditionUpdateList struct {
	Updates []UpdateEntry `json:"updates"`
}

type UpdateEntry struct {
	Version      string   `json:"version"`
	From         string   `json:"from"`
	Title        string   `json:"title"`
	Changelog    []string `json:"changelog"`
	URL          string   `json:"url"`
	Mirrors      []string `json:"mirrors"`
	HasMigration bool     `json:"has_migration"`
}

type UpdateMeta struct {
	Version      string   `yaml:"version"`
	From         string   `yaml:"from"`
	Title        string   `yaml:"title"`
	Changelog    []string `yaml:"changelog"`
	HasMigration bool     `yaml:"has_migration"`
	DeletedFiles []string `yaml:"deleted_files"`
}

// ==================== 主入口 ====================

// isVersionMarkerFile 项目根目录的版本标记文件，更新时必须覆盖，不参与冲突交互。
func isVersionMarkerFile(relPath string) bool {
	return filepath.ToSlash(relPath) == "version.json"
}

// RunUpdate 执行在线更新。若成功应用至少一个版本补丁，applied 为 true。
func RunUpdate(ctx context.Context) (applied bool, err error) {
	projectRoot := getProjectRoot()

	// 1. 读取本地版本
	localVersion, err := readLocalVersion(projectRoot)
	if err != nil {
		return false, fmt.Errorf("读取本地版本失败: %v", err)
	}

	fmt.Println()
	fmt.Println("  ════════════════════════════════════════")
	fmt.Println("  XYGo Admin 在线更新")
	fmt.Println("  ════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  当前版本: v%s (%s)\n", localVersion.Version, localVersion.Edition)
	fmt.Print("  检查更新中 ... ")

	// 2. 获取远程更新索引
	index, err := fetchUpdateIndex()
	if err != nil {
		fmt.Println("FAILED")
		fmt.Printf("  错误: %v\n", err)
		fmt.Printf("  请求地址: %s\n", defaultIndexURL)
		return false, err
	}
	fmt.Println("OK")

	// 3. 找到适用的更新
	edition, ok := index.Editions[localVersion.Edition]
	if !ok {
		fmt.Printf("  远程索引中无 %s 版本的更新\n", localVersion.Edition)
		return false, nil
	}

	var pending []UpdateEntry
	for _, u := range edition.Updates {
		if compareVersion(u.Version, localVersion.Version) > 0 {
			pending = append(pending, u)
		}
	}

	if len(pending) == 0 {
		fmt.Printf("  已是最新版本 v%s\n", localVersion.Version)
		return false, nil
	}

	// 4. 展示更新内容
	fmt.Printf("\n  发现 %d 个可用更新:\n", len(pending))
	for _, u := range pending {
		fmt.Printf("\n  v%s - %s\n", u.Version, u.Title)
		for _, cl := range u.Changelog {
			fmt.Printf("    - %s\n", cl)
		}
	}

	// 5. 逐个版本更新
	for _, u := range pending {
		fmt.Printf("\n  ──── 更新到 v%s ────\n", u.Version)

		confirm := gcmd.Scan(fmt.Sprintf("  确认更新到 v%s？[Y/n] ", u.Version))
		if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
			fmt.Println("  跳过此版本")
			continue
		}

		if err := applyUpdate(ctx, projectRoot, localVersion, u); err != nil {
			return applied, fmt.Errorf("更新到 v%s 失败: %v", u.Version, err)
		}

		localVersion.Version = u.Version
		localVersion.UpdatedAt = time.Now().Format("2006-01-02")
		saveLocalVersion(projectRoot, localVersion)
		applied = true
	}

	if applied {
		fmt.Println()
		fmt.Println("  ════════════════════════════════════════")
		fmt.Printf("  更新完成！当前版本: v%s\n", localVersion.Version)
		fmt.Println("  ════════════════════════════════════════")
		fmt.Println()
		fmt.Println("  旧文件备份位置: server/resource/update/backup/")
		fmt.Println()
		fmt.Println("  请依次执行:")
		fmt.Println("    1. gf gen dao")
		fmt.Println("    2. gf gen service")
		fmt.Println("    3. 重启服务")
		fmt.Println()
	}

	return applied, nil
}

// ==================== 单版本更新逻辑 ====================

func applyUpdate(ctx context.Context, projectRoot string, local *VersionInfo, entry UpdateEntry) error {
	// 1. 下载更新包
	fmt.Print("  [1/5] 下载更新包 ... ")
	updateDir := filepath.Join(projectRoot, "server", "resource", "update")
	zipPath := filepath.Join(updateDir, "tmp", entry.Version+".zip")
	os.MkdirAll(filepath.Dir(zipPath), 0755)
	defer os.RemoveAll(filepath.Join(updateDir, "tmp"))

	if err := downloadFile(entry.URL, entry.Mirrors, zipPath); err != nil {
		fmt.Println("FAILED")
		fmt.Printf("  错误: %v\n", err)
		return err
	}
	fmt.Println("OK")

	// 2. 解压
	fmt.Print("  [2/5] 解压更新包 ... ")
	tmpDir := filepath.Join(updateDir, "tmp", entry.Version)
	if err := unzip(zipPath, tmpDir); err != nil {
		fmt.Println("FAILED")
		fmt.Printf("  错误: %v\n", err)
		return err
	}
	fmt.Println("OK")

	// 找到实际内容目录（可能有一层嵌套）
	contentDir := tmpDir
	if entries, _ := os.ReadDir(tmpDir); len(entries) == 1 && entries[0].IsDir() {
		contentDir = filepath.Join(tmpDir, entries[0].Name())
	}

	// 3. 读取更新元信息和校验和
	metaPath := filepath.Join(contentDir, "update.yaml")
	checksumPath := filepath.Join(contentDir, "checksums.json")
	filesDir := filepath.Join(contentDir, "files")

	var meta UpdateMeta
	if gfile.Exists(metaPath) {
		data, _ := os.ReadFile(metaPath)
		yaml.Unmarshal(data, &meta)
	}

	checksums := make(map[string]string)
	if gfile.Exists(checksumPath) {
		data, _ := os.ReadFile(checksumPath)
		json.Unmarshal(data, &checksums)
	}

	// 4. 对比文件、处理冲突
	fmt.Print("  [3/5] 对比文件 ... ")
	if !gfile.Exists(filesDir) {
		fmt.Println("SKIP (无文件变更)")
	} else {
		var newFiles, modifiedFiles, conflictFiles []string
		var skipFiles []string

		filepath.Walk(filesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(filesDir, path)
			if isAddonPath(relPath) {
				return nil
			}
			localPath := filepath.Join(projectRoot, relPath)
			checksumKey := filepath.ToSlash(relPath)

			if !gfile.Exists(localPath) {
				newFiles = append(newFiles, relPath)
			} else {
				origHash, hasChecksum := checksums[checksumKey]
				localHash := fileHash(localPath)

				if hasChecksum && localHash != origHash {
					if isVersionMarkerFile(relPath) {
						modifiedFiles = append(modifiedFiles, relPath)
					} else {
						conflictFiles = append(conflictFiles, relPath)
					}
				} else {
					modifiedFiles = append(modifiedFiles, relPath)
				}
			}
			return nil
		})

		fmt.Printf("OK (新增 %d, 更新 %d, 冲突 %d)\n",
			len(newFiles), len(modifiedFiles), len(conflictFiles))

		// 处理冲突
		for _, f := range conflictFiles {
			fmt.Printf("\n  冲突: %s\n", f)
			choice := gcmd.Scan("    [1] 覆盖  [2] 跳过  : ")
			choice = strings.TrimSpace(choice)
			if choice == "2" {
				skipFiles = append(skipFiles, f)
			}
		}

		// 5. 备份并应用
		fmt.Print("  [4/5] 应用文件更新 ... ")
		backupDir := filepath.Join(updateDir, "backup", entry.Version)
		os.MkdirAll(backupDir, 0755)

		skipMap := make(map[string]bool)
		for _, f := range skipFiles {
			skipMap[f] = true
		}

		applied, skipped := 0, 0
		filepath.Walk(filesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(filesDir, path)
			if isAddonPath(relPath) {
				return nil
			}
			if skipMap[relPath] {
				skipped++
				return nil
			}

			localPath := filepath.Join(projectRoot, relPath)

			// 备份已存在的文件
			if gfile.Exists(localPath) {
				backupPath := filepath.Join(backupDir, relPath)
				os.MkdirAll(filepath.Dir(backupPath), 0755)
				copyFile(localPath, backupPath)
			}

			os.MkdirAll(filepath.Dir(localPath), 0755)
			copyFile(path, localPath)
			applied++
			return nil
		})

		// 删除文件
		for _, f := range meta.DeletedFiles {
			if isAddonPath(f) {
				continue
			}
			localPath := filepath.Join(projectRoot, f)
			if gfile.Exists(localPath) {
				backupPath := filepath.Join(backupDir, f)
				os.MkdirAll(filepath.Dir(backupPath), 0755)
				copyFile(localPath, backupPath)
				os.Remove(localPath)
			}
		}

		fmt.Printf("OK (%d 个更新, %d 个跳过)\n", applied, skipped)

		if skipped > 0 {
			fmt.Println("  提示: 跳过的冲突文件需要手动合并")
		}
	}

	// 6. 执行数据库迁移
	fmt.Print("  [5/5] 数据库迁移 ... ")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("SKIP (数据库未配置或连接失败: %v)\n", r)
				fmt.Println("  提示: 请手动执行 go run tools.go migrate up")
			}
		}()
		if err := migrate.RunUp(ctx, false); err != nil {
			fmt.Printf("WARNING: %v\n", err)
		} else {
			fmt.Println("OK")
		}
	}()

	return nil
}

// ==================== 工具函数 ====================

func getProjectRoot() string {
	if gfile.Exists("main.go") {
		abs, _ := filepath.Abs("..")
		return abs
	}
	abs, _ := filepath.Abs(".")
	return abs
}

func readLocalVersion(root string) (*VersionInfo, error) {
	path := filepath.Join(root, "version.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v VersionInfo
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func saveLocalVersion(root string, v *VersionInfo) {
	data, _ := json.MarshalIndent(v, "", "  ")
	data = append(data, '\n')
	os.WriteFile(filepath.Join(root, "version.json"), data, 0644)
}

func fetchUpdateIndex() (*UpdateIndex, error) {
	resp, err := httpGet(defaultIndexURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var index UpdateIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

func downloadFile(url string, mirrors []string, dest string) error {
	urls := append([]string{url}, mirrors...)
	var lastErr error
	for _, u := range urls {
		if err := doDownload(u, dest); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("所有下载源均失败，最后错误: %v", lastErr)
}

func doDownload(url, dest string) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func httpGet(url string) (*http.Response, error) {
	client := &http.Client{Timeout: httpTimeout}
	return client.Get(url)
}

func fileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil))
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
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(fpath), 0755)
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
