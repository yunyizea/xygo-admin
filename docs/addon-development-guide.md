# XYGo Admin 扩展开发指南

> 本文档面向扩展（Addon）开发者，介绍如何基于 XYGo Admin 框架开发、安装、升级和分发扩展包。

---

## 一、概述

XYGo Admin 采用**物理隔离**的扩展架构，扩展代码与系统核心代码完全分离：

- 后端代码存放在 `server/addons/{扩展名}/`
- 前端代码存放在 `web/src/addons/{扩展名}/`
- 系统升级永远不会覆盖扩展目录

扩展通过 Go 的 `init()` 机制自动注册，无需修改任何核心代码即可挂载路由、注册事件。

---

## 二、快速开始

### 2.1 使用脚手架创建扩展

```bash
cd server
go run tools.go addon create
```

按提示输入：

| 提示 | 说明 | 示例 |
|------|------|------|
| 扩展标识 | 英文小写，全局唯一 | `shop` |
| 扩展名称 | 中文展示名 | `商城管理` |
| 作者 | 开发者名称 | `张三` |
| 描述 | 一句话描述 | `在线商城功能` |
| 示例表名 | 留空跳过，填写则生成完整 CRUD | `shop_order` |

### 2.2 其他 CLI 命令

```bash
go run tools.go addon install <name>     # 安装扩展（从 ZIP）
go run tools.go addon uninstall <name>   # 卸载扩展
go run tools.go addon pack <name>        # 打包扩展为 ZIP
go run tools.go addon create <name>      # 创建扩展骨架
```

---

## 三、目录结构

### 3.1 后端目录

```
server/addons/{name}/
├── addon.yaml                  # 扩展元信息（必须）
├── module.go                   # 入口文件：路由注册（必须）
├── api/                        # API 定义（请求/响应结构体）
│   └── {name}_xxx.go
├── controller/                 # 控制器
│   └── xxx.go
├── logic/                      # 业务逻辑
│   └── xxx.go
├── model/                      # 数据模型
│   └── xxx.go
├── queues/                     # 消息队列消费者（按需创建）
│   └── xxx.go
├── crons/                      # 定时任务（按需创建）
│   └── xxx.go
├── install/                    # 安装 SQL
│   ├── pgsql.sql
│   └── mysql.sql
├── uninstall/                  # 卸载 SQL
│   ├── pgsql.sql
│   └── mysql.sql
└── upgrade/                    # 升级 SQL（幂等写法）
    ├── pgsql.sql
    └── mysql.sql
```

> **目录组织方式**：`queues/` 和 `crons/` 是独立的 Go 子包，需要在 `module.go` 中空导入来触发 `init()` 注册。这与主包的 `internal/queues/`、`internal/crons/` 模式完全一致。如果扩展功能简单（只有一两个消费者/任务），也可以不建子目录，直接在扩展根包下写 `queues.go`、`crons.go`。

### 3.2 前端目录

```
web/src/addons/{name}/
├── api/                        # 前端 API 接口
│   └── xxx.ts
└── views/                      # 页面组件
    └── xxx/
        ├── index.vue
        └── modules/
            └── xxx-dialog.vue
```

---

## 四、addon.yaml 配置说明

`addon.yaml` 是扩展的元信息文件，安装器和打包器都会读取它。

