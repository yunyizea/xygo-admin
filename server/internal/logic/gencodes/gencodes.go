package gencodes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"xygo/internal/dao"
	"xygo/internal/library/dbdialect"
	"xygo/internal/library/genconfig"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/model/input/form"
	"xygo/internal/service"
)

type sGenCodes struct{}

func init() {
	service.RegisterGenCodes(New())
}

// New 构造代码生成器服务
func New() *sGenCodes {
	return &sGenCodes{}
}

// Selects 获取选项
func (s *sGenCodes) Selects(ctx context.Context) (*adminin.GenCodesSelectsModel, error) {
	// 从配置读取默认路径
	tpl := genconfig.GetDefaultTemplate(ctx)
	genPaths := map[string]string{
		"apiPath":        tpl.ApiPath,
		"controllerPath": tpl.ControllerPath,
		"logicPath":      tpl.LogicPath,
		"inputPath":      tpl.InputPath,
		"sqlPath":        tpl.SqlPath,
		"webApiPath":     tpl.WebApiPath,
		"webViewsPath":   tpl.WebViewsPath,
	}

	return &adminin.GenCodesSelectsModel{
		GenType: []adminin.SelectOption{
			{Value: 10, Label: "普通列表"},
			{Value: 11, Label: "树表"},
		},
		FormType: []adminin.SelectOption{
			{Value: "input", Label: "文本框"},
			{Value: "inputNumber", Label: "数字框"},
			{Value: "textarea", Label: "文本域"},
			{Value: "select", Label: "下拉框"},
			{Value: "radio", Label: "单选框"},
			{Value: "checkbox", Label: "复选框"},
			{Value: "switch", Label: "开关"},
			{Value: "date", Label: "日期"},
			{Value: "datetime", Label: "日期时间"},
			{Value: "imageUpload", Label: "图片上传"},
			{Value: "imagesUpload", Label: "多图上传"},
			{Value: "fileUpload", Label: "文件上传"},
			{Value: "richEditor", Label: "富文本"},
			{Value: "colorPicker", Label: "颜色选择"},
			{Value: "iconSelector", Label: "图标选择"},
			{Value: "remoteSelect", Label: "远程下拉"},
		},
		QueryType: []adminin.SelectOption{
			{Value: "eq", Label: "精确匹配(=)"},
			{Value: "neq", Label: "不等于(!=)"},
			{Value: "like", Label: "模糊匹配(LIKE)"},
			{Value: "gt", Label: "大于(>)"},
			{Value: "gte", Label: "大于等于(>=)"},
			{Value: "lt", Label: "小于(<)"},
			{Value: "lte", Label: "小于等于(<=)"},
			{Value: "between", Label: "区间(BETWEEN)"},
			{Value: "in", Label: "包含(IN)"},
		},
		DesignTypes: []adminin.SelectOption{
			{Value: "pk", Label: "主键"},
			{Value: "string", Label: "字符串"},
			{Value: "number", Label: "数字"},
			{Value: "float", Label: "浮点数"},
			{Value: "switch", Label: "开关"},
			{Value: "radio", Label: "单选框"},
			{Value: "checkbox", Label: "复选框"},
			{Value: "select", Label: "下拉选择"},
			{Value: "selects", Label: "下拉多选"},
			{Value: "textarea", Label: "文本域"},
			{Value: "password", Label: "密码"},
			{Value: "datetime", Label: "日期时间"},
			{Value: "date", Label: "日期"},
			{Value: "time", Label: "时间"},
			{Value: "timestamp", Label: "时间戳"},
			{Value: "image", Label: "图片上传"},
			{Value: "images", Label: "多图上传"},
			{Value: "file", Label: "文件上传"},
			{Value: "files", Label: "多文件上传"},
			{Value: "editor", Label: "富文本"},
			{Value: "color", Label: "颜色选择"},
			{Value: "icon", Label: "图标选择"},
			{Value: "city", Label: "城市选择"},
			{Value: "remoteSelect", Label: "远程下拉"},
			{Value: "remoteSelects", Label: "远程多选"},
			{Value: "weigh", Label: "权重(拖拽排序)"},
		},
		GenPaths:  genPaths,
		AddonList: scanInstalledAddons(),
	}, nil
}

// scanInstalledAddons 扫描 server/addons/ 下已安装的扩展，返回选项列表
func scanInstalledAddons() []adminin.SelectOption {
	addonsDir := filepath.Join(gfile.Pwd(), "addons")
	entries, err := os.ReadDir(addonsDir)
	if err != nil {
		return nil
	}
	var list []adminin.SelectOption
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()
		yamlPath := filepath.Join(addonsDir, name, "addon.yaml")
		if !gfile.Exists(yamlPath) {
			continue
		}
		title := name
		data := gfile.GetContents(yamlPath)
		for _, line := range strings.Split(data, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "title:") {
				t := strings.TrimPrefix(line, "title:")
				t = strings.TrimSpace(t)
				t = strings.Trim(t, "\"'")
				if t != "" {
					title = t
				}
				break
			}
		}
		list = append(list, adminin.SelectOption{Value: name, Label: title + " (" + name + ")"})
	}
	return list
}

