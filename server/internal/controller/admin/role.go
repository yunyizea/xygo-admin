// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package admin

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	api "xygo/api/admin"
	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/field"
	"xygo/internal/library/dbdialect"
	"xygo/internal/model"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/model/input/form"
)

// RoleList 角色列表
func (c *ControllerV1) RoleList(ctx context.Context, req *api.RoleListReq) (res *api.RoleListRes, err error) {
	model := dao.AdminRole.Ctx(ctx)

	// 按名称模糊搜索
	if req.Name != "" {
		model = model.WhereLike("name", "%"+req.Name+"%")
	}

	// 状态过滤
	if req.Status == 0 || req.Status == 1 {
		model = model.Where("status", req.Status)
	}

	// 统计总数
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}

	// 分页参数兜底
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 查询列表（平铺数据，包含树相关字段和权限配置）
	var list []adminin.RoleListItem
	err = model.
		Fields("id, name, "+dbdialect.Get().QuoteIdentifier("key")+", pid, level, tree, data_scope as dataScope, custom_depts as customDepts, sort, status, remark, create_time as createdAt").
		Page(req.Page, req.PageSize).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	// 按 pid 关系构建树，保证列表按层级结构返回（children 为子角色集合）
	roleTree := buildRoleTree(list)

	res = new(api.RoleListRes)
	res.RoleListModel = &adminin.RoleListModel{
		List: roleTree,
		PageRes: form.PageRes{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}
	return
}

// RoleDetail 角色详情
func (c *ControllerV1) RoleDetail(ctx context.Context, req *api.RoleDetailReq) (res *api.RoleDetailRes, err error) {
	var role *entity.AdminRole

	if err = dao.AdminRole.Ctx(ctx).
		Where("id", req.Id).
		Scan(&role); err != nil {
		return nil, err
	}

	if role == nil {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "角色不存在")
	}

	res = new(api.RoleDetailRes)
	res.RoleDetailModel = &adminin.RoleDetailModel{
		Id:          role.Id,
		Name:        role.Name,
		Key:         role.Key,
		DataScope:   role.DataScope,
		CustomDepts: role.CustomDepts,
		Status:      role.Status,
		Remark:      role.Remark,
		CreateTime:  int(role.CreateTime),
		UpdateTime:  int(role.UpdateTime),
		CreatedBy:   role.CreatedBy,
		UpdatedBy:   role.UpdatedBy,
	}
	return
}

// RoleSave 角色保存（新增/编辑）
func (c *ControllerV1) RoleSave(ctx context.Context, req *api.RoleSaveReq) (res *api.RoleSaveRes, err error) {
	if req.Id == 0 {
		// 新增：根据 pid 计算 level/tree
		var (
			pid   = req.Pid
			level = 1
			tree  = "0"
		)

		if pid > 0 {
			var parent *entity.AdminRole
			if err = dao.AdminRole.Ctx(ctx).
				Where("id", pid).
				Scan(&parent); err != nil {
				return nil, err
			}
			if parent != nil {
				level = int(parent.Level + 1)
				if parent.Tree != "" {
					tree = parent.Tree + "," + strconv.Itoa(int(parent.Id))
				} else {
					tree = "0," + strconv.Itoa(int(parent.Id))
				}
			} else {
				// 上级不存在则视为根节点
				pid = 0
				level = 1
				tree = "0"
			}
		}

		now := gtime.Now().Timestamp()
		data := do.AdminRole{
			Name:        req.Name,
			Key:         req.Key,
			Pid:         pid,
			Level:       level,
			Tree:        tree,
			Sort:        req.Sort,
			DataScope:   req.DataScope,
			CustomDepts: req.CustomDepts,
			Status:      req.Status,
			Remark:      req.Remark,
			CreateTime:  now,
			UpdateTime:  now,
		}

		// 新增
		r, err := dao.AdminRole.Ctx(ctx).Data(data).OmitNil().Insert()
		if err != nil {
			return nil, err
		}
		lastId, err := r.LastInsertId()
		if err != nil {
			return nil, err
		}
		res = &api.RoleSaveRes{Id: uint64(lastId)}
	} else {
		// 编辑（暂不支持在此接口修改 pid/level/tree）
		data := do.AdminRole{
			Name:        req.Name,
			Key:         req.Key,
			Sort:        req.Sort,
			DataScope:   req.DataScope,
			CustomDepts: req.CustomDepts,
			Status:      req.Status,
			Remark:      req.Remark,
			UpdateTime:  gtime.Now().Timestamp(),
		}

		_, err = dao.AdminRole.Ctx(ctx).
			Data(data).
			OmitNil().
			Where("id", req.Id).
			Update()
		if err != nil {
			return nil, err
		}
		res = &api.RoleSaveRes{Id: req.Id}
	}
	return
}

