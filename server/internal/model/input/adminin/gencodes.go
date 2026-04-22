// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package adminin

import "xygo/internal/model/input/form"

// ==================== 生成记录列表 ====================

// GenCodesListInp 生成记录列表入参
type GenCodesListInp struct {
	form.PageReq
	GenType int    `json:"genType" dc:"生成类型"`
	VarName string `json:"varName" dc:"实体名"`
	Status  int    `json:"status"  dc:"状态"`
}

// GenCodesListItem 生成记录列表项
type GenCodesListItem struct {
	Id           uint64 `json:"id"           dc:"ID"`
	GenType      int    `json:"genType"      dc:"生成类型"`
	DbName       string `json:"dbName"       dc:"数据库名"`
	TableName    string `json:"tableName"    dc:"表名"`
	TableComment string `json:"tableComment" dc:"表注释"`
	VarName      string `json:"varName"      dc:"实体名"`
	Status       int    `json:"status"       dc:"状态"`
	Options      string `json:"options"      dc:"选项JSON"`
	CreatedAt    int64  `json:"createdAt"    dc:"创建时间"`
	UpdatedAt    int64  `json:"updatedAt"    dc:"更新时间"`
}

// GenCodesListModel 生成记录列表出参
type GenCodesListModel struct {
	List []GenCodesListItem `json:"list"`
	form.PageRes
}

// ==================== 查看详情 ====================

type GenCodesViewInp struct {
	Id uint64 `json:"id" v:"required#ID不能为空" dc:"ID"`
}

type GenCodesViewModel struct {
	GenCodesListItem
	Columns []GenCodesColumnItem `json:"columns" dc:"字段列表"`
}

// ==================== 保存配置 ====================

type GenCodesEditInp struct {
	Id           uint64               `json:"id"           dc:"ID(0=新增)"`
	GenType      int                  `json:"genType"      v:"required#生成类型不能为空" dc:"生成类型"`
	DbName       string               `json:"dbName"       dc:"数据库名"`
	TableName    string               `json:"tableName"    v:"required#表名不能为空"    dc:"表名"`
	TableComment string               `json:"tableComment" v:"required#表注释不能为空"   dc:"表注释"`
	VarName      string               `json:"varName"      v:"required#实体名不能为空"   dc:"实体名"`
	Options      string               `json:"options"      dc:"选项JSON"`
	Columns      []GenCodesColumnItem `json:"columns"      dc:"字段列表"`
}

type GenCodesEditModel struct {
	Id uint64 `json:"id" dc:"ID"`
}

// ==================== 删除 ====================

type GenCodesDeleteInp struct {
	Id          uint64 `json:"id" v:"required#ID不能为空" dc:"ID"`
	DeleteFiles bool   `json:"deleteFiles" dc:"是否删除生成的文件"`
	DeleteMenus bool   `json:"deleteMenus" dc:"是否删除生成的菜单"`
}

// ==================== 预览 ====================

type GenCodesPreviewInp struct {
	GenCodesEditInp
}

type GenCodesPreviewFile struct {
	Path    string `json:"path"    dc:"文件路径"`
	Content string `json:"content" dc:"文件内容"`
	Lang    string `json:"lang"    dc:"语言标识"`
}

type GenCodesPreviewModel struct {
	Files []GenCodesPreviewFile `json:"files" dc:"预览文件列表"`
}

// ==================== 生成 ====================

type GenCodesBuildInp struct {
	GenCodesEditInp
}

// ==================== 选项 ====================

type GenCodesSelectsInp struct{}

