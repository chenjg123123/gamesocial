# 格斗玩家社区后端（GameSocial）模块化架构设计

## 1. 项目目标与约束

### 1.1 目标

- 面向微信小程序（uniapp）提供稳定的业务 API
- 支持约 1000 注册用户、常态在线几十人、峰值同时在线约 200 的使用规模
- 重点保证：积分/核销/报名/发奖的正确性、可追溯、可审计

### 1.2 约束与假设

- 并发压力不高，优先保证交付速度与业务一致性
- 后端采用“模块化单体”起步，后续根据实际增长再拆分服务
- 数据以关系型数据库为主，缓存为辅

## 2. 业务范围（先把边界说清楚）

### 2.1 用户侧核心业务

- 微信小程序登录，查看/编辑基础资料
- 社区等级与经验（或“等级积分”）成长
- VIP（月卡）状态展示（开通与续费可后置接入支付）
- 积分获取与消耗：赛事奖励、任务奖励、兑换扣减、管理员赠送/扣减
- 饮品数量存储与线下兑换/核销
- 赛事：查看赛事、报名参赛、查看自己的排名与历史
- 任务：到店打卡（日任务）、周签、月任务，领取奖励

### 2.2 管理员（店长后台）核心业务

- 单店的核销与打卡管理
- 赛事发布、维护、历史查询
- 赛事排名录入/发布
- 给前 N 名发放积分奖励（必须幂等，避免重复发奖）
- 商品/饮品管理（可选：库存、上下架、封面图）
- 积分调整（赠送/扣减）与审计记录

## 3. 核心角色与权限（用户/管理员怎么区分）

### 3.1 角色

- user：小程序普通用户
- admin：后台管理员（先只配置一个账号，后续需要再扩展多管理员）

### 3.2 权限模型（先做最小可用）

- 用户侧与管理员侧使用两套 token（或同一套 token 但区分签发受众）
- 管理员侧先只做“是否管理员”的判断即可
- 管理员所有操作写审计日志，便于追溯（积分调整、赛事发布、发奖、核销等）

## 4. 总体架构（模块化单体）

### 4.1 三层分离（最重要）

1) api：只做 HTTP（路由/中间件/请求解析/响应输出）  
2) modules：只做业务（数据结构 + 业务接口 + 业务实现）  
3) internal：只放基础设施（配置、数据库、缓存、定时任务、第三方客户端）

原则：

- api 层不写 SQL，不直接依赖数据库
- modules 层不依赖 HTTP 语义（不要出现 Request/Response）
- internal 层对外只提供“可复用能力”，不含业务规则

### 4.2 建议目录结构

```
GameSocial/
├── cmd/
│   └── server/                      # 应用入口：加载配置 -> 初始化基础设施 -> 组装业务模块 -> 注册路由 -> 启动
├── api/
│   ├── middleware/                  # Recover/CORS/Logging/Auth/RateLimit 等
│   └── handlers/                    # HTTP handler：参数校验 -> 调用业务接口 -> 返回 JSON
├── internal/
│   ├── config/                      # 配置加载
│   ├── database/                    # MySQL 连接、事务封装
│   ├── cache/                       # Redis（可选）
│   ├── wechat/                      # 微信 code2session、支付回调验签等（可后置）
│   └── media/                       # 媒体文件：本地存储的保存与访问（也可后续替换为 OSS）
└── modules/
    ├── auth/                        # 小程序登录、token、用户鉴权
    ├── admin/                       # 后台登录、鉴权、审计
    ├── user/                        # 用户资料、等级/经验
    ├── vip/                         # 会员订阅（可先只做状态与到期判断）
    ├── points/                      # 积分账户 + 流水账本
    ├── tournament/                  # 赛事发布/报名/排名/发奖
    ├── task/                        # 每日/周/月任务、打卡
    ├── item/                        # 商品/饮品
    └── redeem/                      # 兑换单、核销流程
```

