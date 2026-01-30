-- GameSocial 初始化脚本：建库建表 + 预置演示数据（本地开发/调试用）。
-- 说明：
-- 1) 脚本尽量可重复执行：大多数对象使用 IF NOT EXISTS / ON DUPLICATE KEY UPDATE。
-- 2) admin_user.password_hash 的 'CHANGE_ME' 只是占位，正式使用前需要替换为真实密码哈希。

-- 创建数据库（utf8mb4 便于存中文昵称/表情等）。
CREATE DATABASE IF NOT EXISTS gamesocial
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

-- 选择业务库。
USE gamesocial;

-- user：小程序用户主表（openid 唯一）。
CREATE TABLE IF NOT EXISTS `user` (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  openid VARCHAR(64) NOT NULL,
  unionid VARCHAR(64) NULL,
  nickname VARCHAR(64) NULL,
  avatar_url VARCHAR(512) NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_openid (openid),
  KEY idx_user_unionid (unionid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- user_level：用户等级/经验表（与 user 一对一）。
CREATE TABLE IF NOT EXISTS user_level (
  user_id BIGINT UNSIGNED NOT NULL,
  level INT NOT NULL DEFAULT 1,
  exp BIGINT NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id),
  CONSTRAINT fk_user_level_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- vip_subscription：VIP 订阅记录（可用于月卡等）。
CREATE TABLE IF NOT EXISTS vip_subscription (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  plan VARCHAR(32) NOT NULL,
  start_at DATETIME NOT NULL,
  end_at DATETIME NOT NULL,
  status VARCHAR(16) NOT NULL,
  pay_order_no VARCHAR(64) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_vip_subscription_pay_order_no (pay_order_no),
  KEY idx_vip_subscription_user_end (user_id, end_at),
  CONSTRAINT fk_vip_subscription_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- admin_user：后台管理员账号表（当前只预置一个管理员）。
CREATE TABLE IF NOT EXISTS admin_user (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_user_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- admin_audit_log：管理员关键操作审计日志。
CREATE TABLE IF NOT EXISTS admin_audit_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  admin_id BIGINT UNSIGNED NOT NULL,
  action VARCHAR(64) NOT NULL,
  biz_type VARCHAR(32) NULL,
  biz_id VARCHAR(64) NULL,
  detail_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_admin_audit_log_admin_id (admin_id),
  KEY idx_admin_audit_log_action (action),
  KEY idx_admin_audit_log_biz (biz_type, biz_id),
  CONSTRAINT fk_admin_audit_log_admin FOREIGN KEY (admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- points_account：积分余额快照（用于快速展示）。
CREATE TABLE IF NOT EXISTS points_account (
  user_id BIGINT UNSIGNED NOT NULL,
  balance BIGINT NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id),
  CONSTRAINT fk_points_account_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- points_ledger：积分流水账本（用唯一键保证幂等）。
CREATE TABLE IF NOT EXISTS points_ledger (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  change_amount BIGINT NOT NULL,
  balance_after BIGINT NOT NULL,
  biz_type VARCHAR(32) NOT NULL,
  biz_id VARCHAR(64) NOT NULL,
  remark VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_points_ledger_idempotent (user_id, biz_type, biz_id),
  KEY idx_points_ledger_user_created (user_id, created_at),
  CONSTRAINT fk_points_ledger_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- goods：积分商品（饮品/毛巾等）。
CREATE TABLE IF NOT EXISTS goods (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  cover_url VARCHAR(512) NULL,
  points_price BIGINT NOT NULL,
  stock INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_goods_status (status),
  KEY idx_goods_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- user_drink_balance：用户可用饮品数量（总量，不区分品类）。
CREATE TABLE IF NOT EXISTS user_drink_balance (
  user_id BIGINT UNSIGNED NOT NULL,
  quantity INT NOT NULL DEFAULT 0,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id),
  CONSTRAINT fk_user_drink_balance_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- redeem_order：积分兑换商品订单（饮料兑换不走订单）。
CREATE TABLE IF NOT EXISTS redeem_order (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(16) NOT NULL,
  total_points BIGINT NOT NULL DEFAULT 0,
  used_by_admin_id BIGINT UNSIGNED NULL,
  used_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_redeem_order_order_no (order_no),
  KEY idx_redeem_order_user (user_id),
  KEY idx_redeem_order_status (status),
  KEY idx_redeem_order_created_at (created_at),
  CONSTRAINT fk_redeem_order_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_redeem_order_used_by_admin FOREIGN KEY (used_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- redeem_order_item：兑换订单明细行。
CREATE TABLE IF NOT EXISTS redeem_order_item (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  redeem_order_id BIGINT UNSIGNED NOT NULL,
  goods_id BIGINT UNSIGNED NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  points_price BIGINT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_redeem_order_item_order (redeem_order_id),
  CONSTRAINT fk_redeem_order_item_order FOREIGN KEY (redeem_order_id) REFERENCES redeem_order(id),
  CONSTRAINT fk_redeem_order_item_goods FOREIGN KEY (goods_id) REFERENCES goods(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- tournament：赛事表。
CREATE TABLE IF NOT EXISTS tournament (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(128) NOT NULL,
  content TEXT NULL,
  cover_url VARCHAR(512) NULL,
  start_at DATETIME NOT NULL,
  end_at DATETIME NOT NULL,
  status VARCHAR(16) NOT NULL,
  created_by_admin_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_tournament_start_at (start_at),
  KEY idx_tournament_end_at (end_at),
  KEY idx_tournament_status (status),
  CONSTRAINT fk_tournament_created_by_admin FOREIGN KEY (created_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- tournament_participant：赛事报名关系表（唯一约束防止重复报名）。
CREATE TABLE IF NOT EXISTS tournament_participant (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tournament_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  join_status VARCHAR(16) NOT NULL,
  joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tournament_participant (tournament_id, user_id),
  KEY idx_tournament_participant_user_joined (user_id, joined_at),
  CONSTRAINT fk_tournament_participant_tournament FOREIGN KEY (tournament_id) REFERENCES tournament(id),
  CONSTRAINT fk_tournament_participant_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- tournament_result：赛事排名结果表（tournament_id + user_id 唯一）。
CREATE TABLE IF NOT EXISTS tournament_result (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tournament_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  rank_no INT NOT NULL,
  score INT NULL,
  published_by_admin_id BIGINT UNSIGNED NOT NULL,
  published_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tournament_result (tournament_id, user_id),
  KEY idx_tournament_result_rank (tournament_id, rank_no),
  CONSTRAINT fk_tournament_result_tournament FOREIGN KEY (tournament_id) REFERENCES tournament(id),
  CONSTRAINT fk_tournament_result_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_tournament_result_published_by_admin FOREIGN KEY (published_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- tournament_award：赛事发奖记录（user_id + biz_id 唯一，用于幂等）。
CREATE TABLE IF NOT EXISTS tournament_award (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tournament_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  award_points BIGINT NOT NULL,
  biz_id VARCHAR(64) NOT NULL,
  created_by_admin_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tournament_award_user_biz (user_id, biz_id),
  KEY idx_tournament_award_tournament_created (tournament_id, created_at),
  CONSTRAINT fk_tournament_award_tournament FOREIGN KEY (tournament_id) REFERENCES tournament(id),
  CONSTRAINT fk_tournament_award_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_tournament_award_created_by_admin FOREIGN KEY (created_by_admin_id) REFERENCES admin_user(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- task_def：任务定义（每日/每周/每月）。
CREATE TABLE IF NOT EXISTS task_def (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  task_code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  period_type VARCHAR(16) NOT NULL,
  target_count INT NOT NULL,
  reward_points BIGINT NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  rule_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_task_def_task_code (task_code),
  KEY idx_task_def_status (status),
  KEY idx_task_def_period_type (period_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- user_task_progress：用户任务进度（user_id + task_id + period_key 唯一，用于同周期幂等更新）。
CREATE TABLE IF NOT EXISTS user_task_progress (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  task_id BIGINT UNSIGNED NOT NULL,
  period_key VARCHAR(16) NOT NULL,
  progress_count INT NOT NULL DEFAULT 0,
  status VARCHAR(16) NOT NULL,
  completed_at DATETIME NULL,
  rewarded_at DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_task_progress (user_id, task_id, period_key),
  KEY idx_user_task_progress_user_period (user_id, period_key),
  CONSTRAINT fk_user_task_progress_user FOREIGN KEY (user_id) REFERENCES `user`(id),
  CONSTRAINT fk_user_task_progress_task FOREIGN KEY (task_id) REFERENCES task_def(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- checkin_log：到店打卡明细。
CREATE TABLE IF NOT EXISTS checkin_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  checkin_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  source VARCHAR(16) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_checkin_log_user (user_id),
  KEY idx_checkin_log_checkin_at (checkin_at),
  CONSTRAINT fk_checkin_log_user FOREIGN KEY (user_id) REFERENCES `user`(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
INSERT INTO `user` (id, openid, unionid, nickname, avatar_url, status, created_at, updated_at)
VALUES
  (1001, 'openid_u1001', NULL, '阿哲', 'https://example.com/avatar/u1001.png', 1, NOW(), NOW()),
  (1002, 'openid_u1002', NULL, '小明', 'https://example.com/avatar/u1002.png', 1, NOW(), NOW()),
  (1003, 'openid_u1003', NULL, '阿强', 'https://example.com/avatar/u1003.png', 1, NOW(), NOW()),
  (1004, 'openid_u1004', NULL, '小雨', 'https://example.com/avatar/u1004.png', 1, NOW(), NOW()),
  (1005, 'openid_u1005', NULL, 'Kyo',  'https://example.com/avatar/u1005.png', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  unionid = VALUES(unionid),
  nickname = VALUES(nickname),
  avatar_url = VALUES(avatar_url),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

-- 预置等级数据（开发用）。
INSERT INTO user_level (user_id, level, exp, updated_at)
VALUES
  (1001, 3, 120, NOW()),
  (1002, 1, 10, NOW()),
  (1003, 5, 520, NOW()),
  (1004, 2, 60, NOW()),
  (1005, 4, 260, NOW())
ON DUPLICATE KEY UPDATE
  level = VALUES(level),
  exp = VALUES(exp),
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
INSERT INTO goods (id, name, cover_url, points_price, stock, status, created_at)
VALUES
  (2001, '饮料（兑换 +1 杯）', NULL, 50, 0, 1, NOW()),
  (2002, '拳馆毛巾', NULL, 200, 0, 1, NOW()),
  (2003, '手套消耗品', NULL, 120, 0, 1, NOW()),
  (2004, '能量饮料（实物）', NULL, 100, 0, 1, NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  cover_url = VALUES(cover_url),
  points_price = VALUES(points_price),
  stock = VALUES(stock),
  status = VALUES(status);

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
