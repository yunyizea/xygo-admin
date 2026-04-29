package {{.PkgName}}

import (
	"context"
{{- if .HasInQuery}}
	"strings"
{{- end}}

{{- if .HasView}}
	"github.com/gogf/gf/v2/errors/gerror"
{{- end}}
{{- if or .HasAdd .HasEdit}}
	"github.com/gogf/gf/v2/frame/g"
{{- end}}

	"xygo/internal/dao"
	adminin "{{.GoInputImport}}"
	"xygo/internal/model/input/form"
	"{{.GoServiceImport}}"
)

type s{{.VarName}} struct{}

func init() {
	service.Register{{.VarName}}(New())
}

func New() *s{{.VarName}} {
	return &s{{.VarName}}{}
}

// List {{.TableComment}}列表
func (s *s{{.VarName}}) List(ctx context.Context, in *adminin.{{.VarName}}ListInp) (*adminin.{{.VarName}}ListModel, error) {
{{- if .HasRelations}}
{{- if .HasRelSoftDelete}}
	model := dao.{{.DaoName}}.Ctx(ctx).As("t").Unscoped()
{{- else}}
	model := dao.{{.DaoName}}.Ctx(ctx).As("t")
{{- end}}
	// 关联表 LeftJoin
{{- range $rel := .Relations}}
{{- if not $rel.IsMultiple}}
{{- if $rel.HasSoftDelete}}
	model = model.LeftJoin("{{$rel.RemoteTable}} {{$rel.RelationAlias}}", "{{$rel.RelationAlias}}.{{$rel.RemotePk}} = t.{{$rel.FieldName}} AND {{$rel.RelationAlias}}.deleted_at = 0")
{{- else}}
	model = model.LeftJoin("{{$rel.RemoteTable}} {{$rel.RelationAlias}}", "{{$rel.RelationAlias}}.{{$rel.RemotePk}} = t.{{$rel.FieldName}}")
{{- end}}
{{- end}}
{{- end}}
{{- range .QueryColumns}}
{{- if eq .QueryType "like"}}
	if in.{{.GoName}} != "" {
		model = model.WhereLike("t.{{.Name}}", "%"+in.{{.GoName}}+"%")
	}
{{- else if eq .QueryType "eq"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("t.{{.Name}}", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("t.{{.Name}}", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "neq"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("t.{{.Name}} <>", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("t.{{.Name}} <>", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "gt"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("t.{{.Name}} > ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("t.{{.Name}} > ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "gte"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("t.{{.Name}} >= ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("t.{{.Name}} >= ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "lt"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("t.{{.Name}} < ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("t.{{.Name}} < ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "lte"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("t.{{.Name}} <= ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("t.{{.Name}} <= ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "in"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		arr := strings.Split(in.{{.GoName}}, ",")
		clean := make([]string, 0, len(arr))
		for _, v := range arr {
			v = strings.TrimSpace(v)
			if v != "" {
				clean = append(clean, v)
			}
		}
		if len(clean) > 0 {
			model = model.WhereIn("t.{{.Name}}", clean)
		}
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.WhereIn("t.{{.Name}}", []interface{}{*in.{{.GoName}}})
	}
{{- end}}
{{- else if eq .QueryType "between"}}
	if in.{{.GoName}}Start != "" && in.{{.GoName}}End != "" {
		model = model.WhereBetween("t.{{.Name}}", in.{{.GoName}}Start, in.{{.GoName}}End)
	}
{{- else}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("t.{{.Name}}", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("t.{{.Name}}", *in.{{.GoName}})
	}
{{- end}}
{{- end}}
{{- end}}
{{- if .HasRelations}}
	// 关联表搜索条件
{{- range $rel := .Relations}}
{{- if $rel.FieldConfigs}}
{{- range $fc := $rel.FieldConfigs}}
{{- if $fc.InSearch}}
{{- if eq $fc.SearchType "like"}}
	if in.{{$rel.RelationName}}{{$fc.GoName}} != "" {
		model = model.WhereLike("{{$rel.RelationAlias}}.{{$fc.Field}}", "%"+in.{{$rel.RelationName}}{{$fc.GoName}}+"%")
	}
{{- else if eq $fc.SearchType "eq"}}
	if in.{{$rel.RelationName}}{{$fc.GoName}} != "" {
		model = model.Where("{{$rel.RelationAlias}}.{{$fc.Field}}", in.{{$rel.RelationName}}{{$fc.GoName}})
	}
{{- else if eq $fc.SearchType "between"}}
	if in.{{$rel.RelationName}}{{$fc.GoName}}Start != "" && in.{{$rel.RelationName}}{{$fc.GoName}}End != "" {
		model = model.WhereBetween("{{$rel.RelationAlias}}.{{$fc.Field}}", in.{{$rel.RelationName}}{{$fc.GoName}}Start, in.{{$rel.RelationName}}{{$fc.GoName}}End)
	}
{{- else}}
	if in.{{$rel.RelationName}}{{$fc.GoName}} != "" {
		model = model.WhereLike("{{$rel.RelationAlias}}.{{$fc.Field}}", "%"+in.{{$rel.RelationName}}{{$fc.GoName}}+"%")
	}
{{- end}}
{{- end}}
{{- end}}
{{- else}}
{{- range $f := $rel.SearchFields}}
	if in.{{$rel.RelationName}}{{pascalCase $f}} != "" {
		model = model.WhereLike("{{$rel.RelationAlias}}.{{$f}}", "%"+in.{{$rel.RelationName}}{{pascalCase $f}}+"%")
	}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
	// 先计数（不带 Fields，避免 COUNT + 字段别名冲突）
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
	// 计数后添加 Fields
	model = model.Fields("t.*")
{{- range $rel := .Relations}}
{{- if not $rel.IsMultiple}}
	model = model.Fields("{{$rel.RelationAlias}}.{{$rel.RemoteField}} as {{$rel.RelationAlias}}_{{$rel.RemoteField}}")
{{- range $f := $rel.RelationFields}}
{{- if ne $f $rel.RemoteField}}
	model = model.Fields("{{$rel.RelationAlias}}.{{$f}} as {{$rel.RelationAlias}}_{{$f}}")
{{- end}}
{{- end}}
{{- end}}
{{- end}}
	var list []adminin.{{.VarName}}ListItem
	err = model.Page(in.Page, in.PageSize).OrderDesc("t.{{.PkColumn}}").Scan(&list)
{{- else}}
	model := dao.{{.DaoName}}.Ctx(ctx)
{{- range .QueryColumns}}
{{- if eq .QueryType "like"}}
	if in.{{.GoName}} != "" {
		model = model.WhereLike("{{.Name}}", "%"+in.{{.GoName}}+"%")
	}
{{- else if eq .QueryType "eq"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("{{.Name}}", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("{{.Name}}", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "neq"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("{{.Name}} <>", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("{{.Name}} <>", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "gt"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("{{.Name}} > ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("{{.Name}} > ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "gte"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("{{.Name}} >= ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("{{.Name}} >= ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "lt"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("{{.Name}} < ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("{{.Name}} < ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "lte"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("{{.Name}} <= ?", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("{{.Name}} <= ?", *in.{{.GoName}})
	}
{{- end}}
{{- else if eq .QueryType "in"}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		arr := strings.Split(in.{{.GoName}}, ",")
		clean := make([]string, 0, len(arr))
		for _, v := range arr {
			v = strings.TrimSpace(v)
			if v != "" {
				clean = append(clean, v)
			}
		}
		if len(clean) > 0 {
			model = model.WhereIn("{{.Name}}", clean)
		}
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.WhereIn("{{.Name}}", []interface{}{*in.{{.GoName}}})
	}
{{- end}}
{{- else if eq .QueryType "between"}}
	if in.{{.GoName}}Start != "" && in.{{.GoName}}End != "" {
		model = model.WhereBetween("{{.Name}}", in.{{.GoName}}Start, in.{{.GoName}}End)
	}
{{- else}}
{{- if eq .GoType "string"}}
	if in.{{.GoName}} != "" {
		model = model.Where("{{.Name}}", in.{{.GoName}})
	}
{{- else}}
	if in.{{.GoName}} != nil {
		model = model.Where("{{.Name}}", *in.{{.GoName}})
	}
{{- end}}
{{- end}}
{{- end}}
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
	var list []adminin.{{.VarName}}ListItem
	err = model.Page(in.Page, in.PageSize).OrderDesc("{{.PkColumn}}").Scan(&list)
{{- end}}
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []adminin.{{.VarName}}ListItem{}
	}

	return &adminin.{{.VarName}}ListModel{
		List: list,
		PageRes: form.PageRes{
			Page:     in.Page,
			PageSize: in.PageSize,
			Total:    count,
		},
	}, nil
}
{{- if .HasView}}

// View {{.TableComment}}详情
func (s *s{{.VarName}}) View(ctx context.Context, id uint64) (*adminin.{{.VarName}}ViewModel, error) {
	var item adminin.{{.VarName}}ViewModel
	err := dao.{{.DaoName}}.Ctx(ctx).Where("{{.PkColumn}}", id).Scan(&item)
	if err != nil {
		return nil, err
	}
	if item.{{.PkGoName}} == 0 {
		return nil, gerror.New("记录不存在")
	}
	return &item, nil
}
{{- end}}
{{- if or .HasAdd .HasEdit}}

// Edit 保存{{.TableComment}}
func (s *s{{.VarName}}) Edit(ctx context.Context, in *adminin.{{.VarName}}EditInp) error {
	data := g.Map{
{{- range .EditColumns}}
{{- if and (ne .Name $.PkColumn) (not .IsTimeField)}}
		"{{.Name}}": in.{{.GoName}},
{{- end}}
{{- end}}
	}

	if in.{{.PkGoName}} == 0 {
		// 新增（created_at/updated_at 由 GoFrame 自动维护）
		_, err := dao.{{.DaoName}}.Ctx(ctx).Data(data).Insert()
		return err
	}

	// 更新（updated_at 由 GoFrame 自动维护）
	_, err := dao.{{.DaoName}}.Ctx(ctx).Where("{{.PkColumn}}", in.{{.PkGoName}}).Data(data).Update()
	return err
}
{{- end}}
{{- if or .HasDel .HasBatchDel}}

// Delete 删除{{.TableComment}}
func (s *s{{.VarName}}) Delete(ctx context.Context, id uint64) error {
	_, err := dao.{{.DaoName}}.Ctx(ctx).Where("{{.PkColumn}}", id).Delete()
	return err
}
{{- end}}