```yaml
# 基础信息
name: shop                        # 扩展标识（英文小写，与目录名一致）
version: "1.0.0"                  # 当前版本号（语义化版本）
title: "商城管理"                  # 显示名称
description: "在线商城功能"        # 描述
author: "开发者A"                  # 作者

# 版本控制
min_version: "1.3.0"              # 要求的最低系统版本
min_upgrade_from: ""              # 最低可升级版本（留空不限制）

# 更新日志（升级时展示给用户）
changelog:
  - "新增订单管理"
  - "修复支付回调问题"

# 功能声明（仅用于说明，不影响逻辑）
features:
  routes: true
  websocket: false
  queue: false
  cron: false

# 菜单声明（安装时自动写入数据库）
menus:
  admin:                          # 后台管理菜单
    - title: "商城管理"
      name: "ShopManage"          # 必须以扩展名的 PascalCase 开头
      path: "shop"
      icon: "ri:shopping-cart-line"
      type: 1                     # 1=目录, 2=菜单, 3=按钮
      sort: 60
      children:
        - title: "订单管理"
          name: "ShopOrderList"
          path: "order"
          component: "@addons/shop/views/order"   # 前端组件路径
          type: 2
          sort: 1
          children:
            - title: "新增/编辑"
              name: "ShopOrderEdit"
              type: 3
              perms: "/admin/shop/order/edit"      # 权限标识
              sort: 1
            - title: "删除"
              name: "ShopOrderDelete"
              type: 3
              perms: "/admin/shop/order/delete"
              sort: 2
  tenant: []                      # 租户端菜单（结构相同）
```

### 菜单命名规范

- 所有菜单的 `name` 字段**必须**以扩展名的 PascalCase 为前缀
- 例如扩展名为 `shop`，菜单 name 必须以 `Shop` 开头：`ShopManage`、`ShopOrderList` 等
- 安装器会自动检测命名冲突，防止与系统菜单或其他扩展重名
- 卸载时按 `remark='addon:{name}'` 标记自动清除所有菜单

### 前端组件路径

菜单中的 `component` 使用 `@addons/` 前缀，系统会自动映射到 `web/src/addons/` 目录：

```
@addons/shop/views/order → web/src/addons/shop/views/order/index.vue
```

---

## 五、后端开发

### 5.1 入口文件 module.go

`module.go` 是扩展的入口，在 `init()` 中完成路由注册和系统集成：

```go
package shop

import (
    "xygo/internal/addon"
    "xygo/internal/middleware"
    "xygo/addons/shop/controller"

    // 空导入：触发 queues、crons 子包的 init() 注册
    _ "xygo/addons/shop/queues"
    _ "xygo/addons/shop/crons"

    "github.com/gogf/gf/v2/net/ghttp"
)

func init() {
    // 注册扩展模块（必须）
    addon.Register(addon.Module{
        Name: "shop",
        Mount: func(s *ghttp.Server) {
            s.Group("/", func(group *ghttp.RouterGroup) {
                group.Middleware(
                    middleware.CORS,
                    middleware.ResponseHandler,
                    middleware.AdminAuth,
                )
                group.Bind(controller.NewV1())
            })
        },
    })

    // WebSocket 事件直接在这里注册（不需要子包）
    // websocket.RegisterEvent("shop.orderNotify", handleOrderNotify)
}
```

> **注意**：`queues/` 和 `crons/` 是独立子包，必须通过空导入 `_ "xygo/addons/shop/queues"` 来触发它们的 `init()`。WebSocket 事件因为只是一行注册调用，通常直接写在 `module.go` 里即可。

系统启动时，`addon.MountAll(s)` 会自动调用所有已注册扩展的 `Mount` 函数挂载路由。

### 5.2 API 定义

在 `api/` 目录下定义请求和响应结构体，使用 GoFrame 的规范路由标签：

```go
package api

import "github.com/gogf/gf/v2/frame/g"

// 列表
type ShopOrderListReq struct {
    g.Meta   `path:"/admin/shop/order/list" method:"get" tags:"商城" summary:"订单列表"`
    Page     int    `json:"page" d:"1"`
    PageSize int    `json:"pageSize" d:"20"`
    Status   *int   `json:"status"`
}
type ShopOrderListRes struct {
    g.Meta `mime:"application/json"`
}

// 保存
type ShopOrderEditReq struct {
    g.Meta `path:"/admin/shop/order/edit" method:"post" tags:"商城" summary:"保存订单"`
    Id     uint64 `json:"id"`
    // 业务字段...
}
type ShopOrderEditRes struct{}

// 删除
type ShopOrderDeleteReq struct {
    g.Meta `path:"/admin/shop/order/delete" method:"post" tags:"商城" summary:"删除订单"`
    Id     uint64 `json:"id" v:"required"`
}
type ShopOrderDeleteRes struct{}
```

