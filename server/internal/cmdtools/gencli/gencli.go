// +----------------------------------------------------------------------
// | XYGo Admin 命令行代码生成器（gen 子命令）
// +----------------------------------------------------------------------
// | 复用 web 设计器同一套 service（ColumnList/Edit/Build/PublishFrontend/Preview）
// | 执行流程与 web 完全一致：先 Edit 落库（写 xy_sys_gen_codes + 字段表），
// | 再 Build 生成代码并记录文件路径，最后 PublishFrontend 发布前端。
// | 这样命令行生成的模块在后台「代码生成器」里有完整记录，可二次调整/删除清理。
// +----------------------------------------------------------------------

package gencli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"xygo/internal/library/dbdialect"
	_ "xygo/internal/logic/gencodes" // 触发 init() 注册 GenCodes 服务
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
)

// Run gen 命令入口，args 为 "gen" 之后的全部参数。
func Run(ctx context.Context, args []string) error {
	f, err := parseFlags(args)
	if err != nil {
		printUsage()
		return err
	}

	// --list-tables：列出可生成的表后退出
	if f.listTables {
		return listTables(ctx)
	}

	// 组装生成入参（快捷模式 / spec 模式）
	in, fileList, err := buildEditInp(ctx, f)
	if err != nil {
		return err
	}

	// 干跑预览：只渲染不落盘、不落库
	if f.preview {
		return previewOnly(ctx, in)
	}

	// 二次确认（AI 可用 --yes 跳过）
	if !f.yes {
		fmt.Printf("\n  即将生成模块：表=%s  实体=%s  类型=%d\n", in.TableName, in.VarName, in.GenType)
		fmt.Printf("  将写入 %d 个文件，并在后台代码生成器中创建记录。\n", len(fileList))
		ans := strings.ToLower(strings.TrimSpace(gcmd.Scan("  确认执行？[y/N]: ")))
		if ans != "y" && ans != "yes" {
			fmt.Println("  已取消。")
			return nil
		}
	}

	return doBuild(ctx, in, fileList)
}

// ==================== 列出可生成的表 ====================

func listTables(ctx context.Context) error {
	model, err := service.GenCodes().TableSelect(ctx)
	if err != nil {
		return err
	}
	if len(model.List) == 0 {
		fmt.Println("  没有可生成的表（可能均已导入或被配置禁用）。")
		return nil
	}
	fmt.Printf("\n  可生成的表（共 %d 个）：\n", len(model.List))
	fmt.Println("  " + strings.Repeat("-", 70))
	fmt.Printf("  %-30s %-20s %s\n", "表名", "推荐实体名", "表注释")
	fmt.Println("  " + strings.Repeat("-", 70))
	for _, t := range model.List {
		fmt.Printf("  %-30s %-20s %s\n", t.TableName, t.VarName, t.TableComment)
	}
	fmt.Println()
	return nil
}

// ==================== 预览 ====================

func previewOnly(ctx context.Context, in *adminin.GenCodesEditInp) error {
	model, err := service.GenCodes().Preview(ctx, &adminin.GenCodesPreviewInp{GenCodesEditInp: *in})
	if err != nil {
		return err
	}
	fmt.Printf("\n  [预览] 将生成 %d 个文件（未写盘、未落库）：\n", len(model.Files))
	for _, file := range model.Files {
		fmt.Printf("    - %s\n", file.Path)
	}
	fmt.Println()
	return nil
}

// ==================== 正式生成 ====================