// RoleDelete 角色删除
func (c *ControllerV1) RoleDelete(ctx context.Context, req *api.RoleDeleteReq) (res *api.RoleDeleteRes, err error) {
	// TODO: 后续增加“有关联用户/菜单时不允许删除”的约束
	_, err = dao.AdminRole.Ctx(ctx).
		Where("id", req.Id).
		Delete()
	if err != nil {
		return nil, err
	}
	return &api.RoleDeleteRes{}, nil
}

// RoleMenuIds 获取角色已绑定的菜单ID列表
func (c *ControllerV1) RoleMenuIds(ctx context.Context, req *api.RoleMenuIdsReq) (res *api.RoleMenuIdsRes, err error) {
	// 查询该角色已绑定的菜单ID
	var rows []struct {
		MenuId uint64 `json:"menuId"`
	}
	if err = dao.AdminRoleMenu.Ctx(ctx).
		Fields("menu_id AS menuId").
		Where("role_id", req.RoleId).
		Scan(&rows); err != nil {
		return nil, err
	}

	menuIds := make([]uint64, 0, len(rows))
	for _, r := range rows {
		menuIds = append(menuIds, r.MenuId)
	}

	res = new(api.RoleMenuIdsRes)
	res.RoleMenuIdsModel = &adminin.RoleMenuIdsModel{
		MenuIds: menuIds,
	}
	return
}

// RoleBindMenus 为角色绑定菜单
func (c *ControllerV1) RoleBindMenus(ctx context.Context, req *api.RoleBindMenusReq) (res *api.RoleBindMenusRes, err error) {
	if req.RoleId == 0 {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "角色ID不能为空")
	}

	// ✅ 查询角色并检查是否为超管
	var role *entity.AdminRole
	if err = dao.AdminRole.Ctx(ctx).Where("id", req.RoleId).Scan(&role); err != nil {
		return nil, err
	}
	if role == nil {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "角色不存在")
	}
	if consts.IsSuperRole(role.Key) {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "超级管理员角色不允许编辑菜单权限")
	}

	// 简单实现：清空后重建
	// 如需更强一致性，可后续引入事务处理与乐观锁

	// 先删除旧绑定关系
	if _, err = dao.AdminRoleMenu.Ctx(ctx).
		Where("role_id", req.RoleId).
		Delete(); err != nil {
		return nil, err
	}

	// 如果没有菜单ID，直接返回
	if len(req.MenuIds) == 0 {
		return &api.RoleBindMenusRes{}, nil
	}

	// 批量插入新关系
	list := make([]do.AdminRoleMenu, 0, len(req.MenuIds))
	for _, menuId := range req.MenuIds {
		if menuId == 0 {
			continue
		}
		list = append(list, do.AdminRoleMenu{
			RoleId: req.RoleId,
			MenuId: menuId,
		})
	}

	if len(list) > 0 {
		if _, err = dao.AdminRoleMenu.Ctx(ctx).
			Data(list).
			Insert(); err != nil {
			return nil, err
		}
	}

	return &api.RoleBindMenusRes{}, nil
}

// buildRoleTree 使用通用树工具按 pid 关系构建角色树。
// 说明：
// - 返回值为根节点列表，每个节点的 Children 字段包含子角色；
// - 便于前端直接作为树表或树形选择组件使用。
func buildRoleTree(list []adminin.RoleListItem) []adminin.RoleListItem {
	if len(list) == 0 {
		return list
	}

	// 转为指针切片以便通用树工具操作
	nodes := make([]*adminin.RoleListItem, 0, len(list))
	for i := range list {
		nodes = append(nodes, &list[i])
	}

	// 构建树
	rootPtrs := model.BuildTree(
		nodes,
		func(n *adminin.RoleListItem) uint { return n.Id },
		func(n *adminin.RoleListItem) uint { return n.Pid },
		func(n *adminin.RoleListItem, children []*adminin.RoleListItem) { n.Children = children },
	)

	result := make([]adminin.RoleListItem, 0, len(rootPtrs))
	for _, n := range rootPtrs {
		if n == nil {
			continue
		}
		result = append(result, *n)
	}
	return result
}