**路由路径规范**：`/admin/{扩展名}/{实体名}/{操作}`，如 `/admin/shop/order/list`。

### 5.3 Controller

```go
package controller

import (
    "context"
    api "xygo/addons/shop/api"
    "xygo/addons/shop/logic"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }

func (c *ControllerV1) ShopOrderList(ctx context.Context, req *api.ShopOrderListReq) (res *api.ShopOrderListRes, err error) {
    return logic.OrderList(ctx, req)
}

func (c *ControllerV1) ShopOrderEdit(ctx context.Context, req *api.ShopOrderEditReq) (res *api.ShopOrderEditRes, err error) {
    return logic.OrderEdit(ctx, req)
}

func (c *ControllerV1) ShopOrderDelete(ctx context.Context, req *api.ShopOrderDeleteReq) (res *api.ShopOrderDeleteRes, err error) {
    return logic.OrderDelete(ctx, req)
}
```

### 5.4 Logic

```go
package logic

import (
    "context"
    api "xygo/addons/shop/api"
)

func OrderList(ctx context.Context, req *api.ShopOrderListReq) (res *api.ShopOrderListRes, err error) {
    // 使用 g.DB() 或 dao 层查询数据库
    // 表名规范：xy_{扩展名}_xxx，如 xy_shop_order
    res = &api.ShopOrderListRes{}
    return
}

func OrderEdit(ctx context.Context, req *api.ShopOrderEditReq) (res *api.ShopOrderEditRes, err error) {
    // 实现新增/编辑逻辑
    return
}

func OrderDelete(ctx context.Context, req *api.ShopOrderDeleteReq) (res *api.ShopOrderDeleteRes, err error) {
    // 实现删除逻辑
    return
}
```

---

## 六、前端开发

### 6.1 API 接口

在 `web/src/addons/{name}/api/` 下创建 TypeScript 接口文件：

```typescript
import { adminRequest } from '@/utils/http'

const prefix = '/admin/shop/order'

/** 列表 */
export function fetchShopOrderList(params: any) {
  return adminRequest.get<Record<string, any>>({
    url: `${prefix}/list`,
    params
  })
}

/** 保存 */
export function fetchShopOrderEdit(params: any) {
  return adminRequest.post<any>({
    url: `${prefix}/edit`,
    params
  })
}

/** 删除 */
export function fetchShopOrderDelete(id: number) {
  return adminRequest.post<any>({
    url: `${prefix}/delete`,
    params: { id }
  })
}
```

### 6.2 页面组件

页面放在 `web/src/addons/{name}/views/` 下，使用框架提供的 `ArtTableHeader` 等公共组件：

```vue
<!-- web/src/addons/shop/views/order/index.vue -->
<template>
  <div class="shop-order-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton type="primary" @click="openDialog()">新增</ElButton>
        </template>
      </ArtTableHeader>
      <ArtTable :data="tableData" :columns="columns" :loading="loading" />
      <ArtPagination :total="total" v-model:page="page" @change="loadData" />
    </ElCard>
    <OrderDialog ref="dialogRef" @success="loadData" />
  </div>
</template>
```

### 6.3 组件路径映射

菜单中配置的 `component: "@addons/shop/views/order"` 会被 `ComponentLoader` 自动解析为：

```
web/src/addons/shop/views/order/index.vue
```

支持两种路径格式：
- `@addons/shop/views/order` → `order/index.vue`
- `@addons/shop/views/order.vue` → `order.vue`（单文件）

---

## 七、WebSocket 事件注册

如果你的扩展需要实时推送功能，可以注册自定义 WebSocket 事件。

### 7.1 注册方式

在 `module.go` 的 `init()` 中注册：

