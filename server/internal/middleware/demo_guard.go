// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// demoGuardConfig 演示模式拦截配置
type demoGuardConfig struct {
	writeMethods map[string]bool
	allowPaths   map[string]bool
	readSuffixes []string
}

var demoConfig = demoGuardConfig{
	writeMethods: map[string]bool{
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
	},
	allowPaths: map[string]bool{
		// admin 端
		"/admin/auth/login":          true,
		"/admin/auth/logout":         true,
		"/admin/auth/info":           true,
		"/admin/chat/send":           true,
		"/admin/chat/read":           true,
		"/admin/chat/session/create": true,
		"/admin/notice/read":         true,
		"/admin/notice/readAll":      true,
		// member 端
		"/member/auth/login":           true,
		"/member/auth/register":        true,
		"/member/auth/logout":          true,
		"/member/user/checkin":         true,
		"/member/user/notice/read":     true,
		"/member/user/notice/read-all": true,
		// 公共
		"/captcha/checkClick": true,
	},
	readSuffixes: []string{
		"/list",
		"/detail",
		"/info",
		"/tree",
		"/select",
		"/option",
		"/options",
	},
}

var codeDemoMode = gcode.New(10005, "演示模式", nil)

// DemoGuard 演示模式中间件
// 当 config.yaml 中 system.demoMode=true 时，拦截所有写操作并返回提示，白名单内的接口放行
func DemoGuard(r *ghttp.Request) {
	enabled := g.Cfg().MustGet(r.GetCtx(), "system.demoMode").Bool()
	if !enabled {
		r.Middleware.Next()
		return
	}

	if !demoConfig.writeMethods[r.Method] {
		r.Middleware.Next()
		return
	}

	path := r.URL.Path

	if demoConfig.allowPaths[path] {
		r.Middleware.Next()
		return
	}

	for _, suffix := range demoConfig.readSuffixes {
		if strings.HasSuffix(path, suffix) {
			r.Middleware.Next()
			return
		}
	}

	r.SetError(gerror.NewCode(codeDemoMode, "当前为演示模式，禁止修改操作"))
}
