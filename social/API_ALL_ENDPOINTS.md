# GameSocial 功能接口总览（按模块拆分）

本文档基于《模块化架构设计》与当前后端实际已注册的路由整理，覆盖“已实现 + 规划待实现”的全部 HTTP 接口；已实现的接口在标题后标注 `√`，未实现标注 `×`。

## 接口目录

- √ [健康检查模块](#module-health)
  - √ [GET /health](#api-health)
- √ [Auth 模块（小程序登录）](#module-auth)
  - √ [POST /api/auth/wechat/login](#api-auth-wechat-login)
- √ [Item 模块（管理员：积分商品管理）](#module-item)
  - √ [POST /admin/goods](#api-admin-goods-create)
  - √ [GET /admin/goods](#api-admin-goods-list)
  - √ [GET /admin/goods/{id}](#api-admin-goods-get)
  - √ [PUT /admin/goods/{id}](#api-admin-goods-update)
  - √ [DELETE /admin/goods/{id}](#api-admin-goods-delete)
- √ [Tournament 模块（管理员：赛事管理）](#module-tournament)
  - √ [POST /admin/tournaments](#api-admin-tournaments-create)
  - √ [GET /admin/tournaments](#api-admin-tournaments-list)
  - √ [GET /admin/tournaments/{id}](#api-admin-tournaments-get)
  - √ [PUT /admin/tournaments/{id}](#api-admin-tournaments-update)
  - √ [DELETE /admin/tournaments/{id}](#api-admin-tournaments-delete)
- √ [Task 模块（管理员：任务定义管理）](#module-task)
  - √ [POST /admin/task-defs](#api-admin-task-defs-create)
  - √ [GET /admin/task-defs](#api-admin-task-defs-list)
  - √ [GET /admin/task-defs/{id}](#api-admin-task-defs-get)
  - √ [PUT /admin/task-defs/{id}](#api-admin-task-defs-update)
  - √ [DELETE /admin/task-defs/{id}](#api-admin-task-defs-delete)
- √ [User 模块（管理员：用户管理）](#module-user)
  - √ [GET /admin/users](#api-admin-users-list)
  - √ [GET /admin/users/{id}](#api-admin-users-get)
  - √ [PUT /admin/users/{id}](#api-admin-users-update)
- √ [Redeem 模块（管理员：兑换订单管理）](#module-redeem)
  - √ [POST /admin/redeem/orders](#api-admin-redeem-orders-create)
  - √ [GET /admin/redeem/orders](#api-admin-redeem-orders-list)
  - √ [GET /admin/redeem/orders/{id}](#api-admin-redeem-orders-get)
  - √ [PUT /admin/redeem/orders/{id}/use](#api-admin-redeem-orders-use)
  - √ [PUT /admin/redeem/orders/{id}/cancel](#api-admin-redeem-orders-cancel)
- × [小程序侧接口（部分未完成）](#module-unimplemented)
  - √ [User 模块（小程序：个人资料）](#module-user-app)
    - √ [GET /api/users/me](#api-users-me-get)
    - √ [PUT /api/users/me](#api-users-me-update)
  - √ [Points 模块（小程序：积分账户与流水）](#module-points)
    - √ [GET /api/points/balance](#api-points-balance)
    - √ [GET /api/points/ledgers](#api-points-ledgers)
  - √ [VIP 模块（小程序：会员订阅）](#module-vip)
    - √ [GET /api/vip/status](#api-vip-status)
  - × [Tournament 模块（小程序：赛事）](#module-tournament-app)
    - √ [GET /api/tournaments](#api-tournaments-list)
    - √ [GET /api/tournaments/{id}](#api-tournaments-get)
    - × [POST /api/tournaments/{id}/join](#api-tournaments-join)
    - × [PUT /api/tournaments/{id}/cancel](#api-tournaments-cancel)
    - × [GET /api/tournaments/{id}/results](#api-tournaments-results)
  - × [Task 模块（小程序：任务与打卡）](#module-task-app)
    - √ [GET /api/tasks](#api-tasks-list)
    - × [POST /api/tasks/checkin](#api-tasks-checkin)
    - × [POST /api/tasks/{taskCode}/claim](#api-tasks-claim)
  - √ [Item 模块（小程序：积分商品）](#module-item-app)
    - √ [GET /api/goods](#api-goods-list)
    - √ [GET /api/goods/{id}](#api-goods-get)
  - √ [Redeem 模块（小程序：兑换订单）](#module-redeem-app)
    - √ [POST /api/redeem/orders](#api-redeem-orders-create)
    - √ [GET /api/redeem/orders](#api-redeem-orders-list)
    - √ [GET /api/redeem/orders/{id}](#api-redeem-orders-get)
    - √ [PUT /api/redeem/orders/{id}/cancel](#api-redeem-orders-cancel)
  - × [Admin 模块（管理员：登录/审计/关键操作）](#module-admin)
    - × [POST /admin/auth/login](#api-admin-auth-login)
    - × [GET /admin/auth/me](#api-admin-auth-me)
    - × [POST /admin/auth/logout](#api-admin-auth-logout)
    - √ [GET /admin/audit/logs](#api-admin-audit-logs)
    - × [POST /admin/points/adjust](#api-admin-points-adjust)
    - × [PUT /admin/users/{id}/drinks/use](#api-admin-users-drinks-use)
    - × [POST /admin/tournaments/{id}/results/publish](#api-admin-tournament-results-publish)
    - × [POST /admin/tournaments/{id}/awards/grant](#api-admin-tournament-awards-grant)
  - × [Media 模块（媒体上传/访问）](#module-media)
    - × [POST /admin/media/upload](#api-admin-media-upload)

## 已注册路由清单（与 cmd/server/main.go#L127-L164 一致）

说明：

- 这一表格展示“当前已注册路由”，与代码注册保持一致；“完成”列用于标识该路由的业务实现完成度。
- “详情”列可直接跳转到该接口的详细说明段落。

| 完成 | 模块 | Method | Path | 详情 |
|---|---|---|---|---|
| √ | 健康检查 | GET | /health | [GET /health](#api-health) |
| √ | Auth（小程序登录） | POST | /api/auth/wechat/login | [POST /api/auth/wechat/login](#api-auth-wechat-login) |
| √ | Item（管理员：积分商品） | POST | /admin/goods | [POST /admin/goods](#api-admin-goods-create) |
| √ | Item（管理员：积分商品） | GET | /admin/goods | [GET /admin/goods](#api-admin-goods-list) |
| √ | Item（管理员：积分商品） | GET | /admin/goods/{id} | [GET /admin/goods/{id}](#api-admin-goods-get) |
| √ | Item（管理员：积分商品） | PUT | /admin/goods/{id} | [PUT /admin/goods/{id}](#api-admin-goods-update) |
| √ | Item（管理员：积分商品） | DELETE | /admin/goods/{id} | [DELETE /admin/goods/{id}](#api-admin-goods-delete) |
| √ | Tournament（管理员：赛事） | POST | /admin/tournaments | [POST /admin/tournaments](#api-admin-tournaments-create) |
| √ | Tournament（管理员：赛事） | GET | /admin/tournaments | [GET /admin/tournaments](#api-admin-tournaments-list) |
| √ | Tournament（管理员：赛事） | GET | /admin/tournaments/{id} | [GET /admin/tournaments/{id}](#api-admin-tournaments-get) |
| √ | Tournament（管理员：赛事） | PUT | /admin/tournaments/{id} | [PUT /admin/tournaments/{id}](#api-admin-tournaments-update) |
| √ | Tournament（管理员：赛事） | DELETE | /admin/tournaments/{id} | [DELETE /admin/tournaments/{id}](#api-admin-tournaments-delete) |
| √ | Task（管理员：任务定义） | POST | /admin/task-defs | [POST /admin/task-defs](#api-admin-task-defs-create) |
| √ | Task（管理员：任务定义） | GET | /admin/task-defs | [GET /admin/task-defs](#api-admin-task-defs-list) |
| √ | Task（管理员：任务定义） | GET | /admin/task-defs/{id} | [GET /admin/task-defs/{id}](#api-admin-task-defs-get) |
| √ | Task（管理员：任务定义） | PUT | /admin/task-defs/{id} | [PUT /admin/task-defs/{id}](#api-admin-task-defs-update) |
| √ | Task（管理员：任务定义） | DELETE | /admin/task-defs/{id} | [DELETE /admin/task-defs/{id}](#api-admin-task-defs-delete) |
| √ | User（管理员：用户） | GET | /admin/users | [GET /admin/users](#api-admin-users-list) |
| √ | User（管理员：用户） | GET | /admin/users/{id} | [GET /admin/users/{id}](#api-admin-users-get) |
| √ | User（管理员：用户） | PUT | /admin/users/{id} | [PUT /admin/users/{id}](#api-admin-users-update) |
| √ | Redeem（管理员：兑换订单） | POST | /admin/redeem/orders | [POST /admin/redeem/orders](#api-admin-redeem-orders-create) |
| √ | Redeem（管理员：兑换订单） | GET | /admin/redeem/orders | [GET /admin/redeem/orders](#api-admin-redeem-orders-list) |
| √ | Redeem（管理员：兑换订单） | GET | /admin/redeem/orders/{id} | [GET /admin/redeem/orders/{id}](#api-admin-redeem-orders-get) |
| √ | Redeem（管理员：兑换订单） | PUT | /admin/redeem/orders/{id}/use | [PUT /admin/redeem/orders/{id}/use](#api-admin-redeem-orders-use) |
| √ | Redeem（管理员：兑换订单） | PUT | /admin/redeem/orders/{id}/cancel | [PUT /admin/redeem/orders/{id}/cancel](#api-admin-redeem-orders-cancel) |
| √ | User（小程序：个人资料） | GET | /api/users/me | [GET /api/users/me](#api-users-me-get) |
| √ | User（小程序：个人资料） | PUT | /api/users/me | [PUT /api/users/me](#api-users-me-update) |
| √ | Item（小程序：积分商品） | GET | /api/goods | [GET /api/goods](#api-goods-list) |
| √ | Item（小程序：积分商品） | GET | /api/goods/{id} | [GET /api/goods/{id}](#api-goods-get) |
| √ | Tournament（小程序：赛事） | GET | /api/tournaments | [GET /api/tournaments](#api-tournaments-list) |
| √ | Tournament（小程序：赛事） | GET | /api/tournaments/{id} | [GET /api/tournaments/{id}](#api-tournaments-get) |
| × | Tournament（小程序：赛事） | POST | /api/tournaments/{id}/join | [POST /api/tournaments/{id}/join](#api-tournaments-join) |
| × | Tournament（小程序：赛事） | PUT | /api/tournaments/{id}/cancel | [PUT /api/tournaments/{id}/cancel](#api-tournaments-cancel) |
| × | Tournament（小程序：赛事） | GET | /api/tournaments/{id}/results | [GET /api/tournaments/{id}/results](#api-tournaments-results) |
| √ | Redeem（小程序：兑换订单） | GET | /api/redeem/orders | [GET /api/redeem/orders](#api-redeem-orders-list) |
| √ | Redeem（小程序：兑换订单） | POST | /api/redeem/orders | [POST /api/redeem/orders](#api-redeem-orders-create) |
| √ | Redeem（小程序：兑换订单） | GET | /api/redeem/orders/{id} | [GET /api/redeem/orders/{id}](#api-redeem-orders-get) |
| √ | Redeem（小程序：兑换订单） | PUT | /api/redeem/orders/{id}/cancel | [PUT /api/redeem/orders/{id}/cancel](#api-redeem-orders-cancel) |
| √ | Points（小程序：积分） | GET | /api/points/balance | [GET /api/points/balance](#api-points-balance) |
| √ | Points（小程序：积分） | GET | /api/points/ledgers | [GET /api/points/ledgers](#api-points-ledgers) |
| √ | VIP（小程序：会员） | GET | /api/vip/status | [GET /api/vip/status](#api-vip-status) |
| √ | Task（小程序：任务） | GET | /api/tasks | [GET /api/tasks](#api-tasks-list) |
| × | Task（小程序：任务） | POST | /api/tasks/checkin | [POST /api/tasks/checkin](#api-tasks-checkin) |
| × | Task（小程序：任务） | POST | /api/tasks/{taskCode}/claim | [POST /api/tasks/{taskCode}/claim](#api-tasks-claim) |
| × | Admin（管理员） | POST | /admin/auth/login | [POST /admin/auth/login](#api-admin-auth-login) |
| × | Admin（管理员） | GET | /admin/auth/me | [GET /admin/auth/me](#api-admin-auth-me) |
| × | Admin（管理员） | POST | /admin/auth/logout | [POST /admin/auth/logout](#api-admin-auth-logout) |
| √ | Admin（管理员） | GET | /admin/audit/logs | [GET /admin/audit/logs](#api-admin-audit-logs) |
| × | Admin（管理员） | POST | /admin/points/adjust | [POST /admin/points/adjust](#api-admin-points-adjust) |
| × | Admin（管理员） | PUT | /admin/users/{id}/drinks/use | [PUT /admin/users/{id}/drinks/use](#api-admin-users-drinks-use) |
| × | Admin（管理员） | POST | /admin/tournaments/{id}/results/publish | [POST /admin/tournaments/{id}/results/publish](#api-admin-tournament-results-publish) |
| × | Admin（管理员） | POST | /admin/tournaments/{id}/awards/grant | [POST /admin/tournaments/{id}/awards/grant](#api-admin-tournament-awards-grant) |
| × | Media（管理员） | POST | /admin/media/upload | [POST /admin/media/upload](#api-admin-media-upload) |

## 0. 通用约定

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

- 目前接口未接入统一的鉴权中间件；`/admin/*` 管理端接口也暂未做 token 校验。
- 小程序登录接口会返回 `token`，用于后续接入鉴权时作为凭证。

---

## module-health
健康检查模块 √

### api-health
GET /health √

用途：用于部署探活与检查服务是否存活。

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

用途：小程序登录。前端传 `wx.login()` 得到的 `code`，后端通过 code2session 换取 openid/unionid，并创建/更新用户后签发 token。

请求：

- Method：`POST`
- Path：`/api/auth/wechat/login`
- Body：JSON

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| code | string | 是 | `wx.login()` 返回的临时登录凭证 |

请求示例：

```bash
curl -X POST "http://localhost:8080/api/auth/wechat/login" \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"wx_login_code_here\"}"
```

成功响应 `data` 字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| token | string | 访问 token（当前仅签发，未被路由鉴权使用） |
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

失败场景示例（code 为空）：

```json
{
  "code": 201,
  "message": "code 不能为空"
}
```

---

## module-item
Item 模块（管理员：积分商品管理） √

### api-admin-goods-create
POST /admin/goods √

用途：创建积分商品（goods）。

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

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| nickname | string | 否 | 昵称（可为空字符串） |
| avatarUrl | string | 否 | 头像 URL（可为空字符串） |
| status | number | 否 | 1=正常，0=封禁 |

注意：

- 当前实现中，如果 `status` 解析为 0，会被后端默认写入 1（因此“封禁=0”在当前实现下不会生效）。

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

### api-users-me-update
PUT /api/users/me √

用途：更新当前登录用户的个人资料（昵称、头像等）。

---

## module-points
Points 模块（小程序：积分账户与流水） √

### api-points-balance
GET /api/points/balance √

用途：获取当前登录用户的积分余额。

### api-points-ledgers
GET /api/points/ledgers √

用途：获取当前登录用户的积分流水列表。

---

## module-vip
VIP 模块（小程序：会员订阅） √

### api-vip-status
GET /api/vip/status √

用途：获取当前登录用户的会员状态（是否会员、到期时间等）。

---

## module-tournament-app
Tournament 模块（小程序：赛事） ×

### api-tournaments-list
GET /api/tournaments √

用途：赛事列表。

### api-tournaments-get
GET /api/tournaments/{id} √

用途：赛事详情。

### api-tournaments-join
POST /api/tournaments/{id}/join ×

用途：报名参赛。

### api-tournaments-cancel
PUT /api/tournaments/{id}/cancel ×

用途：取消报名。

### api-tournaments-results
GET /api/tournaments/{id}/results ×

用途：查看赛事排名/成绩。

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

### api-redeem-orders-create
POST /api/redeem/orders √

用途：创建兑换订单（扣减积分并生成订单号）。

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

### api-admin-auth-me
GET /admin/auth/me ×

用途：获取当前管理员信息（用于后台鉴权/展示）。

### api-admin-auth-logout
POST /admin/auth/logout ×

用途：管理员登出（可选：token 失效/黑名单）。

### api-admin-audit-logs
GET /admin/audit/logs √

用途：查询管理员关键操作审计日志。

### api-admin-points-adjust
POST /admin/points/adjust ×

用途：管理员给用户赠送/扣减积分（需落审计与流水）。

### api-admin-users-drinks-use
PUT /admin/users/{id}/drinks/use ×

用途：管理员核销用户饮品数量（点一次 -1）。

### api-admin-tournament-results-publish
POST /admin/tournaments/{id}/results/publish ×

用途：发布赛事排名（tournament_result）并记录发布人。

### api-admin-tournament-awards-grant
POST /admin/tournaments/{id}/awards/grant ×

用途：给赛事前 N 名发积分奖励（需幂等，避免重复发奖）。

---

## module-media
Media 模块（媒体上传/访问） ×

### api-admin-media-upload
POST /admin/media/upload ×

用途：上传封面/图片等媒体文件，返回可访问 URL。

---

## module-unimplemented
小程序侧接口汇总（部分未完成） ×

本段用于保留一个“规划接口”总入口；具体接口已拆分到上面的模块段落（User/Points/VIP/Tournament/Task/Item/Redeem/Admin/Media）。

