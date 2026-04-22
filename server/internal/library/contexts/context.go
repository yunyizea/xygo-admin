// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package contexts

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"xygo/internal/consts"
	"xygo/internal/model"
)

// Context 自定义上下文结构（存储请求相关的用户信息）
// EndpointUser 通用端点用户（扩展认证端使用）
type EndpointUser struct {
	Id       uint64
	Endpoint string
	App      string
	Data     map[string]interface{}
}

type Context struct {
	User       *model.AuthUser       // 当前登录管理员（后台）
	Member     *model.MemberUser     // 当前登录会员（前台）
	TenantUser *model.TenantAuthUser // 当前登录租户管理员（扩展预留）
	Module     string                // 应用模块（admin/member/tenant/home等）
	TenantId   uint64                // 当前租户ID（0=平台，>0=租户；扩展预留）
	AddonAuth  map[string]*EndpointUser // 扩展端点认证信息
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

// ==================== 租户相关方法（扩展预留） ====================

// SetTenantId 设置当前租户ID
func SetTenantId(ctx context.Context, tenantId uint64) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetTenantId: context is nil")
		return
	}
	c.TenantId = tenantId
}

// GetTenantId 获取当前租户ID
func GetTenantId(ctx context.Context) uint64 {
	c := Get(ctx)
	if c == nil {
		return 0
	}
	return c.TenantId
}

// SetTenantUser 将租户管理员信息设置到上下文中
func SetTenantUser(ctx context.Context, tu *model.TenantAuthUser) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetTenantUser: context is nil")
		return
	}
	c.TenantUser = tu
}

// GetTenantUser 获取租户管理员信息
func GetTenantUser(ctx context.Context) *model.TenantAuthUser {
	c := Get(ctx)
	if c == nil {
		return nil
	}
	return c.TenantUser
}

// GetTenantUserId 获取租户管理员ID
func GetTenantUserId(ctx context.Context) uint64 {
	tu := GetTenantUser(ctx)
	if tu == nil {
		return 0
	}
	return tu.Id
}

// SetEndpointUser 将扩展端点用户信息设置到上下文中
func SetEndpointUser(ctx context.Context, endpoint string, eu *EndpointUser) {
	c := Get(ctx)
	if c == nil {
		g.Log().Warning(ctx, "contexts.SetEndpointUser: context is nil")
		return
	}
	if c.AddonAuth == nil {
		c.AddonAuth = make(map[string]*EndpointUser)
	}
	c.AddonAuth[endpoint] = eu
}

// GetEndpointUser 获取扩展端点用户信息
func GetEndpointUser(ctx context.Context, endpoint string) *EndpointUser {
	c := Get(ctx)
	if c == nil || c.AddonAuth == nil {
		return nil
	}
	return c.AddonAuth[endpoint]
}

// GetEndpointUserId 获取扩展端点用户ID
func GetEndpointUserId(ctx context.Context, endpoint string) uint64 {
	eu := GetEndpointUser(ctx, endpoint)
	if eu == nil {
		return 0
	}
	return eu.Id
}