func doBuild(ctx context.Context, in *adminin.GenCodesEditInp, fileList []string) error {
	// 第一步：Edit 落库（写 xy_sys_gen_codes + 字段表），与 web 流程一致，返回记录 Id。
	editRes, err := service.GenCodes().Edit(ctx, in)
	if err != nil {
		return fmt.Errorf("保存生成器记录失败: %w", err)
	}
	in.Id = editRes.Id
	fmt.Printf("\n  [1/3] 已创建生成器记录 #%d（后台可查看/调整/删除）\n", in.Id)

	// 第二步：Build 生成代码（带 Id，会记录生成文件路径并把状态置为已生成）。
	if err := service.GenCodes().Build(ctx, &adminin.GenCodesBuildInp{GenCodesEditInp: *in}); err != nil {
		return fmt.Errorf("生成代码失败: %w", err)
	}
	fmt.Println("  [2/3] 后端文件、菜单权限已生成")

	// 第三步：发布前端（命令行无 Vite HMR 顾虑，直接发布）。
	if err := service.GenCodes().PublishFrontend(ctx); err != nil {
		return fmt.Errorf("发布前端文件失败: %w", err)
	}
	fmt.Println("  [3/3] 前端文件已发布")

	fmt.Printf("\n  ✅ 生成完成，共 %d 个文件：\n", len(fileList))
	for _, p := range fileList {
		fmt.Printf("    - %s\n", p)
	}
	fmt.Println("\n  提示：如需重新生成或删除，请在后台「代码生成器」中操作记录 #" + fmt.Sprint(in.Id))
	fmt.Println()
	return nil
}

// ==================== 入参组装 ====================

// buildEditInp 根据 flags 组装 GenCodesEditInp，并返回将生成的文件路径清单（用于展示）。
func buildEditInp(ctx context.Context, f *flags) (*adminin.GenCodesEditInp, []string, error) {
	prefix := tablePrefix(ctx)

	// 解析 spec 文件（精确模式）
	var spec *specFile
	if f.spec != "" {
		s, err := loadSpec(f.spec)
		if err != nil {
			return nil, nil, err
		}
		spec = s
	}

	// 1. 确定表名
	rawTable := f.table
	if spec != nil && spec.Table != "" {
		rawTable = spec.Table
	}
	if rawTable == "" {
		return nil, nil, fmt.Errorf("必须指定表名（位置参数、--table 或 spec.table）")
	}
	tableName := rawTable
	if !strings.HasPrefix(tableName, prefix) {
		tableName = prefix + tableName
	}

	// 2. 确定实体名
	varName := f.varName
	if spec != nil && spec.Var != "" {
		varName = spec.Var
	}
	if varName == "" {
		varName = tableToVarName(tableName, prefix)
	}

	// 3. 生成类型
	genType := f.genType
	if spec != nil && spec.GenType > 0 {
		genType = spec.GenType
	}
	if genType == 0 {
		genType = 10
	}

	// 4. 表注释（用于菜单标题）
	tableComment := tableCommentOf(ctx, tableName)
	if tableComment == "" {
		tableComment = varName
	}

	// 5. 取数据库字段（完整集合，自动推断 designType/queryType/关联）
	//    注意：必须传完整字段集，spec 仅按字段名覆盖属性，绝不删减，
	//    否则 autoSyncFieldsToDb 会把"库有但 Columns 没有"的列 DROP 掉。
	colModel, err := service.GenCodes().ColumnList(ctx, &adminin.GenCodesColumnListInp{TableName: tableName})
	if err != nil {
		return nil, nil, fmt.Errorf("读取表字段失败: %w", err)
	}
	columns := colModel.List
	if len(columns) == 0 {
		return nil, nil, fmt.Errorf("表 %s 没有字段或不存在", tableName)
	}

	// 6. 应用 spec 字段覆盖
	if spec != nil && len(spec.Columns) > 0 {
		if err := applyColumnOverrides(columns, spec.Columns); err != nil {
			return nil, nil, err
		}
	}

	// 7. 组装 options
	opts := buildOptions(f, spec, genType)
	optsJSON, _ := json.Marshal(opts)

	in := &adminin.GenCodesEditInp{
		GenType:      genType,
		TableName:    tableName,
		TableComment: tableComment,
		VarName:      varName,
		Options:      string(optsJSON),
		Columns:      columns,
	}

	// 8. 通过 Preview 拿到将生成的文件清单（无副作用），用于展示与确认
	previewModel, err := service.GenCodes().Preview(ctx, &adminin.GenCodesPreviewInp{GenCodesEditInp: *in})
	if err != nil {
		return nil, nil, fmt.Errorf("预渲染失败: %w", err)
	}
	fileList := make([]string, 0, len(previewModel.Files))
	for _, file := range previewModel.Files {
		fileList = append(fileList, file.Path)
	}

	return in, fileList, nil
}

