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
- √ [Media 模块（小程序：临时直传凭证）](#module-media-app)
  - √ [POST /api/media/temp-upload-infos](#api-media-temp-upload-infos)
- √ [Points 模块（小程序：积分账户与流水）](#module-points)
  - √ [GET /api/points/balance](#api-points-balance)
  - √ [GET /api/points/ledgers](#api-points-ledgers)
- √ [VIP 模块（小程序：会员订阅）](#module-vip)
  - √ [GET /api/vip/status](#api-vip-status)
- √ [Tournament 模块（小程序：赛事）](#module-tournament-app)
  - √ [GET /api/tournaments](#api-tournaments-list)
  - √ [GET /api/tournaments/joined](#api-tournaments-joined)
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

- 请求：默认 `Content-Type: application/json`；涉及图片上传的接口会使用 `multipart/form-data`
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

### 0.7 数据库升级（9527：遇到 Unknown column 必看）

如果前端调用赛事/商品相关接口时出现：

```json
{
  "code": 201,
  "message": "Error 1054 (42S22): Unknown column 't.image_urls_json' in 'field list'"
}
```

说明后端 SQL 已经在读取 `image_urls_json`（用于多图），但你的数据库还是旧表结构。

在目标库执行以下 SQL 补齐字段（不要整库 DROP 重建）：

```sql
ALTER TABLE tournament
  ADD COLUMN image_urls_json JSON NULL COMMENT '赛事图片 URL 列表 JSON（可为空）' AFTER cover_url;

ALTER TABLE goods
  ADD COLUMN image_urls_json JSON NULL COMMENT '商品图片 URL 列表 JSON（可为空）' AFTER cover_url;
```

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

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L129-L146)
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

## module-user-app
User 模块（小程序：个人资料） √

### api-users-me-get
GET /api/users/me √

用途：获取当前登录用户的个人资料（昵称、头像等）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L133-L136)
- Handler：[AppUserMeGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_users.go#L10-L41)
- Service：[user.Get](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go#L55-L80)

请求头：

- `Authorization: Bearer <token>`

成功响应 `data`：个人资料对象

| 字段 | 类型 | 说明 |
|---|---|---|
| nickname | string | 昵称（为空表示未设置） |
| avatarUrl | string | 头像 URL（为空表示未设置） |
| level | number | 用户等级（默认 1） |
| exp | number | 经验值（累积） |
| createdAt | string | 注册时间（RFC3339） |

请求示例：

```bash
curl -X GET "http://localhost:8080/api/users/me" \
  -H "Authorization: Bearer <token>"
```

响应示例：

```json
{
  "code": 200,
  "data": {
    "nickname": "小明",
    "avatarUrl": "",
    "level": 1,
    "exp": 0,
    "createdAt": "2026-02-05T04:48:47Z"
  },
  "message": "ok"
}
```

### api-users-me-update
PUT /api/users/me √

用途：更新当前登录用户的个人资料（昵称、头像等）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L133-L136)
- Handler：[AppUserMeUpdate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_users.go#L43-L87)
- Service：[user.Update](file:///e:/VUE3/新建文件夹/GameSocial/modules/user/service.go#L134-L160)

请求头：

- `Authorization: Bearer <token>`

请求体支持两种格式：

1) `application/json`：仅更新文字字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| nickname | string | 否 | 昵称（可为空字符串） |
| avatarUrl | string | 否 | 头像：可传 URL；也可传 base64 图片数据（data URL 或纯 base64），服务端会上传并写入 URL |

请求示例：

```bash
curl -X PUT "http://localhost:8080/api/users/me" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d "{\"nickname\":\"9527\",\"avatarUrl\":\"data:image/png;base64,AAAA...\"}"
```

2) `multipart/form-data`：提交表单并在同一个请求里上传头像（仅当用户确认保存资料时才上传；头像 URL 由服务端根据上传结果写入）

注意：这种“把 file 直接传给后端”的方式，只有在后端启用了「服务端 COS SDK 上传」时可用；如果你希望前端直传图片到 COS，请先调用 [POST /api/media/temp-upload-infos](#api-media-temp-upload-infos) 拿到 `downloadUrl`，再把该 URL 回填到 `avatarUrl`。

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| nickname | string | 否 | 昵称（可为空字符串） |
| file | file | 否 | 头像图片文件（仅允许 `image/*`） |

请求示例：

```bash
curl -X PUT "http://localhost:8080/api/users/me" \
  -H "Authorization: Bearer <token>" \
  -F "nickname=9527" \
  -F "file=@./avatar.png"
```

成功响应 `data`：个人资料对象（同 GET）

---

## module-media-app
Media 模块（小程序：临时直传凭证） √

### api-media-temp-upload-infos
POST /api/media/temp-upload-infos √

用途：让前端在“用户确认提交之前”先把图片直传到后端指定的临时目录（`temp/...`），避免用户取消导致正式图片目录被占用。

特点：

- 前端一次申请 N 个上传凭证（最多 10 个），后端返回每个 objectId 对应的直传信息
- 仅允许已登录用户调用（必须带 `Authorization: Bearer <token>`）
- 后端为每个 `objectId` 计算一次 PUT 签名（有效期 3 分钟）；前端必须在有效期内完成上传
- `objectId` 强制固定在 `temp/...` 目录下，避免前端绕过目录隔离写入正式目录

实现位置：

- 路由：[main.go](file:///w:/GOProject/gamesocial/GameSocial/cmd/server/main.go)
- Handler：[AppMediaTempUploadInfos](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/app_users.go)
- Media：COS 直传签名生成（[COSStore.GetObjectsUploadInfo](file:///w:/GOProject/gamesocial/GameSocial/internal/media/store.go)）

请求头：

- `Authorization: Bearer <token>`
- `Content-Type: application/json`

请求体字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| count | number | 是 | 申请的上传“坑位”数量，范围 1~10 |
| contentType | string | 否 | 图片类型，默认 `image/png` |
| scene | string | 否 | 场景目录：`goods`/`tournament`/`user`/`common`，默认 `common` |

前端对接要点：

- 后端不接收 `items` 入参；需要由前端先算出 `count=待上传图片数量`
- `contentType` 建议同一批次保持一致（例如全是 `image/png`）；如果同一批图片类型不同，建议按类型分批调用本接口

常见错误（会返回 `count 参数错误`）：

```json
{
  "items": [
    { "contentType": "image/png" }
  ]
}
```

正确请求（把 `items.length` 转成 `count`，并把 `contentType` 提到顶层）：

```json
{
  "count": 1,
  "contentType": "image/png",
  "scene": "tournament"
}
```

请求示例（9527 一次申请 3 张赛事临时图）：

```bash
curl -X POST "http://localhost:8080/api/media/temp-upload-infos" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d "{\"count\":3,\"contentType\":\"image/png\",\"scene\":\"tournament\"}"
```

成功响应 `data` 字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| sessionId | string | 本次临时上传会话 ID（建议前端保存，用于后续“提交/绑定”） |
| scene | string | 实际使用的场景目录 |
| items | array | 每个坑位的直传信息数组 |

`items[]` 字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| objectId | string | 目标对象路径（已固定在 `temp/...`） |
| uploadUrl | string | 直传 PUT URL（不带签名 query） |
| downloadUrl | string | 上传完成后可访问的 URL（可能为空时由后端兜底拼接） |
| authorization | string | PUT 请求 `Authorization` 值（必填） |
| token | string | 预留字段（当前为空） |
| cloudObjectMeta | string | 预留字段（当前为空） |

响应示例（截断）：

```json
{
  "code": 200,
  "data": {
    "sessionId": "e3b0c44298fc1c149afbf4c8996fb924",
    "scene": "tournament",
    "items": [
      {
        "objectId": "temp/tournament/u1/e3b0c44298fc1c149afbf4c8996fb924/20260208/2aa0...f3.png",
        "uploadUrl": "https://<bucket>.cos.<region>.myqcloud.com/temp/tournament/u1/e3b0c44298fc1c149afbf4c8996fb924/20260208/2aa0...f3.png",
        "downloadUrl": "https://.../temp/...",
        "authorization": "q-sign-algorithm=sha1&q-ak=...&q-sign-time=...&q-key-time=...&q-header-list=host&q-url-param-list=&q-signature=...",
        "token": "",
        "cloudObjectMeta": ""
      }
    ]
  },
  "message": "ok"
}
```

前端 PUT 上传时的请求头（对每个 items[i] 执行一次 PUT）：

- `Authorization: <authorization>`
- `Content-Type: <contentType>`

备注：

- 需要在 COS 控制台配置 CORS，至少允许 `PUT`，并放行请求头 `Authorization` 与 `Content-Type`

---

## module-points
Points 模块（小程序：积分账户与流水） √

### api-points-balance
GET /api/points/balance √

用途：获取当前登录用户的积分余额。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L143-L149)
- Handler：[AppPointsBalance](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_points.go#L21-L57)

请求头：

- `Authorization: Bearer <token>`

### api-points-ledgers
GET /api/points/ledgers √

用途：获取当前登录用户的积分流水列表。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L147-L149)
- Handler：[AppPointsLedgers](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_points.go#L59-L121)

请求头：

- `Authorization: Bearer <token>`

---

## module-vip
VIP 模块（小程序：会员订阅） √

### api-vip-status
GET /api/vip/status √

用途：获取当前登录用户的会员状态（是否会员、到期时间等）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L147-L150)
- Handler：[AppVipStatus](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_vip.go#L9-L54)

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

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L138-L143)
- Handler：[AppTournamentsList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tournaments.go#L10-L39)
- Service：[tournament.List](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L228-L280)

### api-tournaments-joined
GET /api/tournaments/joined √

用途：查询当前登录用户已报名参加的赛事列表（按报名时间倒序）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L138-L143)
- Handler：[AppTournamentsJoined](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tournaments.go#L41-L78)
- Service：[tournament.ListJoined](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L295-L357)

请求：

- Method：`GET`
- Path：`/api/tournaments/joined`
- Query：
  - `offset`：默认 0
  - `limit`：默认 20，最大 200
  - `status`：可选，过滤赛事状态（例如 PUBLISHED/FINISHED）
  - `q`：可选，按赛事标题模糊搜索

响应 `data`：赛事列表（每项包含赛事字段 + 报名信息）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 赛事 ID |
| title | string | 标题 |
| content | string | 详情（可能为空） |
| coverUrl | string | 封面 URL（可能为空） |
| startAt | string | 开始时间（RFC3339） |
| endAt | string | 结束时间（RFC3339） |
| status | string | 赛事状态 |
| createdByAdminId | number | 创建人管理员 ID |
| createdAt | string | 创建时间（RFC3339） |
| updatedAt | string | 更新时间（RFC3339） |
| joinStatus | string | 报名状态（JOINED） |
| joinedAt | string | 报名时间（RFC3339） |

请求示例：

```bash
curl -X GET "http://localhost:8080/api/tournaments/joined?offset=0&limit=20&q=周末" \
  -H "Authorization: Bearer <token>"
```

响应示例：

```json
{
  "code": 200,
  "data": [
    {
      "id": 4001,
      "title": "周末友谊赛",
      "content": "周末店内友谊赛，欢迎报名",
      "coverUrl": "",
      "startAt": "2026-02-01T06:00:00Z",
      "endAt": "2026-02-01T10:00:00Z",
      "status": "PUBLISHED",
      "createdByAdminId": 1,
      "createdAt": "2026-02-01T00:00:00Z",
      "updatedAt": "2026-02-01T00:00:00Z",
      "joinStatus": "JOINED",
      "joinedAt": "2026-02-05T04:48:47Z"
    }
  ],
  "message": "ok"
}
```

### api-tournaments-get
GET /api/tournaments/{id} √

用途：赛事详情。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L138-L143)
- Handler：[AppTournamentsGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tournaments.go#L80-L105)
- Service：[tournament.Get](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L198-L227)

### api-tournaments-join
POST /api/tournaments/{id}/join √

用途：当前登录用户报名参加指定赛事。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L138-L143)
- Handler：[AppTournamentsJoin](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tournaments.go#L107-L138)
- Service：[tournament.Join](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L281-L310)

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

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L138-L143)
- Handler：[AppTournamentsCancel](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tournaments.go#L140-L171)
- Service：[tournament.Cancel](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L311-L336)

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

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L138-L143)
- Handler：[AppTournamentsResults](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tournaments.go#L173-L204)
- Service：[tournament.GetResults](file:///e:/VUE3/新建文件夹/GameSocial/modules/tournament/service.go#L338-L415)

说明：携带 `Authorization: Bearer <token>` 时，后端从 token 解析当前用户并返回 `my` 字段。

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

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L150-L152)
- Handler：[AppTasksList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tasks.go#L10-L39)
- Service：[task.ListTaskDef](file:///e:/VUE3/新建文件夹/GameSocial/modules/task/service.go#L218-L274)

### api-tasks-checkin
POST /api/tasks/checkin ×

用途：到店打卡（写入 checkin_log 并推动任务进度）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L150-L152)
- Handler：[AppTasksCheckin](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tasks.go#L41-L51)

### api-tasks-claim
POST /api/tasks/{taskCode}/claim ×

用途：领取任务奖励（需幂等，避免重复发奖）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L150-L152)
- Handler：[AppTasksClaim](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_tasks.go#L53-L64)

---

## module-item-app
Item 模块（小程序：积分商品） √

### api-goods-list
GET /api/goods √

用途：商品列表（用户侧展示）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L134-L138)
- Handler：[AppGoodsList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_goods.go#L10-L38)
- Service：[item.ListGoods](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go#L197-L250)

### api-goods-get
GET /api/goods/{id} √

用途：商品详情（用户侧展示）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L136-L138)
- Handler：[AppGoodsGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_goods.go#L40-L65)
- Service：[item.GetGoods](file:///e:/VUE3/新建文件夹/GameSocial/modules/item/service.go#L168-L195)

---

## module-redeem-app
Redeem 模块（小程序：兑换订单） √

请求头（以下接口通用）：

- `Authorization: Bearer <token>`

### api-redeem-orders-create
POST /api/redeem/orders √

用途：创建兑换订单（userId 从 token 获取；扣减积分并生成订单号）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L143-L146)
- Handler：[AppRedeemOrderCreate](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_redeem.go#L11-L44)
- Service：[redeem.CreateOrder](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L76-L146)

### api-redeem-orders-list
GET /api/redeem/orders √

用途：查询我的兑换订单列表。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L143-L146)
- Handler：[AppRedeemOrderList](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_redeem.go#L46-L82)
- Service：[redeem.ListOrders](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L212-L274)

### api-redeem-orders-get
GET /api/redeem/orders/{id} √

用途：查询我的兑换订单详情（含 items）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L143-L146)
- Handler：[AppRedeemOrderGet](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_redeem.go#L84-L115)
- Service：[redeem.GetOrder](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L148-L210)

### api-redeem-orders-cancel
PUT /api/redeem/orders/{id}/cancel √

用途：取消我的兑换订单（仅允许 CREATED -> CANCELED）。

实现位置：

- 路由：[main.go](file:///e:/VUE3/新建文件夹/GameSocial/cmd/server/main.go#L143-L146)
- Handler：[AppRedeemOrderCancel](file:///e:/VUE3/新建文件夹/GameSocial/api/handlers/app_redeem.go#L117-L148)
- Service：[redeem.CancelOrder](file:///e:/VUE3/新建文件夹/GameSocial/modules/redeem/service.go#L302-L336)
