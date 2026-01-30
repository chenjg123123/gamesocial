# GameSocial 接口清单（进度 + 设计）

统一响应包裹：

```json
{
  "code": 200,
  "data": {},
  "message": ""
}
```

## 目录

| 接口 | 状态 |
|---|---|
| [GET /health](#get-health) | ✅ |
| [POST /api/auth/wechat/login](#post-apiauthwechatlogin) | ❌ |
| [POST /admin/auth/login](#post-adminauthlogin) | ❌ |
| [GET /api/debug/users](#get-apidebugusers) | ✅ |
| [GET /api/goods](#get-apigoods) | ✅ |
| [GET /api/goods/{id}](#get-apigoodsid) | ✅ |
| [POST /admin/goods](#post-admingoods) | ✅ |
| [PUT /admin/goods/{id}](#put-admingoodsid) | ✅ |
| [DELETE /admin/goods/{id}](#delete-admingoodsid) | ✅ |
| [GET /api/tournaments](#get-apitournaments) | ✅ |
| [GET /api/tournaments/{id}](#get-apitournamentsid) | ✅ |
| [POST /admin/tournaments](#post-admintournaments) | ✅ |
| [PUT /admin/tournaments/{id}](#put-admintournamentsid) | ✅ |
| [DELETE /admin/tournaments/{id}](#delete-admintournamentsid) | ✅ |
| [GET /api/tasks](#get-apitasks) | ✅ |
| [GET /api/tasks/{id}](#get-apitasksid) | ✅ |
| [POST /admin/tasks](#post-admintasks) | ✅ |
| [PUT /admin/tasks/{id}](#put-admintasksid) | ✅ |
| [DELETE /admin/tasks/{id}](#delete-admintasksid) | ✅ |
| [GET /api/me](#get-apime) | ❌ |
| [PUT /api/me](#put-apime) | ❌ |
| [GET /api/points](#get-apipoints) | ❌ |
| [GET /api/points/ledger](#get-apipointsledger) | ❌ |
| [POST /api/redeem/orders](#post-apiredeemorders) | ❌ |
| [GET /api/redeem/orders](#get-apiredeemorders) | ❌ |
| [POST /admin/redeem/use](#post-adminredeemuse) | ❌ |
| [POST /api/checkin](#post-apicheckin) | ❌ |

<a id="get-health"></a>
## GET /health ✅

无参数。

返回：

```json
{
  "code": 200,
  "data": {
    "status": "ok",
    "started_at": "2026-01-30T00:00:00Z",
    "now": "2026-01-30T00:00:01Z"
  }
}
```

<a id="post-apiauthwechatlogin"></a>
## POST /api/auth/wechat/login ❌

未实现。

<a id="post-adminauthlogin"></a>
## POST /admin/auth/login ❌

未实现。

<a id="get-apidebugusers"></a>
## GET /api/debug/users ✅

无参数。

返回 data：用户列表（最多 200 条）。

<a id="get-apigoods"></a>
## GET /api/goods ✅

查询参数：

| 名称 | 类型 | 必填 | 说明 |
|---|---|---|---|
| status | string | 否 | 过滤状态（0/1） |

返回 data：

```json
[
  {
    "id": 2001,
    "name": "拳馆毛巾",
    "cover_url": "",
    "points_price": 200,
    "stock": 0,
    "status": 1,
    "created_at": "2026-01-30T12:00:00+08:00"
  }
]
```

<a id="get-apigoodsid"></a>
## GET /api/goods/{id} ✅

路径参数：

| 名称 | 类型 | 必填 | 说明 |
|---|---|---|---|
| id | number | 是 | 商品ID |

返回 data：单条商品（结构同上）。

<a id="post-admingoods"></a>
## POST /admin/goods ✅

请求 JSON：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| name | string | 是 | 商品名 |
| cover_url | string | 否 | 封面图 URL（空字符串表示 NULL） |
| points_price | number | 是 | 所需积分（>=0） |
| stock | number | 是 | 库存（>=0） |
| status | number | 否 | 0下架/1上架（缺省=1） |

返回 data：

```json
{ "id": 2005 }
```

<a id="put-admingoodsid"></a>
## PUT /admin/goods/{id} ✅

路径参数：id（同上）。

请求 JSON（全量更新）：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| name | string | 是 | 商品名 |
| cover_url | string | 否 | 封面图 URL（空字符串表示 NULL） |
| points_price | number | 是 | 所需积分（>=0） |
| stock | number | 是 | 库存（>=0） |
| status | number | 是 | 0下架/1上架 |

返回：`code=200`。

<a id="delete-admingoodsid"></a>
## DELETE /admin/goods/{id} ✅

路径参数：id（同上）。

返回：`code=200`。

<a id="get-apitournaments"></a>
## GET /api/tournaments ✅

查询参数：

| 名称 | 类型 | 必填 | 说明 |
|---|---|---|---|
| status | string | 否 | 过滤状态（DRAFT/PUBLISHED/FINISHED/CANCELED） |

返回 data：

```json
[
  {
    "id": 4001,
    "title": "周末友谊赛",
    "content": "周末店内友谊赛，欢迎报名",
    "cover_url": "",
    "start_at": "2026-02-01T14:00:00+08:00",
    "end_at": "2026-02-01T18:00:00+08:00",
    "status": "PUBLISHED",
    "created_by_admin_id": 1,
    "created_at": "2026-01-30T12:00:00+08:00",
    "updated_at": "2026-01-30T12:00:00+08:00"
  }
]
```

<a id="get-apitournamentsid"></a>
## GET /api/tournaments/{id} ✅

路径参数：id（赛事ID）。

返回 data：单条赛事（结构同上）。

<a id="post-admintournaments"></a>
## POST /admin/tournaments ✅

请求 JSON：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| title | string | 是 | 标题 |
| content | string | 否 | 内容（空字符串表示 NULL） |
| cover_url | string | 否 | 封面图 URL（空字符串表示 NULL） |
| start_at | string | 是 | RFC3339 时间 |
| end_at | string | 是 | RFC3339 时间，必须 >= start_at |
| status | string | 是 | DRAFT/PUBLISHED/FINISHED/CANCELED |
| created_by_admin_id | number | 是 | 创建人管理员ID |

返回 data：

```json
{ "id": 4003 }
```

<a id="put-admintournamentsid"></a>
## PUT /admin/tournaments/{id} ✅

路径参数：id（赛事ID）。

请求 JSON（全量更新）：字段同创建。

返回：`code=200`。

<a id="delete-admintournamentsid"></a>
## DELETE /admin/tournaments/{id} ✅

路径参数：id（赛事ID）。

返回：`code=200`。

<a id="get-apitasks"></a>
## GET /api/tasks ✅

查询参数：

| 名称 | 类型 | 必填 | 说明 |
|---|---|---|---|
| status | string | 否 | 过滤状态（0/1） |
| period_type | string | 否 | DAILY/WEEKLY/MONTHLY |

返回 data：

```json
[
  {
    "id": 1,
    "task_code": "DAILY_CHECKIN",
    "name": "到店打卡（每日）",
    "period_type": "DAILY",
    "target_count": 1,
    "reward_points": 1,
    "status": 1,
    "rule_json": null,
    "created_at": "2026-01-30T12:00:00+08:00"
  }
]
```

<a id="get-apitasksid"></a>
## GET /api/tasks/{id} ✅

路径参数：id（任务ID）。

返回 data：单条任务（结构同上）。

<a id="post-admintasks"></a>
## POST /admin/tasks ✅

请求 JSON：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| task_code | string | 是 | 任务编码（唯一） |
| name | string | 是 | 名称 |
| period_type | string | 是 | DAILY/WEEKLY/MONTHLY |
| target_count | number | 是 | 目标次数（>0） |
| reward_points | number | 是 | 奖励积分（>=0） |
| status | number | 否 | 0停用/1启用（缺省=1） |
| rule_json | object | 否 | 扩展规则（任意 JSON，缺省为 NULL） |

返回 data：

```json
{ "id": 10 }
```

<a id="put-admintasksid"></a>
## PUT /admin/tasks/{id} ✅

路径参数：id（任务ID）。

请求 JSON（全量更新）：字段同创建，但 status 必填。

返回：`code=200`。

<a id="delete-admintasksid"></a>
## DELETE /admin/tasks/{id} ✅

路径参数：id（任务ID）。

返回：`code=200`。

<a id="get-apime"></a>
## GET /api/me ❌

未实现。

<a id="put-apime"></a>
## PUT /api/me ❌

未实现。

<a id="get-apipoints"></a>
## GET /api/points ❌

未实现。

<a id="get-apipointsledger"></a>
## GET /api/points/ledger ❌

未实现。

<a id="post-apiredeemorders"></a>
## POST /api/redeem/orders ❌

未实现。

<a id="get-apiredeemorders"></a>
## GET /api/redeem/orders ❌

未实现。

<a id="post-adminredeemuse"></a>
## POST /admin/redeem/use ❌

未实现。

<a id="post-apicheckin"></a>
## POST /api/checkin ❌

未实现。
