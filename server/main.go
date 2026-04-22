// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package main

import (
	_ "xygo/internal/packed"

	// 数据库/缓存驱动（按需启用）
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	// ✅ 统一导入logic（触发所有Service注册）
	_ "xygo/internal/logic"

	// ✅ 统一导入field（触发所有资源注册）
	_ "xygo/internal/field"

	// ✅ 统一导入crons（触发所有定时任务注册）
	_ "xygo/internal/crons"

	// ✅ 统一导入queues（触发所有消费者注册）
	_ "xygo/internal/queues"

	// ✅ 统一导入addons（触发所有扩展模块注册）
	_ "xygo/addons"

	"github.com/gogf/gf/v2/os/gctx"

	"xygo/internal/cmd"
	"xygo/internal/global"
)

func main() {
	ctx := gctx.GetInitCtx()
	global.Init(ctx)
	cmd.Main.Run(ctx)
}