```go
package shop

import (
    "xygo/internal/addon"
    "xygo/internal/websocket"
    // ...
)

func init() {
    addon.Register(addon.Module{
        Name: "shop",
        Mount: func(s *ghttp.Server) {
            // 路由注册...
        },
    })

    // 注册 WebSocket 事件
    websocket.RegisterEvent("shop.orderNotify", handleOrderNotify)
    websocket.RegisterEvent("shop.stockAlert", handleStockAlert)
}
```

### 7.2 事件处理函数

```go
func handleOrderNotify(client *websocket.Client, req *websocket.WsRequest) {
    // req.Event = "shop.orderNotify"
    // req.Data  = map[string]interface{}{...} 前端发送的数据

    // 处理业务逻辑...

    // 回复当前客户端
    client.SendMsg(websocket.NewResponse("shop.orderNotify", map[string]interface{}{
        "orderId": 12345,
        "status":  "paid",
    }))
}

func handleStockAlert(client *websocket.Client, req *websocket.WsRequest) {
    // 可以广播给所有在线用户
    websocket.Manager.Broadcast(websocket.NewResponse("shop.stockAlert", map[string]interface{}{
        "message": "库存不足",
    }))
}
```

### 7.3 核心 API

| 函数 | 说明 |
|------|------|
| `websocket.RegisterEvent(event, handler)` | 注册事件处理器 |
| `client.SendMsg(response)` | 向当前客户端发消息 |
| `websocket.Manager.Broadcast(response)` | 广播给所有在线客户端 |
| `websocket.SendToUser(userType, userId, response)` | 向指定用户发消息 |
| `websocket.NewResponse(event, data)` | 构造成功响应 |
| `websocket.NewErrorResponse(event, code, msg)` | 构造错误响应 |

### 7.4 事件命名规范

事件名建议使用 `{扩展名}.{事件名}` 格式，避免与其他扩展冲突：

```
shop.orderNotify       ✓ 正确
shop.stockAlert        ✓ 正确
orderNotify            ✗ 避免（可能与其他扩展冲突）
```

---

## 八、消息队列集成

如果你的扩展需要异步处理任务（如发送邮件、生成报表），可以使用消息队列。

**扩展与主包的关系**：扩展的消息队列使用的是同一套基础设施 `xygo/internal/library/queue`，注册方式、Consumer 接口、投递 API 完全相同。唯一区别是代码存放位置：

| 对比项 | 主包 | 扩展 |
|-------|------|------|
| 消费者位置 | `server/internal/queues/*.go` | `server/addons/{name}/queues/*.go` |
| 包名 | `package queues` | `package queues` |
| 触发方式 | `main.go` 空导入 `_ "xygo/internal/queues"` | `module.go` 空导入 `_ "xygo/addons/{name}/queues"` |
| Topic 命名 | `login_log`、`operation_log` | 加扩展前缀：`shop.order.payment` |
| 常量定义 | 每个文件内 `const TopicXxx` | 同样，每个文件内 `const TopicXxx` |

> **原则**：扩展不修改主包的任何消费者，只通过同一套 API 注册自己的消费者。所有消费者在同一个进程内运行，共享同一个队列驱动。

### 8.1 定义消费者

在 `queues/` 子目录下创建消费者文件，每个消费者一个文件，与主包 `internal/queues/` 的模式完全一致：

```go
// server/addons/shop/queues/order_payment.go
package queues

import (
    "context"
    "encoding/json"

    "github.com/gogf/gf/v2/frame/g"
    "xygo/internal/library/queue"
)

// Topic 常量（供生产者引用）
const TopicOrderPayment = "shop.order.payment"

func init() {
    queue.Register(&OrderPaymentConsumer{})
}

// OrderPaymentConsumer 订单支付异步处理
type OrderPaymentConsumer struct{}

func (c *OrderPaymentConsumer) GetTopic() string {
    return TopicOrderPayment
}

func (c *OrderPaymentConsumer) Handle(ctx context.Context, msg *queue.Message) error {
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(msg.Body), &data); err != nil {
        g.Log().Errorf(ctx, "[queue:%s] unmarshal failed: %v", TopicOrderPayment, err)
        return nil // 格式错误不重试
    }

    // 处理业务逻辑...

    // 如果需要重试，返回 RetryError
    // return queue.NewRetryError("支付确认失败，稍后重试")

    return nil
}
```

