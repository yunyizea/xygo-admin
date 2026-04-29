package {{.GoApiPkg}}

import (
	"github.com/gogf/gf/v2/frame/g"
	adminin "{{.GoInputImport}}"
)

// {{.VarName}}ListReq {{.TableComment}}列表请求
type {{.VarName}}ListReq struct {
	g.Meta `path:"/admin/{{.RouteName}}/list" method:"get" tags:"{{.VarName}}" summary:"{{.TableComment}}列表"`
	adminin.{{.VarName}}ListInp
}

type {{.VarName}}ListRes struct {
	*adminin.{{.VarName}}ListModel
}
{{- if .HasView}}

// {{.VarName}}ViewReq {{.TableComment}}详情请求
type {{.VarName}}ViewReq struct {
	g.Meta `path:"/admin/{{.RouteName}}/view" method:"get" tags:"{{.VarName}}" summary:"{{.TableComment}}详情"`
	Id uint64 `json:"id" v:"required#ID不能为空" dc:"ID"`
}

type {{.VarName}}ViewRes struct {
	*adminin.{{.VarName}}ViewModel
}
{{- end}}
{{- if or .HasAdd .HasEdit}}

// {{.VarName}}EditReq {{.TableComment}}保存请求
type {{.VarName}}EditReq struct {
	g.Meta `path:"/admin/{{.RouteName}}/edit" method:"post" tags:"{{.VarName}}" summary:"保存{{.TableComment}}"`
	adminin.{{.VarName}}EditInp
}

type {{.VarName}}EditRes struct{}
{{- end}}
{{- if or .HasDel .HasBatchDel}}

// {{.VarName}}DeleteReq {{.TableComment}}删除请求
type {{.VarName}}DeleteReq struct {
	g.Meta `path:"/admin/{{.RouteName}}/delete" method:"post" tags:"{{.VarName}}" summary:"删除{{.TableComment}}"`
	Id uint64 `json:"id" v:"required#ID不能为空" dc:"ID"`
}

type {{.VarName}}DeleteRes struct{}
{{- end}}