// TableSelect 获取数据库表列表
func (s *sGenCodes) TableSelect(ctx context.Context) (*adminin.GenCodesTableSelectModel, error) {
	dbName := getDbName(ctx)
	tablePrefix := getTablePrefix(ctx)
	dialect := dbdialect.Get()

	type tableInfo struct {
		TableName    string `json:"tableName"`
		TableComment string `json:"tableComment"`
	}

	var tables []tableInfo
	err := g.DB().Ctx(ctx).Raw(dialect.ListTablesSQL(dbName)).Scan(&tables)
	if err != nil {
		return nil, err
	}

	// 过滤掉已导入的表
	var existTables []string
	err = dao.SysGenCodes.Ctx(ctx).Fields("table_name").Scan(&existTables)
	if err != nil {
		return nil, err
	}
	existMap := make(map[string]bool, len(existTables))
	for _, t := range existTables {
		existMap[t] = true
	}

	list := make([]adminin.GenCodesTableSelectItem, 0, len(tables))
	for _, t := range tables {
		if existMap[t.TableName] {
			continue
		}
		// 过滤配置中禁用的表
		if genconfig.IsTableDisabled(ctx, t.TableName) {
			continue
		}
		varName := tableNameToVarName(t.TableName, tablePrefix)
		list = append(list, adminin.GenCodesTableSelectItem{
			TableName:    t.TableName,
			TableComment: t.TableComment,
			VarName:      varName,
		})
	}

	return &adminin.GenCodesTableSelectModel{List: list}, nil
}

// ColumnList 获取表字段列表（自动推断组件类型）
func (s *sGenCodes) ColumnList(ctx context.Context, in *adminin.GenCodesColumnListInp) (*adminin.GenCodesColumnListModel, error) {
	dbName := getDbName(ctx)
	dialect := dbdialect.Get()

	type columnInfo struct {
		ColumnName    string `json:"columnName"`
		ColumnType    string `json:"columnType"`
		DataType      string `json:"dataType"`
		ColumnComment string `json:"columnComment"`
		ColumnKey     string `json:"columnKey"`
		IsNullable    string `json:"isNullable"`
		Extra         string `json:"extra"`
		OrdinalPos    int    `json:"ordinalPos"`
	}

	var columns []columnInfo
	err := g.DB().Ctx(ctx).Raw(dialect.ListColumnsSQL(dbName, in.TableName)).Scan(&columns)
	if err != nil {
		return nil, err
	}

	list := make([]adminin.GenCodesColumnItem, 0, len(columns))
	for i, col := range columns {
		item := adminin.GenCodesColumnItem{
			Name:    col.ColumnName,
			GoName:  snakeToPascal(col.ColumnName),
			TsName:  snakeToCamel(col.ColumnName),
			DbType:  col.ColumnType,
			GoType:  dialect.TypeToGoType(col.DataType, col.ColumnType),
			TsType:  dialect.TypeToTsType(col.DataType),
			Comment: col.ColumnComment,
			IsPk:    boolToInt(col.ColumnKey == "PRI"),
			Sort:    i + 1,
		}

		// 推断 designType（核心）
		item.DesignType = inferDesignType(col.ColumnName, col.DataType, col.ColumnType, col.ColumnComment, col.ColumnKey, col.Extra)

		// 基于 designType 推断表单组件和查询方式
		item.FormType = designTypeToFormType(item.DesignType)
		item.QueryType = inferQueryType(col.ColumnName, col.DataType, item.FormType)

		// 推断是否必填
		item.IsRequired = boolToInt(col.IsNullable == "NO" && col.ColumnKey != "PRI" && col.Extra != "auto_increment")

		// 推断是否在列表/编辑/搜索中显示
		item.IsList = boolToInt(!isHiddenColumn(col.ColumnName))
		item.IsEdit = boolToInt(!isAutoColumn(col.ColumnName) && col.ColumnKey != "PRI")
		item.IsQuery = boolToInt(isQueryColumn(col.ColumnName))

		list = append(list, item)
	}

	return &adminin.GenCodesColumnListModel{List: list}, nil
}

// List 生成记录列表
func (s *sGenCodes) List(ctx context.Context, in *adminin.GenCodesListInp) (*adminin.GenCodesListModel, error) {
	model := dao.SysGenCodes.Ctx(ctx)

	if in.GenType > 0 {
		model = model.Where("gen_type", in.GenType)
	}
	if in.VarName != "" {
		model = model.WhereLike("var_name", "%"+in.VarName+"%")
	}
	if in.Status > 0 {
		model = model.Where("status", in.Status)
	}

	count, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}

	var list []adminin.GenCodesListItem
	err = model.Page(in.Page, in.PageSize).OrderDesc("id").Scan(&list)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []adminin.GenCodesListItem{}
	}

	return &adminin.GenCodesListModel{
		List: list,
		PageRes: form.PageRes{
			Page:     in.Page,
			PageSize: in.PageSize,
			Total:    count,
		},
	}, nil
}

