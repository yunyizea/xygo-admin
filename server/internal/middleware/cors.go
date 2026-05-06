package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CORS 跨域中间件（用于 admin / member 等前端对接的路由组）
// - security.corsOrigins 为空：允许所有来源（开发模式）
// - security.corsOrigins 有值：只允许白名单中的域名（生产模式）
func CORS(r *ghttp.Request) {
	origin := r.Header.Get("Origin")

	// 读取白名单配置
	allowedOrigins := g.Cfg().MustGet(r.GetCtx(), "security.corsOrigins").Strings()

	if len(allowedOrigins) == 0 {
		// 未配置白名单 → 开发模式，允许所有来源
		r.Response.CORSDefault()
	} else {
		// 生产模式 → 检查 Origin 是否在白名单中
		allowed := false
		for _, o := range allowedOrigins {
			if strings.EqualFold(o, origin) {
				allowed = true
				break
			}
		}

		if allowed {
			corsAllow(r, origin)
		} else if origin != "" {
			// Origin 不在白名单中，不设置 CORS 头（浏览器会拒绝）
			if r.Method == "OPTIONS" {
				r.Response.WriteStatus(403)
				return
			}
		}
	}

	handleOptionsAndNext(r)
}

// CORSOpen 开放 API 跨域中间件（用于 /api 等开放接口路由组）
// 允许任何来源调用（服务端调用本就不受限，这里主要为第三方前端放行）
// 安全性由 API Key / 签名机制保证，不依赖 CORS
func CORSOpen(r *ghttp.Request) {
	r.Response.CORSDefault()
	handleOptionsAndNext(r)
}

// corsAllow 设置允许的 CORS 响应头
func corsAllow(r *ghttp.Request, origin string) {
	r.Response.CORS(ghttp.CORSOptions{
		AllowOrigin:      origin,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,Xy-User-Token,X-Requested-With",
		ExposeHeaders:    "Content-Length,Content-Type",
		MaxAge:           3600,
		AllowCredentials: "true",
	})
}

// handleOptionsAndNext 处理预检请求并放行
func handleOptionsAndNext(r *ghttp.Request) {
	if r.Method == "OPTIONS" {
		r.Response.WriteStatus(204)
		return
	}
	r.Middleware.Next()
}