## 5. 关键业务流程（把“最容易出错”的链路写清楚）

### 5.1 微信登录（小程序）

1) 小程序端 `wx.login()` 获取 code  
2) 后端调用 `code2session` 获取 openid（和可选 unionid）  
3) 若首次登录则创建 user  
4) 后端签发自己的 token，后续请求用 token 鉴权

### 5.2 积分记账（必须账本化 + 幂等）

规则：

- points_account 只存余额快照，用于快速展示
- points_ledger 存每一笔变动流水，必须携带 biz_type + biz_id
- 对 `(user_id, biz_type, biz_id)` 建唯一约束，实现幂等（重复请求不会重复加/扣积分）

### 5.3 兑换与核销（防重复核销）

- 积分兑换商品：创建 redeem_order（CREATED）+ redeem_order_item + 写扣减积分流水
- 积分兑换饮料：不创建订单，直接把 user_drink_balance.quantity + 1，并写 points_ledger（来源为 DRINK_EXCHANGE）
- 饮料核销：管理员在后台对某个用户“点一次 -1”，将 user_drink_balance.quantity - 1，并写 admin_audit_log（DRINK_USE）
- 商品核销：管理员在后台手动输入订单号完成核销，将 redeem_order 状态从 CREATED -> USED，写核销时间与核销人
- 所有更新要保证“只成功一次”（扣减积分用 points_ledger 幂等约束，核销/点减用状态或条件更新保证）

### 5.4 赛事排名与发奖（防重复发奖）

- 排名发布：写 tournament_result（tournament_id + user_id 唯一）
- 发奖：写 tournament_award（幂等）+ 写 points_ledger（幂等）+ 更新 points_account

### 5.5 任务与打卡（日/周/月）

- task_def：任务定义（周期、目标次数、奖励积分）
- user_task_progress：按 period_key 记录用户某周期的进度与领奖状态
  - period_key 示例：日 20260128、周 2026W05、月 202601
- checkin_log：到店打卡明细
- 领奖触发：写 points_ledger（幂等）

### 5.6 图片/封面等媒体（本地存储优先）

你可以先不使用 OSS，把赛事封面/商品封面存到服务器本地磁盘上，通过 Nginx 提供静态访问。

- 用户头像：微信侧可以拿到 avatarUrl，后端存一份 URL 字符串即可，不必下载保存
- 赛事/商品封面：后端提供上传接口，保存文件到本地目录，业务表只存一个可访问的 URL（或相对路径）
- 后续如果要上 OSS：只需要把“存储实现”从本地换成 OSS，业务表字段不需要改（仍然是 URL）

## 6. 数据库表设计（单店版，可落地）

约定：

- created_at / updated_at 为业务需要的常用时间字段
- 关键写入链路（积分、发奖、兑换、核销）都要可追溯，因此会保留“操作者”和“业务来源”
- 管理员只有一个账号也要保留审计表，后期扩展多管理员不需要改模型

### 6.1 user（用户表）

用途：存储微信小程序用户的身份标识与基础资料，是所有用户业务的主表。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 用户ID |
| openid | VARCHAR(64) | UNIQUE, NOT NULL | 小程序唯一标识 |
| unionid | VARCHAR(64) | NULL | 可选 |
| nickname | VARCHAR(64) | NULL | 昵称 |
| avatar_url | VARCHAR(512) | NULL | 头像URL（微信返回 avatarUrl 或自定义上传地址） |
| status | TINYINT | NOT NULL | 1正常/0封禁 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

索引与约束：

- UNIQUE(openid)
- INDEX(unionid)

### 6.2 user_level（用户等级表）

用途：记录用户等级与等级经验（或“等级积分”），便于后续扩展等级规则。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| user_id | BIGINT | PK | 用户ID |
| level | INT | NOT NULL | 等级 |
| exp | BIGINT | NOT NULL | 经验/等级积分 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

