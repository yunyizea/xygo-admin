package {{.PkgName}}

import (
	"context"
{{- if .HasInQuery}}
	"strings"
{{- end}}
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/dao"
	adminin "{{.GoInputImport}}"
	"{{.GoServiceImport}}"
)

type s{{.VarName}} struct{}

func init() {
	service.Register{{.VarName}}(New())
}

func New() *s{{.VarName}} {
	return &s{{.VarName}}{}
}

// List {{.TableComment}}列表（返回全部，前端构建树）
func (s *s{{.VarName}}) List(ctx context.Context, in *adminin.{{.VarName}}ListInp) (*adminin.{{.VarName}}ListModel, error) {
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

	var list []adminin.{{.VarName}}ListItem
	err := model.OrderAsc("sort").OrderAsc("{{.PkColumn}}").Scan(&list)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []adminin.{{.VarName}}ListItem{}
	}

	return &adminin.{{.VarName}}ListModel{
		List: list,
	}, nil
}

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

// Delete 删除{{.TableComment}}
func (s *s{{.VarName}}) Delete(ctx context.Context, id uint64) error {
	// 检查是否有子项
	count, err := dao.{{.DaoName}}.Ctx(ctx).Where("{{.TreePidColumn}}", id).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("请先删除子项后再删除此记录")
	}
	_, err = dao.{{.DaoName}}.Ctx(ctx).Where("{{.PkColumn}}", id).Delete()
	return err
}
