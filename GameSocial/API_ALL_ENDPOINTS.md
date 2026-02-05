# GameSocial 功能接口总览（按模块拆分）

本文档基于《模块化架构设计》与当前后端实际已注册的路由整理，覆盖“已实现 + 规划待实现”的全部 HTTP 接口；已实现的接口在标题后标注 `√`，未实现标注 `×`。

- 快速入口：后台管理端接口见 [API_ADMIN_ENDPOINTS.md](API_ADMIN_ENDPOINTS.md)
- 快速入口：客户端接口见 [API_CLIENT_ENDPOINTS.md](API_CLIENT_ENDPOINTS.md)

## 接口目录

- 后台管理端接口（Admin）：[API_ADMIN_ENDPOINTS.md](API_ADMIN_ENDPOINTS.md)
- 客户端接口（API）：[API_CLIENT_ENDPOINTS.md](API_CLIENT_ENDPOINTS.md)

## 已注册路由清单（与 cmd/server/main.go#L127-L164 一致）

说明：

- 这一表格展示“当前已注册路由”，与代码注册保持一致；“完成”列用于标识该路由的业务实现完成度。
- “详情”列跳转到拆分后的接口文档段落。

| 完成 | 模块 | Method | Path | 详情 |
|---|---|---|---|---|
| √ | 健康检查 | GET | /health | [GET /health](API_CLIENT_ENDPOINTS.md#api-health) |
| √ | Auth（小程序登录） | POST | /api/auth/wechat/login | [POST /api/auth/wechat/login](API_CLIENT_ENDPOINTS.md#api-auth-wechat-login) |
| √ | Item（管理员：积分商品） | POST | /admin/goods | [POST /admin/goods](API_ADMIN_ENDPOINTS.md#api-admin-goods-create) |
| √ | Item（管理员：积分商品） | GET | /admin/goods | [GET /admin/goods](API_ADMIN_ENDPOINTS.md#api-admin-goods-list) |
| √ | Item（管理员：积分商品） | GET | /admin/goods/{id} | [GET /admin/goods/{id}](API_ADMIN_ENDPOINTS.md#api-admin-goods-get) |
| √ | Item（管理员：积分商品） | PUT | /admin/goods/{id} | [PUT /admin/goods/{id}](API_ADMIN_ENDPOINTS.md#api-admin-goods-update) |
| √ | Item（管理员：积分商品） | DELETE | /admin/goods/{id} | [DELETE /admin/goods/{id}](API_ADMIN_ENDPOINTS.md#api-admin-goods-delete) |
| √ | Tournament（管理员：赛事） | POST | /admin/tournaments | [POST /admin/tournaments](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-create) |
| √ | Tournament（管理员：赛事） | GET | /admin/tournaments | [GET /admin/tournaments](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-list) |
| √ | Tournament（管理员：赛事） | GET | /admin/tournaments/{id} | [GET /admin/tournaments/{id}](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-get) |
| √ | Tournament（管理员：赛事） | PUT | /admin/tournaments/{id} | [PUT /admin/tournaments/{id}](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-update) |
| √ | Tournament（管理员：赛事） | DELETE | /admin/tournaments/{id} | [DELETE /admin/tournaments/{id}](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-delete) |
| √ | Task（管理员：任务定义） | POST | /admin/task-defs | [POST /admin/task-defs](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-create) |
| √ | Task（管理员：任务定义） | GET | /admin/task-defs | [GET /admin/task-defs](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-list) |
| √ | Task（管理员：任务定义） | GET | /admin/task-defs/{id} | [GET /admin/task-defs/{id}](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-get) |
| √ | Task（管理员：任务定义） | PUT | /admin/task-defs/{id} | [PUT /admin/task-defs/{id}](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-update) |
| √ | Task（管理员：任务定义） | DELETE | /admin/task-defs/{id} | [DELETE /admin/task-defs/{id}](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-delete) |
| √ | User（管理员：用户） | GET | /admin/users | [GET /admin/users](API_ADMIN_ENDPOINTS.md#api-admin-users-list) |
| √ | User（管理员：用户） | GET | /admin/users/{id} | [GET /admin/users/{id}](API_ADMIN_ENDPOINTS.md#api-admin-users-get) |
| √ | User（管理员：用户） | PUT | /admin/users/{id} | [PUT /admin/users/{id}](API_ADMIN_ENDPOINTS.md#api-admin-users-update) |
| √ | Redeem（管理员：兑换订单） | POST | /admin/redeem/orders | [POST /admin/redeem/orders](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-create) |
| √ | Redeem（管理员：兑换订单） | GET | /admin/redeem/orders | [GET /admin/redeem/orders](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-list) |
| √ | Redeem（管理员：兑换订单） | GET | /admin/redeem/orders/{id} | [GET /admin/redeem/orders/{id}](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-get) |
| √ | Redeem（管理员：兑换订单） | PUT | /admin/redeem/orders/{id}/use | [PUT /admin/redeem/orders/{id}/use](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-use) |
| √ | Redeem（管理员：兑换订单） | PUT | /admin/redeem/orders/{id}/cancel | [PUT /admin/redeem/orders/{id}/cancel](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-cancel) |
| √ | User（小程序：个人资料） | GET | /api/users/me | [GET /api/users/me](API_CLIENT_ENDPOINTS.md#api-users-me-get) |
| √ | User（小程序：个人资料） | PUT | /api/users/me | [PUT /api/users/me](API_CLIENT_ENDPOINTS.md#api-users-me-update) |
| √ | Item（小程序：积分商品） | GET | /api/goods | [GET /api/goods](API_CLIENT_ENDPOINTS.md#api-goods-list) |
| √ | Item（小程序：积分商品） | GET | /api/goods/{id} | [GET /api/goods/{id}](API_CLIENT_ENDPOINTS.md#api-goods-get) |
| √ | Tournament（小程序：赛事） | GET | /api/tournaments | [GET /api/tournaments](API_CLIENT_ENDPOINTS.md#api-tournaments-list) |
| √ | Tournament（小程序：赛事） | GET | /api/tournaments/{id} | [GET /api/tournaments/{id}](API_CLIENT_ENDPOINTS.md#api-tournaments-get) |
| √ | Tournament（小程序：赛事） | POST | /api/tournaments/{id}/join | [POST /api/tournaments/{id}/join](API_CLIENT_ENDPOINTS.md#api-tournaments-join) |
| √ | Tournament（小程序：赛事） | PUT | /api/tournaments/{id}/cancel | [PUT /api/tournaments/{id}/cancel](API_CLIENT_ENDPOINTS.md#api-tournaments-cancel) |
| √ | Tournament（小程序：赛事） | GET | /api/tournaments/{id}/results | [GET /api/tournaments/{id}/results](API_CLIENT_ENDPOINTS.md#api-tournaments-results) |
| √ | Redeem（小程序：兑换订单） | GET | /api/redeem/orders | [GET /api/redeem/orders](API_CLIENT_ENDPOINTS.md#api-redeem-orders-list) |
| √ | Redeem（小程序：兑换订单） | POST | /api/redeem/orders | [POST /api/redeem/orders](API_CLIENT_ENDPOINTS.md#api-redeem-orders-create) |
| √ | Redeem（小程序：兑换订单） | GET | /api/redeem/orders/{id} | [GET /api/redeem/orders/{id}](API_CLIENT_ENDPOINTS.md#api-redeem-orders-get) |
| √ | Redeem（小程序：兑换订单） | PUT | /api/redeem/orders/{id}/cancel | [PUT /api/redeem/orders/{id}/cancel](API_CLIENT_ENDPOINTS.md#api-redeem-orders-cancel) |
| √ | Points（小程序：积分） | GET | /api/points/balance | [GET /api/points/balance](API_CLIENT_ENDPOINTS.md#api-points-balance) |
| √ | Points（小程序：积分） | GET | /api/points/ledgers | [GET /api/points/ledgers](API_CLIENT_ENDPOINTS.md#api-points-ledgers) |
| √ | VIP（小程序：会员） | GET | /api/vip/status | [GET /api/vip/status](API_CLIENT_ENDPOINTS.md#api-vip-status) |
| √ | Task（小程序：任务） | GET | /api/tasks | [GET /api/tasks](API_CLIENT_ENDPOINTS.md#api-tasks-list) |
| × | Task（小程序：任务） | POST | /api/tasks/checkin | [POST /api/tasks/checkin](API_CLIENT_ENDPOINTS.md#api-tasks-checkin) |
| × | Task（小程序：任务） | POST | /api/tasks/{taskCode}/claim | [POST /api/tasks/{taskCode}/claim](API_CLIENT_ENDPOINTS.md#api-tasks-claim) |
| × | Admin（管理员） | POST | /admin/auth/login | [POST /admin/auth/login](API_ADMIN_ENDPOINTS.md#api-admin-auth-login) |
| × | Admin（管理员） | GET | /admin/auth/me | [GET /admin/auth/me](API_ADMIN_ENDPOINTS.md#api-admin-auth-me) |
| × | Admin（管理员） | POST | /admin/auth/logout | [POST /admin/auth/logout](API_ADMIN_ENDPOINTS.md#api-admin-auth-logout) |
| √ | Admin（管理员） | GET | /admin/audit/logs | [GET /admin/audit/logs](API_ADMIN_ENDPOINTS.md#api-admin-audit-logs) |
| × | Admin（管理员） | POST | /admin/points/adjust | [POST /admin/points/adjust](API_ADMIN_ENDPOINTS.md#api-admin-points-adjust) |
| × | Admin（管理员） | PUT | /admin/users/{id}/drinks/use | [PUT /admin/users/{id}/drinks/use](API_ADMIN_ENDPOINTS.md#api-admin-users-drinks-use) |
| × | Admin（管理员） | POST | /admin/tournaments/{id}/results/publish | [POST /admin/tournaments/{id}/results/publish](API_ADMIN_ENDPOINTS.md#api-admin-tournament-results-publish) |
| × | Admin（管理员） | POST | /admin/tournaments/{id}/awards/grant | [POST /admin/tournaments/{id}/awards/grant](API_ADMIN_ENDPOINTS.md#api-admin-tournament-awards-grant) |
| × | Media（管理员） | POST | /admin/media/upload | [POST /admin/media/upload](API_ADMIN_ENDPOINTS.md#api-admin-media-upload) |

## 详细说明

本文件只保留接口目录与“已注册路由清单”。详细的请求/响应/设计思路已拆分到：

- 后台管理端接口：[API_ADMIN_ENDPOINTS.md](API_ADMIN_ENDPOINTS.md)
- 客户端接口：[API_CLIENT_ENDPOINTS.md](API_CLIENT_ENDPOINTS.md)
*** End of File

### 0.1 Base URL

- 本地开发：`http://localhost:<SERVER_PORT>`

### 0.2 Content-Type

- 请求：`Content-Type: application/json`
- 响应：`application/json; charset=utf-8`

### 0.3 时间格式

- `time.Time` 序列化为 RFC3339，例如：`"2026-01-29T12:34:56Z"`

### 0.4 统一响应结构

后端所有接口统一返回：

```json
{
  "code": 200,
  "data": {},
  "message": "ok"
}
```

字段说明：

- `code`：业务码（BizCode）
- `data`：成功时返回的数据
- `message`：失败提示或成功提示

是否成功以 `code` 判断：

- `code=200`：业务成功
- `code!=200`：业务失败（通常为 `201`）

BizCode 枚举（当前实现）：

| code | 含义 | 默认 message |
|---:|---|---|
| 200 | 成功 | ok |
| 201 | 业务未完成（业务失败） | 业务未完成 |
| 401 | 登录异常 | 登录异常 |
| 403 | 无权限 | 无权限 |
| 404 | 资源不存在 | 资源不存在 |
| 500 | 服务器异常 | 服务器异常 |

### 0.5 HTTP 状态码约定

- 业务成功：通常 `HTTP 200` + `code=200`
- 业务失败（参数/校验/不存在等）：通常 `HTTP 200` + `code=201`
- 系统错误：通常 `HTTP 5xx` + `code=500`
- 方法不允许：当前实现为 `HTTP 405` + `code=201`

### 0.6 鉴权说明

- 小程序端需要登录的接口会解析 `Authorization: Bearer <token>`，从 token 的 `sub` 字段得到 `userId`；不需要再额外传 `userId` 参数。
- `/admin/*` 管理端接口仍暂未接入 token 校验。

---

## module-health
健康检查模块 √

### api-health
GET /health √

用途：用于部署探活与检查服务是否存活。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L129-L133)
- Handler：[health.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/health.go#L1-L21)

实现逻辑：

1. handler 初始化时生成 `startedAt`（闭包捕获）。
2. 每次请求返回 `status/startedAt/now`。
3. 使用统一响应 `SendJSuccess` 返回 `data`。

请求：

- Method：`GET`
- Path：`/health`

请求示例：

```bash
curl -X GET "http://localhost:8080/health"
```

响应 `data` 字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| status | string | 固定为 `ok` |
| startedAt | string | 服务启动时间（RFC3339） |
| now | string | 当前时间（RFC3339） |

响应示例：

```json
{
  "code": 200,
  "data": {
    "now": "2026-01-29T12:00:00Z",
    "startedAt": "2026-01-29T11:55:00Z",
    "status": "ok"
  },
  "message": "ok"
}
```

---

## module-auth
Auth 模块（小程序登录） √

### api-auth-wechat-login
POST /api/auth/wechat/login √

用途：小程序登录（临时方案：直接用 openid 登录并签发 token）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L129-L142)
- Handler：[wechat.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/wechat.go#L1-L57)
- Middleware：解析 `Authorization: Bearer <token>` 写入 `X-User-Id`（[middleware.go](file:///e:/VUE3/新建文件夹/GameSocial/api/middleware/middleware.go#L52-L83)）
- Service：`auth.Service.OpenIDLogin`（[service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/auth/service.go)）

实现逻辑：

1. 校验 HTTP 方法必须为 `POST`，并校验 `svc` 已注入。
2. 解析 JSON body，读取 `openId/openid`（兼容字段）。
3. 调用 `svc.OpenIDLogin(ctx, openID)`：按 openid 获取/创建用户并签发 token。
4. 成功使用 `SendJSuccess` 返回；失败使用 `SendJBizFail/SendJError` 返回。

请求：

- Method：`POST`
- Path：`/api/auth/wechat/login`
- Body：JSON

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| openId | string | 是 | 小程序 openid |
| openid | string | 否 | 兼容字段（同 openId） |

请求示例：

```bash
curl -X POST "http://localhost:8080/api/auth/wechat/login" \
  -H "Content-Type: application/json" \
  -d "{\"openId\":\"o_xxxxxxx\"}"
```

成功响应 `data` 字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| token | string | 访问 token（后续请求放到 `Authorization: Bearer <token>`） |
| user | object | 用户信息 |

`user` 字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 用户 ID |
| openId | string | 小程序 openid |
| unionId | string | 可选 unionid |
| nickname | string | 昵称（当前创建时为空字符串） |
| avatarUrl | string | 头像 URL（当前创建时为空字符串） |
| status | number | 1=正常，0=封禁 |

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1001,
      "openId": "o_xxxxxxx",
      "unionId": "",
      "nickname": "",
      "avatarUrl": "",
      "status": 1
    }
  },
  "message": "ok"
}
```

失败场景示例（openId 为空）：

```json
{
  "code": 201,
  "message": "openId 不能为空"
}
```

---

## module-item
Item 模块（管理员：积分商品管理） √

### api-admin-goods-create
POST /admin/goods √

用途：创建积分商品（goods）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L154-L160)
- Handler：[admin_goods.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go)
- Service：[item/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go)

实现逻辑：

1. 校验方法为 `POST`，并校验 `svc` 已注入。
2. 解析 JSON body 为创建请求对象，进行基础校验（必填字段、数值范围等）。
3. 写入 `goods` 表并返回创建后的商品对象。
4. 返回 `SendJSuccess`。

请求：

- Method：`POST`
- Path：`/admin/goods`
- Body：JSON

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| name | string | 是 | 商品名 |
| coverUrl | string | 否 | 封面 URL（可为空字符串） |
| pointsPrice | number | 是 | 所需积分（必须 >= 0） |
| stock | number | 是 | 库存（必须 >= 0） |
| status | number | 否 | 1=上架，0=下架；不传/传 0 会默认写入 1 |

请求示例：

```bash
curl -X POST "http://localhost:8080/admin/goods" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"可乐\",\"coverUrl\":\"\",\"pointsPrice\":30,\"stock\":100,\"status\":1}"
```

成功响应 `data` 为商品对象：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 商品 ID |
| name | string | 商品名 |
| coverUrl | string | 封面 URL |
| pointsPrice | number | 所需积分 |
| stock | number | 库存 |
| status | number | 1=上架，0=下架（软删除会置 0） |
| createdAt | string | 创建时间（RFC3339） |

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "可乐",
    "coverUrl": "",
    "pointsPrice": 30,
    "stock": 100,
    "status": 1,
    "createdAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-goods-list
GET /admin/goods √

用途：商品列表（默认排除已软删除的商品）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L154-L160)
- Handler：[admin_goods.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go)
- Service：[item/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 解析 `offset/limit/status`（并做分页兜底）。
3. 从 `goods` 表查询列表（默认排除软删除记录）。
4. 返回 `SendJSuccess`。

请求：

- Method：`GET`
- Path：`/admin/goods`
- Query：
  - `offset`：默认 0
  - `limit`：默认 20，最大 200
  - `status`：0 表示不过滤（但仍排除已删除）；1/其它值表示按 status 过滤

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/goods?offset=0&limit=20&status=1"
```

响应 `data`：`Goods[]` 数组（结构同创建响应的商品对象）。

响应示例：

```json
{
  "code": 200,
  "data": [
    {
      "id": 2,
      "name": "雪碧",
      "coverUrl": "",
      "pointsPrice": 20,
      "stock": 50,
      "status": 1,
      "createdAt": "2026-01-29T12:10:00Z"
    },
    {
      "id": 1,
      "name": "可乐",
      "coverUrl": "",
      "pointsPrice": 30,
      "stock": 100,
      "status": 1,
      "createdAt": "2026-01-29T12:00:00Z"
    }
  ],
  "message": "ok"
}
```

### api-admin-goods-get
GET /admin/goods/{id} √

用途：获取单个商品详情。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L154-L160)
- Handler：[admin_goods.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go)
- Service：[item/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，非法直接返回业务失败。
3. 从 `goods` 表读取单条记录并返回。
4. 返回 `SendJSuccess`。

请求：

- Method：`GET`
- Path：`/admin/goods/{id}`
- Path 参数：
  - `id`：商品 ID（uint64）

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/goods/1"
```

响应 `data`：商品对象（结构同创建响应的商品对象）。

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "可乐",
    "coverUrl": "",
    "pointsPrice": 30,
    "stock": 100,
    "status": 1,
    "createdAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-goods-update
PUT /admin/goods/{id} √

用途：更新商品（全量更新可变字段）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L154-L160)
- Handler：[admin_goods.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go)
- Service：[item/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go)

实现逻辑：

1. 校验方法为 `PUT`，并校验 `svc` 已注入。
2. 解析 path 参数 `id`，并解析 JSON body 为更新请求对象。
3. 更新 `goods` 表的可变字段并返回更新后的商品对象。
4. 返回 `SendJSuccess`。

请求：

- Method：`PUT`
- Path：`/admin/goods/{id}`
- Body：JSON（字段同创建）

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| name | string | 是 | 商品名 |
| coverUrl | string | 否 | 封面 URL（可为空字符串） |
| pointsPrice | number | 是 | 所需积分（必须 >= 0） |
| stock | number | 是 | 库存（必须 >= 0） |
| status | number | 否 | 1=上架，0=下架；不传/传 0 会默认写入 1 |

请求示例：

```bash
curl -X PUT "http://localhost:8080/admin/goods/1" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"可乐(大)\",\"coverUrl\":\"\",\"pointsPrice\":35,\"stock\":80,\"status\":1}"
```

响应 `data`：更新后的商品对象（结构同创建响应的商品对象）。

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "可乐(大)",
    "coverUrl": "",
    "pointsPrice": 35,
    "stock": 80,
    "status": 1,
    "createdAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-goods-delete
DELETE /admin/goods/{id} √

用途：删除商品（软删除：`status=0`）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L154-L160)
- Handler：[admin_goods.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go)
- Service：[item/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go)

实现逻辑：

1. 校验方法为 `DELETE`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，非法直接返回业务失败。
3. 调用 `svc.DeleteGoods(ctx, id)`，把 `goods.status` 更新为 0（软删除）。
4. 返回 `data.deleted=true`。

请求：

- Method：`DELETE`
- Path：`/admin/goods/{id}`

请求示例：

```bash
curl -X DELETE "http://localhost:8080/admin/goods/1"
```

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "deleted": true
  },
  "message": "ok"
}
```

---

## module-tournament
Tournament 模块（管理员：赛事管理） √

### api-admin-tournaments-create
POST /admin/tournaments √

用途：创建赛事（tournament）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L161-L166)
- Handler：[admin_tournaments.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go#L1-L44)
- Service：[tournament/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L1-L131)

实现逻辑：

1. 校验方法为 `POST`，并校验 `svc` 已注入。
2. 解析 JSON body 为 `CreateTournamentRequest`。
3. 调用 `svc.Create(ctx, req)` 写入 `tournament` 表，并返回创建后的赛事详情。
4. 返回 `SendJSuccess`。

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| title | string | 是 | 标题 |
| content | string | 否 | 详情（可为空字符串） |
| coverUrl | string | 否 | 封面 URL（可为空字符串） |
| startAt | string | 是 | 开始时间（RFC3339） |
| endAt | string | 是 | 结束时间（RFC3339，必须 >= startAt） |
| status | string | 否 | DRAFT/PUBLISHED/FINISHED/CANCELED；不传默认 DRAFT |
| createdByAdminId | number | 否 | 创建人管理员 ID；不传默认 1 |

请求示例：

```bash
curl -X POST "http://localhost:8080/admin/tournaments" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"周赛\",\"content\":\"\",\"coverUrl\":\"\",\"startAt\":\"2026-02-01T12:00:00Z\",\"endAt\":\"2026-02-01T16:00:00Z\",\"status\":\"PUBLISHED\",\"createdByAdminId\":1}"
```

成功响应 `data`：赛事对象。

赛事对象字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 赛事 ID |
| title | string | 标题 |
| content | string | 详情 |
| coverUrl | string | 封面 |
| startAt | string | 开始时间 |
| endAt | string | 结束时间 |
| status | string | 状态 |
| createdByAdminId | number | 创建人管理员 ID |
| createdAt | string | 创建时间 |
| updatedAt | string | 更新时间 |

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "title": "周赛",
    "content": "",
    "coverUrl": "",
    "startAt": "2026-02-01T12:00:00Z",
    "endAt": "2026-02-01T16:00:00Z",
    "status": "PUBLISHED",
    "createdByAdminId": 1,
    "createdAt": "2026-01-29T12:00:00Z",
    "updatedAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-tournaments-list
GET /admin/tournaments √

用途：赛事列表。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L161-L166)
- Handler：[admin_tournaments.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go)
- Service：[tournament/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 解析 query：`offset/limit/status`（并做分页兜底）。
3. 调用 `svc.List(ctx, req)` 查询 `tournament` 列表（未指定 status 时默认排除 `CANCELED`）。
4. 返回 `SendJSuccess`。

Query：

- `offset`：默认 0
- `limit`：默认 20，最大 200
- `status`：可选；不传则默认排除 `CANCELED`

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/tournaments?offset=0&limit=20&status=PUBLISHED"
```

响应 `data`：`Tournament[]`

响应示例：

```json
{
  "code": 200,
  "data": [
    {
      "id": 2,
      "title": "周赛(第二期)",
      "content": "",
      "coverUrl": "",
      "startAt": "2026-02-08T12:00:00Z",
      "endAt": "2026-02-08T16:00:00Z",
      "status": "PUBLISHED",
      "createdByAdminId": 1,
      "createdAt": "2026-01-29T12:30:00Z",
      "updatedAt": "2026-01-29T12:30:00Z"
    },
    {
      "id": 1,
      "title": "周赛",
      "content": "",
      "coverUrl": "",
      "startAt": "2026-02-01T12:00:00Z",
      "endAt": "2026-02-01T16:00:00Z",
      "status": "PUBLISHED",
      "createdByAdminId": 1,
      "createdAt": "2026-01-29T12:00:00Z",
      "updatedAt": "2026-01-29T12:00:00Z"
    }
  ],
  "message": "ok"
}
```

### api-admin-tournaments-get
GET /admin/tournaments/{id} √

用途：赛事详情。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L161-L166)
- Handler：[admin_tournaments.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go)
- Service：[tournament/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，非法直接返回业务失败。
3. 调用 `svc.Get(ctx, id)` 读取赛事详情。
4. 返回 `SendJSuccess`。

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/tournaments/1"
```

响应 `data`：`Tournament`

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "title": "周赛",
    "content": "",
    "coverUrl": "",
    "startAt": "2026-02-01T12:00:00Z",
    "endAt": "2026-02-01T16:00:00Z",
    "status": "PUBLISHED",
    "createdByAdminId": 1,
    "createdAt": "2026-01-29T12:00:00Z",
    "updatedAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-tournaments-update
PUT /admin/tournaments/{id} √

用途：更新赛事（全量更新可变字段）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L161-L166)
- Handler：[admin_tournaments.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go)
- Service：[tournament/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go)

实现逻辑：

1. 校验方法为 `PUT`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，并解析 JSON body 为 `UpdateTournamentRequest`。
3. 调用 `svc.Update(ctx, id, req)` 更新赛事并返回最新详情。
4. 返回 `SendJSuccess`。

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| title | string | 是 | 标题 |
| content | string | 否 | 详情 |
| coverUrl | string | 否 | 封面 |
| startAt | string | 是 | 开始时间 |
| endAt | string | 是 | 结束时间 |
| status | string | 否 | DRAFT/PUBLISHED/FINISHED/CANCELED；不传默认 DRAFT |

请求示例：

```bash
curl -X PUT "http://localhost:8080/admin/tournaments/1" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"周赛(更新)\",\"content\":\"\",\"coverUrl\":\"\",\"startAt\":\"2026-02-01T12:00:00Z\",\"endAt\":\"2026-02-01T16:00:00Z\",\"status\":\"PUBLISHED\"}"
```

响应 `data`：更新后的 `Tournament`

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "title": "周赛(更新)",
    "content": "",
    "coverUrl": "",
    "startAt": "2026-02-01T12:00:00Z",
    "endAt": "2026-02-01T16:00:00Z",
    "status": "PUBLISHED",
    "createdByAdminId": 1,
    "createdAt": "2026-01-29T12:00:00Z",
    "updatedAt": "2026-01-29T12:40:00Z"
  },
  "message": "ok"
}
```

### api-admin-tournaments-delete
DELETE /admin/tournaments/{id} √

用途：删除赛事（软删除：将 `status` 置为 `CANCELED`）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L161-L166)
- Handler：[admin_tournaments.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go)
- Service：[tournament/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go)

实现逻辑：

1. 校验方法为 `DELETE`，并校验 `svc` 已注入。
2. 从 path 解析 `id`。
3. 调用 `svc.Delete(ctx, id)` 将赛事状态更新为 `CANCELED`（软删除）。
4. 返回 `data.deleted=true`。

请求示例：

```bash
curl -X DELETE "http://localhost:8080/admin/tournaments/1"
```

成功响应：`data.deleted=true`

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "deleted": true
  },
  "message": "ok"
}
```

---

## module-task
Task 模块（管理员：任务定义管理） √

### api-admin-task-defs-create
POST /admin/task-defs √

用途：创建任务定义（task_def）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L168-L174)
- Handler：[admin_task_defs.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go)
- Service：[task/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go)

实现逻辑：

1. 校验方法为 `POST`，并校验 `svc` 已注入。
2. 解析 JSON body 为 `CreateTaskDefRequest`。
3. 写入 `task_def` 表并返回创建后的任务定义对象。
4. 返回 `SendJSuccess`。

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| taskCode | string | 是 | 任务编码（建议唯一） |
| name | string | 是 | 名称 |
| periodType | string | 是 | DAILY/WEEKLY/MONTHLY |
| targetCount | number | 是 | 目标次数（必须 > 0） |
| rewardPoints | number | 是 | 奖励积分（必须 >= 0） |
| status | number | 否 | 1=启用，0=停用；不传/传 0 会默认写入 1 |
| ruleJson | object | 否 | 扩展规则（可不传/传空对象） |

请求示例：

```bash
curl -X POST "http://localhost:8080/admin/task-defs" \
  -H "Content-Type: application/json" \
  -d "{\"taskCode\":\"CHECKIN_DAILY\",\"name\":\"到店打卡\",\"periodType\":\"DAILY\",\"targetCount\":1,\"rewardPoints\":10,\"status\":1,\"ruleJson\":{}}"
```

成功响应 `data`：任务定义对象。

任务定义对象字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 任务 ID |
| taskCode | string | 任务编码 |
| name | string | 名称 |
| periodType | string | 周期类型 |
| targetCount | number | 目标次数 |
| rewardPoints | number | 奖励积分 |
| status | number | 1=启用，0=停用（软删除会置 0） |
| ruleJson | object | 规则 JSON |
| createdAt | string | 创建时间 |

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "taskCode": "CHECKIN_DAILY",
    "name": "到店打卡",
    "periodType": "DAILY",
    "targetCount": 1,
    "rewardPoints": 10,
    "status": 1,
    "ruleJson": {},
    "createdAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-task-defs-list
GET /admin/task-defs √

用途：任务定义列表。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L168-L174)
- Handler：[admin_task_defs.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go)
- Service：[task/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 解析 query：`offset/limit/status`（并做分页兜底）。
3. 查询 `task_def` 表返回列表（默认排除软删除/停用记录取决于传参）。
4. 返回 `SendJSuccess`。

Query：

- `offset`：默认 0
- `limit`：默认 20，最大 200
- `status`：0 表示不过滤（但仍排除已删除）；非 0 表示按 status 过滤

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/task-defs?offset=0&limit=20&status=1"
```

响应 `data`：`TaskDef[]`

响应示例：

```json
{
  "code": 200,
  "data": [
    {
      "id": 2,
      "taskCode": "SPEND_POINTS",
      "name": "积分兑换",
      "periodType": "DAILY",
      "targetCount": 1,
      "rewardPoints": 5,
      "status": 1,
      "ruleJson": {},
      "createdAt": "2026-01-29T12:10:00Z"
    },
    {
      "id": 1,
      "taskCode": "CHECKIN_DAILY",
      "name": "到店打卡",
      "periodType": "DAILY",
      "targetCount": 1,
      "rewardPoints": 10,
      "status": 1,
      "ruleJson": {},
      "createdAt": "2026-01-29T12:00:00Z"
    }
  ],
  "message": "ok"
}
```

### api-admin-task-defs-get
GET /admin/task-defs/{id} √

用途：任务定义详情。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L168-L174)
- Handler：[admin_task_defs.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go)
- Service：[task/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 从 path 解析 `id`。
3. 查询 `task_def` 表返回单条任务定义对象（`rule_json` 扫描为 JSON）。
4. 返回 `SendJSuccess`。

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/task-defs/1"
```

响应 `data`：`TaskDef`

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "taskCode": "CHECKIN_DAILY",
    "name": "到店打卡",
    "periodType": "DAILY",
    "targetCount": 1,
    "rewardPoints": 10,
    "status": 1,
    "ruleJson": {},
    "createdAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-task-defs-update
PUT /admin/task-defs/{id} √

用途：更新任务定义（全量更新可变字段）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L168-L174)
- Handler：[admin_task_defs.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go)
- Service：[task/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go)

实现逻辑：

1. 校验方法为 `PUT`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，并解析 JSON body 为 `UpdateTaskDefRequest`。
3. 更新 `task_def` 表并返回更新后的任务定义对象。
4. 返回 `SendJSuccess`。

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| name | string | 是 | 名称 |
| periodType | string | 是 | DAILY/WEEKLY/MONTHLY |
| targetCount | number | 是 | 目标次数 |
| rewardPoints | number | 是 | 奖励积分 |
| status | number | 否 | 1=启用，0=停用；不传/传 0 会默认写入 1 |
| ruleJson | object | 否 | 规则 JSON |

请求示例：

```bash
curl -X PUT "http://localhost:8080/admin/task-defs/1" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"到店打卡(更新)\",\"periodType\":\"DAILY\",\"targetCount\":1,\"rewardPoints\":15,\"status\":1,\"ruleJson\":{}}"
```

响应 `data`：更新后的 `TaskDef`

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "taskCode": "CHECKIN_DAILY",
    "name": "到店打卡(更新)",
    "periodType": "DAILY",
    "targetCount": 1,
    "rewardPoints": 15,
    "status": 1,
    "ruleJson": {},
    "createdAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-task-defs-delete
DELETE /admin/task-defs/{id} √

用途：删除任务定义（软删除：`status=0`）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L168-L174)
- Handler：[admin_task_defs.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go)
- Service：[task/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go)

实现逻辑：

1. 校验方法为 `DELETE`，并校验 `svc` 已注入。
2. 从 path 解析 `id`。
3. 更新 `task_def.status=0` 作为软删除。
4. 返回 `data.deleted=true`。

请求示例：

```bash
curl -X DELETE "http://localhost:8080/admin/task-defs/1"
```

成功响应：`data.deleted=true`

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "deleted": true
  },
  "message": "ok"
}
```

---

## module-user
User 模块（管理员：用户管理） √

### api-admin-users-list
GET /admin/users √

用途：用户列表。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L175-L178)
- Handler：[admin_users.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_users.go)
- Service：[user/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 解析 query：`offset/limit/status`。
3. 查询 `user` 表返回用户数组。
4. 返回 `SendJSuccess`。

Query：

- `offset`：默认 0
- `limit`：默认 20，最大 200
- `status`：可选；非 0 则按 status 过滤（1=正常，0=封禁）

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/users?offset=0&limit=20&status=1"
```

响应 `data`：`User[]`

用户对象字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 用户 ID |
| openId | string | openid |
| unionId | string | unionid |
| nickname | string | 昵称 |
| avatarUrl | string | 头像 |
| status | number | 1=正常，0=封禁 |
| createdAt | string | 创建时间 |
| updatedAt | string | 更新时间 |

响应示例：

```json
{
  "code": 200,
  "data": [
    {
      "id": 1002,
      "openId": "o_yyyyyyy",
      "unionId": "",
      "nickname": "小明",
      "avatarUrl": "",
      "status": 1,
      "createdAt": "2026-01-29T12:05:00Z",
      "updatedAt": "2026-01-29T12:05:00Z"
    },
    {
      "id": 1001,
      "openId": "o_xxxxxxx",
      "unionId": "",
      "nickname": "",
      "avatarUrl": "",
      "status": 1,
      "createdAt": "2026-01-29T12:00:00Z",
      "updatedAt": "2026-01-29T12:00:00Z"
    }
  ],
  "message": "ok"
}
```

### api-admin-users-get
GET /admin/users/{id} √

用途：用户详情。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L175-L178)
- Handler：[admin_users.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_users.go#L1-L118)
- Service：[user/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 从 path 解析 `id`。
3. 调用 `svc.Get(ctx, id)` 查询用户并返回。
4. 返回 `SendJSuccess`。

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/users/1001"
```

响应 `data`：`User`

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1001,
    "openId": "o_xxxxxxx",
    "unionId": "",
    "nickname": "",
    "avatarUrl": "",
    "status": 1,
    "createdAt": "2026-01-29T12:00:00Z",
    "updatedAt": "2026-01-29T12:00:00Z"
  },
  "message": "ok"
}
```

### api-admin-users-update
PUT /admin/users/{id} √

用途：更新用户资料/状态。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L175-L178)
- Handler：[admin_users.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_users.go#L1-L118)
- Service：[user/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go)

实现逻辑：

1. 校验方法为 `PUT`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，并解析 JSON body 为 `UpdateUserRequest`。
3. 调用 `svc.Update(ctx, id, req)` 更新用户可变字段并返回最新对象。
4. 返回 `SendJSuccess`。

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| nickname | string | 否 | 昵称（可为空字符串） |
| avatarUrl | string | 否 | 头像 URL（可为空字符串） |
| status | number | 否 | 1=正常，0=封禁 |

注意：`status` 不传则不修改当前用户状态。

请求示例：

```bash
curl -X PUT "http://localhost:8080/admin/users/1001" \
  -H "Content-Type: application/json" \
  -d "{\"nickname\":\"9527\",\"avatarUrl\":\"\",\"status\":1}"
```

响应 `data`：更新后的 `User`

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 1001,
    "openId": "o_xxxxxxx",
    "unionId": "",
    "nickname": "9527",
    "avatarUrl": "",
    "status": 1,
    "createdAt": "2026-01-29T12:00:00Z",
    "updatedAt": "2026-01-29T12:45:00Z"
  },
  "message": "ok"
}
```

---

## module-redeem
Redeem 模块（管理员：兑换订单管理） √

### api-admin-redeem-orders-create
POST /admin/redeem/orders √

用途：创建兑换订单（最小 CRUD：当前不接入积分扣减流水）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L180-L185)
- Handler：[admin_redeem_orders.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go)
- Service：[redeem/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go)

实现逻辑：

1. 校验方法为 `POST`，并校验 `svc` 已注入。
2. 解析 JSON body（`userId/items`），并做基础校验（items 至少 1 条、数量/积分非负等）。
3. 调用 `svc.CreateOrder(ctx, req)`：写 `redeem_order` 与 `redeem_order_item`，生成 `orderNo`。
4. 返回 `SendJSuccess`。

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| userId | number | 是 | 下单用户 ID |
| items | array | 是 | 订单明细数组（至少 1 条） |

items 元素字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| goodsId | number | 是 | 商品 ID |
| quantity | number | 是 | 数量（必须 > 0） |
| pointsPrice | number | 是 | 下单时单价积分快照（必须 >= 0） |

请求示例：

```bash
curl -X POST "http://localhost:8080/admin/redeem/orders" \
  -H "Content-Type: application/json" \
  -d "{\"userId\":1001,\"items\":[{\"goodsId\":1,\"quantity\":2,\"pointsPrice\":30}]}"
```

成功响应 `data`：订单详情（包含 items）。

订单对象字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 订单 ID |
| orderNo | string | 订单号（后端生成） |
| userId | number | 用户 ID |
| status | string | CREATED/USED/CANCELED |
| totalPoints | number | 总积分（quantity*pointsPrice 之和） |
| usedByAdminId | number | 核销管理员 ID（已核销时出现） |
| usedAt | string | 核销时间（已核销时出现） |
| createdAt | string | 创建时间 |
| items | array | 明细数组 |

items 元素字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 明细 ID |
| redeemOrderId | number | 订单 ID |
| goodsId | number | 商品 ID |
| quantity | number | 数量 |
| pointsPrice | number | 单价积分快照 |

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 10,
    "orderNo": "R20260129120000a1b2c3d4",
    "userId": 1001,
    "status": "CREATED",
    "totalPoints": 60,
    "createdAt": "2026-01-29T12:00:00Z",
    "items": [
      {
        "id": 100,
        "redeemOrderId": 10,
        "goodsId": 1,
        "quantity": 2,
        "pointsPrice": 30
      }
    ]
  },
  "message": "ok"
}
```

### api-admin-redeem-orders-list
GET /admin/redeem/orders √

用途：兑换订单列表（不包含 items，避免列表响应过重）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L180-L185)
- Handler：[admin_redeem_orders.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go)
- Service：[redeem/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 解析 query：`offset/limit/status/userId`。
3. 调用 `svc.ListOrders(ctx, req)` 查询 `redeem_order` 列表（不带 items）。
4. 返回 `SendJSuccess`。

Query：

- `offset`：默认 0
- `limit`：默认 20，最大 200
- `status`：可选；例如 CREATED/USED/CANCELED
- `userId`：可选；按用户筛选

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/redeem/orders?offset=0&limit=20&status=CREATED&userId=1001"
```

响应 `data`：`RedeemOrder[]`（不带 items）

列表元素字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 订单 ID |
| orderNo | string | 订单号 |
| userId | number | 用户 ID |
| status | string | CREATED/USED/CANCELED |
| totalPoints | number | 总积分 |
| usedByAdminId | number | 核销管理员 ID（已核销时为非 0） |
| usedAt | string | 核销时间（已核销时出现） |
| createdAt | string | 创建时间 |

响应示例：

```json
{
  "code": 200,
  "data": [
    {
      "id": 11,
      "orderNo": "R20260129121000b1b2c3d4",
      "userId": 1001,
      "status": "CREATED",
      "totalPoints": 20,
      "createdAt": "2026-01-29T12:10:00Z"
    },
    {
      "id": 10,
      "orderNo": "R20260129120000a1b2c3d4",
      "userId": 1001,
      "status": "USED",
      "totalPoints": 60,
      "usedByAdminId": 1,
      "usedAt": "2026-01-29T12:20:00Z",
      "createdAt": "2026-01-29T12:00:00Z"
    }
  ],
  "message": "ok"
}
```

### api-admin-redeem-orders-get
GET /admin/redeem/orders/{id} √

用途：兑换订单详情（包含 items）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L180-L185)
- Handler：[admin_redeem_orders.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go)
- Service：[redeem/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go)

实现逻辑：

1. 校验方法为 `GET`，并校验 `svc` 已注入。
2. 从 path 解析 `id`。
3. 调用 `svc.GetOrder(ctx, id)`：读取订单主表 + 明细表并组装返回。
4. 返回 `SendJSuccess`。

请求示例：

```bash
curl -X GET "http://localhost:8080/admin/redeem/orders/10"
```

响应 `data`：`RedeemOrder`（带 items）

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 10,
    "orderNo": "R20260129120000a1b2c3d4",
    "userId": 1001,
    "status": "CREATED",
    "totalPoints": 60,
    "createdAt": "2026-01-29T12:00:00Z",
    "items": [
      {
        "id": 100,
        "redeemOrderId": 10,
        "goodsId": 1,
        "quantity": 2,
        "pointsPrice": 30
      }
    ]
  },
  "message": "ok"
}
```

### api-admin-redeem-orders-use
PUT /admin/redeem/orders/{id}/use √

用途：核销订单（仅允许 `CREATED -> USED`，避免重复核销）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L180-L185)
- Handler：[admin_redeem_orders.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go#L118-L160)
- Service：[redeem/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L270-L323)

实现逻辑：

1. 校验方法为 `PUT`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，校验为正整数。
3. 解析可选 JSON body：`adminId/admin_id`；未提供则默认 `1`。
4. 调用 `svc.UseOrder(ctx, id, adminId)`：仅允许 `CREATED -> USED`，更新 `used_by_admin_id/used_at`。
5. 返回更新后的 `RedeemOrder`（包含 items）。

请求：

- Method：`PUT`
- Path：`/admin/redeem/orders/{id}/use`
- Body：可选 JSON

请求体字段（可选）：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| adminId | number | 否 | 核销管理员 ID；不传默认 1 |

请求示例：

```bash
curl -X PUT "http://localhost:8080/admin/redeem/orders/10/use" \
  -H "Content-Type: application/json" \
  -d "{\"adminId\":1}"
```

响应 `data`：更新后的 `RedeemOrder`（带 items，status 变为 USED）

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 10,
    "orderNo": "R20260129120000a1b2c3d4",
    "userId": 1001,
    "status": "USED",
    "totalPoints": 60,
    "usedByAdminId": 1,
    "usedAt": "2026-01-29T12:20:00Z",
    "createdAt": "2026-01-29T12:00:00Z",
    "items": [
      {
        "id": 100,
        "redeemOrderId": 10,
        "goodsId": 1,
        "quantity": 2,
        "pointsPrice": 30
      }
    ]
  },
  "message": "ok"
}
```

### api-admin-redeem-orders-cancel
PUT /admin/redeem/orders/{id}/cancel √

用途：取消订单（仅允许 `CREATED -> CANCELED`）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L180-L185)
- Handler：[admin_redeem_orders.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go#L162-L193)
- Service：[redeem/service.go](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L299-L323)

实现逻辑：

1. 校验方法为 `PUT`，并校验 `svc` 已注入。
2. 从 path 解析 `id`，校验为正整数。
3. 调用 `svc.CancelOrder(ctx, id)`：仅允许 `CREATED -> CANCELED`。
4. 返回更新后的 `RedeemOrder`（包含 items）。

请求示例：

```bash
curl -X PUT "http://localhost:8080/admin/redeem/orders/10/cancel"
```

响应 `data`：更新后的 `RedeemOrder`（带 items，status 变为 CANCELED）

响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 10,
    "orderNo": "R20260129120000a1b2c3d4",
    "userId": 1001,
    "status": "CANCELED",
    "totalPoints": 60,
    "createdAt": "2026-01-29T12:00:00Z",
    "items": [
      {
        "id": 100,
        "redeemOrderId": 10,
        "goodsId": 1,
        "quantity": 2,
        "pointsPrice": 30
      }
    ]
  },
  "message": "ok"
}
```

---

## module-user-app
User 模块（小程序：个人资料） √

### api-users-me-get
GET /api/users/me √

用途：获取当前登录用户的个人资料（昵称、头像等）。

请求头：

- `Authorization: Bearer <token>`

成功响应 `data`：个人资料对象

| 字段 | 类型 | 说明 |
|---|---|---|
| nickname | string | 昵称（为空表示未设置） |
| avatarUrl | string | 头像 URL（为空表示未设置） |

### api-users-me-update
PUT /api/users/me √

用途：更新当前登录用户的个人资料（昵称、头像等）。

请求头：

- `Authorization: Bearer <token>`

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| nickname | string | 否 | 昵称（可为空字符串） |
| avatarUrl | string | 否 | 头像 URL（可为空字符串） |

---

## module-points
Points 模块（小程序：积分账户与流水） √

### api-points-balance
GET /api/points/balance √

用途：获取当前登录用户的积分余额。

请求头：

- `Authorization: Bearer <token>`

### api-points-ledgers
GET /api/points/ledgers √

用途：获取当前登录用户的积分流水列表。

请求头：

- `Authorization: Bearer <token>`

---

## module-vip
VIP 模块（小程序：会员订阅） √

### api-vip-status
GET /api/vip/status √

用途：获取当前登录用户的会员状态（是否会员、到期时间等）。

请求头：

- `Authorization: Bearer <token>`

---

## module-tournament-app
Tournament 模块（小程序：赛事） √

请求头（报名/取消报名必需；查询排名可选）：

- `Authorization: Bearer <token>`

### api-tournaments-list
GET /api/tournaments √

用途：赛事列表。

### api-tournaments-get
GET /api/tournaments/{id} √

用途：赛事详情。

### api-tournaments-join
POST /api/tournaments/{id}/join √

用途：当前登录用户报名参加指定赛事。

请求：

- Method：`POST`
- Path：`/api/tournaments/{id}/join`
- Path 参数：
  - `id`：赛事 ID

成功响应：`data.joined=true`

请求示例：

```bash
curl -X POST "http://localhost:8080/api/tournaments/4001/join" \
  -H "Authorization: Bearer <token>"
```

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "joined": true
  },
  "message": "ok"
}
```

### api-tournaments-cancel
PUT /api/tournaments/{id}/cancel √

用途：当前登录用户取消指定赛事的报名（幂等；重复取消仍返回成功）。

请求：

- Method：`PUT`
- Path：`/api/tournaments/{id}/cancel`
- Path 参数：
  - `id`：赛事 ID

成功响应：`data.canceled=true`

请求示例：

```bash
curl -X PUT "http://localhost:8080/api/tournaments/4001/cancel" \
  -H "Authorization: Bearer <token>"
```

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "canceled": true
  },
  "message": "ok"
}
```

### api-tournaments-results
GET /api/tournaments/{id}/results √

用途：查看赛事排名/成绩列表；如携带登录态会额外返回当前用户名次。

请求：

- Method：`GET`
- Path：`/api/tournaments/{id}/results`
- Path 参数：
  - `id`：赛事 ID
- Query：
  - `offset`：默认 0
  - `limit`：默认 50，最大 200

响应 `data`：

- `items`：排名列表（按 rankNo 升序）
- `my`：当前用户名次（可选；仅在携带登录态且有成绩时返回）

`items/my` 字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| userId | number | 用户 ID |
| rankNo | number | 名次 |
| score | number | 分数（无则为 0） |
| nickname | string | 昵称 |
| avatarUrl | string | 头像 URL |

请求示例：

```bash
curl -X GET "http://localhost:8080/api/tournaments/4001/results?offset=0&limit=50"
```

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "items": [
      {
        "userId": 1003,
        "rankNo": 1,
        "score": 10,
        "nickname": "小王",
        "avatarUrl": ""
      }
    ],
    "my": {
      "userId": 1003,
      "rankNo": 1,
      "score": 10,
      "nickname": "小王",
      "avatarUrl": ""
    }
  },
  "message": "ok"
}
```

---

## module-task-app
Task 模块（小程序：任务与打卡） ×

### api-tasks-list
GET /api/tasks √

用途：查询任务列表（含当前周期进度与领奖状态）。

### api-tasks-checkin
POST /api/tasks/checkin ×

用途：到店打卡（写入 checkin_log 并推动任务进度）。

### api-tasks-claim
POST /api/tasks/{taskCode}/claim ×

用途：领取任务奖励（需幂等，避免重复发奖）。

---

## module-item-app
Item 模块（小程序：积分商品） √

### api-goods-list
GET /api/goods √

用途：商品列表（用户侧展示）。

### api-goods-get
GET /api/goods/{id} √

用途：商品详情（用户侧展示）。

---

## module-redeem-app
Redeem 模块（小程序：兑换订单） √

请求头（以下接口通用）：

- `Authorization: Bearer <token>`

### api-redeem-orders-create
POST /api/redeem/orders √

用途：创建兑换订单（userId 从 token 获取；扣减积分并生成订单号）。

### api-redeem-orders-list
GET /api/redeem/orders √

用途：查询我的兑换订单列表。

### api-redeem-orders-get
GET /api/redeem/orders/{id} √

用途：查询我的兑换订单详情（含 items）。

### api-redeem-orders-cancel
PUT /api/redeem/orders/{id}/cancel √

用途：取消我的兑换订单（仅允许 CREATED -> CANCELED）。

---

## module-admin
Admin 模块（管理员：登录/审计/关键操作） ×

### api-admin-auth-login
POST /admin/auth/login ×

用途：管理员登录并签发管理员 token。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L682-L694)

实现逻辑：

1. 校验方法为 `POST`。
2. 当前为占位实现：直接返回固定 token（`admin_token_placeholder`）。

### api-admin-auth-me
GET /admin/auth/me ×

用途：获取当前管理员信息（用于后台鉴权/展示）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L696-L709)

实现逻辑：

1. 校验方法为 `GET`。
2. 当前为占位实现：直接返回固定管理员信息（`id=1, username=admin`）。

### api-admin-auth-logout
POST /admin/auth/logout ×

用途：管理员登出（可选：token 失效/黑名单）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L711-L721)

实现逻辑：

1. 校验方法为 `POST`。
2. 当前为占位实现：直接返回 `logout=true`。

### api-admin-audit-logs
GET /admin/audit/logs √

用途：查询管理员关键操作审计日志。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L723-L792)

实现逻辑：

1. 校验方法为 `GET`，并校验 `db` 已注入。
2. 解析 query：`offset/limit/adminId`（分页兜底 limit 默认 20，最大 200）。
3. 查询 `admin_audit_log` 表，按 `id DESC` 返回列表。
4. `detail_json` 以 JSON 字节读取并回填到响应 `detailJson`。
5. 返回 `SendJSuccess`。

### api-admin-points-adjust
POST /admin/points/adjust ×

用途：管理员给用户赠送/扣减积分（需落审计与流水）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L794-L804)

实现逻辑：

1. 校验方法为 `POST`。
2. 当前为占位实现：直接返回 `adjusted=true`。

### api-admin-users-drinks-use
PUT /admin/users/{id}/drinks/use ×

用途：管理员核销用户饮品数量（点一次 -1）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L806-L817)

实现逻辑：

1. 校验方法为 `PUT`。
2. 解析 path 参数 `id`（当前未做业务校验）。
3. 当前为占位实现：直接返回 `used=true`。

### api-admin-tournament-results-publish
POST /admin/tournaments/{id}/results/publish ×

用途：发布赛事排名（tournament_result）并记录发布人。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L819-L830)

实现逻辑：

1. 校验方法为 `POST`。
2. 解析 path 参数 `id`（当前未做业务校验）。
3. 当前为占位实现：直接返回 `published=true`。

### api-admin-tournament-awards-grant
POST /admin/tournaments/{id}/awards/grant ×

用途：给赛事前 N 名发积分奖励（需幂等，避免重复发奖）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L832-L843)

实现逻辑：

1. 校验方法为 `POST`。
2. 解析 path 参数 `id`（当前未做业务校验）。
3. 当前为占位实现：直接返回 `granted=true`。

---

## module-media
Media 模块（媒体上传/访问） ×

### api-admin-media-upload
POST /admin/media/upload ×

用途：上传封面/图片等媒体文件，返回可访问 URL。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[app_endpoints.go](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_endpoints.go#L845-L858)

实现逻辑：

1. 校验方法为 `POST`。
2. 当前为占位实现：直接返回空 `url` 与 `createdAt`（RFC3339）。

---

## module-unimplemented
小程序侧接口汇总（部分未完成） ×

本段用于保留一个“规划接口”总入口；具体接口已拆分到上面的模块段落（User/Points/VIP/Tournament/Task/Item/Redeem/Admin/Media）。