### 8.2 生产消息

在 controller 或 logic 中投递消息，引用消费者定义的 Topic 常量：

```go
import (
    "xygo/internal/library/queue"
    shopQueues "xygo/addons/shop/queues"  // 引用 Topic 常量
)

// 投递即时消息
err := queue.Push(shopQueues.TopicOrderPayment, []byte(`{"orderId": 12345}`))

// 投递延迟消息（如果驱动支持）
err := queue.PushDelay("shop.order.timeout", []byte(`{"orderId": 12345}`), 30*60) // 30分钟后
```

### 8.3 Consumer 接口

```go
type Consumer interface {
    GetTopic() string                                      // 消息主题
    Handle(ctx context.Context, msg *Message) error        // 处理函数
}
```

### 8.4 Topic 命名规范

```
shop.order.payment     ✓ 正确（扩展名.模块.动作）
shop.stock.sync        ✓ 正确
order.payment          ✗ 避免（缺少扩展名前缀）
```

---

## 九、定时任务集成

扩展可以注册定时任务，调度周期在后台管理界面配置。

**扩展与主包的关系**：与消息队列一样，使用同一套 `xygo/internal/library/cron` 基础设施：

| 对比项 | 主包 | 扩展 |
|-------|------|------|
| 任务位置 | `server/internal/crons/*.go` | `server/addons/{name}/crons/*.go` |
| 包名 | `package crons` | `package crons` |
| 触发方式 | `main.go` 空导入 `_ "xygo/internal/crons"` | `module.go` 空导入 `_ "xygo/addons/{name}/crons"` |
| 任务命名 | `queue_alert`、`log_clean` | 加扩展前缀：`shop.order_timeout` |
| cron 别名 | `cronlib "xygo/internal/library/cron"` | 同样用 `cronlib` 别名避免包名冲突 |

### 9.1 注册任务

在 `crons/` 子目录下创建任务文件，每个任务一个文件：

```go
// server/addons/shop/crons/order_timeout.go
package crons

import (
    "context"
    "fmt"

    cronlib "xygo/internal/library/cron"
)

func init() {
    cronlib.Register(&OrderTimeoutTask{})
}

// OrderTimeoutTask 订单超时自动取消
type OrderTimeoutTask struct{}

func (t *OrderTimeoutTask) GetName() string {
    return "shop.order_timeout"
}

func (t *OrderTimeoutTask) Execute(ctx context.Context, params []string) (string, error) {
    // params 是后台配置的运行参数（可选）

    // 业务逻辑：查找超时订单并取消...
    cancelCount := 5

    // 返回执行结果（会记录到日志）
    return fmt.Sprintf("取消了 %d 个超时订单", cancelCount), nil
}
```

> **注意**：因为子目录包名是 `crons`，与系统库 `cron` 同名，需要用别名 `cronlib` 导入，与主包 `internal/crons/` 的写法一致。

```go
type Task interface {
    GetName() string                                              // 任务唯一标识
    Execute(ctx context.Context, params []string) (string, error) // 执行任务
}
```

### 9.4 工作流程

1. **代码注册**：扩展在 `init()` 中调用 `cron.Register()` 将任务放入内存注册表
2. **后台配置**：管理员在 系统管理 → 定时任务 中添加任务，配置 cron 表达式和参数
3. **启动调度**：系统启动时从数据库读取已启用的任务，匹配注册表后启动调度

> 注意：只在代码中 `Register` 不会自动运行任务，还需要在后台管理界面中添加并启用。

