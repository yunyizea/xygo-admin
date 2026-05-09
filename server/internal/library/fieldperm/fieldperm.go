package fieldperm

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/library/contexts"
	"xygo/internal/model/entity"
)

// PermType 字段权限类型
const (
	PermHidden   = 0 // 不可见
	PermReadonly = 1 // 只读
	PermEditable = 2 // 可编辑
)

// GetUserFieldPerms 获取当前用户在指定 resource 上的字段权限
// 返回 map[fieldName]permType，未配置的字段默认为可编辑(2)
func GetUserFieldPerms(ctx context.Context, resource string) map[string]int {
	if resource == "" {
		return nil
	}

	user := contexts.GetUser(ctx)
	if user == nil {
		return nil
	}

	if consts.IsSuperRole(user.RoleKey) {
		return nil
	}

	var userRoles []entity.AdminUserRole
	if err := dao.AdminUserRole.Ctx(ctx).
		Where(dao.AdminUserRole.Columns().UserId, user.Id).
		Scan(&userRoles); err != nil {
		g.Log().Warningf(ctx, "获取用户角色失败: %v", err)
		return nil
	}

	if len(userRoles) == 0 {
		return nil
	}

	roleIds := make([]uint64, 0, len(userRoles))
	for _, r := range userRoles {
		roleIds = append(roleIds, r.RoleId)
	}

	var items []entity.AdminFieldPerm
	if err := dao.AdminFieldPerm.Ctx(ctx).
		Where(dao.AdminFieldPerm.Columns().Status, 1).
		Where(dao.AdminFieldPerm.Columns().RoleId, roleIds).
		Where(dao.AdminFieldPerm.Columns().Resource, resource).
		Scan(&items); err != nil {
		g.Log().Warningf(ctx, "获取字段权限失败: %v", err)
		return nil
	}

	if len(items) == 0 {
		return nil
	}

	perms := make(map[string]int, len(items))
	for _, item := range items {
		existing, ok := perms[item.FieldName]
		if !ok || item.PermType > existing {
			perms[item.FieldName] = item.PermType
		}
	}
	return perms
}

// FilterResponseFields 过滤响应数据中的不可见字段
// data 可以是 map/struct/slice，会被就地修改
func FilterResponseFields(ctx context.Context, resource string, data any) any {
	perms := GetUserFieldPerms(ctx, resource)
	if perms == nil {
		return data
	}

	hiddenFields := make([]string, 0)
	for field, perm := range perms {
		if perm == PermHidden {
			hiddenFields = append(hiddenFields, field)
		}
	}

	if len(hiddenFields) == 0 {
		return data
	}

	switch v := data.(type) {
	case []map[string]any:
		for _, row := range v {
			for _, f := range hiddenFields {
				delete(row, f)
			}
		}
	case map[string]any:
		for _, f := range hiddenFields {
			delete(v, f)
		}
	default:
		m := gconv.Map(data)
		if m != nil {
			for _, f := range hiddenFields {
				delete(m, f)
			}
			return m
		}
	}
	return data
}

// FilterWriteFields 过滤写入请求中的不可见和只读字段
// 返回需要剔除的字段名列表
func FilterWriteFields(ctx context.Context, resource string) []string {
	perms := GetUserFieldPerms(ctx, resource)
	if perms == nil {
		return nil
	}

	blocked := make([]string, 0)
	for field, perm := range perms {
		if perm <= PermReadonly {
			blocked = append(blocked, field)
		}
	}
	return blocked
}
