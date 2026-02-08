-- GameSocial 初始化脚本：建库建表 + 预置演示数据（本地开发/调试用）。
-- 说明：
-- 1) 脚本可重复执行：建表前会 DROP TABLE IF EXISTS；预置数据使用 ON DUPLICATE KEY UPDATE。
-- 2) admin_user.password_hash 的 'CHANGE_ME' 只是占位，正式使用前需要替换为真实密码哈希。

-- 创建数据库（utf8mb4 便于存中文昵称/表情等）。
CREATE DATABASE IF NOT EXISTS gamesocial
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

-- 选择业务库。
USE `gamesocial`;

-- 数据库升级提示（9527：遇到 Unknown column 必看）
-- 如果你在接口调用时遇到：
-- Error 1054 (42S22): Unknown column 't.image_urls_json' in 'field list'
-- 说明数据库还是旧结构，缺少多图字段。请在“目标库”单独执行下面两条 ALTER（不要整库 DROP 重建）：
--
-- ALTER TABLE tournament
--   ADD COLUMN image_urls_json JSON NULL COMMENT '赛事图片 URL 列表 JSON（可为空）' AFTER cover_url;

-- ALTER TABLE goods
--   ADD COLUMN image_urls_json JSON NULL COMMENT '商品图片 URL 列表 JSON（可为空）' AFTER cover_url;
--
-- ALTER TABLE goods
--   ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间' AFTER created_at;
--
-- 重置表结构：如果表已存在则先删除再创建（开发/调试用）。
DROP TABLE IF EXISTS
  checkin_log,
  user_task_progress,
  task_def,
  tournament_award,
  tournament_result,
  tournament_participant,
  tournament,
  redeem_order_item,
  redeem_order,
  user_drink_balance,
  goods,
  points_ledger,
  points_account,
  admin_audit_log,
  vip_subscription,
  user_level,
  admin_user,
  `user`;

