# GameSocial 前端对接指南（接口总览）

本文档用于“快速对接当前所有接口”，包含通用约定、鉴权、媒体直传、以及客户端/管理端接口索引。接口的完整字段说明与示例请以现有文档为准：

- 客户端接口（小程序/APP）：[API_CLIENT_ENDPOINTS.md](API_CLIENT_ENDPOINTS.md)
- 管理端接口（后台管理）：[API_ADMIN_ENDPOINTS.md](API_ADMIN_ENDPOINTS.md)

## 1. 通用约定

### 1.1 Base URL

- 本地开发：`http://localhost:<SERVER_PORT>`

### 1.2 统一响应结构

后端接口统一返回：

```json
{
  "code": 200,
  "data": {},
  "message": "ok"
}
```

前端判断是否成功只看 `code`：

- `code=200`：成功
- `code!=200`：失败（通常 `message` 有可直接提示的文案）

注意：很多业务错误会返回 `HTTP 200`，但 `code!=200`。

### 1.3 鉴权（客户端 /api）

需要登录的客户端接口使用：

- `Authorization: Bearer <token>`

登录成功后把 token 持久化（storage）即可，后续请求统一带上。

### 1.4 鉴权（管理端 /admin）

当前管理端接口文档里说明为“暂未接入 token 校验”，以实际实现为准。

## 2. 媒体直传（COS PUT）

目标：让前端在用户“确认提交之前”把图片直传到 COS 的 `temp/...`，避免占用正式目录。

对应接口：

- 申请上传坑位：`POST /api/media/temp-upload-infos`
- 完整说明：见 [API_CLIENT_ENDPOINTS.md 的 Media 模块](API_CLIENT_ENDPOINTS.md#api-media-temp-upload-infos)

### 2.1 调用流程

1) 前端统计这次要上传的图片数量 `count`（范围 1~10），请求后端生成每个文件的 `objectId` 与 PUT 授权信息。

2) 对 `items[]` 的每一项执行一次 PUT 上传：

- URL：`items[i].uploadUrl`
- Header：`Authorization: items[i].authorization`
- Header：`Content-Type: <文件实际类型>`
- Body：文件二进制

3) 上传成功后，将 `items[i].downloadUrl` 当作图片 URL，回填到后续业务接口（例如用户头像、商品图片、赛事图片等）。

### 2.2 fetch 示例

```js
export async function cosPutUpload(item, file) {
  const res = await fetch(item.uploadUrl, {
    method: "PUT",
    headers: {
      Authorization: item.authorization,
      "Content-Type": file.type || "application/octet-stream",
    },
    body: file,
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`upload failed: ${res.status} ${text}`);
  }

  return {
    objectId: item.objectId,
    downloadUrl: item.downloadUrl,
    etag: res.headers.get("etag") || "",
  };
}
```

### 2.3 CORS（COS 控制台必须配置）

至少确保：

- Allowed Methods：`PUT`（以及你需要的 `GET/HEAD`）
- Allowed Headers：`Authorization, Content-Type`
- Expose Headers：`ETag`
- Allowed Origin：你的 Web 站点域名（开发期可先放宽，线上收紧）

## 3. 接口索引（客户端 /api）

完整字段与示例见：[API_CLIENT_ENDPOINTS.md](API_CLIENT_ENDPOINTS.md)

