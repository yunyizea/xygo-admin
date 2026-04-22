package addon

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Module 扩展模块注册信息
type Module struct {
	Name  string                  // 扩展唯一标识
	Mount func(s *ghttp.Server)   // 路由挂载函数
}

var modules []Module

// Register 注册扩展模块（在各扩展 init() 中调用）
func Register(m Module) {
	modules = append(modules, m)
}

// MountAll 挂载所有已注册扩展的路由（在 cmd.go 中 s.Run() 前调用）
func MountAll(s *ghttp.Server) {
	for _, m := range modules {
		if m.Mount != nil {
			g.Log().Infof(nil, "[addon] mounting: %s", m.Name)
			m.Mount(s)
		}
	}
	if len(modules) > 0 {
		g.Log().Infof(nil, "[addon] %d addon(s) mounted", len(modules))
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