// View 查看详情
func (s *sGenCodes) View(ctx context.Context, in *adminin.GenCodesViewInp) (*adminin.GenCodesViewModel, error) {
	var genCode adminin.GenCodesListItem
	err := dao.SysGenCodes.Ctx(ctx).Where("id", in.Id).Scan(&genCode)
	if err != nil {
		return nil, err
	}
	if genCode.Id == 0 {
		return nil, gerror.New("记录不存在")
	}

	// 查询关联字段
	var columns []adminin.GenCodesColumnItem
	err = dao.SysGenCodesColumn.Ctx(ctx).Where("gen_id", in.Id).OrderAsc("sort").Scan(&columns)
	if err != nil {
		return nil, err
	}
	if columns == nil {
		columns = []adminin.GenCodesColumnItem{}
	}

	return &adminin.GenCodesViewModel{
		GenCodesListItem: genCode,
		Columns:          columns,
	}, nil
}

// Edit 保存配置
func (s *sGenCodes) Edit(ctx context.Context, in *adminin.GenCodesEditInp) (*adminin.GenCodesEditModel, error) {
	now := time.Now().Unix()

	if in.DbName == "" {
		in.DbName = getDbName(ctx)
	}

	// 如果 id=0，先查找是否已存在同表记录（避免唯一索引冲突）
	if in.Id == 0 && in.TableName != "" {
		existId, _ := dao.SysGenCodes.Ctx(ctx).
			Where("db_name", in.DbName).
			Where("table_name", in.TableName).
			Value("id")
		if !existId.IsEmpty() {
			in.Id = existId.Uint64()
		}
	}

	if in.Id == 0 {
		// 新增
		id, err := dao.SysGenCodes.Ctx(ctx).Data(g.Map{
			"gen_type":      in.GenType,
			"db_name":       in.DbName,
			"table_name":    in.TableName,
			"table_comment": in.TableComment,
			"var_name":      in.VarName,
			"options":       in.Options,
			"status":        2,
			"created_at":    now,
			"updated_at":    now,
		}).InsertAndGetId()
		if err != nil {
			return nil, err
		}
		in.Id = uint64(id)

		// 保存字段
		if err := saveColumns(ctx, in.Id, in.Columns); err != nil {
			return nil, err
		}
	} else {
		// 更新
		_, err := dao.SysGenCodes.Ctx(ctx).Where("id", in.Id).Data(g.Map{
			"gen_type":      in.GenType,
			"table_comment": in.TableComment,
			"var_name":      in.VarName,
			"options":       in.Options,
			"updated_at":    now,
		}).Update()
		if err != nil {
			return nil, err
		}

		// 重新保存字段
		if err := saveColumns(ctx, in.Id, in.Columns); err != nil {
			return nil, err
		}
	}

	return &adminin.GenCodesEditModel{Id: in.Id}, nil
}

// Delete 删除配置（支持同时删除生成文件和菜单）
func (s *sGenCodes) Delete(ctx context.Context, in *adminin.GenCodesDeleteInp) error {
	// 查询配置详情（用于定位文件和菜单）
	var record entity.SysGenCodes
	err := dao.SysGenCodes.Ctx(ctx).Where("id", in.Id).Scan(&record)
	if err != nil {
		return err
	}
	if record.Id == 0 {
		return gerror.New("记录不存在")
	}

	// 解析 options 获取路径配置
	optStr := ""
	if record.Options != nil {
		optStr = record.Options.String()
	}
	opts := parseOptions(optStr)
	varName := record.VarName
	routeName := camelToKebab(lcFirst(varName))

	// 1. 删除生成的文件
	if in.DeleteFiles {
		deleteGeneratedFiles(ctx, &record, opts)
	}

	// 2. 删除生成的菜单
	if in.DeleteMenus {
		deleteGeneratedMenus(ctx, routeName, varName)
	}

	// 3. 删除配置记录
	_, err = dao.SysGenCodes.Ctx(ctx).Where("id", in.Id).Delete()
	if err != nil {
		return err
	}
	_, err = dao.SysGenCodesColumn.Ctx(ctx).Where("gen_id", in.Id).Delete()
	return err
}

// Preview 预览代码（gen-3 实现）
func (s *sGenCodes) Preview(ctx context.Context, in *adminin.GenCodesPreviewInp) (*adminin.GenCodesPreviewModel, error) {
	return generateCode(ctx, &in.GenCodesEditInp, false)
}

// Build 执行生成（gen-3 实现）
func (s *sGenCodes) Build(ctx context.Context, in *adminin.GenCodesBuildInp) error {
	_, err := generateCode(ctx, &in.GenCodesEditInp, true)
	if err != nil {
		return err
	}
	// 更新状态为已生成
	if in.Id > 0 {
		_, _ = dao.SysGenCodes.Ctx(ctx).Where("id", in.Id).Data(g.Map{
			"status":     1,
			"updated_at": time.Now().Unix(),
		}).Update()
	}
	return nil
}

// CreateTable 创建数据表（gen-5 实现）
func (s *sGenCodes) CreateTable(ctx context.Context, in *adminin.GenCodesCreateTableInp) (*adminin.GenCodesCreateTableModel, error) {
	return createTableFromDesign(ctx, in)
}