type SelectOption struct {
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

type GenCodesSelectsModel struct {
	GenType     []SelectOption    `json:"genType"     dc:"生成类型"`
	FormType    []SelectOption    `json:"formType"    dc:"表单组件"`
	QueryType   []SelectOption    `json:"queryType"   dc:"查询方式"`
	DesignTypes []SelectOption    `json:"designTypes" dc:"设计类型"`
	GenPaths    map[string]string `json:"genPaths"    dc:"默认生成路径配置"`
	AddonList   []SelectOption    `json:"addonList"   dc:"已安装扩展列表"`
}

// ==================== 数据库表选项 ====================

type GenCodesTableSelectInp struct{}

type GenCodesTableSelectItem struct {
	TableName    string `json:"tableName"    dc:"表名"`
	TableComment string `json:"tableComment" dc:"表注释"`
	VarName      string `json:"varName"      dc:"推荐实体名"`
}

type GenCodesTableSelectModel struct {
	List []GenCodesTableSelectItem `json:"list"`
}

// ==================== 表字段列表 ====================

type GenCodesColumnListInp struct {
	TableName string `json:"tableName" v:"required#表名不能为空" dc:"表名"`
}

// GenCodesColumnItem 字段配置项
type GenCodesColumnItem struct {
	Id         uint64 `json:"id"         dc:"ID"`
	GenId      uint64 `json:"genId"      dc:"关联ID"`
	Name       string `json:"name"       dc:"字段名"`
	GoName     string `json:"goName"     dc:"Go字段名"`
	TsName     string `json:"tsName"     dc:"TS字段名"`
	DbType     string `json:"dbType"     dc:"数据库类型"`
	GoType     string `json:"goType"     dc:"Go类型"`
	TsType     string `json:"tsType"     dc:"TS类型"`
	Comment    string `json:"comment"    dc:"注释"`
	IsPk       int    `json:"isPk"       dc:"主键"`
	IsRequired int    `json:"isRequired" dc:"必填"`
	IsList     int    `json:"isList"     dc:"表格显示"`
	IsEdit     int    `json:"isEdit"     dc:"表单显示"`
	IsQuery    int    `json:"isQuery"    dc:"搜索"`
	QueryType  string `json:"queryType"  dc:"查询方式"`
	FormType   string `json:"formType"   dc:"表单组件"`
	DesignType string `json:"designType" dc:"设计类型"`
	Extra      string `json:"extra"      dc:"扩展配置JSON"`
	DictType   string `json:"dictType"   dc:"字典类型"`
	Sort       int    `json:"sort"       dc:"排序"`
}

type GenCodesColumnListModel struct {
	List []GenCodesColumnItem `json:"list"`
}

// ==================== 同步字段到数据库 ====================

type GenCodesSyncFieldsInp struct {
	TableName string              `json:"tableName" v:"required#表名不能为空" dc:"表名"`
	Columns   []CreateTableColumn `json:"columns"   v:"required#字段不能为空" dc:"设计器字段列表"`
}

// FieldDiff 单个字段的差异
type FieldDiff struct {
	Name      string `json:"name"      dc:"字段名"`
	Action    string `json:"action"    dc:"操作: add/drop/modify"`
	Detail    string `json:"detail"    dc:"变更详情"`
	SQL       string `json:"sql"       dc:"DDL SQL"`
	IsRisky   bool   `json:"isRisky"   dc:"是否有风险（如删除、缩小类型）"`
}

type GenCodesSyncFieldsModel struct {
	Diffs     []FieldDiff `json:"diffs"     dc:"字段差异列表"`
	HasChange bool        `json:"hasChange" dc:"是否有变更"`
}

type GenCodesExecuteDDLInp struct {
	TableName string   `json:"tableName" v:"required#表名不能为空" dc:"表名"`
	SQLs      []string `json:"sqls"      v:"required#SQL不能为空" dc:"要执行的DDL SQL列表"`
}

// ==================== 创建数据表 ====================

type GenCodesCreateTableInp struct {
	TableName    string               `json:"tableName"    v:"required#表名不能为空"  dc:"表名"`
	TableComment string               `json:"tableComment" v:"required#表注释不能为空" dc:"表注释"`
	Columns      []CreateTableColumn  `json:"columns"      v:"required#字段不能为空"  dc:"字段列表"`
}

type CreateTableColumn struct {
	Name         string `json:"name"         v:"required#字段名不能为空" dc:"字段名"`
	Type         string `json:"type"         v:"required#类型不能为空"  dc:"字段类型"`
	Comment      string `json:"comment"      dc:"注释"`
	IsPk         int    `json:"isPk"         dc:"主键"`
	IsNullable   int    `json:"isNullable"   dc:"可空"`
	DefaultValue string `json:"defaultValue" dc:"默认值"`
}

type GenCodesCreateTableModel struct {
	TableName string `json:"tableName" dc:"创建的表名"`
}
