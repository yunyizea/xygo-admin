package admin

import (
	"context"

	api "xygo/api/admin"
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
)

func (c *ControllerV1) SmsLogList(ctx context.Context, req *api.SmsLogListReq) (res *api.SmsLogListRes, err error) {
	model, err := service.Sms().LogList(ctx, &adminin.SmsLogListInp{
		Page: req.Page, Size: req.Size, Phone: req.Phone,
		TemplateCode: req.TemplateCode, Status: req.Status, Driver: req.Driver,
	})
	if err != nil {
		return nil, err
	}
	res = &api.SmsLogListRes{List: model.List, Total: model.Total, Page: model.Page, Size: model.Size}
	return
}
