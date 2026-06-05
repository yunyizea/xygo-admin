// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package dept

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/model"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
	"xygo/utility"
)

type sAdminDept struct{}

func init() {
	service.RegisterAdminDept(New())
}

// New 构造部门服务
func New() *sAdminDept {
	return &sAdminDept{}
}

// List 获取部门列表（树形结构）
func (s *sAdminDept) List(ctx context.Context, in *adminin.DeptListInp) ([]*adminin.DeptListItem, error) {
	builder := dao.AdminDept.Ctx(ctx)

	// 按名称模糊搜索
	if in.Name != "" {
		builder = builder.WhereLike("name", "%"+in.Name+"%")
	}

	// 状态过滤
	if in.Status == 0 || in.Status == 1 {
		builder = builder.Where("status", in.Status)
	}

	// 查询所有部门
	var list []adminin.DeptListItem
	err := builder.
		Fields("id, parent_id as parentId, name, sort, status, remark, create_time, update_time").
		OrderAsc("sort, id").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	nodes := make([]*adminin.DeptListItem, 0, len(list))
	for i := range list {
		nodes = append(nodes, &list[i])
	}

	rootPtrs := model.BuildTree(
		nodes,
		func(n *adminin.DeptListItem) uint { return uint(n.Id) },
		func(n *adminin.DeptListItem) uint { return uint(n.ParentId) },
		func(n *adminin.DeptListItem, children []*adminin.DeptListItem) { n.Children = children },
	)

	roots := make([]*adminin.DeptListItem, 0, len(rootPtrs))
	for _, n := range rootPtrs {
		if n != nil {
			roots = append(roots, n)
		}
	}

	return roots, nil
}

// Detail 获取部门详情
func (s *sAdminDept) Detail(ctx context.Context, id uint64) (*adminin.DeptListItem, error) {
	var dept *entity.AdminDept

	if err := dao.AdminDept.Ctx(ctx).
		Where("id", id).
		Scan(&dept); err != nil {
		return nil, err
	}

	if dept == nil {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "部门不存在")
	}

	return &adminin.DeptListItem{
		Id:         dept.Id,
		ParentId:   dept.ParentId,
		Name:       dept.Name,
		Sort:       dept.Sort,
		Status:     dept.Status,
		Remark:     dept.Remark,
		CreateTime: int(dept.CreateTime),
		UpdateTime: int(dept.UpdateTime),
	}, nil
}

// Save 保存部门（新增/编辑）
func (s *sAdminDept) Save(ctx context.Context, in *adminin.DeptSaveInp) (uint, error) {
	// 父部门校验
	if in.ParentId != 0 {
		var parent *entity.AdminDept
		if err := dao.AdminDept.Ctx(ctx).
			Where("id", in.ParentId).
			Scan(&parent); err != nil {
			return 0, err
		}
		if parent == nil {
			return 0, gerror.NewCode(consts.CodeDataNotFound, "上级部门不存在")
		}
		// 不能选择自己作为父部门
		if in.Id != 0 && in.Id == in.ParentId {
			return 0, gerror.NewCode(consts.CodeInvalidParam, "上级部门不能选择自己")
		}
	}

	// 同级部门名称唯一性校验
	count, err := dao.AdminDept.Ctx(ctx).
		Where("parent_id", in.ParentId).
		Where("name", in.Name).
		WhereNot("id", in.Id).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.NewCode(consts.CodeInvalidParam, "同级已存在相同名称的部门")
	}

	now := utility.NowUnix()
	data := do.AdminDept{
		ParentId:   in.ParentId,
		Name:       in.Name,
		Sort:       in.Sort,
		Status:     in.Status,
		Remark:     in.Remark,
		UpdateTime: now,
	}

	if in.Id == 0 {
		// 新增
		data.CreateTime = now
		r, err := dao.AdminDept.Ctx(ctx).Data(data).OmitNil().Insert()
		if err != nil {
			return 0, err
		}
		lastId, err := r.LastInsertId()
		if err != nil {
			return 0, err
		}
		return uint(lastId), nil
	}

	// 编辑
	_, err = dao.AdminDept.Ctx(ctx).
		Data(data).
		OmitNil().
		Where("id", in.Id).
		Update()
	if err != nil {
		return 0, err
	}
	return uint(in.Id), nil
}

// Delete 删除部门
func (s *sAdminDept) Delete(ctx context.Context, id uint64) error {
	// 检查是否有子部门
	childCount, err := dao.AdminDept.Ctx(ctx).
		Where("parent_id", id).
		Count()
	if err != nil {
		return err
	}
	if childCount > 0 {
		return gerror.NewCode(consts.CodeInvalidParam, "该部门下还有子部门，无法删除")
	}

	// 检查是否有用户绑定该部门
	userCount, err := dao.AdminUser.Ctx(ctx).
		Where("dept_id", id).
		Count()
	if err != nil {
		return err
	}
	if userCount > 0 {
		return gerror.NewCode(consts.CodeInvalidParam, "该部门下还有用户，无法删除")
	}

	_, err = dao.AdminDept.Ctx(ctx).
		Where("id", id).
		Delete()
	return err
}
