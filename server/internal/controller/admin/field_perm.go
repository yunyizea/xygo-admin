package admin

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	api "xygo/api/admin"
	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/field"
	"xygo/internal/library/contexts"
	"xygo/internal/library/dbdialect"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/utility"
)

// FieldPermList 查询字段权限列表
func (c *ControllerV1) FieldPermList(ctx context.Context, req *api.FieldPermListReq) (res *api.FieldPermListRes, err error) {
	var (
		items []entity.AdminFieldPerm
		m     = dao.AdminFieldPerm.Ctx(ctx)
	)

	if req.RoleId > 0 {
		if _, err = getManageableRole(ctx, req.RoleId); err != nil {
			return nil, err
		}
		m = m.Where(dao.AdminFieldPerm.Columns().RoleId, req.RoleId)
	} else {
		manageableIds, isSuper, err := getManageableRoleIds(ctx)
		if err != nil {
			return nil, err
		}
		if !isSuper {
			if len(manageableIds) == 0 {
				return &api.FieldPermListRes{List: []api.FieldPermItem{}}, nil
			}
			m = m.WhereIn(dao.AdminFieldPerm.Columns().RoleId, manageableIds)
		}
	}
	if req.Module != "" {
		m = m.Where(dao.AdminFieldPerm.Columns().Module, req.Module)
	}
	if req.Resource != "" {
		m = m.Where(dao.AdminFieldPerm.Columns().Resource, req.Resource)
	}

	err = m.Where(dao.AdminFieldPerm.Columns().Status, 1).
		Order(dao.AdminFieldPerm.Columns().Id + " ASC").
		Scan(&items)
	if err != nil {
		return nil, err
	}

	list := make([]api.FieldPermItem, 0, len(items))
	for _, item := range items {
		list = append(list, api.FieldPermItem{
			Id:         uint64(item.Id),
			RoleId:     uint64(item.RoleId),
			Module:     item.Module,
			Resource:   item.Resource,
			FieldName:  item.FieldName,
			FieldLabel: item.FieldLabel,
			PermType:   item.PermType,
			Status:     item.Status,
			Remark:     item.Remark,
		})
	}

	res = &api.FieldPermListRes{List: list}
	return
}

// FieldPermBatchSave 批量保存字段权限
func (c *ControllerV1) FieldPermBatchSave(ctx context.Context, req *api.FieldPermBatchSaveReq) (res *api.FieldPermBatchSaveRes, err error) {
	role, err := getManageableRole(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if consts.IsSuperRole(role.Key) {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "超级管理员角色不允许编辑字段权限")
	}

	now := utility.NowUnix()

	// 先删除该角色+资源的所有旧配置
	_, err = dao.AdminFieldPerm.Ctx(ctx).
		Where(dao.AdminFieldPerm.Columns().RoleId, req.RoleId).
		Where(dao.AdminFieldPerm.Columns().Resource, req.Resource).
		Delete()
	if err != nil {
		return nil, err
	}

	// 批量插入新配置
	for _, field := range req.Fields {
		_, err = dao.AdminFieldPerm.Ctx(ctx).Data(do.AdminFieldPerm{
			RoleId:     req.RoleId,
			Module:     "", // 可选，从 resource 自动推导
			Resource:   req.Resource,
			FieldName:  field.FieldName,
			FieldLabel: field.FieldLabel,
			PermType:   field.PermType,
			Status:     1,
			CreateTime: now,
			UpdateTime: now,
		}).Insert()
		if err != nil {
			return nil, err
		}
	}

	res = &api.FieldPermBatchSaveRes{}
	return
}

// FieldPermGetByRole 获取角色的字段权限映射（用于前端一次性加载）
func (c *ControllerV1) FieldPermGetByRole(ctx context.Context, req *api.FieldPermGetByRoleReq) (res *api.FieldPermGetByRoleRes, err error) {
	if _, err = getManageableRole(ctx, req.RoleId); err != nil {
		return nil, err
	}

	var (
		items []entity.AdminFieldPerm
		m     = dao.AdminFieldPerm.Ctx(ctx).Where(dao.AdminFieldPerm.Columns().RoleId, req.RoleId)
	)

	if req.Resource != "" {
		m = m.Where(dao.AdminFieldPerm.Columns().Resource, req.Resource)
	}

	err = m.Where(dao.AdminFieldPerm.Columns().Status, 1).Scan(&items)
	if err != nil {
		return nil, err
	}

	// 构建 resource -> field -> permType 的映射
	fieldPerms := make(map[string]map[string]int)
	for _, item := range items {
		if fieldPerms[item.Resource] == nil {
			fieldPerms[item.Resource] = make(map[string]int)
		}
		fieldPerms[item.Resource][item.FieldName] = item.PermType
	}

	res = &api.FieldPermGetByRoleRes{
		FieldPerms: fieldPerms,
	}
	return
}

