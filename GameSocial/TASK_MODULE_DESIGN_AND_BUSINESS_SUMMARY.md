# 任务系统优先 + 比赛系统规划（前端转后端可读版）

你现在最关心的是两块：

1) 任务系统：到店打卡、每日/周/月任务、领奖发积分（以及你提到的“社区经验”）。  
2) 比赛系统：以后经常有比赛，怎么把它设计得更完善（人工配置/随机匹配/赛程/发奖/复盘）。

本文目标不是“写得高级”，而是让你这种从前端转后端的人能看懂：**数据从哪来、落哪张表、哪一步会改哪些字段、哪里必须做幂等**。

---

## 0. 读懂任务系统，你只要先记住这 4 个概念

把任务系统先粗暴拆成“3 张表 + 1 个积分账本”：

1) `task_def`：任务配置（运营在后台配出来的）。  
2) `checkin_log`：用户行为日志（用户每天到店打卡写一条）。  
3) `user_task_progress`：用户在“当前周期”对某个任务的进度与领奖状态。  
4) `points_ledger` + `points_account`：积分账本 + 余额快照（领奖真正发积分的地方）。

表定义都在：[gamesocial_init.sql](file:///w:/GOProject/gamesocial/GameSocial/gamesocial_init.sql)

---

## 1. 先把“数据从哪来”讲清楚（给你对照看的）

### 1.1 数据来源路线图（从请求到落库）

你可以把后端当成固定的三段式：

- 前端请求某个接口（例如 `/api/tasks/checkin`）
- handler 解析参数 + 拿到 `userId`（从 token 注入，不允许前端传 userId）
- service/SQL 改数据库（写日志、改进度、发积分）

这里 `userId` 的来源很关键：  
中间件把 token 解析结果写到请求里，然后 handler 读出来。参考：

- [middleware.go](file:///w:/GOProject/gamesocial/GameSocial/api/middleware/middleware.go)
- [app_endpoints.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/app_endpoints.go)

### 1.2 任务系统涉及的接口在哪里

路由注册入口在：[main.go](file:///w:/GOProject/gamesocial/GameSocial/cmd/server/main.go)

当前任务相关 handler 在：

- [app_tasks.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/app_tasks.go)（小程序侧：任务列表/打卡/领奖）
- [admin_task_defs.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/admin_task_defs.go)（后台侧：任务定义 CRUD）

任务定义 service 在：

- [task/service.go](file:///w:/GOProject/gamesocial/GameSocial/modules/task/service.go)

---

## 2. 任务系统（优先级最高）：你可以这样理解它

### 2.1 任务系统要解决的 3 个业务问题

1) 运营能配置任务：今天想做“每日到店 +1 分”，明天想做“每周到店 3 次 +5 分”，不改代码。  
2) 用户能看到任务：我这个周期做了几次？还差几次？现在能不能领？领过没有？  
3) 发奖必须正确：不管用户点多少次“领奖”，积分最多发一次（幂等）。

### 2.2 三张核心表分别存什么（用人话解释字段）

#### 2.2.1 task_def（任务定义：运营配置的“模板”）

你可以把它理解成“任务的静态说明书”：

- `task_code`：任务编号（接口里会用到，比如 `/api/tasks/{taskCode}/claim`）  
- `period_type`：周期类型（DAILY/WEEKLY/MONTHLY）  
- `target_count`：完成门槛（到店打卡几次算完成）  
- `reward_points`：奖励积分（领奖后给多少积分）  
- `rule_json`：高级扩展（以后可以放“奖励经验”“触发条件”等，不改表也能扩展）

#### 2.2.2 checkin_log（打卡日志：用户行为证据）

这张表只记录“发生了什么”，不记录“进度是多少”。最简单就是：

- `user_id`：谁打卡的  
- `checkin_at`：什么时候打卡  
- `source`：打卡来源（先 MANUAL，后面可以做 GPS/核销等）

为什么要留日志：以后运营/风控/对账都会靠它回查。

#### 2.2.3 user_task_progress（用户进度：每个周期一条进度）

这张表记录“某用户、某任务、在某个周期内的当前进度与领奖状态”：

- `user_id + task_id + period_key` 唯一：这就是“同周期幂等”的根基  
- `progress_count`：当前进度（比如本周打卡 2 次）  
- `status`：`IN_PROGRESS/COMPLETED/REWARDED`  
- `completed_at/rewarded_at`：用于展示与对账

你只要记住一句话：  
`user_task_progress` 是给“任务列表展示 + 领奖校验”用的，不是给“记录行为”用的（行为用 `checkin_log`）。

### 2.3 period_key 是什么：为什么它很重要

你可以把 `period_key` 理解成“本次任务统计的桶”：

- 日任务：桶是“今天”，`period_key = 20260206`
- 周任务：桶是“本周”，`period_key = 2026W06`
- 月任务：桶是“本月”，`period_key = 202602`

这个桶的口径必须统一（尤其时区 + 周从哪天开始）。  
文档里建议用服务端时区（比如 Asia/Shanghai）+ ISO week（跨年周最稳定）。

---

## 3. 三个接口怎么设计（你看接口就能知道改了哪些数据）

### 3.1 GET /api/tasks：任务列表（从“返回任务定义”升级成“任务中心”）

你可以把它理解成：把 `task_def`（任务模板）和 `user_task_progress`（当前周期进度）拼在一起返回。

目标态建议返回字段（给前端最友好的形状）：

- `taskCode/name/periodType/targetCount/rewardPoints`
- `periodKey`：当前周期键（前端展示“本周/本月”时也能用）
- `progressCount/status/completedAt/rewardedAt`
- `claimStatus`：建议做成一个枚举：`LOCKED/CAN_CLAIM/CLAIMED`

`claimStatus` 的口径（按最直觉的规则）：

- `LOCKED`：进度没达到 `target_count`
- `CAN_CLAIM`：达到门槛且 `status=COMPLETED`
- `CLAIMED`：已经领奖，`status=REWARDED`

### 3.2 POST /api/tasks/checkin：到店打卡（写日志 + 推进进度）

它会做两件事：

1) 写一条 `checkin_log`（行为证据）  
2) 找到所有“由打卡触发”的任务，把本周期的 `user_task_progress.progress_count` 往上加（进度）

