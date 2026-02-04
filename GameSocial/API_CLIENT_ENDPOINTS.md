# GameSocial 客户端接口（API）

本文档整理客户端（小程序/APP，`/api/*`）接口说明，包含：用途、实现位置、设计思路、请求格式、响应格式。已实现接口标注 `√`，未实现标注 `×`。

快速入口：

- 后台管理端接口：[API_ADMIN_ENDPOINTS.md](API_ADMIN_ENDPOINTS.md)
- 总览：[API_ALL_ENDPOINTS.md](API_ALL_ENDPOINTS.md)

## 接口目录

- √ [健康检查模块](#module-health)
  - √ [GET /health](#api-health)
- √ [Auth 模块（小程序登录）](#module-auth)
  - √ [POST /api/auth/wechat/login](#api-auth-wechat-login)
- √ [User 模块（小程序：个人资料）](#module-user-app)
  - √ [GET /api/users/me](#api-users-me-get)
  - √ [PUT /api/users/me](#api-users-me-update)
- √ [Points 模块（小程序：积分账户与流水）](#module-points)
  - √ [GET /api/points/balance](#api-points-balance)
  - √ [GET /api/points/ledgers](#api-points-ledgers)
- √ [VIP 模块（小程序：会员订阅）](#module-vip)
  - √ [GET /api/vip/status](#api-vip-status)
- √ [Tournament 模块（小程序：赛事）](#module-tournament-app)
  - √ [GET /api/tournaments](#api-tournaments-list)
  - √ [GET /api/tournaments/{id}](#api-tournaments-get)
  - √ [POST /api/tournaments/{id}/join](#api-tournaments-join)
  - √ [PUT /api/tournaments/{id}/cancel](#api-tournaments-cancel)
  - √ [GET /api/tournaments/{id}/results](#api-tournaments-results)
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

- 小程序端需要登录的接口会解析 `Authorization: Bearer <token>`，从 token 的 `sub` 字段得到 `userId`；不需要再额外传 `userId` 参数。
- `/admin/*` 管理端接口仍暂未接入 token 校验。

---

## module-health
健康检查模块 √

### api-health
GET /health √

用途：用于部署探活与检查服务是否存活。

实现位置：

- 路由：[main.go](file:///w:/GOProject/gamesocial/GameSocial/cmd/server/main.go#L129-L133)
- Handler：[health.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/health.go#L1-L21)

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

- 路由：[main.go](file:///w:/GOProject/gamesocial/GameSocial/cmd/server/main.go#L129-L142)
- Handler：[wechat.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/wechat.go#L1-L57)
- Middleware：解析 `Authorization: Bearer <token>` 写入 `X-User-Id`（[middleware.go](file:///w:/GOProject/gamesocial/GameSocial/api/middleware/middleware.go#L52-L83)）
- Service：`auth.Service.OpenIDLogin`（[service.go](file:///w:/GOProject/gamesocial/GameSocial/modules/auth/service.go)）

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

## module-user-app
User 模块（小程序：个人资料） √

### api-users-me-get
GET /api/users/me √

用途：获取当前登录用户的个人资料（昵称、头像等）。

请求头：

- `Authorization: Bearer <token>`

### api-users-me-update
PUT /api/users/me √

用途：更新当前登录用户的个人资料（昵称、头像等）。

请求头：

- `Authorization: Bearer <token>`

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