### 6.3 vip_subscription（VIP 订阅表）

用途：记录月卡（或其它 VIP 套餐）的有效期与状态，用于判断用户是否为 VIP。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 订阅ID |
| user_id | BIGINT | INDEX, NOT NULL | 用户ID |
| plan | VARCHAR(32) | NOT NULL | MONTH_CARD 等 |
| start_at | DATETIME | NOT NULL | 开始时间 |
| end_at | DATETIME | NOT NULL | 结束时间 |
| status | VARCHAR(16) | NOT NULL | ACTIVE/EXPIRED/CANCELED |
| pay_order_no | VARCHAR(64) | UNIQUE, NULL | 关联支付单号（后置接入支付时用） |
| created_at | DATETIME | NOT NULL | 创建时间 |

索引与约束：

- INDEX(user_id, end_at)
- UNIQUE(pay_order_no)

### 6.4 admin_user（后台管理员表）

用途：后台登录账号表。当前只配一个管理员，后续要加多个管理员无需改表。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 管理员ID |
| username | VARCHAR(64) | UNIQUE, NOT NULL | 登录名 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希 |
| status | TINYINT | NOT NULL | 1正常/0禁用 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

### 6.5 admin_audit_log（管理员操作审计表）

用途：记录管理员在后台的关键操作，用于排查“积分为何变化、谁发了奖、谁核销了饮品”等问题。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 日志ID |
| admin_id | BIGINT | INDEX, NOT NULL | 操作人（admin_user.id） |
| action | VARCHAR(64) | NOT NULL | 操作类型（如 TOURNAMENT_PUBLISH/POINTS_ADJUST/REDEEM_USE） |
| biz_type | VARCHAR(32) | NULL | 业务类型（可选） |
| biz_id | VARCHAR(64) | NULL | 业务ID（可选） |
| detail_json | JSON | NULL | 操作详情（可选） |
| created_at | DATETIME | NOT NULL | 时间 |

### 6.6 points_account（积分账户表）

用途：存储用户当前可用积分余额，用于快速展示与校验扣减是否足够。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| user_id | BIGINT | PK | 用户ID |
| balance | BIGINT | NOT NULL | 可用积分 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

### 6.7 points_ledger（积分流水表）

用途：积分“账本”。每一笔积分变化都必须落在这里，并写明来源，便于追溯与对账。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 流水ID |
| user_id | BIGINT | INDEX, NOT NULL | 用户ID |
| change_amount | BIGINT | NOT NULL | 变动值（+/-） |
| balance_after | BIGINT | NOT NULL | 变动后余额快照 |
| biz_type | VARCHAR(32) | NOT NULL | 来源类型（赛事/任务/兑换/管理员调整等） |
| biz_id | VARCHAR(64) | NOT NULL | 来源ID（用于幂等） |
| remark | VARCHAR(255) | NULL | 备注（可选） |
| created_at | DATETIME | NOT NULL | 时间 |

索引与约束：

- UNIQUE(user_id, biz_type, biz_id)
- INDEX(user_id, created_at)

### 6.8 goods（积分商品表）

用途：积分可兑换的商品列表（不是饮品数量）。是否启用库存由你决定，先不做也能上线。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 商品ID |
| name | VARCHAR(128) | NOT NULL | 商品名 |
| cover_url | VARCHAR(512) | NULL | 封面图URL（本地静态地址或 OSS 地址） |
| points_price | BIGINT | NOT NULL | 所需积分 |
| stock | INT | NOT NULL | 库存（不需要可固定为0） |
| status | TINYINT | NOT NULL | 1上架/0下架 |
| created_at | DATETIME | NOT NULL | 创建时间 |

### 6.9 user_drink_balance（用户饮品数量表）

用途：记录用户在店里“可兑换的饮品数量”。当前只做总数量，不做不同饮品/不同价值的区分。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| user_id | BIGINT | PK | 用户ID |
| quantity | INT | NOT NULL | 可用饮品数量 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