你提到“经常到店、每日打卡加积分/社区经验”，这一步是进度推进入口。

打卡最关键的业务口径是：一天允许打几次？

- 如果“一天只能算一次”：打卡接口就必须做去重（否则用户一天连点 10 次就刷满周任务）  
- 如果“一天可以多次”：那就要在任务定义里清晰表达“每天计数上限”，否则运营和用户都会困惑

### 3.3 POST /api/tasks/{taskCode}/claim：领奖（发积分，必须幂等）

领奖可以理解为：把“进度完成”转成“真实资产入账”。

它必须满足：同一用户、同一任务、同一周期，重复点多少次都最多发一次积分。

建议的双保险幂等：

1) `user_task_progress.status` 从 `COMPLETED -> REWARDED`（同周期只会成功一次）  
2) `points_ledger` 用唯一键兜底：`UNIQUE(user_id, biz_type, biz_id)`  
   - `biz_type = TASK`  
   - `biz_id = TASK-{taskCode}-{periodKey}`

积分相关表：

- `points_ledger`：账本（每次变动都写，带幂等键，能追溯）  
- `points_account`：余额快照（用于快速查询余额）

定义在：[gamesocial_init.sql](file:///w:/GOProject/gamesocial/GameSocial/gamesocial_init.sql)

---

## 4. 你提到的“社区经验/等级”：当前库里有什么？任务系统怎么接上？

目前 `user` 表里已经有：

- `exp`：经验值（累积）  
- `level`：等级（由经验映射）  

并且有一张 `user_level` 表记录“等级阈值”。定义在：

- [gamesocial_init.sql](file:///w:/GOProject/gamesocial/GameSocial/gamesocial_init.sql)

但当前项目里“经验如何增长”还没有一套像 `points_ledger` 那样的账本机制。

任务系统要支持“打卡同时加积分 + 加经验”，建议先按演进分两步想：

### 4.1 先能上线（最小可用）

- 任务定义的奖励仍以积分为主（`reward_points`）  
- 经验奖励先放到 `rule_json`（例如 `{"rewardExp": 3}`），然后领奖时同时更新 `user.exp`  
- 等级 `level` 可以用“查 user_level 表找到最新阈值”来算（或者后续做成定时/触发式）

### 4.2 做到可追溯（更完善，也是你的卖点）

经验也做账本（和积分同一套思路）：

- 新增 `exp_ledger`（或统一成 `wallet_ledger` 支持多币种：POINTS/EXP/DRINK 等）  
- 经验增长也有幂等键：同一个领奖不会重复加经验  
- 这样一来：积分、经验、饮品次数等“资产/权益”都能用同一套路解释（面试很加分）

---

## 5. 比赛系统怎么做更完善（先给规划，不着急落地）

你说的点非常对：比赛经常有，系统“完善”主要体现在两件事：

1) 运营效率：创建/发布/报名/赛程/排名/发奖尽量标准化，少人工重复劳动。  
2) 业务正确性：报名防重复、成绩发布防篡改、发奖幂等、可追溯可审计。