### 9.5 Cron 表达式

采用标准 6 位 cron 格式（秒 分 时 日 月 周）：

```
*/30 * * * * *     每30秒
0 */5 * * * *      每5分钟
0 0 2 * * *        每天凌晨2点
0 0 0 1 * *        每月1号0点
```

### 9.6 任务名规范

```
shop.order_timeout     ✓ 正确（扩展名.任务描述）
shop.daily_report      ✓ 正确
order_timeout          ✗ 避免（缺少扩展名前缀）
```

---

## 十、自定义认证端点（Auth Endpoint Factory）

如果你的扩展需要独立的用户认证体系（如供应商端、代理商端），可以使用 Auth Endpoint Factory 创建完全独立的 JWT 认证端点，无需修改核心代码。

### 10.1 创建认证端点

```go
package shop

import (
    "xygo/internal/addon"
    "xygo/internal/library/token"
    "xygo/addons/shop/controller"

    "github.com/gogf/gf/v2/net/ghttp"
)

// 创建供应商认证端点
var supplierAuth = token.NewEndpoint(token.EndpointConfig{
    Name:       "supplier",                        // 端点标识
    Secret:     "",                                 // 留空则复用系统密钥
    Expires:    86400,                              // Token 有效期（秒），0 复用系统默认
    MultiLogin: true,                               // 是否允许多端登录
    HeaderKey:  "Authorization",                    // 请求头名称
    HeaderType: "Bearer",                           // 请求头前缀
    LoginPaths: []string{"/supplier/auth/login"},   // 免鉴权路径
})

func init() {
    addon.Register(addon.Module{
        Name: "shop",
        Mount: func(s *ghttp.Server) {
            // 后台管理路由（使用系统 AdminAuth）
            s.Group("/", func(group *ghttp.RouterGroup) {
                group.Middleware(middleware.CORS, middleware.ResponseHandler, middleware.AdminAuth)
                group.Bind(controller.NewAdminV1())
            })

            // 供应商端路由（使用自定义认证）
            s.Group("/supplier", func(group *ghttp.RouterGroup) {
                group.Middleware(middleware.CORS, middleware.ResponseHandler)
                group.Middleware(supplierAuth.Middleware())  // 使用端点中间件
                group.Bind(controller.NewSupplierV1())
            })
        },
    })
}
```

### 10.2 登录接口实现

```go
func SupplierLogin(ctx context.Context, username, password string) (tokenStr string, expiresIn int64, err error) {
    // 1. 验证用户名密码...
    supplierId := uint64(1001)

    // 2. 生成 Token
    tokenStr, expiresIn, err = supplierAuth.Generate(ctx, supplierId, map[string]any{
        "name":  "供应商A",
        "level": 2,
    })
    return
}
```

### 10.3 在业务中获取当前用户

```go
import "xygo/internal/library/contexts"

func SomeLogic(ctx context.Context) {
    // 获取当前认证的供应商用户
    user := contexts.GetEndpointUser(ctx, "supplier")
    if user != nil {
        fmt.Println("供应商ID:", user.Id)
        fmt.Println("供应商名:", user.Data["name"])
    }

    // 快捷获取 ID
    userId := contexts.GetEndpointUserId(ctx, "supplier")
}
```

### 10.4 Endpoint 完整 API

| 方法 | 签名 | 说明 |
|------|------|------|
| `NewEndpoint` | `func NewEndpoint(cfg EndpointConfig) *Endpoint` | 创建端点实例 |
| `Generate` | `func (e *Endpoint) Generate(ctx, userId, data) (token, expiresIn, err)` | 生成 Token |
| `Parse` | `func (e *Endpoint) Parse(ctx, token) (userId, data, err)` | 解析 Token |
| `Delete` | `func (e *Endpoint) Delete(ctx, token) error` | 删除 Token（登出） |
| `KickByUserId` | `func (e *Endpoint) KickByUserId(ctx, userId) error` | 踢下线 |
| `Middleware` | `func (e *Endpoint) Middleware() func(r *ghttp.Request)` | 生成路由中间件 |