// ==================== 内部辅助 ====================

// saveColumns 保存字段配置（先删后插）
func saveColumns(ctx context.Context, genId uint64, columns []adminin.GenCodesColumnItem) error {
	if len(columns) == 0 {
		return nil
	}
	// 先删除旧数据
	_, _ = dao.SysGenCodesColumn.Ctx(ctx).Where("gen_id", genId).Delete()

	// 批量插入
	for i, col := range columns {
		_, err := dao.SysGenCodesColumn.Ctx(ctx).Data(g.Map{
			"gen_id":      genId,
			"name":        col.Name,
			"go_name":     col.GoName,
			"ts_name":     col.TsName,
			"db_type":     col.DbType,
			"go_type":     col.GoType,
			"ts_type":     col.TsType,
			"comment":     col.Comment,
			"is_pk":       col.IsPk,
			"is_required": col.IsRequired,
			"is_list":     col.IsList,
			"is_edit":     col.IsEdit,
			"is_query":    col.IsQuery,
			"query_type":  col.QueryType,
			"form_type":   col.FormType,
			"design_type": col.DesignType,
			"extra":       col.Extra,
			"dict_type":   col.DictType,
			"sort":        i + 1,
		}).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

// getDbName 获取当前数据库名（通过方言层适配 MySQL/PG）
func getDbName(ctx context.Context) string {
	dialect := dbdialect.Get()
	name, err := dialect.GetDbName(ctx)
	if err != nil || name == "" {
		return "xygonew"
	}
	return name
}

// getTablePrefix 获取表前缀
func getTablePrefix(ctx context.Context) string {
	cfg, err := g.Cfg().Get(ctx, "database.default.Prefix")
	if err != nil || cfg.IsEmpty() {
		return "xy_"
	}
	return cfg.String()
}

// tableNameToVarName 表名 -> PascalCase 实体名
// xy_biz_article -> BizArticle
func tableNameToVarName(tableName, prefix string) string {
	name := strings.TrimPrefix(tableName, prefix)
	return snakeToPascal(name)
}

// snakeToPascal snake_case -> PascalCase
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// snakeToCamel snake_case -> camelCase
func snakeToCamel(s string) string {
	pascal := snakeToPascal(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// mysqlTypeToGoType / mysqlTypeToTsType 已迁移至 dbdialect 方言层
// 调用方统一使用 dialect.TypeToGoType() / dialect.TypeToTsType()

// ==================== 设计类型推断（对齐 BuildAdmin $inputTypeRule） ====================

// designRule 单条推断规则，所有条件取交集（AND）。
// 未指定的切片视为"不限"。
type designRule struct {
	types       []string // DATA_TYPE 匹配（如 tinyint, varchar）
	suffixes    []string // 字段名后缀匹配（同 BuildAdmin isMatchSuffix）
	columnTypes []string // COLUMN_TYPE 精确匹配（如 tinyint(1)）
	value       string   // 命中后返回的 designType
}

// designRules 有序规则列表，直接翻译自 BuildAdmin Helper::$inputTypeRule。
// 首条匹配即返回，顺序决定优先级。
var designRules = []designRule{
	// ---- 开关 ----
	{types: []string{"tinyint", "int", "enum"}, suffixes: []string{"switch", "toggle"}, value: "switch"},
	{columnTypes: []string{"tinyint(1)", "char(1)", "tinyint(1) unsigned", "int2", "smallint"}, suffixes: []string{"switch", "toggle"}, value: "switch"},
	// ---- 富文本（高优先级，type+suffix 同时命中） ----
	{types: []string{"longtext", "text", "mediumtext", "tinytext"}, suffixes: []string{"content", "editor"}, value: "editor"},
	// ---- textarea（varchar + suffix） ----
	{types: []string{"varchar"}, suffixes: []string{"textarea", "multiline", "rows"}, value: "textarea"},
	// ---- Array ----
	{suffixes: []string{"array"}, value: "array"},
	// ---- 时间戳（int/bigint + time/datetime 后缀） ----
	{types: []string{"int", "bigint"}, suffixes: []string{"time", "datetime"}, value: "timestamp"},
	// ---- 日期时间类型 ----
	{types: []string{"datetime", "timestamp"}, value: "datetime"},
	{types: []string{"date"}, value: "date"},
	{types: []string{"year"}, value: "year"},
	{types: []string{"time"}, value: "time"},
	// ---- 远程多选（_ids 先于 _id） ----
	{suffixes: []string{"_ids"}, value: "remoteSelects"},
	// ---- 远程单选 ----
	{suffixes: []string{"_id"}, value: "remoteSelect"},
	// ---- 多选 select ----
	{suffixes: []string{"selects", "multi", "lists"}, value: "selects"},
	// ---- 单选 select ----
	{suffixes: []string{"select", "list", "data"}, value: "select"},
	// ---- 城市选择器 ----
	{suffixes: []string{"city"}, value: "city"},
	// ---- 多图（images 先于 image） ----
	{suffixes: []string{"images", "avatars"}, value: "images"},
	// ---- 单图 ----
	{suffixes: []string{"image", "avatar"}, value: "image"},
	// ---- 多文件 ----
	{suffixes: []string{"files"}, value: "files"},
	// ---- 单文件 ----
	{suffixes: []string{"file"}, value: "file"},
	// ---- 图标 ----
	{suffixes: []string{"icon"}, value: "icon"},
	// ---- 单选框（column_type + suffix，如 status tinyint(1)） ----
	{columnTypes: []string{"tinyint(1)", "char(1)", "tinyint(1) unsigned", "int2", "smallint"}, suffixes: []string{"status", "state", "type"}, value: "radio"},
	// ---- 数字输入框（后缀） ----
	{suffixes: []string{"number", "int", "num"}, value: "number"},
	// ---- 数字输入框（类型兜底） ----
	{types: []string{"bigint", "int", "mediumint", "smallint", "tinyint", "decimal", "double", "float"}, value: "number"},
	// ---- textarea（低优先级，纯类型） ----
	{types: []string{"longtext", "text", "mediumtext", "tinytext"}, value: "textarea"},
	// ---- 单选框（enum 低优先级） ----
	{types: []string{"enum"}, value: "radio"},
	// ---- 多选框（set） ----
	{types: []string{"set"}, value: "checkbox"},
	// ---- 颜色选择器 ----
	{suffixes: []string{"color"}, value: "color"},
}

// matchSuffix 检查 name 是否以 suffixes 中任一后缀结尾（不区分大小写）
func matchSuffix(name string, suffixes []string) bool {
	nl := strings.ToLower(name)
	for _, s := range suffixes {
		if strings.HasSuffix(nl, strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// normalizePgDataType 将 PG 底层类型名标准化为 MySQL 类型名，供 designRules 规则表匹配
func normalizePgDataType(pgType string) string {
	switch strings.ToLower(pgType) {
	case "int2", "smallserial":
		return "smallint"
	case "int4", "serial":
		return "int"
	case "int8", "bigserial":
		return "bigint"
	case "float4", "real":
		return "float"
	case "float8", "double precision":
		return "double"
	case "numeric":
		return "decimal"
	case "bool", "boolean":
		return "tinyint" // PG bool 对应 MySQL tinyint(1)
	case "varchar", "character varying":
		return "varchar"
	case "char", "character", "bpchar":
		return "char"
	case "timestamptz", "timestamp without time zone", "timestamp with time zone":
		return "timestamp"
	case "timetz", "time without time zone", "time with time zone":
		return "time"
	case "jsonb":
		return "json"
	case "bytea":
		return "blob"
	default:
		return strings.ToLower(pgType)
	}
}

// inferDesignType 根据字段名/类型/注释智能推断 designType（对齐 BuildAdmin Helper::getTableColumnsDataType）
func inferDesignType(name, dataType, columnType, comment, columnKey, extra string) string {
	nameLower := strings.ToLower(name)
	dtLower := strings.ToLower(dataType)
	ctLower := strings.ToLower(columnType)

	// PG 类型标准化：将 int4/int8/bool 等映射为规则表能识别的 MySQL 类型名
	if dbdialect.IsPgsql() {
		dtLower = normalizePgDataType(dtLower)
	}

	// ---------- 预匹配（同 BuildAdmin 在规则表前的特殊检测） ----------
	// 主键
	if columnKey == "PRI" || (strings.Contains(nameLower, "id") && strings.Contains(strings.ToLower(extra), "auto_increment")) {
		return "pk"
	}
	// 权重
	if nameLower == "weigh" || nameLower == "weight" || nameLower == "sort" || nameLower == "order_num" {
		return "weigh"
	}
	// 自动时间字段：仅 int/bigint 存储的时间戳才预判为 timestamp，
	// datetime/timestamp 类型的交给规则表走 "datetime" 分支。
	switch nameLower {
	case "create_time", "update_time", "createtime", "updatetime", "created_at", "updated_at":
		if dtLower == "int" || dtLower == "bigint" {
			return "timestamp"
		}
		// datetime/timestamp 类型不预判，继续走规则表
	}
	// 密码字段（BuildAdmin 未做，但语义明确保留）
	if strings.Contains(nameLower, "password") || strings.Contains(nameLower, "passwd") || strings.Contains(nameLower, "secret") {
		return "password"
	}

	// ---------- 规则表匹配（BuildAdmin $inputTypeRule 翻译） ----------
	for _, rule := range designRules {
		typeOK := true
		suffixOK := true
		ctOK := true

		if len(rule.types) > 0 {
			typeOK = false
			for _, t := range rule.types {
				if dtLower == t {
					typeOK = true
					break
				}
			}
		}
		if len(rule.suffixes) > 0 {
			suffixOK = matchSuffix(name, rule.suffixes)
		}
		if len(rule.columnTypes) > 0 {
			ctOK = false
			for _, ct := range rule.columnTypes {
				if ctLower == ct {
					ctOK = true
					break
				}
			}
		}
		if typeOK && suffixOK && ctOK {
			return rule.value
		}
	}

	// ---------- JSON ----------
	if dtLower == "json" {
		return "textarea"
	}

	return "string"
}

// designTypeToFormType 根据 designType 映射到 formType
func designTypeToFormType(designType string) string {
	switch designType {
	case "pk":
		return "input"
	case "string":
		return "input"
	case "number":
		return "inputNumber"
	case "float":
		return "inputNumber"
	case "switch":
		return "switch"
	case "radio":
		return "radio"
	case "checkbox":
		return "checkbox"
	case "select":
		return "select"
	case "selects":
		return "select"
	case "textarea":
		return "textarea"
	case "password":
		return "input"
	case "datetime":
		return "datetime"
	case "date":
		return "date"
	case "time":
		return "input"
	case "timestamp":
		return "datetime"
	case "image":
		return "imageUpload"
	case "images":
		return "imagesUpload"
	case "file":
		return "fileUpload"
	case "files":
		return "fileUpload"
	case "editor":
		return "richEditor"
	case "color":
		return "colorPicker"
	case "icon":
		return "iconSelector"
	case "city":
		return "input"
	case "remoteSelect":
		return "remoteSelect"
	case "remoteSelects":
		return "remoteSelect"
	case "weigh":
		return "inputNumber"
	default:
		return "input"
	}
}

// inferFormType 推断表单组件类型
func inferFormType(name, dataType, columnType, comment string) string {
	nameLower := strings.ToLower(name)

	// 字段名后缀规则
	if strings.HasSuffix(nameLower, "_id") || strings.HasSuffix(nameLower, "_ids") {
		return "remoteSelect"
	}
	if strings.Contains(nameLower, "image") || strings.Contains(nameLower, "avatar") || strings.Contains(nameLower, "logo") || strings.Contains(nameLower, "thumb") || strings.Contains(nameLower, "pic") {
		if strings.Contains(nameLower, "images") || strings.Contains(nameLower, "pics") {
			return "imagesUpload"
		}
		return "imageUpload"
	}
	if strings.Contains(nameLower, "file") || strings.Contains(nameLower, "attachment") {
		return "fileUpload"
	}
	if strings.Contains(nameLower, "content") || strings.Contains(nameLower, "editor") || strings.Contains(nameLower, "body") {
		return "richEditor"
	}
	if strings.Contains(nameLower, "color") {
		return "colorPicker"
	}
	if strings.Contains(nameLower, "icon") {
		return "iconSelector"
	}

	// 字段类型规则
	dtLower := strings.ToLower(dataType)
	ctLower := strings.ToLower(columnType)

	if dtLower == "text" || dtLower == "mediumtext" || dtLower == "longtext" {
		return "textarea"
	}
	if dtLower == "datetime" || dtLower == "timestamp" {
		return "datetime"
	}
	if dtLower == "date" {
		return "date"
	}
	if dtLower == "json" {
		return "textarea"
	}
	if dtLower == "decimal" || dtLower == "float" || dtLower == "double" {
		return "inputNumber"
	}
	if (nameLower == "status" || nameLower == "state" || nameLower == "type") &&
		((dtLower == "tinyint" && strings.Contains(ctLower, "tinyint(1)")) || strings.Contains(ctLower, "char(1)") ||
			dtLower == "smallint" || dtLower == "int2") {
		return "radio"
	}
	if (dtLower == "tinyint" && strings.Contains(ctLower, "tinyint(1)")) ||
		((dtLower == "smallint" || dtLower == "int2") && strings.HasPrefix(nameLower, "is_")) {
		return "switch"
	}

	// 注释解析：格式如 "状态:0=禁用,1=启用"
	if strings.Contains(comment, ":") && strings.Contains(comment, "=") {
		return "radio"
	}
	if strings.Contains(comment, "：") && strings.Contains(comment, "=") {
		return "radio"
	}

	// 特定字段名推断
	if nameLower == "status" || nameLower == "state" || nameLower == "type" {
		return "radio"
	}
	if nameLower == "gender" || nameLower == "sex" {
		return "radio"
	}
	if nameLower == "sort" || nameLower == "weight" || nameLower == "order" {
		return "inputNumber"
	}
	if strings.Contains(nameLower, "switch") || strings.Contains(nameLower, "toggle") || strings.Contains(nameLower, "enable") || strings.Contains(nameLower, "visible") {
		return "switch"
	}
	if strings.Contains(nameLower, "remark") || strings.Contains(nameLower, "desc") || strings.Contains(nameLower, "note") || strings.Contains(nameLower, "memo") {
		return "textarea"
	}

	return "input"
}

// inferQueryType 推断查询方式
func inferQueryType(name, dataType, formType string) string {
	nameLower := strings.ToLower(name)

	if strings.Contains(nameLower, "name") || strings.Contains(nameLower, "title") || nameLower == "remark" || nameLower == "desc" {
		return "like"
	}
	if formType == "datetime" || formType == "date" {
		return "between"
	}
	if strings.HasSuffix(nameLower, "_time") || strings.HasSuffix(nameLower, "_at") || strings.HasSuffix(nameLower, "_date") {
		return "between"
	}
	return "eq"
}

// isHiddenColumn 是否隐藏的列（不在列表中显示）
func isHiddenColumn(name string) bool {
	hidden := map[string]bool{
		"deleted_at": true, "delete_time": true,
		"options": true, "content": true, "body": true,
		"password": true, "salt": true, "secret": true,
	}
	return hidden[strings.ToLower(name)]
}

// isAutoColumn 自动列（不可编辑）
func isAutoColumn(name string) bool {
	auto := map[string]bool{
		"id": true, "created_at": true, "updated_at": true,
		"create_time": true, "update_time": true, "deleted_at": true,
		"delete_time": true, "created_by": true, "updated_by": true,
	}
	return auto[strings.ToLower(name)]
}

// isQueryColumn 是否默认作为搜索条件
func isQueryColumn(name string) bool {
	query := map[string]bool{
		"status": true, "state": true, "type": true,
		"name": true, "title": true,
	}
	return query[strings.ToLower(name)]
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ==================== 删除辅助函数 ====================

// deleteGeneratedFiles 删除生成的文件（优先从 options 读已保存的路径，兜底推导）
func deleteGeneratedFiles(ctx context.Context, record *entity.SysGenCodes, opts OptionsJson) {
	pkgName := strings.ToLower(record.VarName)
	varLower := strings.ToLower(record.VarName[:1]) + record.VarName[1:]
	snakeName := toSnake(varLower)

	// 优先从 options.generatedFiles 读取已保存的路径
	var allFiles []string
	useSaved := false
	if record.Options != nil {
		optMap := record.Options.Map()
		if gf, ok := optMap["generatedFiles"].(map[string]interface{}); ok {
			for _, key := range []string{"server", "frontend"} {
				if arr, ok := gf[key].([]interface{}); ok {
					for _, v := range arr {
						if s, ok := v.(string); ok && s != "" {
							allFiles = append(allFiles, s)
						}
					}
				}
			}
			if len(allFiles) > 0 {
				useSaved = true
			}
		}
	}

	if useSaved {
		g.Log().Infof(ctx, "[Delete] using %d saved file paths", len(allFiles))
		for _, absPath := range allFiles {
			if absPath != "" && gfile.Exists(absPath) {
				if err := os.Remove(absPath); err != nil {
					g.Log().Warningf(ctx, "[Delete] remove %s error: %v", absPath, err)
				} else {
					g.Log().Infof(ctx, "[Delete] removed: %s", absPath)
				}
			}
		}
	} else {
		// 兜底：推导路径（兼容旧数据）
		g.Log().Info(ctx, "[Delete] no saved paths, falling back to inference")
		deleteGeneratedFilesFallback(ctx, record, opts)
	}

	// 额外删除 gf gen 自动生成的文件（不在模板列表中）
	autoFiles := []string{
		fmt.Sprintf("internal/service/%s.go", pkgName),
		fmt.Sprintf("internal/dao/%s.go", snakeName),
		fmt.Sprintf("internal/dao/internal/%s.go", snakeName),
		fmt.Sprintf("internal/model/entity/%s.go", snakeName),
		fmt.Sprintf("internal/model/do/%s.go", snakeName),
	}
	for _, f := range autoFiles {
		absPath := resolveOutputPath(ctx, f, true)
		if absPath != "" && gfile.Exists(absPath) {
			_ = os.Remove(absPath)
		}
	}

	// 删除 logic 目录
	logicDir := resolveOutputPath(ctx, fmt.Sprintf("internal/logic/%s", pkgName), true)
	if logicDir != "" && gfile.IsDir(logicDir) {
		_ = os.RemoveAll(logicDir)
	}

	// 尝试删除前端模块目录（从 options.modulePath 读取）
	if record.Options != nil {
		if mp := record.Options.Map()["modulePath"]; mp != nil {
			tpl := genconfig.GetDefaultTemplate(ctx)
			viewDir := resolveOutputPath(ctx, fmt.Sprintf("%s/%s", tpl.WebViewsPath, mp), false)
			if viewDir != "" && gfile.IsDir(viewDir) {
				_ = os.RemoveAll(viewDir)
				g.Log().Infof(ctx, "[Delete] removed frontend dir: %s", viewDir)
			}
		}
	}

	// 清理 logic/logic.go 中的 import（仅主包模式）
	if opts.AddonName == "" {
		unregisterLogicImport(ctx, pkgName)
	}

	g.Log().Infof(ctx, "[Delete] files cleanup completed for %s", record.VarName)
}

// deleteGeneratedFilesFallback 兜底：通过推导删除文件（旧数据没有保存路径时）
func deleteGeneratedFilesFallback(ctx context.Context, record *entity.SysGenCodes, opts OptionsJson) {
	tpl := genconfig.GetDefaultTemplate(ctx)
	varLower := strings.ToLower(record.VarName[:1]) + record.VarName[1:]
	snakeName := toSnake(varLower)
	pkgName := strings.ToLower(record.VarName)
	routeName := camelToKebab(lcFirst(record.VarName))
	modulePath := routeName

	if opts.Menu.Pid > 0 {
		parentPath := getMenuPath(ctx, opts.Menu.Pid)
		if parentPath != "" {
			modulePath = parentPath + "/" + routeName
		}
	}
	if opts.GenPaths != nil && opts.GenPaths["webApi"] != "" {
		apiPath := opts.GenPaths["webApi"]
		apiPath = strings.TrimPrefix(apiPath, "api/backend/")
		apiPath = strings.TrimSuffix(apiPath, ".ts")
		modulePath = apiPath
	}

	files := []struct {
		path     string
		isServer bool
	}{
		{fmt.Sprintf("%s/admin_%s.go", tpl.ApiPath, snakeName), true},
		{fmt.Sprintf("%s/%s.go", tpl.InputPath, snakeName), true},
		{fmt.Sprintf("%s/%s.go", tpl.ControllerPath, snakeName), true},
		{fmt.Sprintf("%s/%s/%s.go", tpl.LogicPath, pkgName, snakeName), true},
		{fmt.Sprintf("%s/menu_%s.sql", tpl.SqlPath, snakeName), true},
		{fmt.Sprintf("%s/%s.ts", tpl.WebApiPath, modulePath), false},
		{fmt.Sprintf("%s/%s/index.vue", tpl.WebViewsPath, modulePath), false},
	}
	filePrefix := camelToKebab(lcFirst(record.VarName))
	files = append(files,
		struct {
			path     string
			isServer bool
		}{fmt.Sprintf("%s/%s/modules/%s-dialog.vue", tpl.WebViewsPath, modulePath, filePrefix), false},
		struct {
			path     string
			isServer bool
		}{fmt.Sprintf("%s/%s/modules/%s-search.vue", tpl.WebViewsPath, modulePath, filePrefix), false},
		struct {
			path     string
			isServer bool
		}{fmt.Sprintf("%s/%s/modules/%s-detail-drawer.vue", tpl.WebViewsPath, modulePath, filePrefix), false},
		struct {
			path     string
			isServer bool
		}{fmt.Sprintf("%s/%s/detail/index.vue", tpl.WebViewsPath, modulePath), false},
	)

	for _, f := range files {
		absPath := resolveOutputPath(ctx, f.path, f.isServer)
		if absPath != "" && gfile.Exists(absPath) {
			_ = os.Remove(absPath)
			g.Log().Infof(ctx, "[Delete] removed: %s", absPath)
		}
	}

	// 清理前端目录
	viewDir := resolveOutputPath(ctx, fmt.Sprintf("%s/%s", tpl.WebViewsPath, modulePath), false)
	if viewDir != "" && gfile.IsDir(viewDir) {
		_ = os.RemoveAll(viewDir)
	}
}

// deleteGeneratedMenus 删除生成的菜单
func deleteGeneratedMenus(ctx context.Context, routeName, varName string) {
	db := g.DB()

	// 查找页面菜单（type=2, name 匹配）
	pageRecords, err := db.GetAll(ctx,
		"SELECT id FROM xy_admin_menu WHERE name = ? AND type = 2", varName)
	if err != nil {
		g.Log().Warningf(ctx, "[Delete] query page menus error: %v", err)
		return
	}

	for _, row := range pageRecords {
		pageId := row["id"].Int64()
		// 删除所有子菜单（按钮权限 type=3 + 详情页路由 type=2 等）
		_, _ = db.Exec(ctx, "DELETE FROM xy_admin_menu WHERE parent_id = ?", pageId)
		// 删除页面菜单本身
		_, _ = db.Exec(ctx, "DELETE FROM xy_admin_menu WHERE id = ?", pageId)
		g.Log().Infof(ctx, "[Delete] removed menu page #%d and all children", pageId)
	}

	// 删除详情页菜单（name = VarNameDetail，跟列表页同级）
	detailName := varName + "Detail"
	_, _ = db.Exec(ctx, "DELETE FROM xy_admin_menu WHERE name = ?", detailName)

	// 查找目录菜单（type=1, name 匹配 VarNameDir）
	dirName := varName + "Dir"
	dirResult, err := db.GetOne(ctx,
		"SELECT id FROM xy_admin_menu WHERE name = ? AND type = 1", dirName)
	if err == nil && dirResult != nil && !dirResult.IsEmpty() {
		dirId := dirResult["id"].Int64()
		// 检查目录下是否还有其他子菜单
		childCount, _ := db.GetCount(ctx,
			"SELECT COUNT(*) FROM xy_admin_menu WHERE parent_id = ?", dirId)
		if childCount == 0 {
			// 目录下没有其他子菜单了，可以安全删除
			_, _ = db.Exec(ctx, "DELETE FROM xy_admin_menu WHERE id = ?", dirId)
			g.Log().Infof(ctx, "[Delete] removed empty directory menu #%d", dirId)
		} else {
			g.Log().Infof(ctx, "[Delete] directory #%d still has %d children, kept", dirId, childCount)
		}
	}

	g.Log().Infof(ctx, "[Delete] menu cleanup completed for %s", varName)
}
