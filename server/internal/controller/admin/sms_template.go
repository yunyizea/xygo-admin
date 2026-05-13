package admin

import (
	"context"

	api "xygo/api/admin"
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
)

func (c *ControllerV1) SmsTemplateList(ctx context.Context, req *api.SmsTemplateListReq) (res *api.SmsTemplateListRes, err error) {
	model, err := service.Sms().TemplateList(ctx, &adminin.SmsTemplateListInp{
		Page: req.Page, Size: req.Size, Status: req.Status, Code: req.Code, Title: req.Title,
	})
	if err != nil {
		return nil, err
	}
	res = &api.SmsTemplateListRes{List: model.List, Total: model.Total, Page: model.Page, Size: model.Size}
	return
}

func (c *ControllerV1) SmsTemplateSave(ctx context.Context, req *api.SmsTemplateSaveReq) (res *api.SmsTemplateSaveRes, err error) {
	err = service.Sms().TemplateSave(ctx, &adminin.SmsTemplateSaveInp{
		Id: req.Id, Title: req.Title, Code: req.Code, Content: req.Content,
		ProviderTemplateId: req.ProviderTemplateId, Variables: req.Variables,
		RelatedVariableId: req.RelatedVariableId, Status: req.Status,
		Sort: req.Sort, Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &api.SmsTemplateSaveRes{}, nil
}

func (c *ControllerV1) SmsTemplateDelete(ctx context.Context, req *api.SmsTemplateDeleteReq) (res *api.SmsTemplateDeleteRes, err error) {
	if err = service.Sms().TemplateDelete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &api.SmsTemplateDeleteRes{}, nil
}

func (c *ControllerV1) SmsTemplateTest(ctx context.Context, req *api.SmsTemplateTestReq) (res *api.SmsTemplateTestRes, err error) {
	model, err := service.Sms().TemplateTest(ctx, &adminin.SmsTemplateTestInp{Id: req.Id, Phone: req.Phone})
	if err != nil {
		return nil, err
	}
	return &api.SmsTemplateTestRes{Success: model.Success, RequestId: model.RequestId, Message: model.Message}, nil
}
