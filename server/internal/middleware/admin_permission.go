package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/library/contexts"
	"xygo/internal/model/entity"
)

type permItem struct {
	MenuId uint64
	Name   string // 按钮 name（如 add, edit, delete）
}

var permCache struct {
	sync.RWMutex
	data   map[string][]permItem // perm("METHOD /path") -> 关联的菜单列表
	loaded bool
}

// AdminPermission 后台 API 权限校验中间件（放在 AdminAuth 之后）
func AdminPermission(r *ghttp.Request) {
	user := contexts.GetUser(r.Context())
	if user == nil {
		r.Middleware.Next()
		return
	}

	if consts.IsSuperRole(user.RoleKey) {
		r.Middleware.Next()
		return
	}

	perm := strings.ToUpper(r.Method) + " " + r.URL.Path

	items := getPermItems(r.Context(), perm)
	if len(items) == 0 {
		r.Middleware.Next()
		return
	}

	menuIds := resolveMenuIds(r, items)

	allowed, err := userHasMenuPermission(r.Context(), user.Id, menuIds)
	if err != nil {
		g.Log().Warningf(r.Context(), "权限校验异常，放行: %v", err)
		r.Middleware.Next()
		return
	}

	if !allowed {
		r.SetError(gerror.NewCode(consts.CodeNoPermission, "无操作权限"))
		return
	}

	r.Middleware.Next()
}

// resolveMenuIds 根据请求内容区分 add/edit，精确匹配按钮权限
func resolveMenuIds(r *ghttp.Request, items []permItem) []uint64 {
	if len(items) <= 1 {
		return []uint64{items[0].MenuId}
	}

	hasAdd := false
	hasEdit := false
	for _, it := range items {
		n := strings.ToLower(it.Name)
		if strings.Contains(n, "add") {
			hasAdd = true
		}
		if strings.Contains(n, "edit") {
			hasEdit = true
		}
	}

	if hasAdd && hasEdit {
		idVal := r.Get("id")
		isAddOp := idVal == nil || idVal.IsEmpty() || idVal.Int64() == 0

		target := "add"
		if !isAddOp {
			target = "edit"
		}

		for _, it := range items {
			if strings.Contains(strings.ToLower(it.Name), target) {
				return []uint64{it.MenuId}
			}
		}
	}

	ids := make([]uint64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.MenuId)
	}
	return ids
}

func getPermItems(ctx context.Context, perm string) []permItem {
	ensurePermCacheLoaded(ctx)
	permCache.RLock()
	defer permCache.RUnlock()
	return permCache.data[perm]
}

func userHasMenuPermission(ctx context.Context, userId uint64, menuIds []uint64) (bool, error) {
	var userRoles []entity.AdminUserRole
	if err := dao.AdminUserRole.Ctx(ctx).
		Where(dao.AdminUserRole.Columns().UserId, userId).
		Scan(&userRoles); err != nil {
		return false, err
	}
	if len(userRoles) == 0 {
		return false, nil
	}

	roleIds := make([]uint64, 0, len(userRoles))
	for _, r := range userRoles {
		if r.RoleId > 0 {
			roleIds = append(roleIds, r.RoleId)
		}
	}

	allRoleIds, err := collectParentRoleIds(ctx, roleIds)
	if err != nil {
		return false, err
	}

	count, err := dao.AdminRoleMenu.Ctx(ctx).
		Where(dao.AdminRoleMenu.Columns().RoleId, allRoleIds).
		Where(dao.AdminRoleMenu.Columns().MenuId, menuIds).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// collectParentRoleIds 递归收集角色及其所有父角色 ID
func collectParentRoleIds(ctx context.Context, roleIds []uint64) ([]uint64, error) {
	if len(roleIds) == 0 {
		return roleIds, nil
	}

	allIds := make(map[uint64]struct{})
	for _, id := range roleIds {
		allIds[id] = struct{}{}
	}

	pending := make([]uint64, len(roleIds))
	copy(pending, roleIds)

	for len(pending) > 0 {
		var roles []entity.AdminRole
		if err := dao.AdminRole.Ctx(ctx).WhereIn("id", pending).Scan(&roles); err != nil {
			return nil, err
		}
		pending = pending[:0]
		for _, role := range roles {
			if role.Pid > 0 {
				if _, exists := allIds[role.Pid]; !exists {
					allIds[role.Pid] = struct{}{}
					pending = append(pending, role.Pid)
				}
			}
		}
	}

	result := make([]uint64, 0, len(allIds))
	for id := range allIds {
		result = append(result, id)
	}
	return result, nil
}

// ==================== 缓存管理 ====================

func ensurePermCacheLoaded(ctx context.Context) {
	permCache.RLock()
	loaded := permCache.loaded
	permCache.RUnlock()
	if !loaded {
		loadPermCache(ctx)
	}
}

func loadPermCache(ctx context.Context) {
	permCache.Lock()
	defer permCache.Unlock()
	if permCache.loaded {
		return
	}

	bgCtx := context.Background()
	var menus []*entity.AdminMenu
	err := dao.AdminMenu.Ctx(bgCtx).
		Where(dao.AdminMenu.Columns().Status, 1).
		Scan(&menus)
	if err != nil {
		g.Log().Errorf(bgCtx, "加载权限缓存失败: %v", err)
		return
	}

	data := make(map[string][]permItem)
	for _, m := range menus {
		if m.Perms == "" {
			continue
		}
		for _, perm := range parsePerms(m.Perms) {
			if perm != "" {
				data[perm] = append(data[perm], permItem{
					MenuId: m.Id,
					Name:   m.Name,
				})
			}
		}
	}

	permCache.data = data
	permCache.loaded = true
	g.Log().Infof(bgCtx, "API权限缓存已加载，共 %d 条权限点映射", len(data))
}

func parsePerms(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var perms []string
	if err := json.Unmarshal([]byte(raw), &perms); err != nil {
		perms = []string{raw}
	}

	result := make([]string, 0, len(perms))
	for _, p := range perms {
		p = strings.TrimSpace(p)
		if p == "" || strings.Contains(p, ":") && !strings.Contains(p, "/") {
			continue
		}
		p = strings.ToUpper(p[:strings.Index(p, " ")+1]) + p[strings.Index(p, " ")+1:]
		result = append(result, normalizePermKey(p))
	}
	return result
}

func normalizePermKey(perm string) string {
	perm = strings.TrimSpace(perm)
	if perm == "" {
		return ""
	}
	parts := strings.SplitN(perm, " ", 2)
	if len(parts) == 2 {
		return strings.ToUpper(parts[0]) + " " + parts[1]
	}
	return perm
}

// RefreshPermCache 刷新权限缓存（菜单或角色菜单变更时调用）
func RefreshPermCache(ctx context.Context) {
	permCache.Lock()
	permCache.loaded = false
	permCache.Unlock()
	loadPermCache(ctx)
}