// DataScopeSelect 获取数据范围选项（供前端下拉框使用）
func (c *ControllerV1) DataScopeSelect(ctx context.Context, req *api.DataScopeSelectReq) (res *api.DataScopeSelectRes, err error) {
	res = &api.DataScopeSelectRes{
		List: consts.DataScopeSelect,
	}
	return
}

// RoleAvailableResources 获取角色可配置字段权限的资源列表
func (c *ControllerV1) RoleAvailableResources(ctx context.Context, req *api.RoleAvailableResourcesReq) (res *api.RoleAvailableResourcesRes, err error) {
	// 1. 查询角色已绑定的菜单ID
	var menuIds []uint
	var menuRows []struct {
		MenuId uint `json:"menuId"`
	}
	if err = dao.AdminRoleMenu.Ctx(ctx).
		Fields("menu_id AS menuId").
		Where("role_id", req.RoleId).
		Scan(&menuRows); err != nil {
		return nil, err
	}
	for _, row := range menuRows {
		if row.MenuId > 0 {
			menuIds = append(menuIds, row.MenuId)
		}
	}

	if len(menuIds) == 0 {
		res = new(api.RoleAvailableResourcesRes)
		res.RoleAvailableResourcesModel = &adminin.RoleAvailableResourcesModel{
			List: []adminin.AvailableResource{},
		}
		return res, nil
	}

	// 2. 查询这些菜单的resource字段
	var menus []struct {
		Resource string `json:"resource"`
	}
	if err = dao.AdminMenu.Ctx(ctx).
		Fields("resource").
		WhereIn("id", menuIds).
		Where("resource !=", ""). // 只要有resource的菜单
		Scan(&menus); err != nil {
		return nil, err
	}

	// 3. 去重并获取资源信息
	resourceMap := make(map[string]bool)
	resourceList := make([]adminin.AvailableResource, 0)

	for _, menu := range menus {
		if menu.Resource == "" {
			continue
		}
		if !resourceMap[menu.Resource] {
			resourceMap[menu.Resource] = true

			// ✅ 从field registry获取资源中文名称
			resource := field.Get(menu.Resource)
			label := menu.Resource // 默认只显示表名
			if resource != nil && resource.Label != "" {
				// 中文名称（表名）格式
				label = resource.Label + "（" + menu.Resource + "）"
			}

			resourceList = append(resourceList, adminin.AvailableResource{
				Code:  menu.Resource,
				Label: label, // ✅ 显示"用户管理（admin_user）"
			})
		}
	}

	res = new(api.RoleAvailableResourcesRes)
	res.RoleAvailableResourcesModel = &adminin.RoleAvailableResourcesModel{
		List: resourceList,
	}
	return
}

// DataScopeEdit 编辑角色数据权限
func (c *ControllerV1) DataScopeEdit(ctx context.Context, req *api.DataScopeEditReq) (res *api.DataScopeEditRes, err error) {
	// 查询角色是否存在
	var role *entity.AdminRole
	err = dao.AdminRole.Ctx(ctx).Where(dao.AdminRole.Columns().Id, req.Id).Scan(&role)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "角色不存在")
	}

	// ✅ 超级管理员不允许编辑权限
	if consts.IsSuperRole(role.Key) {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "超级管理员角色不允许编辑权限配置")
	}

	// 自定义部门权限处理
	var customDepts string
	if req.DataScope == consts.RoleDataDeptCustom {
		// 如果是自定义部门，必须选择至少一个部门
		if len(req.CustomDepts) == 0 {
			return nil, gerror.NewCode(consts.CodeInvalidParam, "自定义部门时至少选择一个部门")
		}
		// 序列化为 JSON 字符串
		customDepts = "[" + func() string {
			s := ""
			for i, deptId := range req.CustomDepts {
				if i > 0 {
					s += ","
				}
				s += strconv.FormatUint(deptId, 10)
			}
			return s
		}() + "]"
	}

	// 更新数据范围
	_, err = dao.AdminRole.Ctx(ctx).
		Data(do.AdminRole{
			DataScope:   req.DataScope,
			CustomDepts: customDepts,
		}).
		Where(dao.AdminRole.Columns().Id, req.Id).
		Update()

	if err != nil {
		return nil, err
	}

	res = &api.DataScopeEditRes{}
	return
}
