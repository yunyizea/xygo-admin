package middleware

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"xygo/internal/consts"
	"xygo/internal/library/contexts"
	"xygo/internal/library/token"
)

// AdminAuth 后台接口鉴权中间件
// - 放行 /admin/auth/login
// - 其它 /admin/** 必须携带有效 accessToken（Authorization: Bearer XXX）
// - ✨ 验证通过后，将用户信息注入上下文，供 Handler 使用
// - ✨ 区分"过期/未登录"与"被踢下线"，返回不同错误码
func AdminAuth(r *ghttp.Request) {
	path := r.URL.Path

	// 初始化自定义上下文
	customCtx := &contexts.Context{
		Module: "admin",
	}
	contexts.Init(r, customCtx)

	// 登录和刷新接口放行
	if path == "/admin/auth/login" || path == "/admin/auth/refresh" {
		r.Middleware.Next()
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.SetError(gerror.NewCode(consts.CodeNotAuthorized, "未登录"))
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		r.SetError(gerror.NewCode(consts.CodeNotAuthorized, "未登录"))
		return
	}

	// ✨ 解析 Token 并获取完整用户信息
	authUser, err := token.Parse(r.Context(), tokenStr)
	if err != nil {
		// ✨ 区分"被踢下线"和"普通过期/失效"
		if errors.Is(err, token.ErrTokenKicked) {
			r.SetError(gerror.NewCode(consts.CodeKickedOut, "您的账号已在其他设备登录，请重新登录"))
			return
		}
		r.SetError(gerror.NewCode(consts.CodeNotAuthorized, "登录已失效，请重新登录"))
		return
	}

	// ✨ 将用户信息注入到上下文中
	contexts.SetUser(r.Context(), authUser)

	r.Middleware.Next()
}
