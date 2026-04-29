package {{.GoControllerPkg}}

import (
	"context"

	api "{{.GoApiImport}}"
	"{{.GoServiceImport}}"
)

// {{.VarName}}List {{.TableComment}}列表
func (c *ControllerV1) {{.VarName}}List(ctx context.Context, req *api.{{.VarName}}ListReq) (res *api.{{.VarName}}ListRes, err error) {
	result, err := service.{{.VarName}}().List(ctx, &req.{{.VarName}}ListInp)
	if err != nil {
		return nil, err
	}
	return &api.{{.VarName}}ListRes{result}, nil
}
{{- if .HasView}}

// {{.VarName}}View {{.TableComment}}详情
func (c *ControllerV1) {{.VarName}}View(ctx context.Context, req *api.{{.VarName}}ViewReq) (res *api.{{.VarName}}ViewRes, err error) {
	result, err := service.{{.VarName}}().View(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.{{.VarName}}ViewRes{result}, nil
}
{{- end}}
{{- if or .HasAdd .HasEdit}}

// {{.VarName}}Edit 保存{{.TableComment}}
func (c *ControllerV1) {{.VarName}}Edit(ctx context.Context, req *api.{{.VarName}}EditReq) (res *api.{{.VarName}}EditRes, err error) {
	err = service.{{.VarName}}().Edit(ctx, &req.{{.VarName}}EditInp)
	return &api.{{.VarName}}EditRes{}, err
}
{{- end}}
{{- if or .HasDel .HasBatchDel}}

// {{.VarName}}Delete 删除{{.TableComment}}
func (c *ControllerV1) {{.VarName}}Delete(ctx context.Context, req *api.{{.VarName}}DeleteReq) (res *api.{{.VarName}}DeleteRes, err error) {
	err = service.{{.VarName}}().Delete(ctx, req.Id)
	return &api.{{.VarName}}DeleteRes{}, err
}
{{- end}}
