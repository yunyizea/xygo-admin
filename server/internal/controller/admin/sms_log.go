package admin

import (
	"context"

	api "xygo/api/admin"
	"xygo/internal/dao"
	"xygo/internal/model/entity"
)

// SmsLogList 短信发送日志列表
func (c *ControllerV1) SmsLogList(ctx context.Context, req *api.SmsLogListReq) (res *api.SmsLogListRes, err error) {
	m := dao.SmsLog.Ctx(ctx)

	if req.Phone != "" {
		m = m.WhereLike("phone", "%"+req.Phone+"%")
	}
	if req.TemplateCode != "" {
		m = m.Where("template_code", req.TemplateCode)
	}
	if req.Status >= 0 {
		m = m.Where("status", req.Status)
	}
	if req.Driver != "" {
		m = m.Where("driver", req.Driver)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []entity.SmsLog
	err = m.OrderDesc("id").
		Page(req.Page, req.Size).
		Scan(&items)
	if err != nil {
		return nil, err
	}

	list := make([]api.SmsLogListItem, 0, len(items))
	for _, it := range items {
		list = append(list, api.SmsLogListItem{
			Id:           it.Id,
			Phone:        it.Phone,
			TemplateCode: it.TemplateCode,
			Driver:       it.Driver,
			Content:      it.Content,
			Params:       it.Params,
			Status:       it.Status,
			RequestId:    it.RequestId,
			ErrorMsg:     it.ErrorMsg,
			CreateTime:   it.CreateTime,
		})
	}

	res = &api.SmsLogListRes{
		List:  list,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}
	return
}
