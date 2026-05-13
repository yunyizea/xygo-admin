package admin

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"

	api "xygo/api/admin"
	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/library/sms"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/utility"
)

// SmsTemplateList 短信模板列表
func (c *ControllerV1) SmsTemplateList(ctx context.Context, req *api.SmsTemplateListReq) (res *api.SmsTemplateListRes, err error) {
	m := dao.SmsTemplate.Ctx(ctx)

	if req.Status >= 0 {
		m = m.Where("status", req.Status)
	}
	if req.Code != "" {
		m = m.WhereLike("code", "%"+req.Code+"%")
	}
	if req.Title != "" {
		m = m.WhereLike("title", "%"+req.Title+"%")
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []entity.SmsTemplate
	err = m.OrderAsc("sort").OrderAsc("id").
		Page(req.Page, req.Size).
		Scan(&items)
	if err != nil {
		return nil, err
	}

	list := make([]api.SmsTemplateListItem, 0, len(items))
	for _, it := range items {
		list = append(list, api.SmsTemplateListItem{
			Id:                 it.Id,
			Title:              it.Title,
			Code:               it.Code,
			Content:            it.Content,
			ProviderTemplateId: it.ProviderTemplateId,
			Variables:          it.Variables,
			RelatedVariableId:  it.RelatedVariableId,
			Status:             it.Status,
			Sort:               it.Sort,
			Remark:             it.Remark,
			CreateTime:         it.CreateTime,
			UpdateTime:         it.UpdateTime,
		})
	}

	res = &api.SmsTemplateListRes{
		List:  list,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}
	return
}

// SmsTemplateSave 新增/编辑短信模板
func (c *ControllerV1) SmsTemplateSave(ctx context.Context, req *api.SmsTemplateSaveReq) (res *api.SmsTemplateSaveRes, err error) {
	now := utility.NowUnix()

	var variables *gjson.Json
	if req.Variables != "" {
		variables, err = gjson.LoadContent([]byte(req.Variables))
		if err != nil {
			return nil, gerror.NewCode(consts.CodeInvalidParam, "variables 非法 JSON")
		}
	}

	if req.Id > 0 {
		count, err := dao.SmsTemplate.Ctx(ctx).
			Where("code", req.Code).
			WhereNot("id", req.Id).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("模板标识已存在：%s", req.Code))
		}

		_, err = dao.SmsTemplate.Ctx(ctx).
			Data(do.SmsTemplate{
				Title:              req.Title,
				Code:               req.Code,
				Content:            req.Content,
				ProviderTemplateId: req.ProviderTemplateId,
				Variables:          variables,
				RelatedVariableId:  req.RelatedVariableId,
				Status:             req.Status,
				Sort:               req.Sort,
				Remark:             req.Remark,
				UpdateTime:         now,
			}).
			Where("id", req.Id).
			Update()
		if err != nil {
			return nil, err
		}
	} else {
		count, err := dao.SmsTemplate.Ctx(ctx).
			Where("code", req.Code).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("模板标识已存在：%s", req.Code))
		}

		_, err = dao.SmsTemplate.Ctx(ctx).
			Data(do.SmsTemplate{
				Title:              req.Title,
				Code:               req.Code,
				Content:            req.Content,
				ProviderTemplateId: req.ProviderTemplateId,
				Variables:          variables,
				RelatedVariableId:  req.RelatedVariableId,
				Status:             req.Status,
				Sort:               req.Sort,
				Remark:             req.Remark,
				CreateTime:         now,
				UpdateTime:         now,
			}).
			Insert()
		if err != nil {
			return nil, err
		}
	}

	res = &api.SmsTemplateSaveRes{}
	return
}

// SmsTemplateDelete 删除短信模板
func (c *ControllerV1) SmsTemplateDelete(ctx context.Context, req *api.SmsTemplateDeleteReq) (res *api.SmsTemplateDeleteRes, err error) {
	_, err = dao.SmsTemplate.Ctx(ctx).Where("id", req.Id).Delete()
	if err != nil {
		return nil, err
	}
	res = &api.SmsTemplateDeleteRes{}
	return
}

// SmsTemplateTest 测试发送短信
func (c *ControllerV1) SmsTemplateTest(ctx context.Context, req *api.SmsTemplateTestReq) (res *api.SmsTemplateTestRes, err error) {
	var tpl entity.SmsTemplate
	err = dao.SmsTemplate.Ctx(ctx).Where("id", req.Id).Scan(&tpl)
	if err != nil {
		return nil, err
	}
	if tpl.Id == 0 {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "模板不存在")
	}

	mgr := sms.Instance(ctx)
	result, err := mgr.Send(ctx, &sms.SendRequest{
		Phone:      req.Phone,
		TemplateId: tpl.ProviderTemplateId,
		Params:     nil,
	})
	if err != nil {
		return &api.SmsTemplateTestRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	_, _ = dao.SmsLog.Ctx(ctx).Data(do.SmsLog{
		Phone:        req.Phone,
		TemplateCode: tpl.Code,
		Driver:       result.Driver,
		Content:      tpl.Content,
		Status: func() int {
			if result.Success {
				return 1
			}
			return 0
		}(),
		RequestId: result.RequestId,
		ErrorMsg: func() string {
			if !result.Success {
				return result.Message
			}
			return ""
		}(),
		CreateTime: utility.NowUnix(),
	}).Insert()

	res = &api.SmsTemplateTestRes{
		Success:   result.Success,
		RequestId: result.RequestId,
		Message:   result.Message,
	}
	return
}
