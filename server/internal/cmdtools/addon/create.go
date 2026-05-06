package addon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"
)

type scaffoldData struct {
	Name        string // shop
	Title       string // 商城管理
	Author      string // 开发者A
	Description string // 在线商城
	TableName   string // shop_order (without xy_ prefix)
	Entity      string // order (table name minus addon prefix)
	HasFrontend bool   // 是否有独立前台（双控制器模式）

	PascalName   string // ShopOrder
	PascalEntity string // Order
	CamelName    string // shopOrder
	CamelEntity  string // order
	KebabEntity  string // order
	ModulePath   string // xygo/addons/shop
}

func Create(ctx context.Context, name string) error {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║  XYGo Admin 扩展脚手架               ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()

	if name == "" {
		name = gcmd.Scan("  扩展标识 (英文小写，如 shop): ")
		name = strings.TrimSpace(name)
		if name == "" {
			fmt.Println("  取消创建")
			return nil
		}
	}
	if !isValidAddonName(name) {
		return fmt.Errorf("扩展标识只能包含小写字母、数字和下划线，且以字母开头")
	}

	projectRoot := getProjectRoot()
	serverAddonDir := filepath.Join(projectRoot, "server", "addons", name)
	if gfile.Exists(serverAddonDir) {
		confirm := gcmd.Scan(fmt.Sprintf("  目录 server/addons/%s/ 已存在，是否覆盖？[y/N] ", name))
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			fmt.Println("  取消创建")
			return nil
		}
	}

	title := gcmd.Scan("  扩展名称 (中文，如 商城管理): ")
	title = strings.TrimSpace(title)
	if title == "" {
		title = name
	}

	author := gcmd.Scan("  作者: ")
	author = strings.TrimSpace(author)
	if author == "" {
		author = "XYGo Developer"
	}

	desc := gcmd.Scan("  描述: ")
	desc = strings.TrimSpace(desc)

	tableName := gcmd.Scan(fmt.Sprintf("  示例表名 (如 %s_order，留空跳过): ", name))
	tableName = strings.TrimSpace(tableName)

	hasFrontend := false
	frontendAnswer := gcmd.Scan("  是否有独立前台 (双控制器模式)？[y/N] ")
	if strings.ToLower(strings.TrimSpace(frontendAnswer)) == "y" {
		hasFrontend = true
	}

	d := &scaffoldData{
		Name:        name,
		Title:       title,
		Author:      author,
		Description: desc,
		HasFrontend: hasFrontend,
		ModulePath:  "xygo/addons/" + name,
	}

	if tableName != "" {
		d.TableName = tableName
		if strings.HasPrefix(tableName, name+"_") {
			d.Entity = strings.TrimPrefix(tableName, name+"_")
		} else {
			d.Entity = tableName
		}
		d.PascalName = toPascal(tableName)
		d.PascalEntity = toPascal(d.Entity)
		d.CamelName = toCamel(tableName)
		d.CamelEntity = toCamel(d.Entity)
		d.KebabEntity = toKebab(d.Entity)
	}

	// [1/5] 后端目录结构
	fmt.Print("  [1/5] 创建后端目录结构 ... ")
	serverDirs := []string{
		serverAddonDir,
	}
	if tableName != "" {
		serverDirs = append(serverDirs,
			filepath.Join(serverAddonDir, "api"),
			filepath.Join(serverAddonDir, "controller"),
			filepath.Join(serverAddonDir, "logic"),
			filepath.Join(serverAddonDir, "model"),
		)
	}
	serverDirs = append(serverDirs,
		filepath.Join(serverAddonDir, "install"),
		filepath.Join(serverAddonDir, "uninstall"),
		filepath.Join(serverAddonDir, "upgrade"),
		filepath.Join(serverAddonDir, "queues"),
		filepath.Join(serverAddonDir, "crons"),
	)
	for _, dir := range serverDirs {
		os.MkdirAll(dir, 0755)
	}
	fmt.Println("OK")

	// [2/5] 后端模板文件
	fmt.Print("  [2/5] 生成后端模板文件 ... ")
	fileCount := 0

	files := []struct {
		path string
		tpl  string
	}{
		{filepath.Join(serverAddonDir, "addon.yaml"), tplAddonYaml},
		{filepath.Join(serverAddonDir, "module.go"), tplModuleGo},
		{filepath.Join(serverAddonDir, "install", "pgsql.sql"), tplInstallPgsql},
		{filepath.Join(serverAddonDir, "install", "mysql.sql"), tplInstallMysql},
		{filepath.Join(serverAddonDir, "uninstall", "pgsql.sql"), tplUninstallPgsql},
		{filepath.Join(serverAddonDir, "uninstall", "mysql.sql"), tplUninstallMysql},
		{filepath.Join(serverAddonDir, "upgrade", "pgsql.sql"), tplUpgradePgsql},
		{filepath.Join(serverAddonDir, "upgrade", "mysql.sql"), tplUpgradeMysql},
		{filepath.Join(serverAddonDir, "queues", "example.go"), tplQueuesExample},
		{filepath.Join(serverAddonDir, "crons", "example.go"), tplCronsExample},
	}
	if tableName != "" {
		files = append(files,
			struct{ path, tpl string }{filepath.Join(serverAddonDir, "api", d.TableName+".go"), tplApiGo},
			struct{ path, tpl string }{filepath.Join(serverAddonDir, "controller", "controller.go"), tplControllerBaseGo},
			struct{ path, tpl string }{filepath.Join(serverAddonDir, "controller", d.Entity+".go"), tplControllerGo},
			struct{ path, tpl string }{filepath.Join(serverAddonDir, "logic", d.Entity+".go"), tplLogicGo},
			struct{ path, tpl string }{filepath.Join(serverAddonDir, "model", d.Entity+".go"), tplModelGo},
		)
	}
	if hasFrontend {
		files = append(files,
			struct{ path, tpl string }{filepath.Join(serverAddonDir, "middleware.go"), tplMiddlewareGo},
		)
	}

	for _, f := range files {
		if err := renderTemplate(f.path, f.tpl, d); err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("生成 %s 失败: %v", filepath.Base(f.path), err)
		}
		fileCount++
	}
	fmt.Printf("OK (%d 个文件)\n", fileCount)

	// [3/5] 前端目录
	fmt.Print("  [3/5] 创建前端目录结构 ... ")
	webAddonDir := filepath.Join(projectRoot, "web", "src", "addons", name)
	webDirs := []string{webAddonDir}
	if tableName != "" {
		webDirs = append(webDirs,
			filepath.Join(webAddonDir, "views", d.KebabEntity),
			filepath.Join(webAddonDir, "views", d.KebabEntity, "modules"),
			filepath.Join(webAddonDir, "api"),
		)
	}
	for _, dir := range webDirs {
		os.MkdirAll(dir, 0755)
	}
	fmt.Println("OK")

	// [4/5] 前端模板文件
	fmt.Print("  [4/5] 生成前端模板文件 ... ")
	webFileCount := 0
	if tableName != "" {
		webFiles := []struct {
			path string
			tpl  string
		}{
			{filepath.Join(webAddonDir, "api", d.KebabEntity+".ts"), tplWebApi},
			{filepath.Join(webAddonDir, "views", d.KebabEntity, "index.vue"), tplWebIndex},
			{filepath.Join(webAddonDir, "views", d.KebabEntity, "modules", d.KebabEntity+"-dialog.vue"), tplWebDialog},
		}
		for _, f := range webFiles {
			if err := renderTemplate(f.path, f.tpl, d); err != nil {
				fmt.Println("FAILED")
				return fmt.Errorf("生成 %s 失败: %v", filepath.Base(f.path), err)
			}
			webFileCount++
		}
	}
	fmt.Printf("OK (%d 个文件)\n", webFileCount)

	// [5/5] 更新 addons/addons.go
	fmt.Print("  [5/5] 更新扩展导入文件 ... ")
	addonsDir := filepath.Join(projectRoot, "server", "addons")
	if err := updateAddonsImport(addonsDir); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("更新 addons.go 失败: %v", err)
	}
	fmt.Println("OK")

	// 打印结果
	fmt.Println()
	fmt.Println("  ════════════════════════════════════════")
	fmt.Printf("  扩展 %s 创建完成！\n", name)
	fmt.Println("  ════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  生成的文件：")
	fmt.Printf("    server/addons/%s/\n", name)
	fmt.Println("    ├── addon.yaml")
	fmt.Println("    ├── module.go")
	if hasFrontend {
		fmt.Println("    ├── middleware.go          (独立前台鉴权骨架)")
	}
	if tableName != "" {
		fmt.Printf("    ├── api/%s.go\n", d.TableName)
		if hasFrontend {
			fmt.Printf("    ├── controller/%s.go       (含 AdminControllerV1 + ControllerV1)\n", d.Entity)
		} else {
			fmt.Printf("    ├── controller/%s.go\n", d.Entity)
		}
		fmt.Printf("    ├── logic/%s.go\n", d.Entity)
		fmt.Printf("    ├── model/%s.go\n", d.Entity)
	}
	fmt.Println("    ├── install/pgsql.sql")
	fmt.Println("    ├── install/mysql.sql")
	fmt.Println("    ├── uninstall/pgsql.sql")
	fmt.Println("    └── uninstall/mysql.sql")
	if tableName != "" {
		fmt.Println()
		fmt.Printf("    web/src/addons/%s/\n", name)
		fmt.Printf("    ├── api/%s.ts\n", d.KebabEntity)
		fmt.Printf("    ├── views/%s/index.vue\n", d.KebabEntity)
		fmt.Printf("    └── views/%s/modules/%s-dialog.vue\n", d.KebabEntity, d.KebabEntity)
	}
	fmt.Println()
	fmt.Println("  下一步：")
	fmt.Println("    1. 编辑 install/pgsql.sql 定义你的表结构")
	fmt.Println("    2. 执行安装 SQL 建表")
	fmt.Println("    3. gf gen dao        (生成数据模型)")
	fmt.Println("    4. 修改 controller/logic 实现业务逻辑")
	if hasFrontend {
		fmt.Println("    5. 完善 middleware.go 中的鉴权逻辑")
		fmt.Printf("    6. admin_*.go 的方法 receiver 用 *AdminControllerV1\n")
		fmt.Println("       其余文件的方法 receiver 用 *ControllerV1")
		fmt.Println("    7. 重启后端，访问页面验证")
	} else {
		fmt.Println("    5. 重启后端，访问页面验证")
	}
	fmt.Println()

	return nil
}