// GetResourceFields 获取资源的字段列表
func (c *ControllerV1) GetResourceFields(ctx context.Context, req *api.GetResourceFieldsReq) (res *api.GetResourceFieldsRes, err error) {
	// 1) 优先使用 field registry 中已注册的字段定义
	resourceFields := field.GetFields(req.Resource)
	if len(resourceFields) > 0 {
		list := make([]api.ResourceFieldItem, 0, len(resourceFields))
		for _, f := range resourceFields {
			list = append(list, api.ResourceFieldItem{
				FieldName:   f.Name,
				FieldLabel:  f.Label,
				IsSensitive: f.IsSensitive,
			})
		}
		res = &api.GetResourceFieldsRes{Fields: list}
		return
	}

	// 2) 未注册时动态回退：按 resource 映射表结构自动读取字段
	list, err := getDynamicResourceFields(ctx, req.Resource)
	if err != nil {
		return nil, err
	}
	// 动态模式下，资源可能暂未建表或无可见字段，返回空列表给前端展示占位文案
	return &api.GetResourceFieldsRes{Fields: list}, nil
}

func getDynamicResourceFields(ctx context.Context, resource string) ([]api.ResourceFieldItem, error) {
	dialect := dbdialect.Get()
	tableName, err := resolveResourceTableName(ctx, resource)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return []api.ResourceFieldItem{}, nil
	}

	dbName, err := dialect.GetDbName(ctx)
	if err != nil {
		return nil, err
	}

	var columns []struct {
		ColumnName    string `json:"columnName"`
		ColumnComment string `json:"columnComment"`
	}
	err = g.DB().Ctx(ctx).Raw(dialect.ListColumnsSimpleSQL(dbName, tableName)).Scan(&columns)
	if err != nil {
		return nil, err
	}

	list := make([]api.ResourceFieldItem, 0, len(columns))
	for _, col := range columns {
		name := strings.TrimSpace(col.ColumnName)
		if name == "" {
			continue
		}
		label := strings.TrimSpace(col.ColumnComment)
		if label == "" {
			label = name
		}
		list = append(list, api.ResourceFieldItem{
			FieldName:   name,
			FieldLabel:  label,
			IsSensitive: isSensitiveColumn(name),
		})
	}
	return list, nil
}

func resolveResourceTableName(ctx context.Context, resource string) (string, error) {
	dialect := dbdialect.Get()
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return "", nil
	}

	dbName, err := dialect.GetDbName(ctx)
	if err != nil {
		return "", err
	}

	candidates := make([]string, 0, 3)
	candidates = append(candidates, resource)
	if strings.HasPrefix(resource, "xy_") {
		candidates = append(candidates, strings.TrimPrefix(resource, "xy_"))
	} else {
		candidates = append(candidates, "xy_"+resource)
	}

	seen := make(map[string]bool, len(candidates))
	for _, table := range candidates {
		table = strings.TrimSpace(table)
		if table == "" || seen[table] {
			continue
		}
		seen[table] = true

		var rows []struct {
			TableName string `json:"tableName"`
		}
		err := g.DB().Ctx(ctx).Raw(dialect.TableExistsSQL(dbName, table)).Scan(&rows)
		if err != nil {
			return "", err
		}
		if len(rows) > 0 && rows[0].TableName != "" {
			return rows[0].TableName, nil
		}
	}
	return "", nil
}

func isSensitiveColumn(column string) bool {
	name := strings.ToLower(strings.TrimSpace(column))
	if name == "" {
		return false
	}
	keywords := []string{
		"password", "passwd", "pwd", "salt",
		"token", "secret", "key",
		"mobile", "phone", "email",
		"id_card", "identity", "cert",
		"ip", "user_agent",
	}
	for _, kw := range keywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

// FieldPermMine 获取当前登录用户的字段权限（合并所有角色，取最高权限）
func (c *ControllerV1) FieldPermMine(ctx context.Context, req *api.FieldPermMineReq) (res *api.FieldPermMineRes, err error) {
	user := contexts.GetUser(ctx)
	if user == nil {
		return nil, gerror.NewCode(consts.CodeNotAuthorized, "未登录")
	}

	if consts.IsSuperRole(user.RoleKey) {
		res = &api.FieldPermMineRes{FieldPerms: map[string]map[string]int{}}
		return
	}

	var userRoles []entity.AdminUserRole
	if err = dao.AdminUserRole.Ctx(ctx).
		Where(dao.AdminUserRole.Columns().UserId, user.Id).
		Scan(&userRoles); err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		res = &api.FieldPermMineRes{FieldPerms: map[string]map[string]int{}}
		return
	}

	roleIds := make([]uint64, 0, len(userRoles))
	for _, r := range userRoles {
		roleIds = append(roleIds, r.RoleId)
	}

	var items []entity.AdminFieldPerm
	err = dao.AdminFieldPerm.Ctx(ctx).
		Where(dao.AdminFieldPerm.Columns().Status, 1).
		Where(dao.AdminFieldPerm.Columns().RoleId, roleIds).
		Scan(&items)
	if err != nil {
		return nil, err
	}

	fieldPerms := make(map[string]map[string]int)
	for _, item := range items {
		if fieldPerms[item.Resource] == nil {
			fieldPerms[item.Resource] = make(map[string]int)
		}
		existing, ok := fieldPerms[item.Resource][item.FieldName]
		if !ok || item.PermType > existing {
			fieldPerms[item.Resource][item.FieldName] = item.PermType
		}
	}

	res = &api.FieldPermMineRes{FieldPerms: fieldPerms}
	return
}