### 6.10 redeem_order（积分兑换商品订单表）

用途：记录“积分兑换商品”的订单。饮料不走订单：积分兑换饮料直接把用户饮品数量 +1；管理员核销饮料直接对用户饮品数量 -1。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 订单ID |
| order_no | VARCHAR(64) | UNIQUE, NOT NULL | 订单号（给用户展示/管理员录入） |
| user_id | BIGINT | INDEX, NOT NULL | 用户ID |
| status | VARCHAR(16) | INDEX, NOT NULL | CREATED/USED/CANCELED |
| total_points | BIGINT | NOT NULL | 订单总扣减积分 |
| used_by_admin_id | BIGINT | NULL | 核销人（admin_user.id） |
| used_at | DATETIME | NULL | 核销时间 |
| created_at | DATETIME | NOT NULL | 创建时间 |

### 6.11 redeem_order_item（兑换订单明细表）

用途：记录“积分兑换商品”的明细行。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 明细ID |
| redeem_order_id | BIGINT | INDEX, NOT NULL | 订单ID |
| goods_id | BIGINT | NOT NULL | 商品ID（goods.id） |
| quantity | INT | NOT NULL | 数量 |
| points_price | BIGINT | NOT NULL | 下单时商品单价积分快照 |

### 6.12 tournament（赛事表）

用途：赛事基础信息表，管理员发布赛事并维护封面图、时间与状态。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 赛事ID |
| title | VARCHAR(128) | NOT NULL | 标题 |
| content | TEXT | NULL | 详情 |
| cover_url | VARCHAR(512) | NULL | 封面图URL（本地静态地址或 OSS 地址） |
| start_at | DATETIME | INDEX, NOT NULL | 开始时间 |
| end_at | DATETIME | INDEX, NOT NULL | 结束时间 |
| status | VARCHAR(16) | INDEX, NOT NULL | DRAFT/PUBLISHED/FINISHED/CANCELED |
| created_by_admin_id | BIGINT | NOT NULL | 创建人 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

### 6.13 tournament_participant（赛事报名表）

用途：记录用户报名赛事的关系表，用于报名列表与资格校验。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 记录ID |
| tournament_id | BIGINT | NOT NULL | 赛事ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| join_status | VARCHAR(16) | NOT NULL | JOINED/CANCELED |
| joined_at | DATETIME | NOT NULL | 报名时间 |

索引与约束：

- UNIQUE(tournament_id, user_id)
- INDEX(user_id, joined_at)

### 6.14 tournament_result（赛事排名表）

用途：记录赛事排名结果。用户可查看自己的名次，管理员可回查历史成绩。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 结果ID |
| tournament_id | BIGINT | NOT NULL | 赛事ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| rank_no | INT | NOT NULL | 名次 |
| score | INT | NULL | 可选分数 |
| published_by_admin_id | BIGINT | NOT NULL | 发布人 |
| published_at | DATETIME | NOT NULL | 发布时间 |

索引与约束：

- UNIQUE(tournament_id, user_id)
- INDEX(tournament_id, rank_no)

### 6.15 tournament_award（赛事发奖记录表）

用途：记录赛事发奖（积分奖励）的落库凭证，用于防重复发奖与审计回查。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 发奖记录ID |
| tournament_id | BIGINT | NOT NULL | 赛事ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| award_points | BIGINT | NOT NULL | 发放积分 |
| biz_id | VARCHAR(64) | NOT NULL | 幂等键 |
| created_by_admin_id | BIGINT | NOT NULL | 发奖人 |
| created_at | DATETIME | NOT NULL | 时间 |

索引与约束：

- UNIQUE(user_id, biz_id)
- INDEX(tournament_id, created_at)

### 6.16 task_def（任务定义表）