// ==================== 命名转换 ====================

func isValidAddonName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 && !unicode.IsLetter(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

// toPascal converts snake_case to PascalCase: "shop_order" -> "ShopOrder"
func toPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// toCamel converts snake_case to camelCase: "shop_order" -> "shopOrder"
func toCamel(s string) string {
	p := toPascal(s)
	if p == "" {
		return ""
	}
	return strings.ToLower(p[:1]) + p[1:]
}

// toKebab converts snake_case to kebab-case: "shop_order" -> "shop-order"
func toKebab(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}

// ==================== 模板渲染 ====================

func renderTemplate(filePath, tplContent string, data *scaffoldData) error {
	t, err := template.New(filepath.Base(filePath)).Parse(tplContent)
	if err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}

// ==================== 后端模板 ====================

var tplAddonYaml = `name: {{.Name}}
version: "1.0.0"
title: "{{.Title}}"
description: "{{.Description}}"
author: "{{.Author}}"
min_version: "1.3.0"
min_upgrade_from: ""
changelog:
  - "初始版本"

features:
  routes: true
  websocket: false
  queue: false
  cron: false
{{if .TableName}}
menus:
  admin:
    - title: "{{.Title}}"
      name: "{{.PascalName}}Manage"
      path: "{{.Name}}"
      icon: "ri:apps-line"
      type: 1
      sort: 60
      children:
        - title: "{{.PascalEntity}}管理"
          name: "{{.PascalName}}List"
          path: "{{.Entity}}"
          component: "@addons/{{.Name}}/views/{{.KebabEntity}}"
          type: 2
          sort: 1
          children:
            - title: "新增/编辑"
              name: "{{.PascalName}}Edit"
              type: 3
              perms: "/admin/{{.Name}}/{{.Entity}}/edit"
              sort: 1
            - title: "删除"
              name: "{{.PascalName}}Delete"
              type: 3
              perms: "/admin/{{.Name}}/{{.Entity}}/delete"
              sort: 2
{{else}}
# menus:
#   admin:
#     - title: "扩展名称"
#       name: "扩展PascalName"
#       path: "addon-path"
#       icon: "ri:apps-line"
#       type: 1
#       sort: 60
#       children:
#         - title: "子菜单"
#           name: "扩展PascalNameList"
#           path: "list"
#           component: "@addons/{{.Name}}/views/list"
#           type: 2
#           sort: 1
{{end}}`

var tplModuleGo = `package {{.Name}}

import (
	"xygo/internal/addon"
	"xygo/internal/middleware"
{{- if .TableName}}
	"{{.ModulePath}}/controller"
{{- end}}

	// 空导入：触发 queues、crons 子包的 init() 注册
	_ "{{.ModulePath}}/queues"
	_ "{{.ModulePath}}/crons"

	"github.com/gogf/gf/v2/net/ghttp"
)

func init() {
	addon.Register(addon.Module{
		Name:  "{{.Name}}",
		Mount: mountRoutes,
	})
}

func mountRoutes(s *ghttp.Server) {
{{- if .HasFrontend}}
	// 路由组 1：平台管理端 — 复用核心 AdminAuth
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middleware.CORS,
			middleware.ResponseHandler,
		)
		group.Group("/", func(ag *ghttp.RouterGroup) {
			ag.Middleware(middleware.AdminAuth, middleware.DemoGuard, middleware.OperationLog)
{{- if .TableName}}
			ag.Bind(controller.NewAdminV1())
{{- else}}
			// ag.Bind(controller.NewAdminV1())
{{- end}}
		})
	})

	// 路由组 2：Addon 自身端 — 独立鉴权
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middleware.CORS,
			middleware.ResponseHandler,
			{{.Name}}Resolve,
			{{.Name}}Auth,
			middleware.DemoGuard,
		)
{{- if .TableName}}
		group.Bind(controller.NewV1())
{{- else}}
		// group.Bind(controller.NewV1())
{{- end}}
	})
{{- else}}
	// 普通 addon — 仅平台管理端
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middleware.CORS,
			middleware.ResponseHandler,
		)
		group.Group("/", func(ag *ghttp.RouterGroup) {
			ag.Middleware(middleware.AdminAuth, middleware.DemoGuard, middleware.OperationLog)
{{- if .TableName}}
			ag.Bind(controller.NewV1())
{{- else}}
			// ag.Bind(controller.NewV1())
{{- end}}
		})
	})
{{- end}}
}
`

var tplApiGo = `package api

import "github.com/gogf/gf/v2/frame/g"

// ========== {{.PascalName}} 列表 ==========

type {{.PascalName}}ListReq struct {
	g.Meta   ` + "`" + `path:"/admin/{{.Name}}/{{.Entity}}/list" method:"get" tags:"{{.Title}}" summary:"{{.PascalEntity}}列表"` + "`" + `
	Page     int    ` + "`" + `json:"page" d:"1"` + "`" + `
	PageSize int    ` + "`" + `json:"pageSize" d:"20"` + "`" + `
	Status   *int   ` + "`" + `json:"status"` + "`" + `
}
type {{.PascalName}}ListRes struct {
	g.Meta ` + "`" + `mime:"application/json"` + "`" + `
}

// ========== {{.PascalName}} 保存 ==========

type {{.PascalName}}EditReq struct {
	g.Meta ` + "`" + `path:"/admin/{{.Name}}/{{.Entity}}/edit" method:"post" tags:"{{.Title}}" summary:"保存{{.PascalEntity}}"` + "`" + `
	Id     uint64 ` + "`" + `json:"id"` + "`" + `
	// TODO: 添加业务字段
}
type {{.PascalName}}EditRes struct{}

// ========== {{.PascalName}} 删除 ==========

type {{.PascalName}}DeleteReq struct {
	g.Meta ` + "`" + `path:"/admin/{{.Name}}/{{.Entity}}/delete" method:"post" tags:"{{.Title}}" summary:"删除{{.PascalEntity}}"` + "`" + `
	Id     uint64 ` + "`" + `json:"id" v:"required"` + "`" + `
}
type {{.PascalName}}DeleteRes struct{}
`

var tplControllerBaseGo = `package controller
{{- if .HasFrontend}}

type AdminControllerV1 struct{}

func NewAdminV1() *AdminControllerV1 { return &AdminControllerV1{} }

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }
{{- else}}

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }
{{- end}}
`

var tplControllerGo = `package controller

import (
	"context"
	api "{{.ModulePath}}/api"
	"{{.ModulePath}}/logic"
)

func (c *ControllerV1) {{.PascalName}}List(ctx context.Context, req *api.{{.PascalName}}ListReq) (res *api.{{.PascalName}}ListRes, err error) {
	return logic.{{.PascalEntity}}List(ctx, req)
}

func (c *ControllerV1) {{.PascalName}}Edit(ctx context.Context, req *api.{{.PascalName}}EditReq) (res *api.{{.PascalName}}EditRes, err error) {
	return logic.{{.PascalEntity}}Edit(ctx, req)
}

func (c *ControllerV1) {{.PascalName}}Delete(ctx context.Context, req *api.{{.PascalName}}DeleteReq) (res *api.{{.PascalName}}DeleteRes, err error) {
	return logic.{{.PascalEntity}}Delete(ctx, req)
}
`

var tplLogicGo = `package logic

import (
	"context"
	api "{{.ModulePath}}/api"
)

func {{.PascalEntity}}List(ctx context.Context, req *api.{{.PascalName}}ListReq) (res *api.{{.PascalName}}ListRes, err error) {
	// TODO: 实现列表查询
	res = &api.{{.PascalName}}ListRes{}
	return
}

func {{.PascalEntity}}Edit(ctx context.Context, req *api.{{.PascalName}}EditReq) (res *api.{{.PascalName}}EditRes, err error) {
	// TODO: 实现新增/编辑
	return
}

func {{.PascalEntity}}Delete(ctx context.Context, req *api.{{.PascalName}}DeleteReq) (res *api.{{.PascalName}}DeleteRes, err error) {
	// TODO: 实现删除
	return
}
`

var tplModelGo = `package model

// {{.PascalName}}ListItem 列表项
type {{.PascalName}}ListItem struct {
	Id     uint64 ` + "`" + `json:"id"` + "`" + `
	Status int    ` + "`" + `json:"status"` + "`" + `
	Sort   int    ` + "`" + `json:"sort"` + "`" + `
	// TODO: 添加业务字段
}
`

// ==================== SQL 模板 ====================

var tplInstallPgsql = `-- ============================================================
-- 扩展: {{.Name}} ({{.Title}})
-- 安装 SQL - PostgreSQL
-- 注意: 菜单由 addon.yaml 声明，installer 自动写入，无需在此手写
-- ============================================================
{{if .TableName}}
-- 建表（表名必须以 xy_{{.Name}}_ 为前缀）
CREATE TABLE IF NOT EXISTS xy_{{.TableName}} (
    id         bigserial PRIMARY KEY,
    title      varchar(255) NOT NULL DEFAULT '',
    status     smallint     NOT NULL DEFAULT 1,
    sort       int          NOT NULL DEFAULT 0,
    created_by bigint       NOT NULL DEFAULT 0,
    updated_by bigint       NOT NULL DEFAULT 0,
    created_at bigint       NOT NULL DEFAULT 0,
    updated_at bigint       NOT NULL DEFAULT 0,
    deleted_at bigint       NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_{{.TableName}}_deleted_at ON xy_{{.TableName}}(deleted_at);
CREATE INDEX IF NOT EXISTS idx_{{.TableName}}_status ON xy_{{.TableName}}(status);

COMMENT ON TABLE xy_{{.TableName}} IS '{{.Title}}-{{.PascalEntity}}表';
{{else}}
-- TODO: 在此编写建表语句
-- 表名规范: xy_{{.Name}}_xxx
{{end}}`

var tplInstallMysql = `-- ============================================================
-- 扩展: {{.Name}} ({{.Title}})
-- 安装 SQL - MySQL
-- 注意: 菜单由 addon.yaml 声明，installer 自动写入，无需在此手写
-- ============================================================
{{if .TableName}}
-- 建表（表名必须以 xy_{{.Name}}_ 为前缀）
CREATE TABLE IF NOT EXISTS ` + "`" + `xy_{{.TableName}}` + "`" + ` (
    ` + "`" + `id` + "`" + `         bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    ` + "`" + `title` + "`" + `      varchar(255) NOT NULL DEFAULT '',
    ` + "`" + `status` + "`" + `     tinyint NOT NULL DEFAULT 1,
    ` + "`" + `sort` + "`" + `       int NOT NULL DEFAULT 0,
    ` + "`" + `created_by` + "`" + ` bigint NOT NULL DEFAULT 0,
    ` + "`" + `updated_by` + "`" + ` bigint NOT NULL DEFAULT 0,
    ` + "`" + `created_at` + "`" + ` bigint NOT NULL DEFAULT 0,
    ` + "`" + `updated_at` + "`" + ` bigint NOT NULL DEFAULT 0,
    ` + "`" + `deleted_at` + "`" + ` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (` + "`" + `id` + "`" + `),
    INDEX ` + "`" + `idx_deleted_at` + "`" + ` (` + "`" + `deleted_at` + "`" + `),
    INDEX ` + "`" + `idx_status` + "`" + ` (` + "`" + `status` + "`" + `)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='{{.Title}}-{{.PascalEntity}}表';
{{else}}
-- TODO: 在此编写建表语句
-- 表名规范: xy_{{.Name}}_xxx
{{end}}`

var tplUninstallPgsql = `-- 扩展: {{.Name}} 卸载 - PostgreSQL
-- 注意: 菜单由 installer 自动删除（通过 remark 标记），无需在此手写
{{if .TableName}}
-- 删除数据表（谨慎：会丢失所有数据）
DROP TABLE IF EXISTS xy_{{.TableName}};
{{end}}`

var tplUninstallMysql = `-- 扩展: {{.Name}} 卸载 - MySQL
-- 注意: 菜单由 installer 自动删除（通过 remark 标记），无需在此手写
{{if .TableName}}
-- 删除数据表（谨慎：会丢失所有数据）
DROP TABLE IF EXISTS ` + "`" + `xy_{{.TableName}}` + "`" + `;
{{end}}`

var tplUpgradePgsql = `-- ============================================================
-- 扩展: {{.Name}} 增量升级 - PostgreSQL
-- ============================================================
-- 重要：此文件使用幂等写法，所有语句必须可重复执行。
-- 升级时 installer 会直接执行此文件，覆盖所有历史版本变更。
-- ============================================================

-- 示例：新增字段（IF NOT EXISTS 保证幂等）
-- ALTER TABLE xy_{{.Name}}_xxx ADD COLUMN IF NOT EXISTS new_field varchar(255) DEFAULT '';

-- 示例：新增索引
-- CREATE INDEX IF NOT EXISTS idx_{{.Name}}_xxx_field ON xy_{{.Name}}_xxx(field_name);

-- 示例：新增表（CREATE TABLE IF NOT EXISTS 保证幂等）
-- CREATE TABLE IF NOT EXISTS xy_{{.Name}}_yyy (
--     id bigserial PRIMARY KEY,
--     name varchar(100) NOT NULL DEFAULT ''
-- );

-- 示例：更新配置（使用 ON CONFLICT 保证幂等）
-- INSERT INTO xy_sys_config (group, key, value)
-- VALUES ('{{.Name}}', 'new_key', 'default_value')
-- ON CONFLICT (group, key) DO NOTHING;
`

var tplUpgradeMysql = `-- ============================================================
-- 扩展: {{.Name}} 增量升级 - MySQL
-- ============================================================
-- 重要：此文件使用幂等写法，所有语句必须可重复执行。
-- 升级时 installer 会直接执行此文件，覆盖所有历史版本变更。
-- ============================================================

-- 示例：新增字段（MySQL 不支持 IF NOT EXISTS，需用存储过程或忽略错误）
-- ALTER TABLE xy_{{.Name}}_xxx ADD COLUMN new_field varchar(255) DEFAULT '';
-- 注意：如果字段已存在 MySQL 会报错，建议用以下方式：
-- SET @sql = (SELECT IF(
--     (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME='xy_{{.Name}}_xxx' AND COLUMN_NAME='new_field') = 0,
--     'ALTER TABLE xy_{{.Name}}_xxx ADD COLUMN new_field varchar(255) DEFAULT ""',
--     'SELECT 1'));
-- PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 示例：新增表
-- CREATE TABLE IF NOT EXISTS ` + "`" + `xy_{{.Name}}_yyy` + "`" + ` (
--     ` + "`" + `id` + "`" + ` bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
--     ` + "`" + `name` + "`" + ` varchar(100) NOT NULL DEFAULT ''
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 示例：更新配置（使用 INSERT IGNORE 保证幂等）
-- INSERT IGNORE INTO xy_sys_config (` + "`" + `group` + "`" + `, ` + "`" + `key` + "`" + `, ` + "`" + `value` + "`" + `)
-- VALUES ('{{.Name}}', 'new_key', 'default_value');
`

// ==================== 前端模板 ====================

var tplWebApi = `/**
 * {{.Title}} - {{.PascalEntity}}管理 API
 */
import { adminRequest } from '@/utils/http'

const prefix = '/admin/{{.Name}}/{{.Entity}}'

/** 列表 */
export function fetch{{.PascalName}}List(params: any) {
  return adminRequest.get<Record<string, any>>({
    url: ` + "`" + `${prefix}/list` + "`" + `,
    params
  })
}

/** 保存(新增/编辑) */
export function fetch{{.PascalName}}Edit(params: any) {
  return adminRequest.post<any>({
    url: ` + "`" + `${prefix}/edit` + "`" + `,
    params
  })
}

/** 删除 */
export function fetch{{.PascalName}}Delete(id: number) {
  return adminRequest.post<any>({
    url: ` + "`" + `${prefix}/delete` + "`" + `,
    params: { id }
  })
}
`

var tplWebIndex = `<!-- {{.Title}} - {{.PascalEntity}}管理 -->
<template>
  <div class="{{.Name}}-{{.KebabEntity}}-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <{{.CamelEntity}}Dialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :edit-data="currentRow"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { formatTimestamp } from '@/utils/time'
  import { fetch{{.PascalName}}List, fetch{{.PascalName}}Edit, fetch{{.PascalName}}Delete } from '../../api/{{.KebabEntity}}'
  import {{.CamelEntity}}Dialog from './modules/{{.KebabEntity}}-dialog.vue'
  import { ElTag, ElMessageBox } from 'element-plus'
  import { DialogType } from '@/types'

  defineOptions({ name: '{{.PascalName}}' })

  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentRow = ref<any>({})

  const {
    columns, columnChecks, data, loading, pagination,
    getData, searchParams, resetSearchParams,
    handleSizeChange, handleCurrentChange, refreshData
  } = useTable({
    core: {
      apiFn: fetch{{.PascalName}}List,
      apiParams: { page: 1, pageSize: 20 },
      paginationKey: { current: 'page', size: 'pageSize' },
      columnsFactory: () => [
        {
          prop: 'id',
          label: 'ID',
          width: 80
        },
        {
          prop: 'title',
          label: '标题',
          minWidth: 200,
          formatter: (row: any) => row.title ?? '-'
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          align: 'center',
          formatter: (row: any) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'danger', size: 'small' },
              () => row.status === 1 ? '启用' : '禁用')
        },
        {
          prop: 'createdAt',
          label: '创建时间',
          width: 180,
          formatter: (row: any) => formatTimestamp(row.createdAt)
        },
        {
          prop: 'operation',
          label: '操作',
          width: 180,
          fixed: 'right',
          formatter: (row: any) =>
            h('div', { class: 'flex items-center gap-1' }, [
              h(ArtButtonTable, { type: 'edit', onClick: () => showDialog('edit', row) }),
              h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
            ])
        }
      ]
    }
  })

  const showDialog = (type: DialogType, row?: any) => {
    dialogType.value = type
    currentRow.value = row || {}
    nextTick(() => { dialogVisible.value = true })
  }

  const handleDelete = async (row: any) => {
    try {
      await ElMessageBox.confirm('确定要删除该记录吗？删除后无法恢复', '删除确认', {
        confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
      })
      await fetch{{.PascalName}}Delete(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch (e) { if (e !== 'cancel') console.error(e) }
  }

  const handleDialogSubmit = async (formData: any) => {
    try {
      await fetch{{.PascalName}}Edit(formData)
      ElMessage.success(formData.id ? '编辑成功' : '添加成功')
      dialogVisible.value = false
      refreshData()
    } catch (e) { console.error(e) }
  }
</script>
`

var tplWebDialog = `<!-- {{.Title}} - {{.PascalEntity}}编辑弹窗 -->
<template>
  <ElDialog
    v-model="visible"
    :title="type === 'add' ? '新增{{.PascalEntity}}' : '编辑{{.PascalEntity}}'"
    width="600px"
    :close-on-click-modal="false"
    @closed="resetForm"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="80px">
      <ElFormItem label="标题" prop="title">
        <ElInput v-model="form.title" placeholder="请输入标题" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
      </ElFormItem>
      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { DialogType } from '@/types'

  const props = defineProps<{
    visible: boolean
    type: DialogType
    editData: Record<string, any>
  }>()
  const emit = defineEmits(['update:visible', 'submit'])

  const visible = computed({
    get: () => props.visible,
    set: (val) => emit('update:visible', val)
  })

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = ref({
    id: undefined as number | undefined,
    title: '',
    status: 1,
    sort: 0
  })

  const rules: FormRules = {
    title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
  }

  watch(() => props.visible, (val) => {
    if (val && props.type === 'edit' && props.editData) {
      form.value = { ...form.value, ...props.editData }
    }
  })

  const resetForm = () => {
    form.value = { id: undefined, title: '', status: 1, sort: 0 }
    formRef.value?.resetFields()
  }

  const handleSubmit = async () => {
    const valid = await formRef.value?.validate().catch(() => false)
    if (!valid) return
    submitting.value = true
    try {
      emit('submit', { ...form.value })
    } finally {
      submitting.value = false
    }
  }
</script>
`

var tplQueuesExample = `package queues

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
	"xygo/internal/library/queue"
)

// Topic 常量（生产者通过引用此常量投递消息）
const TopicExample = "{{.Name}}.example"

func init() {
	queue.Register(&ExampleConsumer{})
}

type ExampleConsumer struct{}

func (c *ExampleConsumer) GetTopic() string { return TopicExample }

func (c *ExampleConsumer) Handle(ctx context.Context, msg *queue.Message) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Body), &data); err != nil {
		g.Log().Errorf(ctx, "[queue:%s] unmarshal failed: %v", TopicExample, err)
		return nil
	}
	g.Log().Infof(ctx, "[queue:%s] received: %v", TopicExample, data)
	return nil
}
`

var tplCronsExample = `package crons

import (
	"context"

	cronlib "xygo/internal/library/cron"
)

func init() {
	cronlib.Register(&ExampleTask{})
}

// ExampleTask 示例定时任务
// 注册后需在后台 系统管理→定时任务 中添加任务并配置 cron 表达式
type ExampleTask struct{}

func (t *ExampleTask) GetName() string {
	return "{{.Name}}.example"
}

func (t *ExampleTask) Execute(ctx context.Context, params []string) (string, error) {
	return "example task executed", nil
}
`

// ==================== 中间件模板（含独立前台时生成） ====================

var tplMiddlewareGo = `package {{.Name}}

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"xygo/internal/consts"
	"xygo/internal/library/contexts"
	"xygo/internal/library/token"
)

// {{.Name}}Resolve 识别中间件（从请求中解析上下文信息并注入 context）
func {{.Name}}Resolve(r *ghttp.Request) {
	// TODO: 根据业务需求从请求头/域名/token 中解析上下文
	r.Middleware.Next()
}

// {{.Name}}Auth 鉴权中间件
func {{.Name}}Auth(r *ghttp.Request) {
	customCtx := &contexts.Context{Module: "{{.Name}}"}
	contexts.Init(r, customCtx)

	// 跳过登录接口
	path := r.URL.Path
	if path == "/{{.Name}}/auth/login" {
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

	// TODO: 使用 token.Endpoint 解析令牌并注入用户信息到 context
	_ = token.ErrTokenKicked // 示例：可处理踢出场景
	_ = tokenStr

	r.Middleware.Next()
}
`
