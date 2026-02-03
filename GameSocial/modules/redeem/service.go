// redeem 模块负责积分兑换订单的创建、查询与核销等业务能力。
package redeem

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// RedeemOrder 对应数据库 redeem_order 表的数据结构。
type RedeemOrder struct {
	ID            uint64            `json:"id"`
	OrderNo       string            `json:"orderNo"`
	UserID        uint64            `json:"userId"`
	Status        string            `json:"status"`
	TotalPoints   int64             `json:"totalPoints"`
	UsedByAdminID uint64            `json:"usedByAdminId,omitempty"`
	UsedAt        *time.Time        `json:"usedAt,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	Items         []RedeemOrderItem `json:"items,omitempty"`
}

// RedeemOrderItem 对应数据库 redeem_order_item 表的数据结构。
type RedeemOrderItem struct {
	ID            uint64 `json:"id"`
	RedeemOrderID uint64 `json:"redeemOrderId"`
	GoodsID       uint64 `json:"goodsId"`
	Quantity      int    `json:"quantity"`
	PointsPrice   int64  `json:"pointsPrice"`
}

// CreateOrderRequest 创建兑换订单入参（这里是最小 CRUD：不接入积分扣减流水）。
type CreateOrderRequest struct {
	UserID uint64                 `json:"userId"`
	Items  []CreateOrderItemInput `json:"items"`
}

// CreateOrderItemInput 表示创建订单时的单条兑换明细入参。
type CreateOrderItemInput struct {
	GoodsID     uint64 `json:"goodsId"`
	Quantity    int    `json:"quantity"`
	PointsPrice int64  `json:"pointsPrice"`
}

// ListOrderRequest 查询订单列表入参。
type ListOrderRequest struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Status string `json:"status"`
	UserID uint64 `json:"userId"`
}

// Service 定义 redeem 模块对外提供的业务接口（兑换订单 CRUD + 核销）。
type Service interface {
	CreateOrder(ctx context.Context, req CreateOrderRequest) (RedeemOrder, error)
	GetOrder(ctx context.Context, id uint64) (RedeemOrder, error)
	ListOrders(ctx context.Context, req ListOrderRequest) ([]RedeemOrder, error)
	UseOrder(ctx context.Context, id uint64, adminID uint64) (RedeemOrder, error)
	CancelOrder(ctx context.Context, id uint64) (RedeemOrder, error)
}

type service struct {
	db *sql.DB
}

// NewService 创建 redeem 模块服务。
func NewService(db *sql.DB) Service {
	return &service{db: db}
}

// CreateOrder 创建兑换订单并返回订单详情（包含 items）。
func (s *service) CreateOrder(ctx context.Context, req CreateOrderRequest) (RedeemOrder, error) {
	// 1) 基础校验：必须有用户与至少一条明细。
	if s.db == nil {
		return RedeemOrder{}, errors.New("database disabled")
	}
	if req.UserID == 0 {
		return RedeemOrder{}, errors.New("userId is empty")
	}
	if len(req.Items) == 0 {
		return RedeemOrder{}, errors.New("items is empty")
	}

	// 2) 计算总积分：sum(points_price * quantity)。
	var total int64
	for _, it := range req.Items {
		if it.GoodsID == 0 {
			return RedeemOrder{}, errors.New("goodsId is empty")
		}
		if it.Quantity <= 0 {
			return RedeemOrder{}, errors.New("quantity must be > 0")
		}
		if it.PointsPrice < 0 {
			return RedeemOrder{}, errors.New("pointsPrice must be >= 0")
		}
		total += int64(it.Quantity) * it.PointsPrice
	}

	// 3) 开启事务：订单表与明细表需要同时成功写入。
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RedeemOrder{}, err
	}
	defer func() { _ = tx.Rollback() }()

	orderNo, err := newOrderNo()
	if err != nil {
		return RedeemOrder{}, err
	}

	// 4) 写入 redeem_order（初始状态 CREATED）。
	res, err := tx.ExecContext(ctx, `
		INSERT INTO redeem_order (order_no, user_id, status, total_points, used_by_admin_id, used_at, created_at)
		VALUES (?, ?, 'CREATED', ?, NULL, NULL, NOW())
	`, orderNo, req.UserID, total)
	if err != nil {
		return RedeemOrder{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return RedeemOrder{}, err
	}

	// 5) 写入 redeem_order_item。
	for _, it := range req.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO redeem_order_item (redeem_order_id, goods_id, quantity, points_price)
			VALUES (?, ?, ?, ?)
		`, id, it.GoodsID, it.Quantity, it.PointsPrice); err != nil {
			return RedeemOrder{}, err
		}
	}

	// 6) 提交事务。
	if err := tx.Commit(); err != nil {
		return RedeemOrder{}, err
	}

	// 7) 返回订单详情（包含 items）。
	return s.GetOrder(ctx, uint64(id))
}

