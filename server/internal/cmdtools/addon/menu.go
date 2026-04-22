package addon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
)

const menuTable = "xy_admin_menu"

// MenuNode 扩展菜单声明（addon.yaml 中 menus 字段）
type MenuNode struct {
	Title     string     `yaml:"title"`
	Name      string     `yaml:"name"`
	Path      string     `yaml:"path"`
	Component string     `yaml:"component"`
	Icon      string     `yaml:"icon"`
	Type      int        `yaml:"type"`       // 1=目录, 2=菜单, 3=按钮
	Hidden    int        `yaml:"hidden"`     // 0=否, 1=是
	KeepAlive int        `yaml:"keep_alive"` // 0=否, 1=是
	Sort      int        `yaml:"sort"`
	Perms     string     `yaml:"perms"`
	Redirect  string     `yaml:"redirect"`
	Children  []MenuNode `yaml:"children"`
}

// AddonMenus addon.yaml 中菜单声明的顶层结构
type AddonMenus struct {
	Admin  []MenuNode `yaml:"admin"`
	Tenant []MenuNode `yaml:"tenant"`
}

// installMenus 安装扩展菜单：校验 → 冲突检测 → 递归写入
func installMenus(ctx context.Context, db gdb.DB, addonName string, menus AddonMenus) error {
	pascalPrefix := toPascal(addonName)

	// 收集所有待写入菜单，做预校验
	allNodes := collectAllNodes(menus.Admin)
	allNodes = append(allNodes, collectAllNodes(menus.Tenant)...)

	if len(allNodes) == 0 {
		return nil
	}

	// 校验：所有菜单 name 必须以扩展 PascalName 开头
	for _, node := range allNodes {
		if node.Name == "" {
			return fmt.Errorf("菜单声明缺少 name 字段 (title=%s)", node.Title)
		}
		if !strings.HasPrefix(node.Name, pascalPrefix) {
			return fmt.Errorf("菜单 name '%s' 必须以 '%s' 为前缀", node.Name, pascalPrefix)
		}
	}

	// 冲突检测：检查 name 是否已被非本扩展占用
	existingNames, err := getExistingMenuNames(ctx, db, addonName)
	if err != nil {
		return fmt.Errorf("查询菜单失败: %v", err)
	}
	for _, node := range allNodes {
		if owner, exists := existingNames[node.Name]; exists {
			return fmt.Errorf("菜单 name '%s' 已被 '%s' 占用", node.Name, owner)
		}
	}

	// 先清理本扩展旧菜单（覆盖安装/升级时）
	_, _ = db.Exec(ctx, "DELETE FROM "+menuTable+" WHERE remark = $1", "addon:"+addonName)

	// 递归写入 admin 端菜单
	for _, node := range menus.Admin {
		if err := insertMenuTree(ctx, db, addonName, node, 0); err != nil {
			return err
		}
	}

	// 递归写入 tenant 端菜单（如有）
	for _, node := range menus.Tenant {
		if err := insertMenuTree(ctx, db, addonName, node, 0); err != nil {
			return err
		}
	}

	return nil
}

// uninstallMenus 卸载扩展菜单：按 remark 标记批量删除
func uninstallMenus(ctx context.Context, db gdb.DB, addonName string) (int, error) {
	result, err := db.Exec(ctx, "DELETE FROM "+menuTable+" WHERE remark = $1", "addon:"+addonName)
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

// insertMenuTree 递归插入菜单树
func insertMenuTree(ctx context.Context, db gdb.DB, addonName string, node MenuNode, parentId int64) error {
	now := time.Now().Unix()
	dbType := detectDBType(ctx)

	var insertedId int64

	if dbType == "pgsql" {
		result, err := db.Exec(ctx, `
			INSERT INTO `+menuTable+` (parent_id, type, title, name, path, component, icon, hidden, keep_alive, sort, status, perms, redirect, remark, create_time, update_time)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1, $11, $12, $13, $14, $15)
		`, parentId, node.Type, node.Title, node.Name, node.Path, node.Component, node.Icon,
			node.Hidden, node.KeepAlive, node.Sort, node.Perms, node.Redirect,
			"addon:"+addonName, now, now)
		if err != nil {
			return fmt.Errorf("插入菜单 '%s' 失败: %v", node.Name, err)
		}
		insertedId, _ = result.LastInsertId()
		if insertedId == 0 {
			// PgSQL 的 LastInsertId 不一定可靠，回查
			val, err := db.GetValue(ctx, "SELECT id FROM "+menuTable+" WHERE name = $1 AND remark = $2", node.Name, "addon:"+addonName)
			if err == nil && val != nil {
				insertedId = val.Int64()
			}
		}
	} else {
		result, err := db.Exec(ctx, `
			INSERT INTO `+menuTable+` (parent_id, type, title, name, path, component, icon, hidden, keep_alive, sort, status, perms, redirect, remark, create_time, update_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?)
		`, parentId, node.Type, node.Title, node.Name, node.Path, node.Component, node.Icon,
			node.Hidden, node.KeepAlive, node.Sort, node.Perms, node.Redirect,
			"addon:"+addonName, now, now)
		if err != nil {
			return fmt.Errorf("插入菜单 '%s' 失败: %v", node.Name, err)
		}
		insertedId, _ = result.LastInsertId()
	}

	// 递归写入子菜单
	for _, child := range node.Children {
		if err := insertMenuTree(ctx, db, addonName, child, insertedId); err != nil {
			return err
		}
	}

	return nil
}

// getExistingMenuNames 获取已存在的菜单 name → 所属者映射
// 本扩展自身的菜单不算冲突（支持覆盖安装）
func getExistingMenuNames(ctx context.Context, db gdb.DB, addonName string) (map[string]string, error) {
	result := make(map[string]string)
	rows, err := db.GetAll(ctx, "SELECT name, remark FROM "+menuTable+" WHERE name IS NOT NULL AND name != ''")
	if err != nil {
		return nil, err
	}
	ownRemark := "addon:" + addonName
	for _, row := range rows {
		name := row["name"].String()
		remark := row["remark"].String()
		if remark == ownRemark {
			continue // 本扩展的旧菜单不算冲突
		}
		if name != "" {
			owner := "system"
			if strings.HasPrefix(remark, "addon:") {
				owner = remark
			}
			result[name] = owner
		}
	}
	return result, nil
}

// collectAllNodes 递归收集所有菜单节点（用于预校验）
func collectAllNodes(nodes []MenuNode) []MenuNode {
	var all []MenuNode
	for _, n := range nodes {
		all = append(all, n)
		if len(n.Children) > 0 {
			all = append(all, collectAllNodes(n.Children)...)
		}
	}
	return all
}