// applyColumnOverrides 按字段名将 spec 覆盖应用到 ColumnList 结果上。
// 未知字段名直接报错（提示先建表），从根本上杜绝误删列。
func applyColumnOverrides(columns []adminin.GenCodesColumnItem, overrides map[string]specColumn) error {
	index := make(map[string]int, len(columns))
	for i, c := range columns {
		index[c.Name] = i
	}
	for name, ov := range overrides {
		i, ok := index[name]
		if !ok {
			return fmt.Errorf("spec 中的字段 %q 在表里不存在；请先建表/迁移，gen 不负责改表结构", name)
		}
		c := &columns[i]
		if ov.DesignType != nil {
			c.DesignType = *ov.DesignType
		}
		if ov.FormType != nil {
			c.FormType = *ov.FormType
		}
		if ov.QueryType != nil {
			c.QueryType = *ov.QueryType
		}
		if ov.IsList != nil {
			c.IsList = *ov.IsList
		}
		if ov.IsEdit != nil {
			c.IsEdit = *ov.IsEdit
		}
		if ov.IsQuery != nil {
			c.IsQuery = *ov.IsQuery
		}
		if ov.IsRequired != nil {
			c.IsRequired = *ov.IsRequired
		}
		if ov.DictType != nil {
			c.DictType = *ov.DictType
		}
		if len(ov.Extra) > 0 {
			c.Extra = string(ov.Extra)
		}
	}
	return nil
}

// buildOptions 根据 flags 与 spec 组装 options JSON 结构。
func buildOptions(f *flags, spec *specFile, genType int) *optionsJSON {
	o := &optionsJSON{
		GenType:   genType,
		HeadOps:   defaultHeadOps,
		ColumnOps: defaultColumnOps,
		AutoOps:   defaultAutoOps,
		ViewMode:  "drawer",
	}
	o.Menu.Sort = 100
	// 默认给一个合法的 Iconify 图标，避免落入 generate.go 的 ele-Document 兜底（渲染空白）。
	o.Menu.Icon = defaultMenuIcon

	// flags 覆盖
	if f.head != "" {
		o.HeadOps = splitCSV(f.head)
	}
	if f.column != "" {
		o.ColumnOps = splitCSV(f.column)
	}
	if f.auto != "" {
		o.AutoOps = splitCSV(f.auto)
	}
	if f.menuPid != 0 {
		o.Menu.Pid = f.menuPid
	}
	if f.icon != "" {
		o.Menu.Icon = f.icon
	}
	if f.sort != 0 {
		o.Menu.Sort = f.sort
	}
	if f.view != "" {
		o.ViewMode = f.view
	}
	if genType == 11 {
		o.Tree.PidColumn = orDefault(f.treePid, "parent_id")
		o.Tree.TitleColumn = orDefault(f.treeTitle, "name")
	}

	// spec.options 覆盖（优先级最高）
	if spec != nil && spec.Options != nil {
		so := spec.Options
		if len(so.HeadOps) > 0 {
			o.HeadOps = so.HeadOps
		}
		if len(so.ColumnOps) > 0 {
			o.ColumnOps = so.ColumnOps
		}
		if len(so.AutoOps) > 0 {
			o.AutoOps = so.AutoOps
		}
		if so.Menu.Pid != 0 {
			o.Menu.Pid = so.Menu.Pid
		}
		if so.Menu.Icon != "" {
			o.Menu.Icon = so.Menu.Icon
		}
		if so.Menu.Sort != 0 {
			o.Menu.Sort = so.Menu.Sort
		}
		if so.ViewMode != "" {
			o.ViewMode = so.ViewMode
		}
		if so.Tree.PidColumn != "" {
			o.Tree.PidColumn = so.Tree.PidColumn
		}
		if so.Tree.TitleColumn != "" {
			o.Tree.TitleColumn = so.Tree.TitleColumn
		}
		if so.ApiPrefix != "" {
			o.ApiPrefix = so.ApiPrefix
		}
		if so.AddonName != "" {
			o.AddonName = so.AddonName
		}
		if len(so.GenPaths) > 0 {
			o.GenPaths = so.GenPaths
		}
	}

	// --no-gf：移除 runDao/runService
	if f.noGf {
		o.AutoOps = removeFromSlice(o.AutoOps, "runDao", "runService")
	}

	return o
}