// GetOrder 获取兑换订单详情（包含 items）。
func (s *service) GetOrder(ctx context.Context, id uint64) (RedeemOrder, error) {
	// 1) 基础校验。
	if s.db == nil {
		return RedeemOrder{}, errors.New("database disabled")
	}
	if id == 0 {
		return RedeemOrder{}, errors.New("invalid id")
	}

	// 2) 读取订单主表。
	var o RedeemOrder
	var usedAdmin sql.NullInt64
	var usedAt sql.NullTime
	row := s.db.QueryRowContext(ctx, `
		SELECT id, order_no, user_id, status, total_points, used_by_admin_id, used_at, created_at
		FROM redeem_order
		WHERE id = ?
		LIMIT 1
	`, id)
	if err := row.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalPoints, &usedAdmin, &usedAt, &o.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return RedeemOrder{}, fmt.Errorf("redeem_order not found")
		}
		return RedeemOrder{}, err
	}
	if usedAdmin.Valid {
		o.UsedByAdminID = uint64(usedAdmin.Int64)
	}
	if usedAt.Valid {
		t := usedAt.Time
		o.UsedAt = &t
	}

	// 3) 读取订单明细。
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, redeem_order_id, goods_id, quantity, points_price
		FROM redeem_order_item
		WHERE redeem_order_id = ?
		ORDER BY id
	`, id)
	if err != nil {
		return RedeemOrder{}, err
	}
	defer rows.Close()

	items := make([]RedeemOrderItem, 0, 8)
	for rows.Next() {
		var it RedeemOrderItem
		if err := rows.Scan(&it.ID, &it.RedeemOrderID, &it.GoodsID, &it.Quantity, &it.PointsPrice); err != nil {
			return RedeemOrder{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return RedeemOrder{}, err
	}
	o.Items = items

	return o, nil
}

// ListOrders 获取兑换订单列表（不包含 items）。
func (s *service) ListOrders(ctx context.Context, req ListOrderRequest) ([]RedeemOrder, error) {
	// 1) 基础校验与分页兜底。
	if s.db == nil {
		return nil, errors.New("database disabled")
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// 2) 组装筛选条件：支持按 user_id/status 过滤。
	where := "WHERE 1=1"
	args := make([]any, 0, 6)
	if req.UserID != 0 {
		where += " AND user_id = ?"
		args = append(args, req.UserID)
	}
	if req.Status != "" {
		where += " AND status = ?"
		args = append(args, req.Status)
	}
	args = append(args, req.Limit, req.Offset)

	// 3) 查询主表列表（不带 items，避免列表请求过重）。
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, order_no, user_id, status, total_points, IFNULL(used_by_admin_id, 0), used_at, created_at
		FROM redeem_order
		`+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RedeemOrder, 0, req.Limit)
	for rows.Next() {
		var o RedeemOrder
		var usedAt sql.NullTime
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalPoints, &o.UsedByAdminID, &usedAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		if usedAt.Valid {
			t := usedAt.Time
			o.UsedAt = &t
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UseOrder 核销兑换订单（CREATED -> USED）并返回最新订单详情。
func (s *service) UseOrder(ctx context.Context, id uint64, adminID uint64) (RedeemOrder, error) {
	// 1) 基础校验。
	if s.db == nil {
		return RedeemOrder{}, errors.New("database disabled")
	}
	if id == 0 {
		return RedeemOrder{}, errors.New("invalid id")
	}
	if adminID == 0 {
		adminID = 1
	}

	// 2) 条件更新：只有 CREATED 才能被核销为 USED，防止重复核销。
	result, err := s.db.ExecContext(ctx, `
		UPDATE redeem_order
		SET status = 'USED', used_by_admin_id = ?, used_at = NOW()
		WHERE id = ? AND status = 'CREATED'
	`, adminID, id)
	if err != nil {
		return RedeemOrder{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return RedeemOrder{}, fmt.Errorf("order not found or not creatable")
	}
	return s.GetOrder(ctx, id)
}

// CancelOrder 取消兑换订单（CREATED -> CANCELED）并返回最新订单详情。
func (s *service) CancelOrder(ctx context.Context, id uint64) (RedeemOrder, error) {
	// 1) 基础校验。
	if s.db == nil {
		return RedeemOrder{}, errors.New("database disabled")
	}
	if id == 0 {
		return RedeemOrder{}, errors.New("invalid id")
	}

	// 2) 条件更新：只有 CREATED 才能取消。
	result, err := s.db.ExecContext(ctx, `
		UPDATE redeem_order
		SET status = 'CANCELED'
		WHERE id = ? AND status = 'CREATED'
	`, id)
	if err != nil {
		return RedeemOrder{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return RedeemOrder{}, fmt.Errorf("order not found or not cancelable")
	}
	return s.GetOrder(ctx, id)
}

func newOrderNo() (string, error) {
	// 订单号规则：R + yyyymmddhhmmss + 4 字节随机数（hex）。
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "R" + time.Now().Format("20060102150405") + hex.EncodeToString(buf), nil
}