### 10.5 EndpointConfig 字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Name` | string | （必填） | 端点唯一标识 |
| `Secret` | string | 系统密钥 | JWT 签名密钥 |
| `Expires` | int64 | 系统默认 | Token 有效期（秒） |
| `MultiLogin` | bool | false | 是否允许多端登录 |
| `HeaderKey` | string | Authorization | 请求头名称 |
| `HeaderType` | string | Bearer | 请求头前缀 |
| `LoginPaths` | []string | [] | 免鉴权路径列表 |

---

## 十一、数据库规范

### 11.1 表命名

所有扩展数据表必须以 `xy_{扩展名}_` 为前缀：

```
xy_shop_order          ✓ 正确
xy_shop_product        ✓ 正确
xy_order               ✗ 错误（缺少扩展名前缀）
shop_order             ✗ 错误（缺少 xy_ 前缀）
```

### 11.2 安装 SQL

`install/pgsql.sql` 和 `install/mysql.sql` 在**全新安装**时执行，建议使用幂等写法：

```sql
-- PostgreSQL
CREATE TABLE IF NOT EXISTS xy_shop_order (
    id         bigserial PRIMARY KEY,
    order_no   varchar(64) NOT NULL DEFAULT '',
    status     smallint    NOT NULL DEFAULT 1,
    amount     decimal(10,2) NOT NULL DEFAULT 0,
    created_at bigint      NOT NULL DEFAULT 0,
    updated_at bigint      NOT NULL DEFAULT 0,
    deleted_at bigint      NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_shop_order_status ON xy_shop_order(status);
CREATE INDEX IF NOT EXISTS idx_shop_order_deleted_at ON xy_shop_order(deleted_at);
```

### 11.3 升级 SQL

`upgrade/pgsql.sql` 在**版本升级**时执行。必须使用幂等写法，因为同一文件可能被多次执行：

```sql
-- PostgreSQL 幂等写法示例

-- 新增字段
ALTER TABLE xy_shop_order ADD COLUMN IF NOT EXISTS remark varchar(500) DEFAULT '';

-- 新增索引
CREATE INDEX IF NOT EXISTS idx_shop_order_created_at ON xy_shop_order(created_at);

-- 新增表
CREATE TABLE IF NOT EXISTS xy_shop_refund (
    id         bigserial PRIMARY KEY,
    order_id   bigint NOT NULL DEFAULT 0,
    reason     varchar(500) NOT NULL DEFAULT ''
);

-- 新增配置（不覆盖已有值）
INSERT INTO xy_sys_config ("group", "key", "value")
VALUES ('shop', 'payment_timeout', '1800')
ON CONFLICT ("group", "key") DO NOTHING;
```

### 11.4 卸载 SQL

`uninstall/pgsql.sql` 在卸载时执行，通常只做删表操作。菜单由安装器自动清理，无需手写。

```sql
-- 注意：菜单由 installer 自动删除（通过 remark 标记），无需在此手写
DROP TABLE IF EXISTS xy_shop_order;
DROP TABLE IF EXISTS xy_shop_refund;
```

---

## 十二、扩展生命周期

### 12.1 安装流程

```
[1/9] 解压 ZIP 到临时目录
[2/9] 读取 addon.yaml，校验元信息
[3/9] 检查安装状态（全新安装 / 升级 / 覆盖重装 / 降级拒绝）
[4/9] 执行 install SQL（全新安装）或 upgrade SQL（升级）
[5/9] 安装扩展菜单（从 addon.yaml 声明式写入数据库）
[6/9] 备份旧版本文件（仅升级时，备份到 .backup/{name}-{oldVer}/）
[7/9] 复制扩展文件到隔离目录
[8/9] 更新 addons/addons.go（自动重新生成导入声明）
[9/9] 记录安装信息到 xy_addon 表
```