用途：定义每日/周/月任务的规则与奖励。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 任务ID |
| task_code | VARCHAR(64) | UNIQUE, NOT NULL | 任务编码 |
| name | VARCHAR(128) | NOT NULL | 名称 |
| period_type | VARCHAR(16) | NOT NULL | DAILY/WEEKLY/MONTHLY |
| target_count | INT | NOT NULL | 目标次数 |
| reward_points | BIGINT | NOT NULL | 奖励积分 |
| status | TINYINT | NOT NULL | 1启用/0停用 |
| rule_json | JSON | NULL | 扩展规则 |
| created_at | DATETIME | NOT NULL | 创建时间 |

### 6.17 user_task_progress（用户任务进度表）

用途：记录用户在某个周期内对某个任务的进度与领奖状态。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 进度ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| period_key | VARCHAR(16) | NOT NULL | 20260128/2026W05/202601 |
| progress_count | INT | NOT NULL | 当前进度 |
| status | VARCHAR(16) | NOT NULL | IN_PROGRESS/COMPLETED/REWARDED |
| completed_at | DATETIME | NULL | 完成时间 |
| rewarded_at | DATETIME | NULL | 领奖时间 |

索引与约束：

- UNIQUE(user_id, task_id, period_key)
- INDEX(user_id, period_key)

### 6.18 checkin_log（到店打卡记录表）

用途：记录用户每天到店打卡明细，用于任务计算与运营统计。

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT | PK | 打卡ID |
| user_id | BIGINT | INDEX, NOT NULL | 用户ID |
| checkin_at | DATETIME | NOT NULL | 打卡时间 |
| source | VARCHAR(16) | NOT NULL | MANUAL/GPS/QR 等 |

## 7. 功能说明（给非技术人员看的版本）

### 7.1 用户在小程序里能得到什么（用户感知）

你可以把它理解成“一个到店玩格斗、参加比赛、攒积分换东西的小程序”。用户打开小程序后，主要能看到并获得下面这些：

- 我的主页（最常用）
  - 你能看到：你的头像和昵称、你现在的等级、你是不是月卡会员、你现在有多少积分、你现在有多少杯饮料可以兑换
  - 你能解决：不用问店员也能随时知道“我还有多少积分/饮料”“我是不是会员”
- 积分（怎么来的、怎么用的）
  - 你能看到：积分总数，以及每一次积分变化的原因
    - 例如：参加某场比赛获得了多少积分、今天打卡得了多少积分、兑换了某个商品扣了多少积分、管理员给你补了多少积分
  - 你能解决：当你觉得积分不对时，可以自己查清楚“是哪一天、因为什么变了”
- 我的饮料（数量与兑换）
  - 你能看到：你现在还有多少杯饮料可以在店里兑换
  - 你能操作：用积分兑换饮料（一次兑换 +1 杯）
  - 你到店怎么用：到店后店长/管理员给你“点一次核销 -1”，饮料数量会自动减少
  - 你能解决：饮料只维护一个数量，兑换/核销操作更简单
- 赛事（比赛信息与报名）
  - 你能看到：最近有哪些比赛、每场比赛的介绍、时间、封面图
  - 你能操作：报名参加比赛（如果你临时有事，也可以取消报名）
  - 你能解决：不用到群里翻消息，所有比赛信息都在小程序里
- 赛事结果（我打得怎么样）
  - 你能看到：你参加过的比赛、你在每场比赛里的名次（排名）
  - 你能解决：自己成绩一眼能查到，方便复盘和分享
- 每日/每周/每月任务（到店就有奖励）
  - 你能看到：今天要做什么能拿奖励、本周还差几次、本月还差多少
  - 你能操作：到店打卡（表示你今天来过店里）
  - 你能获得：完成任务后系统会给你积分奖励
  - 你能解决：把“到店奖励”变成清晰的目标，用户更愿意持续来店里