-- user：小程序用户主表（openid 唯一）。
CREATE TABLE `user` (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  openid VARCHAR(64) NOT NULL COMMENT '微信 openid（唯一标识小程序用户）',
  unionid VARCHAR(64) NULL COMMENT '微信 unionid（同主体下多应用统一标识，可为空）',
  nickname VARCHAR(64) NULL COMMENT '用户昵称',
  avatar_url VARCHAR(512) NULL COMMENT '头像 URL',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=正常；0=禁用',
  level INT NOT NULL DEFAULT 1 COMMENT '用户等级（由经验值映射，默认 1）',
  exp BIGINT NOT NULL DEFAULT 0 COMMENT '经验值（累积）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_openid (openid),
  KEY idx_user_unionid (unionid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='小程序用户主表（openid 唯一）';

-- user_level：等级配置表（用于管理员配置每个等级所需经验阈值）。
CREATE TABLE user_level (
  level INT NOT NULL COMMENT '等级（从 1 开始）',
  exp_required BIGINT NOT NULL COMMENT '达到该等级所需经验（阈值）',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='等级配置（经验阈值）';

-- vip_subscription：VIP 订阅记录（可用于月卡等）。
CREATE TABLE vip_subscription (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id）',
  plan VARCHAR(32) NOT NULL COMMENT '订阅套餐编码（例如 MONTH_CARD）',
  start_at DATETIME NOT NULL COMMENT '生效时间',
  end_at DATETIME NOT NULL COMMENT '到期时间',
  status VARCHAR(16) NOT NULL COMMENT '订阅状态（例如 ACTIVE/EXPIRED/CANCELED）',
  vip_level INT NOT NULL DEFAULT 1 COMMENT '本次开通的会员等级（>=1）',
  amount_cent BIGINT NOT NULL DEFAULT 0 COMMENT '充值金额（分）',
  pay_channel VARCHAR(32) NULL COMMENT '支付渠道（可为空）',
  pay_order_no VARCHAR(64) NULL COMMENT '支付订单号（用于幂等/对账，可为空）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_vip_subscription_pay_order_no (pay_order_no),
  KEY idx_vip_subscription_user_end (user_id, end_at),
  CONSTRAINT fk_vip_subscription_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VIP 订阅记录（用于月卡/会员等）';

-- admin_user：后台管理员账号表（当前只预置一个管理员）。
CREATE TABLE admin_user (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  username VARCHAR(64) NOT NULL COMMENT '登录用户名（唯一）',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希（禁止存明文）',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=启用；0=禁用',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_user_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理员账号';

-- admin_audit_log：管理员关键操作审计日志。
CREATE TABLE admin_audit_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  admin_id BIGINT UNSIGNED NOT NULL COMMENT '管理员 ID（对应 admin_user.id）',
  action VARCHAR(64) NOT NULL COMMENT '操作动作（如 POINTS_ADJUST/DRINK_USE）',
  biz_type VARCHAR(32) NULL COMMENT '业务类型（如 USER/REDEEM_ORDER，可为空）',
  biz_id VARCHAR(64) NULL COMMENT '业务标识（如 userId/orderNo，可为空）',
  detail_json JSON NULL COMMENT '操作详情 JSON（扩展字段）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_admin_audit_log_admin_id (admin_id),
  KEY idx_admin_audit_log_action (action),
  KEY idx_admin_audit_log_biz (biz_type, biz_id),
  CONSTRAINT fk_admin_audit_log_admin FOREIGN KEY (admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员关键操作审计日志';

-- points_account：积分余额快照（用于快速展示）。
CREATE TABLE points_account (
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id，一对一）',
  balance BIGINT NOT NULL DEFAULT 0 COMMENT '积分余额（快照）',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (user_id),
  CONSTRAINT fk_points_account_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分账户余额快照（用于快速查询）';

-- points_ledger：积分流水账本（用唯一键保证幂等）。
CREATE TABLE points_ledger (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id）',
  change_amount BIGINT NOT NULL COMMENT '本次积分变动（正=增加；负=扣减）',
  balance_after BIGINT NOT NULL COMMENT '变动后的余额（用于展示/校验）',
  biz_type VARCHAR(32) NOT NULL COMMENT '业务类型（用于幂等/追踪，例如 INIT/CHECKIN/REDEEM）',
  biz_id VARCHAR(64) NOT NULL COMMENT '业务唯一标识（同 user_id+biz_type 唯一）',
  remark VARCHAR(255) NULL COMMENT '备注说明（可为空）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_points_ledger_idempotent (user_id, biz_type, biz_id),
  KEY idx_points_ledger_user_created (user_id, created_at),
  CONSTRAINT fk_points_ledger_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分流水账本（含幂等键）';

-- goods：积分商品（饮品/毛巾等）。
CREATE TABLE goods (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  name VARCHAR(128) NOT NULL COMMENT '商品名称',
  cover_url VARCHAR(512) NULL COMMENT '封面图 URL（可为空）',
  image_urls_json JSON NULL COMMENT '商品图片 URL 列表 JSON（可为空）',
  points_price BIGINT NOT NULL COMMENT '兑换所需积分（>=0）',
  stock INT NOT NULL DEFAULT 0 COMMENT '库存（>=0；可用于实物）',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=上架；0=下架/删除',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_goods_status (status),
  KEY idx_goods_created_at (created_at),
  KEY idx_goods_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分商品（饮品/毛巾等）';

-- user_drink_balance：用户可用饮品数量（总量，不区分品类）。
CREATE TABLE user_drink_balance (
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id，一对一）',
  quantity INT NOT NULL DEFAULT 0 COMMENT '可用饮品数量（杯数）',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (user_id),
  CONSTRAINT fk_user_drink_balance_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户饮品数量余额（总量，不区分品类）';

-- redeem_order：积分兑换商品订单（饮料兑换不走订单）。
CREATE TABLE redeem_order (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  order_no VARCHAR(64) NOT NULL COMMENT '订单号（业务唯一，用于展示/幂等）',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '下单用户 ID（对应 user.id）',
  status VARCHAR(16) NOT NULL COMMENT '订单状态（例如 CREATED/USED/CANCELED）',
  total_points BIGINT NOT NULL DEFAULT 0 COMMENT '订单总积分（汇总值）',
  used_by_admin_id BIGINT UNSIGNED NULL COMMENT '核销管理员 ID（对应 admin_user.id，可为空）',
  used_at DATETIME NULL COMMENT '核销时间（可为空）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_redeem_order_order_no (order_no),
  KEY idx_redeem_order_user (user_id),
  KEY idx_redeem_order_status (status),
  KEY idx_redeem_order_created_at (created_at),
  CONSTRAINT fk_redeem_order_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_redeem_order_used_by_admin FOREIGN KEY (used_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分兑换订单（用于实物/券类核销）';

-- redeem_order_item：兑换订单明细行。
CREATE TABLE redeem_order_item (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  redeem_order_id BIGINT UNSIGNED NOT NULL COMMENT '兑换订单 ID（对应 redeem_order.id）',
  goods_id BIGINT UNSIGNED NOT NULL COMMENT '商品 ID（对应 goods.id）',
  quantity INT NOT NULL DEFAULT 1 COMMENT '数量（>=1）',
  points_price BIGINT NOT NULL COMMENT '下单时商品积分单价（快照）',
  PRIMARY KEY (id),
  KEY idx_redeem_order_item_order (redeem_order_id),
  CONSTRAINT fk_redeem_order_item_order FOREIGN KEY (redeem_order_id) REFERENCES redeem_order(id),
  CONSTRAINT fk_redeem_order_item_goods FOREIGN KEY (goods_id) REFERENCES goods(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='兑换订单明细行（商品快照）';

-- tournament：赛事表。
CREATE TABLE tournament (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  title VARCHAR(128) NOT NULL COMMENT '赛事标题',
  content TEXT NULL COMMENT '赛事内容/详情（可为空）',
  cover_url VARCHAR(512) NULL COMMENT '封面图 URL（可为空）',
  image_urls_json JSON NULL COMMENT '赛事图片 URL 列表 JSON（可为空）',
  start_at DATETIME NOT NULL COMMENT '开始时间',
  end_at DATETIME NOT NULL COMMENT '结束时间',
  status VARCHAR(16) NOT NULL COMMENT '赛事状态（例如 DRAFT/PUBLISHED/ENDED/CANCELED）',
  created_by_admin_id BIGINT UNSIGNED NOT NULL COMMENT '创建管理员 ID（对应 admin_user.id）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_tournament_start_at (start_at),
  KEY idx_tournament_end_at (end_at),
  KEY idx_tournament_status (status),
  CONSTRAINT fk_tournament_created_by_admin FOREIGN KEY (created_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='赛事主表';

-- tournament_participant：赛事报名关系表（唯一约束防止重复报名）。
CREATE TABLE tournament_participant (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  tournament_id BIGINT UNSIGNED NOT NULL COMMENT '赛事 ID（对应 tournament.id）',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id）',
  join_status VARCHAR(16) NOT NULL COMMENT '报名状态（例如 JOINED/CANCELED）',
  joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '报名时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_tournament_participant (tournament_id, user_id),
  KEY idx_tournament_participant_user_joined (user_id, joined_at),
  CONSTRAINT fk_tournament_participant_tournament FOREIGN KEY (tournament_id) REFERENCES tournament(id),
  CONSTRAINT fk_tournament_participant_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='赛事报名关系（防重复报名）';

-- tournament_result：赛事排名结果表（tournament_id + user_id 唯一）。
CREATE TABLE tournament_result (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  tournament_id BIGINT UNSIGNED NOT NULL COMMENT '赛事 ID（对应 tournament.id）',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id）',
  rank_no INT NOT NULL COMMENT '名次（从 1 开始）',
  score INT NULL COMMENT '成绩/分数（可为空）',
  published_by_admin_id BIGINT UNSIGNED NOT NULL COMMENT '发布管理员 ID（对应 admin_user.id）',
  published_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_tournament_result (tournament_id, user_id),
  KEY idx_tournament_result_rank (tournament_id, rank_no),
  CONSTRAINT fk_tournament_result_tournament FOREIGN KEY (tournament_id) REFERENCES tournament(id),
  CONSTRAINT fk_tournament_result_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_tournament_result_published_by_admin FOREIGN KEY (published_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='赛事成绩/排名结果';

-- tournament_award：赛事发奖记录（user_id + biz_id 唯一，用于幂等）。
CREATE TABLE tournament_award (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  tournament_id BIGINT UNSIGNED NOT NULL COMMENT '赛事 ID（对应 tournament.id）',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '获奖用户 ID（对应 user.id）',
  award_points BIGINT NOT NULL COMMENT '发放积分数量（>=0）',
  biz_id VARCHAR(64) NOT NULL COMMENT '幂等业务标识（用于防重复发奖）',
  created_by_admin_id BIGINT UNSIGNED NOT NULL COMMENT '发奖管理员 ID（对应 admin_user.id）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_tournament_award_user_biz (user_id, biz_id),
  KEY idx_tournament_award_tournament_created (tournament_id, created_at),
  CONSTRAINT fk_tournament_award_tournament FOREIGN KEY (tournament_id) REFERENCES tournament(id),
  CONSTRAINT fk_tournament_award_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_tournament_award_created_by_admin FOREIGN KEY (created_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='赛事发奖记录（含幂等键）';

-- task_def：任务定义（每日/每周/每月）。
CREATE TABLE task_def (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  task_code VARCHAR(64) NOT NULL COMMENT '任务编码（唯一，用于程序引用）',
  name VARCHAR(128) NOT NULL COMMENT '任务名称',
  period_type VARCHAR(16) NOT NULL COMMENT '周期类型（DAILY/WEEKLY/MONTHLY）',
  target_count INT NOT NULL COMMENT '目标次数（达到后可领奖）',
  reward_points BIGINT NOT NULL COMMENT '奖励积分数量（>=0）',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=启用；0=禁用',
  rule_json JSON NULL COMMENT '任务规则扩展 JSON（可为空）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_task_def_task_code (task_code),
  KEY idx_task_def_status (status),
  KEY idx_task_def_period_type (period_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务定义（周期/目标/奖励）';

-- user_task_progress：用户任务进度（user_id + task_id + period_key 唯一，用于同周期幂等更新）。
CREATE TABLE user_task_progress (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id）',
  task_id BIGINT UNSIGNED NOT NULL COMMENT '任务定义 ID（对应 task_def.id）',
  period_key VARCHAR(16) NOT NULL COMMENT '周期键（如 202601、2026W05，用于同周期去重）',
  progress_count INT NOT NULL DEFAULT 0 COMMENT '当前完成次数',
  status VARCHAR(16) NOT NULL COMMENT '进度状态（例如 IN_PROGRESS/COMPLETED/REWARDED）',
  completed_at DATETIME NULL COMMENT '完成时间（可为空）',
  rewarded_at DATETIME NULL COMMENT '领奖时间（可为空）',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_task_progress (user_id, task_id, period_key),
  KEY idx_user_task_progress_user_period (user_id, period_key),
  CONSTRAINT fk_user_task_progress_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_user_task_progress_task FOREIGN KEY (task_id) REFERENCES task_def(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户任务进度（按周期维度幂等）';

-- checkin_log：到店打卡明细。
CREATE TABLE checkin_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID（对应 user.id）',
  checkin_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '打卡时间',
  source VARCHAR(16) NOT NULL COMMENT '打卡来源（例如 MANUAL/WECHAT/GPS）',
  PRIMARY KEY (id),
  KEY idx_checkin_log_user (user_id),
  KEY idx_checkin_log_checkin_at (checkin_at),
  CONSTRAINT fk_checkin_log_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='到店打卡记录';

-- 预置管理员账号（开发用）。
INSERT INTO admin_user (id, username, password_hash, status, created_at, updated_at)
VALUES (1, 'admin', 'CHANGE_ME', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  username = VALUES(username),
  password_hash = VALUES(password_hash),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

-- 预置任务定义（开发用）。
INSERT INTO task_def (task_code, name, period_type, target_count, reward_points, status, rule_json, created_at)
VALUES
  ('DAILY_CHECKIN', '到店打卡（每日）', 'DAILY', 1, 1, 1, NULL, NOW()),
  ('WEEKLY_CHECKIN_3', '到店打卡（每周3次）', 'WEEKLY', 3, 5, 1, NULL, NOW()),
  ('MONTHLY_CHECKIN_10', '到店打卡（每月10次）', 'MONTHLY', 10, 20, 1, NULL, NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  period_type = VALUES(period_type),
  target_count = VALUES(target_count),
  reward_points = VALUES(reward_points),
  status = VALUES(status),
  rule_json = VALUES(rule_json);

-- 预置演示用户（开发用）。
INSERT INTO `user` (id, openid, unionid, nickname, avatar_url, status, level, exp, created_at, updated_at)
VALUES
  (1001, 'openid_u1001', NULL, '阿哲', 'https://example.com/avatar/u1001.png', 1, 2, 120, NOW(), NOW()),
  (1002, 'openid_u1002', NULL, '小明', 'https://example.com/avatar/u1002.png', 1, 1, 10, NOW(), NOW()),
  (1003, 'openid_u1003', NULL, '阿强', 'https://example.com/avatar/u1003.png', 1, 4, 520, NOW(), NOW()),
  (1004, 'openid_u1004', NULL, '小雨', 'https://example.com/avatar/u1004.png', 1, 2, 60, NOW(), NOW()),
  (1005, 'openid_u1005', NULL, 'Kyo',  'https://example.com/avatar/u1005.png', 1, 3, 260, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  unionid = VALUES(unionid),
  nickname = VALUES(nickname),
  avatar_url = VALUES(avatar_url),
  status = VALUES(status),
  level = VALUES(level),
  exp = VALUES(exp),
  updated_at = VALUES(updated_at);

-- 预置等级数据（开发用）。
INSERT INTO user_level (level, exp_required, updated_at)
VALUES
  (1, 0, NOW()),
  (2, 100, NOW()),
  (3, 200, NOW()),
  (4, 500, NOW()),
  (5, 800, NOW())
ON DUPLICATE KEY UPDATE
  exp_required = VALUES(exp_required),
  updated_at = VALUES(updated_at);

-- 预置积分余额（开发用）。
INSERT INTO points_account (user_id, balance, updated_at)
VALUES
  (1001, 300, NOW()),
  (1002, 80, NOW()),
  (1003, 1200, NOW()),
  (1004, 150, NOW()),
  (1005, 500, NOW())
ON DUPLICATE KEY UPDATE
  balance = VALUES(balance),
  updated_at = VALUES(updated_at);

-- 预置饮品数量（开发用）。
INSERT INTO user_drink_balance (user_id, quantity, updated_at)
VALUES
  (1001, 2, NOW()),
  (1002, 0, NOW()),
  (1003, 5, NOW()),
  (1004, 1, NOW()),
  (1005, 3, NOW())
ON DUPLICATE KEY UPDATE
  quantity = VALUES(quantity),
  updated_at = VALUES(updated_at);

-- 预置商品（开发用）。
INSERT INTO goods (id, name, cover_url, points_price, stock, status, created_at, updated_at)
VALUES
  (2001, '饮料（兑换 +1 杯）', NULL, 50, 0, 1, NOW(), NOW()),
  (2002, '拳馆毛巾', NULL, 200, 0, 1, NOW(), NOW()),
  (2003, '手套消耗品', NULL, 120, 0, 1, NOW(), NOW()),
  (2004, '能量饮料（实物）', NULL, 100, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  cover_url = VALUES(cover_url),
  points_price = VALUES(points_price),
  stock = VALUES(stock),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

-- 预置积分流水（开发用，演示幂等键 biz_type + biz_id）。
INSERT INTO points_ledger (user_id, change_amount, balance_after, biz_type, biz_id, remark, created_at)
VALUES
  (1001, 300, 300, 'INIT', 'INIT-1001', '初始化积分', NOW()),
  (1001, -50, 250, 'DRINK_EXCHANGE', 'DRINK-EX-1001-1', '积分兑换饮料 +1', NOW()),
  (1002, 80, 80, 'INIT', 'INIT-1002', '初始化积分', NOW()),
  (1003, 1200, 1200, 'INIT', 'INIT-1003', '初始化积分', NOW()),
  (1004, 150, 150, 'INIT', 'INIT-1004', '初始化积分', NOW()),
  (1005, 500, 500, 'INIT', 'INIT-1005', '初始化积分', NOW())
ON DUPLICATE KEY UPDATE
  remark = VALUES(remark);

-- 预置管理员操作日志（开发用）。
INSERT INTO admin_audit_log (admin_id, action, biz_type, biz_id, detail_json, created_at)
VALUES
  (1, 'DRINK_USE', 'USER', '1001', JSON_OBJECT('user_id', 1001, 'delta', -1), NOW()),
  (1, 'POINTS_ADJUST', 'USER', '1004', JSON_OBJECT('user_id', 1004, 'delta', 50, 'reason', '线下活动奖励'), NOW()),
  (1, 'GOODS_REDEEM_USE', 'REDEEM_ORDER', 'R202601280001', JSON_OBJECT('order_no', 'R202601280001'), NOW());

-- 预置兑换订单与明细（开发用）。
INSERT INTO redeem_order (id, order_no, user_id, status, total_points, used_by_admin_id, used_at, created_at)
VALUES
  (3001, 'R202601280001', 1002, 'USED', 200, 1, NOW(), NOW()),
  (3002, 'R202601280002', 1005, 'CREATED', 120, NULL, NULL, NOW())
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  total_points = VALUES(total_points),
  used_by_admin_id = VALUES(used_by_admin_id),
  used_at = VALUES(used_at);

INSERT INTO redeem_order_item (redeem_order_id, goods_id, quantity, points_price)
VALUES
  (3001, 2002, 1, 200),
  (3002, 2003, 1, 120);

-- 预置赛事、报名、排名与发奖（开发用）。
INSERT INTO tournament (id, title, content, cover_url, start_at, end_at, status, created_by_admin_id, created_at, updated_at)
VALUES
  (4001, '周末友谊赛', '周末店内友谊赛，欢迎报名', NULL, '2026-02-01 14:00:00', '2026-02-01 18:00:00', 'PUBLISHED', 1, NOW(), NOW()),
  (4002, '月度排位赛', '月度排位赛，按积分发奖', NULL, '2026-02-15 14:00:00', '2026-02-15 19:00:00', 'DRAFT', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  content = VALUES(content),
  cover_url = VALUES(cover_url),
  start_at = VALUES(start_at),
  end_at = VALUES(end_at),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO tournament_participant (tournament_id, user_id, join_status, joined_at)
VALUES
  (4001, 1001, 'JOINED', NOW()),
  (4001, 1002, 'JOINED', NOW()),
  (4001, 1003, 'JOINED', NOW()),
  (4001, 1004, 'JOINED', NOW())
ON DUPLICATE KEY UPDATE
  join_status = VALUES(join_status),
  joined_at = VALUES(joined_at);

INSERT INTO tournament_result (tournament_id, user_id, rank_no, score, published_by_admin_id, published_at)
VALUES
  (4001, 1003, 1, 10, 1, NOW()),
  (4001, 1001, 2, 8, 1, NOW()),
  (4001, 1002, 3, 6, 1, NOW())
ON DUPLICATE KEY UPDATE
  rank_no = VALUES(rank_no),
  score = VALUES(score),
  published_by_admin_id = VALUES(published_by_admin_id),
  published_at = VALUES(published_at);

INSERT INTO tournament_award (tournament_id, user_id, award_points, biz_id, created_by_admin_id, created_at)
VALUES
  (4001, 1003, 100, 'AWARD-4001-1003', 1, NOW()),
  (4001, 1001, 50,  'AWARD-4001-1001', 1, NOW()),
  (4001, 1002, 20,  'AWARD-4001-1002', 1, NOW())
ON DUPLICATE KEY UPDATE
  award_points = VALUES(award_points),
  created_by_admin_id = VALUES(created_by_admin_id);

-- 预置打卡记录（开发用）。
INSERT INTO checkin_log (user_id, checkin_at, source)
VALUES
  (1001, NOW(), 'MANUAL'),
  (1002, NOW(), 'MANUAL'),
  (1003, NOW(), 'MANUAL'),
  (1004, NOW(), 'MANUAL'),
  (1005, NOW(), 'MANUAL');

-- 预置任务进度（开发用）：给用户 1001 初始化一个月任务进度。
INSERT INTO user_task_progress (user_id, task_id, period_key, progress_count, status, completed_at, rewarded_at)
SELECT
  1001,
  t.id,
  '202601',
  3,
  'IN_PROGRESS',
  NULL,
  NULL
FROM task_def t
WHERE t.task_code = 'MONTHLY_CHECKIN_10'
ON DUPLICATE KEY UPDATE
  progress_count = VALUES(progress_count),
  status = VALUES(status);