// ==================== 工具 ====================

func tablePrefix(ctx context.Context) string {
	cfg, err := g.Cfg().Get(ctx, "database.default.Prefix")
	if err != nil || cfg.IsEmpty() {
		return "xy_"
	}
	return cfg.String()
}

// tableCommentOf 通过方言层读取表注释（不受是否已导入影响）。
func tableCommentOf(ctx context.Context, tableName string) string {
	dialect := dbdialect.Get()
	dbName, err := dialect.GetDbName(ctx)
	if err != nil || dbName == "" {
		return ""
	}
	type tableInfo struct {
		TableName    string `json:"tableName"`
		TableComment string `json:"tableComment"`
	}
	var tables []tableInfo
	if err := g.DB().Ctx(ctx).Raw(dialect.ListTablesSQL(dbName)).Scan(&tables); err != nil {
		return ""
	}
	for _, t := range tables {
		if t.TableName == tableName {
			return t.TableComment
		}
	}
	return ""
}

// tableToVarName 表名去前缀转 PascalCase，如 xy_biz_article -> BizArticle。
func tableToVarName(tableName, prefix string) string {
	name := strings.TrimPrefix(tableName, prefix)
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func removeFromSlice(arr []string, drop ...string) []string {
	dropSet := make(map[string]bool, len(drop))
	for _, d := range drop {
		dropSet[d] = true
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if !dropSet[v] {
			out = append(out, v)
		}
	}
	return out
}

func loadSpec(path string) (*specFile, error) {
	if !gfile.Exists(path) {
		return nil, fmt.Errorf("spec 文件不存在: %s", path)
	}
	content := gfile.GetContents(path)
	var s specFile
	if err := json.Unmarshal([]byte(content), &s); err != nil {
		return nil, fmt.Errorf("解析 spec JSON 失败: %w", err)
	}
	return &s, nil
}

func printUsage() {
	fmt.Println(`
  用法:
    go run tools.go gen <表名> [flags]      生成 CRUD 模块
    go run tools.go gen <表名> --preview    干跑预览（不写盘、不落库）
    go run tools.go gen --spec <file.json>  JSON 精确控制（含关联表）
    go run tools.go gen --list-tables       列出可生成的表

  常用 flags:
    --type 10|11      10=普通列表(默认) 11=树表
    --var Name        实体名(PascalCase，默认由表名推导)
    --head a,b,c      头部按钮: add,batchDel,export
    --column a,b,c    行内操作: edit,del,view,status,check
    --auto a,b,c      自动步骤: genMenuPermissions,runDao,runService
    --menu-pid N      挂载父菜单ID(0=新建顶级目录+页面)
    --icon ic         菜单图标(Iconify Remix 名，如 ri:box-3-line；默认 ri:apps-2-line)
    --sort N          菜单排序(默认100)
    --view drawer|page 详情模式
    --tree-pid col    树表父ID列(默认 parent_id)
    --tree-title col  树表标题列(默认 name)
    --no-gf           跳过 gf gen dao/service
    --yes / -y        跳过确认直接执行`)
}
