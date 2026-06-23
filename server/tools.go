//go:build ignore

// +----------------------------------------------------------------------
// | XYGo Admin 工具命令入口
// +----------------------------------------------------------------------
// | 用途：提供交互式工具菜单，支持数据库迁移、模板检查等开发辅助命令
// |
// | 使用方式：
// |   go run tools.go                 -- 交互式选择命令
// |   go run tools.go migrate up      -- 直接执行迁移
// |   go run tools.go migrate status  -- 查看迁移状态
// |   go run tools.go migrate history -- 查看迁移历史
// |   go run tools.go check-tpl       -- 检查模板语法
// |   go run tools.go update          -- 在线更新
// +----------------------------------------------------------------------

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"

	"xygo/internal/cmdtools/addon"
	"xygo/internal/cmdtools/checktpl"
	"xygo/internal/cmdtools/gencli"
	"xygo/internal/cmdtools/migrate"
	"xygo/internal/cmdtools/updater"
)

func main() {
	ctx := gctx.New()

	// 如果有命令行参数，直接执行对应命令
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		// gen 命令参数较多，整体交给 gencli 解析
		if cmd == "gen" {
			if err := gencli.Run(ctx, os.Args[2:]); err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
			return
		}
		sub := ""
		if len(os.Args) > 2 {
			sub = os.Args[2]
		}
		runCommand(cmd, sub)
		return
	}

	// 无参数，进入交互式菜单
	for {
		fmt.Println()
		fmt.Println("  ╔════════════════════════════════════════════════════════════╗")
		fmt.Println("  ║                                                            ║")
		fmt.Println("  ║              XYGo Admin Tools v1.3.1                       ║")
		fmt.Println("  ║        Vue3 + GoFrame 企业级中后台管理系统                 ║")
		fmt.Println("  ║                                                            ║")
		fmt.Println("  ╠════════════════════════════════════════════════════════════╣")
		fmt.Println("  ║                                                            ║")
		fmt.Println("  ║  官网:   https://www.xygoadmin.com                         ║")
		fmt.Println("  ║  Gitee:  https://gitee.com/a751300685a/xygo-admin          ║")
		fmt.Println("  ║  GitHub: https://github.com/z312193608/xygo-admin          ║")
		fmt.Println("  ║                                                            ║")
		fmt.Println("  ╠════════════════════════════════════════════════════════════╣")
		fmt.Println("  ║                                                            ║")
		fmt.Println("  ║  [1] migrate up        执行数据库迁移                      ║")
		fmt.Println("  ║  [2] migrate status    查看迁移状态                        ║")
		fmt.Println("  ║  [3] migrate history   查看迁移历史                        ║")
		fmt.Println("  ║  [4] check-tpl         检查模板语法                        ║")
		fmt.Println("  ║  [5] addon install     安装扩展                            ║")
		fmt.Println("  ║  [6] addon uninstall   卸载扩展                            ║")
		fmt.Println("  ║  [7] addon create      创建扩展骨架                        ║")
		fmt.Println("  ║  [8] addon pack        打包扩展为ZIP                       ║")
		fmt.Println("  ║  [9] update            在线更新                            ║")
		fmt.Println("  ║  [10] gen              代码生成（命令行 CRUD）             ║")
		fmt.Println("  ║  [0] exit              退出                                ║")
		fmt.Println("  ║                                                            ║")
		fmt.Println("  ╚════════════════════════════════════════════════════════════╝")
		fmt.Println()

		choice := gcmd.Scan("  请选择命令 [0-10]: ")
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			_ = migrate.RunUp(ctx)
		case "2":
			_ = migrate.RunStatus(ctx)
		case "3":
			_ = migrate.RunHistory(ctx)
		case "4":
			_ = checktpl.Run(ctx)
		case "5":
			if err := addon.Install(ctx, ""); err != nil {
				fmt.Printf("  错误: %v\n", err)
			} else {
				return
			}
		case "6":
			if err := addon.Uninstall(ctx, ""); err != nil {
				fmt.Printf("  错误: %v\n", err)
			} else {
				return
			}
		case "7":
			if err := addon.Create(ctx, ""); err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
		case "8":
			if err := addon.Pack(ctx, ""); err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
		case "9":
			applied, _ := updater.RunUpdate(ctx)
			if applied {
				return
			}
		case "10", "gen":
			runGenInteractive(ctx)
		case "0", "exit", "quit", "q":
			fmt.Println("  Bye!")
			return
		default:
			fmt.Println("  无效选项，请重新选择")
		}
	}
}

func runCommand(cmd, sub string) {
	ctx := gctx.New()
	switch cmd {
	case "migrate":
		switch sub {
		case "up":
			_ = migrate.RunUp(ctx)
		case "status":
			_ = migrate.RunStatus(ctx)
		case "history":
			_ = migrate.RunHistory(ctx)
		default:
			fmt.Printf("  未知的 migrate 子命令: %s\n", sub)
			fmt.Println("  可用: up / status / history")
		}
	case "check-tpl":
		_ = checktpl.Run(ctx)
	case "addon":
		switch sub {
		case "install":
			addonName := ""
			if len(os.Args) > 3 {
				addonName = os.Args[3]
			}
			if err := addon.Install(ctx, addonName); err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
		case "uninstall":
			addonName := ""
			if len(os.Args) > 3 {
				addonName = os.Args[3]
			}
			if err := addon.Uninstall(ctx, addonName); err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
		case "create":
			addonName := ""
			if len(os.Args) > 3 {
				addonName = os.Args[3]
			}
			if err := addon.Create(ctx, addonName); err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
		case "pack":
			addonName := ""
			if len(os.Args) > 3 {
				addonName = os.Args[3]
			}
			if err := addon.Pack(ctx, addonName); err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
		default:
			fmt.Printf("  未知的 addon 子命令: %s\n", sub)
			fmt.Println("  可用: install / uninstall / create / pack")
		}
	case "update":
		_, _ = updater.RunUpdate(ctx)
	case "gen":
		if err := gencli.Run(ctx, os.Args[2:]); err != nil {
			fmt.Printf("  错误: %v\n", err)
		}
	default:
		fmt.Printf("  未知命令: %s\n", cmd)
		fmt.Println("  可用: migrate / check-tpl / addon / update / gen")
	}
}

// runGenInteractive 交互式代码生成：先列出可生成的表，再输入表名与基本选项。
func runGenInteractive(ctx context.Context) {
	// 先列出可生成的表
	if err := gencli.Run(ctx, []string{"--list-tables"}); err != nil {
		fmt.Printf("  错误: %v\n", err)
		return
	}

	table := strings.TrimSpace(gcmd.Scan("  请输入要生成的表名(留空取消): "))
	if table == "" {
		fmt.Println("  已取消。")
		return
	}

	args := []string{table}

	genType := strings.TrimSpace(gcmd.Scan("  生成类型 [10=普通列表(默认) 11=树表]: "))
	if genType == "11" {
		args = append(args, "--type", "11")
	}

	varName := strings.TrimSpace(gcmd.Scan("  实体名(PascalCase，留空自动推导): "))
	if varName != "" {
		args = append(args, "--var", varName)
	}

	menuPid := strings.TrimSpace(gcmd.Scan("  挂载父菜单ID(留空=新建顶级目录+页面): "))
	if menuPid != "" {
		args = append(args, "--menu-pid", menuPid)
	}

	if err := gencli.Run(ctx, args); err != nil {
		fmt.Printf("  错误: %v\n", err)
	}
}