### 12.2 卸载流程

```
[1/6] 确认卸载
[2/6] 执行 uninstall SQL（删表等）
[3/6] 卸载扩展菜单（按 remark='addon:{name}' 批量删除）
[4/6] 删除扩展目录（server/addons/{name}/ + web/src/addons/{name}/）
[5/6] 更新 addons/addons.go
[6/6] 更新 xy_addon 状态为已卸载
```

### 12.3 升级注意事项

- 升级时执行 `upgrade/` 目录下的 SQL，不是 `install/`
- 升级前自动备份旧版文件到 `server/addons/.backup/{name}-{oldVer}/`
- `addon.yaml` 中的 `min_upgrade_from` 可以限制最低可升级版本
- 升级完成后会显示 `changelog` 中的更新日志
- 菜单会先清除再重新写入（支持菜单变更）

### 12.4 打包分发

```bash
go run tools.go addon pack shop
```

生成 `shop-1.0.0.zip`，包含：

```
shop-1.0.0.zip
├── addon.yaml              # 元信息
├── server/                 # 后端代码
├── web/                    # 前端代码
├── install/                # 安装 SQL
├── uninstall/              # 卸载 SQL
└── upgrade/                # 升级 SQL
```

用户拿到 ZIP 后重命名为 `shop.zip`，放入 `server/addons/` 目录，执行 `addon install shop` 即可。

---

## 十三、CRUD 代码生成器

系统内置的代码生成器支持直接生成代码到扩展目录。

1. 进入后台 → 开发工具 → 代码生成
2. 选择数据表后进入第二步"基础配置"
3. 在"生成目标"下拉中选择已安装的扩展
4. 所有路径自动切换为扩展目录
5. 生成的代码直接放入 `server/addons/{name}/` 和 `web/src/addons/{name}/`

---

## 十四、完整示例：从零开发一个扩展

以下是一个完整的扩展开发流程：

```bash
# 1. 创建扩展骨架
cd server
go run tools.go addon create

# 输入：
#   标识: shop
#   名称: 商城管理
#   作者: 你的名字
#   描述: 在线商城
#   示例表: shop_order

# 2. 编辑建表 SQL
# 修改 server/addons/shop/install/pgsql.sql，完善表结构

# 3. 手动执行建表 SQL（开发阶段）
# 在数据库中执行 install/pgsql.sql

# 4. 生成数据模型
gf gen dao

# 5. 编写业务逻辑
# 修改 controller/ 和 logic/ 中的代码

# 6. 重启后端
# Ctrl+C 停止，重新 go run main.go

# 7. 前端开发
# 编辑 web/src/addons/shop/views/ 下的 Vue 组件

# 8. 测试通过后，打包分发
go run tools.go addon pack shop
```

---

## 十五、注意事项

### 不要做的事

- 不要修改 `server/internal/` 下的核心代码
- 不要在扩展中直接操作非 `xy_{扩展名}_` 前缀的数据表
- 不要注册与系统或其他扩展重名的路由、事件、任务
- 不要在 `install SQL` 中手写菜单插入语句（使用 `addon.yaml` 声明）

### 命名规范汇总

| 项目 | 规范 | 示例 |
|------|------|------|
| 扩展标识 | 英文小写 | `shop` |
| 数据表 | `xy_{扩展名}_xxx` | `xy_shop_order` |
| API 路由 | `/admin/{扩展名}/{实体}/{操作}` | `/admin/shop/order/list` |
| 菜单 name | `{PascalCase扩展名}XxxYyy` | `ShopOrderList` |
| WebSocket 事件 | `{扩展名}.{事件名}` | `shop.orderNotify` |
| 队列 Topic | `{扩展名}.{模块}.{动作}` | `shop.order.payment` |
| 定时任务 name | `{扩展名}.{任务描述}` | `shop.order_timeout` |
| Auth 端点 name | 描述性名称 | `supplier` |
