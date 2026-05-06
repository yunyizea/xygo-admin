package contexts

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"xygo/internal/consts"
	"xygo/internal/model"
)

// Context 自定义上下文结构（存储请求相关的用户信息）
type Context struct {
	User      *model.AuthUser          // 当前登录管理员（后台）
	Member    *model.MemberUser        // 当前登录会员（前台）
	Module    string                   // 应用模块（admin/member/tenant/home等）
	AddonAuth map[string]*EndpointUser // 扩展认证端点用户（由 Auth Endpoint Factory 写入）
	Data      map[string]any           // 通用扩展数据（供 addon 存取，如 tenantId / tenantUser 等）
}

// EndpointUser 扩展认证端点的通用用户信息
type EndpointUser struct {
	Id   uint64         `json:"id"`
	Data map[string]any `json:"data"`
}

// Init 初始化上下文对象到请求中
func Init(r *ghttp.Request, customCtx *Context) {
	r.SetCtxVar(consts.ContextKey, customCtx)
}

// Get 获取上下文变量
func Get(ctx context.Context) *Context {
	value := ctx.Value(consts.ContextKey)
	if value == nil {
		return nil
	}
	if localCtx, ok := value.(*Context); ok {
		return localCtx
	}
	return nil
}

// SetUser 将用户信息设置到上下文中
func SetUser(ctx context.Context, user *model.AuthUser) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetUser: context is nil")
		return
	}
	c.User = user
}

// SetModule 设置应用模块
func SetModule(ctx context.Context, module string) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetModule: context is nil")
		return
	}
	c.Module = module
}

// GetUser 获取用户信息
func GetUser(ctx context.Context) *model.AuthUser {
	c := Get(ctx)
	if c == nil {
		return nil
	}
	return c.User
}

// GetUserId 获取用户ID
func GetUserId(ctx context.Context) uint64 {
	user := GetUser(ctx)
	if user == nil {
		return 0
	}
	return user.Id
}

// GetDeptId 获取用户部门ID
func GetDeptId(ctx context.Context) uint64 {
	user := GetUser(ctx)
	if user == nil {
		return 0
	}
	return user.DeptId
}

// GetRoleId 获取用户角色ID
func GetRoleId(ctx context.Context) uint64 {
	user := GetUser(ctx)
	if user == nil {
		return 0
	}
	return user.RoleId
}

// GetRoleKey 获取用户角色标识
func GetRoleKey(ctx context.Context) string {
	user := GetUser(ctx)
	if user == nil {
		return ""
	}
	return user.RoleKey
}

// IsSuper 判断是否为超级管理员（只基于角色标识）
func IsSuper(ctx context.Context) bool {
	user := GetUser(ctx)
	if user == nil {
		return false
	}
	// ✅ 只判断角色标识（对齐 HotGo）
	return consts.IsSuperRole(user.RoleKey)
}

// ==================== 会员相关方法 ====================

// SetMember 将会员信息设置到上下文中
func SetMember(ctx context.Context, member *model.MemberUser) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetMember: context is nil")
		return
	}
	c.Member = member
}

// GetMember 获取会员信息
func GetMember(ctx context.Context) *model.MemberUser {
	c := Get(ctx)
	if c == nil {
		return nil
	}
	return c.Member
}

// GetMemberId 获取会员ID
func GetMemberId(ctx context.Context) uint64 {
	member := GetMember(ctx)
	if member == nil {
		return 0
	}
	return member.Id
}

// GetMemberGroupId 获取会员分组ID
func GetMemberGroupId(ctx context.Context) uint64 {
	member := GetMember(ctx)
	if member == nil {
		return 0
	}
	return member.GroupId
}

// ==================== 通用扩展数据方法（供 addon 使用） ====================

// SetData 设置通用扩展数据
func SetData(ctx context.Context, key string, value any) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetData: context is nil")
		return
	}
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
	c.Data[key] = value
}

// GetData 获取通用扩展数据
func GetData(ctx context.Context, key string) any {
	c := Get(ctx)
	if c == nil || c.Data == nil {
		return nil
	}
	return c.Data[key]
}

// GetDataUint64 获取通用扩展数据（uint64 类型安全取值）
func GetDataUint64(ctx context.Context, key string) uint64 {
	v := GetData(ctx, key)
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case uint64:
		return val
	case int64:
		return uint64(val)
	case int:
		return uint64(val)
	case float64:
		return uint64(val)
	default:
		return 0
	}
}

// ==================== 扩展认证端点相关方法 ====================

// SetEndpointUser 设置扩展认证端点用户信息
func SetEndpointUser(ctx context.Context, endpointName string, user *EndpointUser) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetEndpointUser: context is nil")
		return
	}
	if c.AddonAuth == nil {
		c.AddonAuth = make(map[string]*EndpointUser)
	}
	c.AddonAuth[endpointName] = user
}

// GetEndpointUser 获取扩展认证端点用户信息
func GetEndpointUser(ctx context.Context, endpointName string) *EndpointUser {
	c := Get(ctx)
	if c == nil || c.AddonAuth == nil {
		return nil
	}
	return c.AddonAuth[endpointName]
}

// GetEndpointUserId 获取扩展认证端点用户ID
func GetEndpointUserId(ctx context.Context, endpointName string) uint64 {
	u := GetEndpointUser(ctx, endpointName)
	if u == nil {
		return 0
	}
	return u.Id
}
