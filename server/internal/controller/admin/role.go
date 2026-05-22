package admin

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	api "xygo/api/admin"
	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/field"
	"xygo/internal/library/contexts"
	"xygo/internal/library/dbdialect"
	"xygo/internal/middleware"
	"xygo/internal/model"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/model/input/form"
)

// RoleList 角色列表
func (c *ControllerV1) RoleList(ctx context.Context, req *api.RoleListReq) (res *api.RoleListRes, err error) {
	model := dao.AdminRole.Ctx(ctx)

	// 分页参数兜底
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	manageableIds, isSuper, err := getManageableRoleIds(ctx)
	if err != nil {
		return nil, err
	}
	if !isSuper {
		if len(manageableIds) == 0 {
			return &api.RoleListRes{
				RoleListModel: &adminin.RoleListModel{
					List: []adminin.RoleListItem{},
					PageRes: form.PageRes{
						Page:     req.Page,
						PageSize: req.PageSize,
						Total:    0,
					},
				},
			}, nil
		}
		model = model.WhereIn("id", manageableIds)
	}

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
	role, err := getManageableRole(ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	res = new(api.RoleDetailRes)
	res.RoleDetailModel = &adminin.RoleDetailModel{
		Id:          role.Id,
		Name:        role.Name,
		Key:         role.Key,
		Pid:         uint(role.Pid),
		Level:       int(role.Level),
		Tree:        role.Tree,
		Sort:        role.Sort,
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
		if err = ensureCanCreateRoleUnder(ctx, req.Pid); err != nil {
			return nil, err
		}

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
		if _, err = getManageableRole(ctx, req.Id); err != nil {
			return nil, err
		}

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
	if _, err = getManageableRole(ctx, req.Id); err != nil {
		return nil, err
	}

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
	if _, err = getManageableRole(ctx, req.RoleId); err != nil {
		return nil, err
	}

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

	role, err := getManageableRole(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if consts.IsSuperRole(role.Key) {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "超级管理员角色不允许编辑菜单权限")
	}
	if err = ensureAssignableMenuIds(ctx, req.MenuIds); err != nil {
		return nil, err
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

	middleware.RefreshPermCache(ctx)
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

func getCurrentDirectRoleIds(ctx context.Context) ([]uint64, bool, error) {
	user := contexts.GetUser(ctx)
	if user == nil {
		return nil, false, gerror.NewCode(consts.CodeNotAuthorized, "未登录")
	}
	if consts.IsSuperRole(user.RoleKey) {
		return nil, true, nil
	}

	var rows []struct {
		RoleId uint64 `json:"roleId"`
	}
	if err := dao.AdminRole.Ctx(ctx).
		LeftJoin(dao.AdminUserRole.Table()+" aur", "aur.role_id = "+dao.AdminRole.Table()+".id").
		Fields(dao.AdminRole.Table()+".id AS roleId").
		Where("aur.user_id", user.Id).
		Where(dao.AdminRole.Table()+".status", 1).
		Scan(&rows); err != nil {
		return nil, false, err
	}

	roleIds := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if row.RoleId > 0 {
			roleIds = append(roleIds, row.RoleId)
		}
	}
	return roleIds, false, nil
}

func getManageableRoleIds(ctx context.Context) ([]uint64, bool, error) {
	roleIds, isSuper, err := getCurrentDirectRoleIds(ctx)
	if err != nil || isSuper {
		return nil, isSuper, err
	}
	if len(roleIds) == 0 {
		return []uint64{}, false, nil
	}

	var roles []entity.AdminRole
	if err = dao.AdminRole.Ctx(ctx).Scan(&roles); err != nil {
		return nil, false, err
	}

	manageableIds := make([]uint64, 0)
	for _, role := range roles {
		if consts.IsSuperRole(role.Key) || uint64In(role.Id, roleIds) {
			continue
		}
		for _, currentRoleId := range roleIds {
			if roleTreeContains(role.Tree, currentRoleId) {
				manageableIds = append(manageableIds, role.Id)
				break
			}
		}
	}
	return manageableIds, false, nil
}

func getManageableRole(ctx context.Context, roleId uint64) (*entity.AdminRole, error) {
	var role *entity.AdminRole
	if err := dao.AdminRole.Ctx(ctx).Where("id", roleId).Scan(&role); err != nil {
		return nil, err
	}
	if role == nil {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "角色不存在")
	}

	currentRoleIds, isSuper, err := getCurrentDirectRoleIds(ctx)
	if err != nil {
		return nil, err
	}
	if isSuper {
		return role, nil
	}
	if consts.IsSuperRole(role.Key) || uint64In(role.Id, currentRoleIds) {
		return nil, gerror.NewCode(consts.CodeNoPermission, "无权管理该角色")
	}
	for _, currentRoleId := range currentRoleIds {
		if roleTreeContains(role.Tree, currentRoleId) {
			return role, nil
		}
	}
	return nil, gerror.NewCode(consts.CodeNoPermission, "无权管理该角色")
}

func ensureCanCreateRoleUnder(ctx context.Context, pid uint) error {
	if pid == 0 {
		_, isSuper, err := getCurrentDirectRoleIds(ctx)
		if err != nil {
			return err
		}
		if isSuper {
			return nil
		}
		return gerror.NewCode(consts.CodeNoPermission, "无权创建根角色")
	}

	currentRoleIds, isSuper, err := getCurrentDirectRoleIds(ctx)
	if err != nil || isSuper {
		return err
	}
	parentId := uint64(pid)
	if uint64In(parentId, currentRoleIds) {
		return nil
	}
	_, err = getManageableRole(ctx, parentId)
	return err
}

func roleTreeContains(tree string, roleId uint64) bool {
	for _, part := range strings.Split(tree, ",") {
		if strings.TrimSpace(part) == strconv.FormatUint(roleId, 10) {
			return true
		}
	}
	return false
}

func uint64In(id uint64, ids []uint64) bool {
	for _, item := range ids {
		if item == id {
			return true
		}
	}
	return false
}

func getAssignableMenuIds(ctx context.Context) (map[uint64]struct{}, bool, error) {
	roleIds, isSuper, err := getCurrentDirectRoleIds(ctx)
	if err != nil || isSuper {
		return nil, isSuper, err
	}
	if len(roleIds) == 0 {
		return map[uint64]struct{}{}, false, nil
	}

	var rows []struct {
		MenuId uint64 `json:"menuId"`
	}
	if err = dao.AdminRoleMenu.Ctx(ctx).
		Fields("menu_id AS menuId").
		WhereIn("role_id", roleIds).
		Scan(&rows); err != nil {
		return nil, false, err
	}

	allowedIds := make(map[uint64]struct{}, len(rows))
	for _, row := range rows {
		if row.MenuId > 0 {
			allowedIds[row.MenuId] = struct{}{}
		}
	}
	if len(allowedIds) == 0 {
		return allowedIds, false, nil
	}

	var menus []entity.AdminMenu
	if err = dao.AdminMenu.Ctx(ctx).Fields("id,parent_id").Scan(&menus); err != nil {
		return nil, false, err
	}
	parentMap := make(map[uint64]uint64, len(menus))
	for _, menu := range menus {
		parentMap[menu.Id] = menu.ParentId
	}
	for menuId := range allowedIds {
		for parentId := parentMap[menuId]; parentId > 0; parentId = parentMap[parentId] {
			if _, exists := allowedIds[parentId]; exists {
				break
			}
			allowedIds[parentId] = struct{}{}
		}
	}
	return allowedIds, false, nil
}

func ensureAssignableMenuIds(ctx context.Context, menuIds []uint64) error {
	allowedIds, isSuper, err := getAssignableMenuIds(ctx)
	if err != nil || isSuper {
		return err
	}
	for _, menuId := range menuIds {
		if menuId == 0 {
			continue
		}
		if _, ok := allowedIds[menuId]; !ok {
			return gerror.NewCode(consts.CodeNoPermission, "不能分配超出当前角色权限范围的菜单")
		}
	}
	return nil
}

func ensureAssignableRoleIds(ctx context.Context, roleIds []uint64) error {
	if len(roleIds) == 0 {
		return nil
	}

	manageableIds, isSuper, err := getManageableRoleIds(ctx)
	if err != nil || isSuper {
		return err
	}

	allowedIds := make(map[uint64]struct{}, len(manageableIds))
	for _, roleId := range manageableIds {
		allowedIds[roleId] = struct{}{}
	}
	for _, roleId := range roleIds {
		if roleId == 0 {
			continue
		}
		if _, ok := allowedIds[roleId]; !ok {
			return gerror.NewCode(consts.CodeNoPermission, "不能分配超出当前管理范围的角色")
		}
	}
	return nil
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
	if _, err = getManageableRole(ctx, req.RoleId); err != nil {
		return nil, err
	}

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
	role, err := getManageableRole(ctx, req.Id)
	if err != nil {
		return nil, err
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