### 5.1 先看当前项目里比赛已经有什么（现状）

当前比赛模块已经具备“人工创建赛事 + 用户报名 + 排名查询 + 发奖落库”的基础结构：

- 业务 service：[tournament/service.go](file:///w:/GOProject/gamesocial/GameSocial/modules/tournament/service.go)
- 管理端接口：[admin_tournaments.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/admin_tournaments.go)
- 用户端接口：[app_tournaments.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/app_tournaments.go)
- 表结构：`tournament / tournament_participant / tournament_result / tournament_award`  
  定义在：[gamesocial_init.sql](file:///w:/GOProject/gamesocial/GameSocial/gamesocial_init.sql)

简单说：现在是“运营手动建比赛、用户手动报名、管理员手动录入/发布成绩并发奖”这一类。

### 5.2 你想要的“完善”，建议按 3 个版本来规划

#### V1（强运营、最稳）：人工配置赛程 + 手动发布成绩

适用：门店比赛、线下活动为主（大多数拳馆/俱乐部都先走这一版）。

- 比赛创建：管理员设置开始/结束、报名截止、人数上限、费用/积分门槛（可选）  
- 报名：`UNIQUE(tournament_id, user_id)` 防重复  
- 成绩：管理员发布排行榜（`tournament_result`），发布后只允许“更正”但要审计  
- 发奖：写 `tournament_award` + `points_ledger`（幂等键保证不重复发）

卖点：闭环完整、账务正确、审计可追溯、运营成本可控。

#### V2（半自动）：引入“比赛阶段”与“结算”

让系统可控地自动推进状态，减少人工漏操作：

- 状态建议拆细：`DRAFT -> PUBLISHED -> REGISTRATION_CLOSED -> ONGOING -> ENDED -> SETTLED`  
- 自动推进：按时间把比赛从报名到进行到结束（可用定时任务/后台 job）  
- 结算：成绩发布后执行“发奖结算”，产出发奖清单 + 幂等发放

卖点：减少运营手工切状态，数据更一致，复盘更容易。

#### V3（体验型）：随机匹配/自动开赛（更产品化）

这部分是“高级玩法”，但要注意：它会引入大量业务分支与复杂状态。

- 随机匹配：需要“匹配池 + 匹配规则（等级/体重/段位/地区）+ 超时取消”  
- 自动开赛：需要“对局/轮次/赛程表”，并且成绩由系统产生（或由裁判录入）  
- 风险点：作弊/逃跑/申诉/补赛/退款（这些都是产品需求，不是纯技术）

卖点：更像一个真正的赛事平台，但成本高，建议等 V1/V2 稳定再做。

---

## 6. 面试怎么讲（给你一套“前端转后端”的讲法）

你可以用“我先把正确性做稳，再逐步产品化”的路线讲：

1) 我先把任务系统做成：3 张表 + 账本 + 幂等键（重复领奖/并发都不怕）。  
2) 我用 `period_key` 解决“每日/每周/每月清零”的复杂度，不依赖定时任务。  
3) 积分用账本 + 余额快照，所有发奖/扣奖都可追溯可对账。  
4) 比赛系统先做 V1 人工配置闭环，后续演进到 V2 自动推进与结算，再考虑 V3 随机匹配。

---

## 7. 任务系统现状 vs 待完成（你接下来要做什么）

### 已完成

- 后台任务定义 CRUD（`/admin/task-defs`）  
  参考：[admin_task_defs.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/admin_task_defs.go)、[task/service.go](file:///w:/GOProject/gamesocial/GameSocial/modules/task/service.go)
- 用户侧任务列表接口（`GET /api/tasks`），当前仅返回启用任务定义  
  参考：[app_tasks.go](file:///w:/GOProject/gamesocial/GameSocial/api/handlers/app_tasks.go)
- DB 表已预置：`task_def / user_task_progress / checkin_log`  
  参考：[gamesocial_init.sql](file:///w:/GOProject/gamesocial/GameSocial/gamesocial_init.sql)

### 待完成（最后模块的核心）

- `POST /api/tasks/checkin`：落 `checkin_log` + 推进 `user_task_progress`（要定清楚“一天几次”口径）  
- `POST /api/tasks/{taskCode}/claim`：校验完成 + 发积分（幂等、事务一致）  
- `GET /api/tasks`：返回“定义 + 当前周期进度 + claimStatus”（从“任务定义列表”升级为“任务中心”）