- GET `/health`（√）详见 [健康检查](API_CLIENT_ENDPOINTS.md#api-health)
- POST `/api/auth/wechat/login`（√）详见 [微信登录](API_CLIENT_ENDPOINTS.md#api-auth-wechat-login)
- GET `/api/users/me`（√）详见 [获取个人资料](API_CLIENT_ENDPOINTS.md#api-users-me-get)
- PUT `/api/users/me`（√）详见 [更新个人资料](API_CLIENT_ENDPOINTS.md#api-users-me-update)
- POST `/api/media/temp-upload-infos`（√）详见 [临时直传凭证](API_CLIENT_ENDPOINTS.md#api-media-temp-upload-infos)
- POST `/api/qrcodes/verify`（√）详见 [二维码校验](API_CLIENT_ENDPOINTS.md#api-qrcodes-verify)
- POST `/api/qrcodes/use`（√）详见 [二维码核销](API_CLIENT_ENDPOINTS.md#api-qrcodes-use)
- GET `/api/points/balance`（√）详见 [积分余额](API_CLIENT_ENDPOINTS.md#api-points-balance)
- GET `/api/points/ledgers`（√）详见 [积分流水](API_CLIENT_ENDPOINTS.md#api-points-ledgers)
- GET `/api/vip/status`（√）详见 [会员状态](API_CLIENT_ENDPOINTS.md#api-vip-status)
- GET `/api/tournaments`（√）详见 [赛事列表](API_CLIENT_ENDPOINTS.md#api-tournaments-list)
- GET `/api/tournaments/joined`（√）详见 [已报名赛事](API_CLIENT_ENDPOINTS.md#api-tournaments-joined)
- GET `/api/tournaments/{id}`（√）详见 [赛事详情](API_CLIENT_ENDPOINTS.md#api-tournaments-get)
- POST `/api/tournaments/{id}/join`（√）详见 [报名赛事](API_CLIENT_ENDPOINTS.md#api-tournaments-join)
- PUT `/api/tournaments/{id}/cancel`（√）详见 [取消报名](API_CLIENT_ENDPOINTS.md#api-tournaments-cancel)
- GET `/api/tournaments/{id}/results`（√）详见 [赛事结果/排名](API_CLIENT_ENDPOINTS.md#api-tournaments-results)
- GET `/api/tasks`（√）详见 [任务列表](API_CLIENT_ENDPOINTS.md#api-tasks-list)
- POST `/api/tasks/checkin`（×）详见 [任务打卡](API_CLIENT_ENDPOINTS.md#api-tasks-checkin)
- POST `/api/tasks/{taskCode}/claim`（×）详见 [领取任务奖励](API_CLIENT_ENDPOINTS.md#api-tasks-claim)
- GET `/api/goods`（√）详见 [商品列表](API_CLIENT_ENDPOINTS.md#api-goods-list)
- GET `/api/goods/{id}`（√）详见 [商品详情](API_CLIENT_ENDPOINTS.md#api-goods-get)
- POST `/api/redeem/orders`（√）详见 [创建兑换订单](API_CLIENT_ENDPOINTS.md#api-redeem-orders-create)
- GET `/api/redeem/orders`（√）详见 [订单列表](API_CLIENT_ENDPOINTS.md#api-redeem-orders-list)
- GET `/api/redeem/orders/{id}`（√）详见 [订单详情](API_CLIENT_ENDPOINTS.md#api-redeem-orders-get)
- PUT `/api/redeem/orders/{id}/cancel`（√）详见 [取消订单](API_CLIENT_ENDPOINTS.md#api-redeem-orders-cancel)

## 4. 接口索引（管理端 /admin）

完整字段与示例见：[API_ADMIN_ENDPOINTS.md](API_ADMIN_ENDPOINTS.md)

- GET `/health`（√）详见 [健康检查](API_ADMIN_ENDPOINTS.md#api-health)
- POST `/admin/goods`（√）详见 [创建商品](API_ADMIN_ENDPOINTS.md#api-admin-goods-create)
- GET `/admin/goods`（√）详见 [商品列表](API_ADMIN_ENDPOINTS.md#api-admin-goods-list)
- GET `/admin/goods/{id}`（√）详见 [商品详情](API_ADMIN_ENDPOINTS.md#api-admin-goods-get)
- PUT `/admin/goods/{id}`（√）详见 [更新商品](API_ADMIN_ENDPOINTS.md#api-admin-goods-update)
- DELETE `/admin/goods/{id}`（√）详见 [删除商品](API_ADMIN_ENDPOINTS.md#api-admin-goods-delete)
- POST `/admin/tournaments`（√）详见 [创建赛事](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-create)
- GET `/admin/tournaments`（√）详见 [赛事列表](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-list)
- GET `/admin/tournaments/{id}`（√）详见 [赛事详情](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-get)
- PUT `/admin/tournaments/{id}`（√）详见 [更新赛事](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-update)
- DELETE `/admin/tournaments/{id}`（√）详见 [删除赛事](API_ADMIN_ENDPOINTS.md#api-admin-tournaments-delete)
- POST `/admin/task-defs`（√）详见 [创建任务](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-create)
- GET `/admin/task-defs`（√）详见 [任务列表](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-list)
- GET `/admin/task-defs/{id}`（√）详见 [任务详情](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-get)
- PUT `/admin/task-defs/{id}`（√）详见 [更新任务](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-update)
- DELETE `/admin/task-defs/{id}`（√）详见 [删除任务](API_ADMIN_ENDPOINTS.md#api-admin-task-defs-delete)
- GET `/admin/users`（√）详见 [用户列表](API_ADMIN_ENDPOINTS.md#api-admin-users-list)
- GET `/admin/users/{id}`（√）详见 [用户详情](API_ADMIN_ENDPOINTS.md#api-admin-users-get)
- PUT `/admin/users/{id}`（√）详见 [更新用户](API_ADMIN_ENDPOINTS.md#api-admin-users-update)
- POST `/admin/redeem/orders`（√）详见 [创建订单](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-create)
- GET `/admin/redeem/orders`（√）详见 [订单列表](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-list)
- GET `/admin/redeem/orders/{id}`（√）详见 [订单详情](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-get)
- PUT `/admin/redeem/orders/{id}/use`（√）详见 [核销订单](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-use)
- PUT `/admin/redeem/orders/{id}/cancel`（√）详见 [取消订单](API_ADMIN_ENDPOINTS.md#api-admin-redeem-orders-cancel)
- POST `/admin/qrcodes`（√）详见 [生成二维码](API_ADMIN_ENDPOINTS.md#api-admin-qrcodes-create)
- POST `/admin/auth/login`（×）详见 [管理端登录](API_ADMIN_ENDPOINTS.md#api-admin-auth-login)
- GET `/admin/auth/me`（×）详见 [管理端个人信息](API_ADMIN_ENDPOINTS.md#api-admin-auth-me)
- POST `/admin/auth/logout`（×）详见 [管理端退出](API_ADMIN_ENDPOINTS.md#api-admin-auth-logout)
- GET `/admin/audit/logs`（√）详见 [审计日志](API_ADMIN_ENDPOINTS.md#api-admin-audit-logs)
- POST `/admin/points/adjust`（×）详见 [积分调整](API_ADMIN_ENDPOINTS.md#api-admin-points-adjust)
- PUT `/admin/users/{id}/drinks/use`（×）详见 [饮品核销](API_ADMIN_ENDPOINTS.md#api-admin-users-drinks-use)
- POST `/admin/tournaments/{id}/results/publish`（×）详见 [发布成绩](API_ADMIN_ENDPOINTS.md#api-admin-tournament-results-publish)
- POST `/admin/tournaments/{id}/awards/grant`（×）详见 [发放奖励](API_ADMIN_ENDPOINTS.md#api-admin-tournament-awards-grant)
