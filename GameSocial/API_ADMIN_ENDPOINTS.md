# GameSocial 后台管理端接口（Admin）

本文档整理后台管理端（`/admin/*`）接口说明，包含：用途、实现位置、设计思路、请求格式、响应格式。已实现接口标注 `√`，未实现标注 `×`。

快速入口：

- 客户端接口：[API_CLIENT_ENDPOINTS.md](API_CLIENT_ENDPOINTS.md)
- 总览：[API_ALL_ENDPOINTS.md](API_ALL_ENDPOINTS.md)

## 接口目录

- √ [健康检查模块](#module-health)
  - √ [GET /health](#api-health)
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
  - √ [POST /admin/media/upload](#api-admin-media-upload)

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

## module-item
Item 模块（管理员：积分商品管理） √

### api-admin-goods-create
POST /admin/goods √

用途：创建积分商品（goods）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L154-L160)
- Handler：[AdminGoodsCreate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go#L12-L43)
- Service：[item.CreateGoods](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go#L66-L99)

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
- Handler：[AdminGoodsList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go#L153-L187)
- Service：[item.ListGoods](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go#L197-L250)

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
- Handler：[AdminGoodsGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go#L119-L151)
- Service：[item.GetGoods](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go#L168-L195)

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
- Handler：[AdminGoodsUpdate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go#L45-L84)
- Service：[item.UpdateGoods](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go#L101-L139)

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
- Handler：[AdminGoodsDelete](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_goods.go#L86-L117)
- Service：[item.DeleteGoods](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go#L141-L166)

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
- Handler：[AdminTournamentCreate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go#L12-L43)
- Service：[tournament.Create](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L96-L131)

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
- Handler：[AdminTournamentList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go#L150-L183)
- Service：[tournament.List](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L227-L279)

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
- Handler：[AdminTournamentGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go#L117-L148)
- Service：[tournament.Get](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L197-L225)

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
- Handler：[AdminTournamentUpdate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go#L45-L83)
- Service：[tournament.Update](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L133-L169)

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
- Handler：[AdminTournamentDelete](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_tournaments.go#L85-L115)
- Service：[tournament.Delete](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L171-L195)

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
- Handler：[AdminTaskDefCreate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go#L12-L42)
- Service：[task.CreateTaskDef](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go#L72-L114)

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
- Handler：[AdminTaskDefList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go#L149-L182)
- Service：[task.ListTaskDef](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go#L218-L274)

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
- Handler：[AdminTaskDefGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go#L116-L147)
- Service：[task.GetTaskDef](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go#L187-L216)

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
- Handler：[AdminTaskDefUpdate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go#L44-L82)
- Service：[task.UpdateTaskDef](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go#L116-L159)

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
- Handler：[AdminTaskDefDelete](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_task_defs.go#L84-L114)
- Service：[task.DeleteTaskDef](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go#L161-L185)

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
- Handler：[AdminUserList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_users.go#L45-L78)
- Service：[user.List](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go#L85-L135)

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
| level | number | 用户等级（默认 1） |
| exp | number | 经验值（累积） |
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
      "level": 1,
      "exp": 10,
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
      "level": 2,
      "exp": 120,
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
- Handler：[AdminUserGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_users.go#L12-L43)
- Service：[user.Get](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go#L54-L83)

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
    "level": 2,
    "exp": 120,
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
- Handler：[AdminUserUpdate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_users.go#L80-L118)
- Service：[user.Update](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go#L134-L160)

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
    "level": 2,
    "exp": 120,
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
- Handler：[AdminRedeemOrderCreate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go#L12-L42)
- Service：[redeem.CreateOrder](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L76-L146)

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
- Handler：[AdminRedeemOrderList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go#L77-L116)
- Service：[redeem.ListOrders](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L214-L274)

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
- Handler：[AdminRedeemOrderGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go#L44-L75)
- Service：[redeem.GetOrder](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L147-L212)

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
- Handler：[AdminRedeemOrderUse](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go#L118-L160)
- Service：[redeem.UseOrder](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L276-L302)

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
- Handler：[AdminRedeemOrderCancel](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_redeem_orders.go#L162-L193)
- Service：[redeem.CancelOrder](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L304-L333)

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

## module-admin
Admin 模块（管理员：登录/审计/关键操作） ×

### api-admin-auth-login
POST /admin/auth/login ×

用途：管理员登录并签发管理员 token。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[AdminAuthLogin](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L8-L20)

实现逻辑：

1. 校验方法为 `POST`。
2. 当前为占位实现：直接返回固定 token（`admin_token_placeholder`）。

### api-admin-auth-me
GET /admin/auth/me ×

用途：获取当前管理员信息（用于后台鉴权/展示）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[AdminAuthMe](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L22-L35)

实现逻辑：

1. 校验方法为 `GET`。
2. 当前为占位实现：直接返回固定管理员信息（`id=1, username=admin`）。

### api-admin-auth-logout
POST /admin/auth/logout ×

用途：管理员登出（可选：token 失效/黑名单）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[AdminAuthLogout](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L37-L47)

实现逻辑：

1. 校验方法为 `POST`。
2. 当前为占位实现：直接返回 `logout=true`。

### api-admin-audit-logs
GET /admin/audit/logs √

用途：查询管理员关键操作审计日志。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[AdminAuditLogs](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_audit_logs.go#L22-L91)

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
- Handler：[AdminPointsAdjust](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L49-L59)

实现逻辑：

1. 校验方法为 `POST`。
2. 当前为占位实现：直接返回 `adjusted=true`。

### api-admin-users-drinks-use
PUT /admin/users/{id}/drinks/use ×

用途：管理员核销用户饮品数量（点一次 -1）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[AdminUsersDrinksUse](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L61-L72)

实现逻辑：

1. 校验方法为 `PUT`。
2. 解析 path 参数 `id`（当前未做业务校验）。
3. 当前为占位实现：直接返回 `used=true`。

### api-admin-tournament-results-publish
POST /admin/tournaments/{id}/results/publish ×

用途：发布赛事排名（tournament_result）并记录发布人。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[AdminTournamentResultsPublish](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L74-L85)

实现逻辑：

1. 校验方法为 `POST`。
2. 解析 path 参数 `id`（当前未做业务校验）。
3. 当前为占位实现：直接返回 `published=true`。

### api-admin-tournament-awards-grant
POST /admin/tournaments/{id}/awards/grant ×

用途：给赛事前 N 名发积分奖励（需幂等，避免重复发奖）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L187-L196)
- Handler：[AdminTournamentAwardsGrant](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L87-L98)

实现逻辑：

1. 校验方法为 `POST`。
2. 解析 path 参数 `id`（当前未做业务校验）。
3. 当前为占位实现：直接返回 `granted=true`。

---

## module-media
Media 模块（媒体上传/访问） √

### api-admin-media-upload
POST /admin/media/upload √

用途：上传封面/图片等媒体文件，返回可访问 URL。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L210-L219)
- Handler：[AdminMediaUpload](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/admin_misc.go#L105-L126)
- Store：[COSStore](file:///e:/VUE3/新建文件夹/GameSocial/internal/media/store.go#L33-L124)

实现逻辑：

1. 校验方法为 `POST`。
2. 解析 multipart 表单中的 `file` 文件字段。
3. 校验文件类型为 `image/*`，并按配置限制文件大小。
4. 上传到腾讯云 COS，返回 `url/key/createdAt`。

请求：

- Method：`POST`
- Path：`/admin/media/upload`
- Content-Type：`multipart/form-data`
- Form：
  - `file`：图片文件（必须；仅允许 `image/*`）

成功响应 `data`：

| 字段 | 类型 | 说明 |
|---|---|---|
| url | string | 可访问 URL |
| key | string | COS 对象 key |
| createdAt | string | 创建时间（RFC3339） |
