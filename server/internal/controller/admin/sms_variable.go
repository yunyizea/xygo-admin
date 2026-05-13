package admin

import (
	"context"

	api "xygo/api/admin"
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
)

func (c *ControllerV1) SmsVariableList(ctx context.Context, req *api.SmsVariableListReq) (res *api.SmsVariableListRes, err error) {
	model, err := service.Sms().VariableList(ctx, &adminin.SmsVariableListInp{
		Page: req.Page, Size: req.Size, Name: req.Name, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	res = &api.SmsVariableListRes{List: model.List, Total: model.Total, Page: model.Page, Size: model.Size}
	return
}

func (c *ControllerV1) SmsVariableSave(ctx context.Context, req *api.SmsVariableSaveReq) (res *api.SmsVariableSaveRes, err error) {
	err = service.Sms().VariableSave(ctx, &adminin.SmsVariableSaveInp{
		Id: req.Id, Title: req.Title, Name: req.Name,
		SourceType: req.SourceType, SqlQuery: req.SqlQuery,
		MethodName: req.MethodName, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &api.SmsVariableSaveRes{}, nil
}

func (c *ControllerV1) SmsVariableDelete(ctx context.Context, req *api.SmsVariableDeleteReq) (res *api.SmsVariableDeleteRes, err error) {
	if err = service.Sms().VariableDelete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &api.SmsVariableDeleteRes{}, nil
}
