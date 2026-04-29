package {{.GoInputPkg}}

import (
{{- if or .NeedsGtime .NeedsGjson}}
{{- if .NeedsGtime}}
	"github.com/gogf/gf/v2/os/gtime"
{{- end}}
{{- if .NeedsGjson}}
	"github.com/gogf/gf/v2/encoding/gjson"
{{- end}}
{{- end}}
	"xygo/internal/model/input/form"
)

// ==================== {{.TableComment}} ====================

// {{.VarName}}ListInp {{.TableComment}}列表入参
type {{.VarName}}ListInp struct {
	form.PageReq
{{- range .QueryColumns}}
{{- if eq .QueryType "between"}}
	{{.GoName}}Start string `json:"{{.TsName}}Start" dc:"{{.Comment}}开始值"`
	{{.GoName}}End string `json:"{{.TsName}}End" dc:"{{.Comment}}结束值"`
{{- else if eq .GoType "string"}}
	{{.GoName}} {{.GoType}} `json:"{{.TsName}}" dc:"{{.Comment}}"`
{{- else if contains .GoType "*"}}
	{{.GoName}} {{.GoType}} `json:"{{.TsName}}" dc:"{{.Comment}}"`
{{- else}}
	{{.GoName}} *{{.GoType}} `json:"{{.TsName}}" dc:"{{.Comment}}"`
{{- end}}
{{- end}}
{{- if .HasRelations}}
	// 关联表搜索字段
{{- range $rel := .Relations}}
{{- if $rel.FieldConfigs}}
{{- range $fc := $rel.FieldConfigs}}
{{- if $fc.InSearch}}
{{- if eq $fc.SearchType "between"}}
	{{$rel.RelationName}}{{$fc.GoName}}Start string `json:"{{$rel.RelationAlias}}_{{$fc.Field}}Start" dc:"{{$fc.Label}}开始"`
	{{$rel.RelationName}}{{$fc.GoName}}End string `json:"{{$rel.RelationAlias}}_{{$fc.Field}}End" dc:"{{$fc.Label}}结束"`
{{- else}}
	{{$rel.RelationName}}{{$fc.GoName}} string `json:"{{$rel.RelationAlias}}_{{$fc.Field}}" dc:"{{$fc.Label}}"`
{{- end}}
{{- end}}
{{- end}}
{{- else}}
{{- range $f := $rel.SearchFields}}
	{{$rel.RelationName}}{{pascalCase $f}} string `json:"{{$rel.RelationAlias}}_{{$f}}" dc:"关联{{$rel.RelationName}}{{$f}}"`
{{- end}}
{{- end}}
{{- end}}
{{- end}}
}

// {{.VarName}}ListItem {{.TableComment}}列表项
type {{.VarName}}ListItem struct {
{{- range .ListColumns}}
	{{.GoName}} {{.GoType}} `json:"{{.TsName}}" dc:"{{.Comment}}"`
{{- end}}
{{- if .HasRelations}}
	// 关联表字段（来自 LeftJoin）
{{- range $rel := .Relations}}
{{- if not $rel.IsMultiple}}
	{{$rel.RelationName}}{{pascalCase $rel.RemoteField}} string `json:"{{$rel.RelationAlias}}_{{$rel.RemoteField}}" dc:"{{$rel.RelationName}}{{$rel.RemoteField}}"`
{{- range $f := $rel.RelationFields}}
{{- if ne $f $rel.RemoteField}}
	{{$rel.RelationName}}{{pascalCase $f}} string `json:"{{$rel.RelationAlias}}_{{$f}}" dc:"{{$rel.RelationName}}{{$f}}"`
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
}

// {{.VarName}}ListModel {{.TableComment}}列表出参
type {{.VarName}}ListModel struct {
	List []{{.VarName}}ListItem `json:"list"`
	form.PageRes
}
{{- if .HasView}}

// {{.VarName}}ViewModel {{.TableComment}}详情出参
type {{.VarName}}ViewModel struct {
{{- range .AllColumns}}
	{{.GoName}} {{.GoType}} `json:"{{.TsName}}" dc:"{{.Comment}}"`
{{- end}}
}
{{- end}}
{{- if or .HasAdd .HasEdit}}

// {{.VarName}}EditInp {{.TableComment}}编辑入参
type {{.VarName}}EditInp struct {
{{- range .EditColumns}}
	{{.GoName}} {{.GoType}} `json:"{{.TsName}}"{{if .Required}} v:"required#{{.Comment}}不能为空"{{end}} dc:"{{.Comment}}"`
{{- end}}
}
{{- end}}
