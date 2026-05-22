package admin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	api "xygo/api/admin"
	"xygo/internal/library/contexts"
	"xygo/internal/library/token"
	"xygo/internal/model/input/adminin"
	"xygo/internal/model/input/form"
	"xygo/internal/service"
)

// UserList 管理员列表
func (c *ControllerV1) UserList(ctx context.Context, req *api.UserListReq) (res *api.UserListRes, err error) {
	list, total, err := service.AdminUser().List(ctx, &req.UserListInp)
	if err != nil {
		return nil, err
	}

	res = &api.UserListRes{
		UserListModel: adminin.UserListModel{
			List: list,
			PageRes: form.PageRes{
				Page:     req.Page,
				PageSize: req.PageSize,
				Total:    total,
			},
		},
	}
	return
}

// UserDetail 用户详情（编辑用，未脱敏）
func (c *ControllerV1) UserDetail(ctx context.Context, req *api.UserDetailReq) (res *api.UserDetailRes, err error) {
	result, err := service.AdminUser().DetailForEdit(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.UserDetailRes{UserDetailModel: result}, nil
}

// UserSave 保存用户（新增/编辑）
func (c *ControllerV1) UserSave(ctx context.Context, req *api.UserSaveReq) (res *api.UserSaveRes, err error) {
	if req.RoleIds != nil {
		if err = ensureAssignableRoleIds(ctx, req.RoleIds); err != nil {
			return nil, err
		}
	}

	id, err := service.AdminUser().Save(ctx, &req.UserSaveInp)
	if err != nil {
		return nil, err
	}
	return &api.UserSaveRes{Id: id}, nil
}

// UserDelete 删除用户
func (c *ControllerV1) UserDelete(ctx context.Context, req *api.UserDeleteReq) (res *api.UserDeleteRes, err error) {
	// 不能删除自己
	currentUserId := contexts.GetUserId(ctx)
	if currentUserId == req.Id {
		return nil, gerror.New("不能删除自己")
	}
	err = service.AdminUser().Delete(ctx, req.Id)
	return &api.UserDeleteRes{}, err
}

// UserKick 强制用户下线
func (c *ControllerV1) UserKick(ctx context.Context, req *api.UserKickReq) (res *api.UserKickRes, err error) {
	// 不能踢自己
	currentUserId := contexts.GetUserId(ctx)
	if currentUserId == req.Id {
		return nil, gerror.New("不能踢自己下线")
	}

	// 执行踢人
	if err = token.KickByUserId(ctx, token.AppAdmin, req.Id); err != nil {
		return nil, gerror.Newf("踢人失败: %v", err)
	}

	return &api.UserKickRes{}, nil
}
