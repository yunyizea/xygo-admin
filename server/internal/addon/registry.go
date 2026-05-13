package addon

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gres"
)

// Module 扩展模块注册信息
type Module struct {
	Name  string                // 扩展唯一标识
	Mount func(s *ghttp.Server) // 路由挂载函数
}

var modules []Module

// Register 注册扩展模块（在各扩展 init() 中调用）
func Register(m Module) {
	modules = append(modules, m)
}

// MountAll 挂载所有已注册扩展的路由和静态资源（在 cmd.go 中 s.Run() 前调用）
func MountAll(s *ghttp.Server) {
	for _, m := range modules {
		if m.Mount != nil {
			g.Log().Infof(nil, "[addon] mounting routes: %s", m.Name)
			m.Mount(s)
		}
	}
	mountStaticPaths(s)
	if len(modules) > 0 {
		g.Log().Infof(nil, "[addon] %d addon(s) mounted", len(modules))
	}
}

// mountStaticPaths 自动为已注册插件挂载静态资源目录
// 约定：addons/{name}/public/ 存在时，映射到 /addons/{name}/
// 静态资源与插件代码同目录，保持插件完全自包含
func mountStaticPaths(s *ghttp.Server) {
	const addonsDir = "addons"
	for _, m := range modules {
		publicDir := addonsDir + "/" + m.Name + "/public"
		if gres.Contains(publicDir) || gfile.Exists(publicDir) {
			prefix := "/addons/" + m.Name
			g.Log().Infof(nil, "[addon] static: %s -> %s", prefix, publicDir)
			s.AddStaticPath(prefix, publicDir)
		}
	}
}

// Installed 返回已注册的扩展名列表
func Installed() []string {
	names := make([]string, 0, len(modules))
	for _, m := range modules {
		names = append(names, m.Name)
	}
	return names
}