- 兑换商品（用积分换东西）
  - 你能看到：店里有哪些东西可以用积分换（带图、带所需积分）
  - 你能操作：下单兑换，系统会生成订单号并扣掉相应积分
  - 你能解决：积分怎么花更透明，减少线下沟通成本
- 我的订单（所有兑换记录）
  - 你能看到：你所有兑换/核销的订单记录、每个订单的状态（还没用/已经用掉/已取消）
  - 你能解决：防止“我到底换过没”“订单给过店长没”这种扯皮

### 7.2 店长/管理员在后台能做什么（店长感知）

你可以把后台理解成“店长的管理工具”，它解决的是：比赛怎么发布、积分怎么发、饮料怎么核销、用户数据怎么查的问题。

- 登录后台
  - 你能做：用后台账号登录
  - 你能解决：避免随便谁都能操作积分、核销等关键动作
- 核销（最核心）
  - 你能做：用户到店后，店长输入用户提供的订单号，确认后完成核销
  - 你能看到：这个订单是谁的、兑换的是什么、数量是多少、是不是已经核销过
  - 你能解决：防止重复兑换、减少口头确认，核销结果系统自动记录
- 饮料核销（更简单）
  - 你能做：找到用户后，点一次“核销 -1”，用户可用饮料数量立刻减少
  - 你能看到：用户当前可用饮料数量
  - 你能解决：不需要饮料订单号，操作更快
- 打卡记录（用户是否到店）
  - 你能做：查看每天有多少用户来过店里打卡
  - 你能看到：打卡的用户列表与时间
  - 你能解决：统计到店情况，评估活动效果
- 赛事管理（发布与维护）
  - 你能做：发布比赛、修改比赛信息、设置开始/结束时间、上传封面图、结束或取消比赛
  - 你能看到：所有比赛列表、每场比赛的报名情况
  - 你能解决：比赛运营不再依赖群公告，统一从后台发出
- 排名录入（比赛结果）
  - 你能做：比赛结束后录入名次（第几名是谁）
  - 你能看到：每场比赛的结果表
  - 你能解决：成绩统一沉淀，后续查询不丢
- 发放奖励（给前几名发积分）
  - 你能做：对前 N 名发放积分奖励（例如第1名100分、第2名50分）
  - 你能看到：这次比赛给谁发了多少分、是否已经发过
  - 你能解决：避免重复发奖；用户问“为什么我没收到”能快速核查
- 积分调整（处理特殊情况）
  - 你能做：给某个用户补积分或扣积分（比如线下消费赠送、纠错）
  - 你能看到：每次调整的原因与记录
  - 你能解决：店铺运营更灵活，同时有记录可查
- 商品管理（积分能换什么）
  - 你能做：新增商品、上架/下架、设置需要多少积分、上传封面图（库存要不要管可以先不管）
  - 你能看到：商品列表、兑换情况
  - 你能解决：积分兑换规则清晰，减少线下解释成本
- 操作记录（出了问题能追）
  - 你能看到：后台所有关键操作的记录
    - 例如：谁在什么时间发布了比赛、给谁发了积分、核销了哪个订单
  - 你能解决：出现争议时可以快速追溯，避免扯皮

## 8. 部署建议（先简单）

- 单机部署（够用）
  - Nginx（反代与 HTTPS）
  - API 服务（单实例）
  - MySQL（数据持久化挂载）
  - Redis（可选：限流/缓存/幂等锁）
- 本地媒体存储
  - 图片文件落在服务器磁盘目录，通过 Nginx 静态目录对外提供访问
  - 需要做目录持久化与备份（避免重装/迁移导致图片丢失）

## 9. 里程碑（建议按这个顺序做）

1) 登录（user/admin）
2) 积分中心（余额 + 流水 + 幂等约束）
3) 赛事：发布/报名/排名/发奖（含封面图）
4) 任务：到店打卡 + 日/周/月任务 + 奖励发放
5) 兑换与核销：积分兑换 + 饮品数量核销 + 手动核销

