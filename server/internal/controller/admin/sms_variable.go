package admin

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"

	api "xygo/api/admin"
	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/utility"
)

// SmsVariableList 短信变量列表
func (c *ControllerV1) SmsVariableList(ctx context.Context, req *api.SmsVariableListReq) (res *api.SmsVariableListRes, err error) {
	m := dao.SmsVariable.Ctx(ctx)

	if req.Name != "" {
		m = m.WhereLike("name", "%"+req.Name+"%")
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []entity.SmsVariable
	err = m.OrderAsc("id").
		Page(req.Page, req.Size).
		Scan(&items)
	if err != nil {
		return nil, err
	}

	list := make([]api.SmsVariableListItem, 0, len(items))
	for _, it := range items {
		list = append(list, api.SmsVariableListItem{
			Id:          it.Id,
			Title:       it.Title,
			Name:        it.Name,
			SourceType:  it.SourceType,
			SqlQuery:    it.SqlQuery,
			MethodName:  it.MethodName,
			SharedCount: it.SharedCount,
			Status:      it.Status,
			CreateTime:  it.CreateTime,
			UpdateTime:  it.UpdateTime,
		})
	}

	res = &api.SmsVariableListRes{
		List:  list,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}
	return
}

// SmsVariableSave 新增/编辑短信变量
func (c *ControllerV1) SmsVariableSave(ctx context.Context, req *api.SmsVariableSaveReq) (res *api.SmsVariableSaveRes, err error) {
	now := utility.NowUnix()

	if req.Id > 0 {
		count, err := dao.SmsVariable.Ctx(ctx).
			Where("name", req.Name).
			WhereNot("id", req.Id).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("变量名已存在：%s", req.Name))
		}

		_, err = dao.SmsVariable.Ctx(ctx).
			Data(do.SmsVariable{
				Title:      req.Title,
				Name:       req.Name,
				SourceType: req.SourceType,
				SqlQuery:   req.SqlQuery,
				MethodName: req.MethodName,
				Status:     req.Status,
				UpdateTime: now,
			}).
			Where("id", req.Id).
			Update()
		if err != nil {
			return nil, err
		}
	} else {
		count, err := dao.SmsVariable.Ctx(ctx).
			Where("name", req.Name).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("变量名已存在：%s", req.Name))
		}

		_, err = dao.SmsVariable.Ctx(ctx).
			Data(do.SmsVariable{
				Title:      req.Title,
				Name:       req.Name,
				SourceType: req.SourceType,
				SqlQuery:   req.SqlQuery,
				MethodName: req.MethodName,
				Status:     req.Status,
				CreateTime: now,
				UpdateTime: now,
			}).
			Insert()
		if err != nil {
			return nil, err
		}
	}

	res = &api.SmsVariableSaveRes{}
	return
}

// SmsVariableDelete 删除短信变量
func (c *ControllerV1) SmsVariableDelete(ctx context.Context, req *api.SmsVariableDeleteReq) (res *api.SmsVariableDeleteRes, err error) {
	_, err = dao.SmsVariable.Ctx(ctx).Where("id", req.Id).Delete()
	if err != nil {
		return nil, err
	}
	res = &api.SmsVariableDeleteRes{}
	return
}
