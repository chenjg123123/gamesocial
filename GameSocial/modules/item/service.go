// item 模块负责商品/饮品的管理与查询等业务能力。
package item

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// Goods 对应数据库 goods 表的数据结构。
type Goods struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	CoverURL    string    `json:"coverUrl,omitempty"`
	ImageURLs   []string  `json:"imageUrls,omitempty"`
	PointsPrice int64     `json:"pointsPrice"`
	Stock       int       `json:"stock"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateGoodsRequest 创建商品的入参。
type CreateGoodsRequest struct {
	Name        string   `json:"name"`
	CoverURL    string   `json:"coverUrl"`
	ImageURLs   []string `json:"imageUrls,omitempty"`
	PointsPrice int64    `json:"pointsPrice"`
	Stock       int      `json:"stock"`
	Status      int      `json:"status"`
}

// UpdateGoodsRequest 更新商品入参（只更新可变字段）。
type UpdateGoodsRequest struct {
	Name        string   `json:"name"`
	CoverURL    string   `json:"coverUrl"`
	ImageURLs   []string `json:"imageUrls,omitempty"`
	PointsPrice int64    `json:"pointsPrice"`
	Stock       int      `json:"stock"`
	Status      int      `json:"status"`
}

// ListGoodsRequest 列表查询入参（分页 + 状态筛选）。
type ListGoodsRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Status int `json:"status"`
}

// Service 定义 item 模块对外提供的业务接口（商品 CRUD）。
type Service interface {
	CreateGoods(ctx context.Context, req CreateGoodsRequest) (Goods, error)
	UpdateGoods(ctx context.Context, id uint64, req UpdateGoodsRequest) (Goods, error)
	DeleteGoods(ctx context.Context, id uint64) error
	GetGoods(ctx context.Context, id uint64) (Goods, error)
	ListGoods(ctx context.Context, req ListGoodsRequest) ([]Goods, error)
}

type service struct {
	db *sql.DB
}

// NewService 创建 item 模块服务。
func NewService(db *sql.DB) Service {
	return &service{db: db}
}

// CreateGoods 创建商品并返回创建后的商品详情。
func (s *service) CreateGoods(ctx context.Context, req CreateGoodsRequest) (Goods, error) {
	// 1) 基础校验：DB 必须可用，参数必须满足最小约束。
	if s.db == nil {
		return Goods{}, errors.New("database disabled")
	}
	if req.Name == "" {
		return Goods{}, errors.New("name is empty")
	}
	if req.PointsPrice < 0 {
		return Goods{}, errors.New("points_price must be >= 0")
	}
	if req.Stock < 0 {
		return Goods{}, errors.New("stock must be >= 0")
	}
	if req.Status == 0 {
		req.Status = 1
	}

	if len(req.ImageURLs) == 0 && req.CoverURL != "" {
		req.ImageURLs = []string{req.CoverURL}
	}
	if len(req.ImageURLs) > 0 {
		req.CoverURL = req.ImageURLs[0]
	}

	imageURLsJSON := ""
	if len(req.ImageURLs) > 0 {
		b, err := json.Marshal(req.ImageURLs)
		if err != nil {
			return Goods{}, errors.New("invalid imageUrls")
		}
		imageURLsJSON = string(b)
	}

	// 2) 写入 goods 表：cover_url 允许为空，因此用 NULLIF(?, '') 转成 NULL。
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO goods (name, cover_url, image_urls_json, points_price, stock, status, created_at)
		VALUES (?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, NOW())
	`, req.Name, req.CoverURL, imageURLsJSON, req.PointsPrice, req.Stock, req.Status)
	if err != nil && isUnknownColumn(err, "image_urls_json") {
		res, err = s.db.ExecContext(ctx, `
			INSERT INTO goods (name, cover_url, points_price, stock, status, created_at)
			VALUES (?, NULLIF(?, ''), ?, ?, ?, NOW())
		`, req.Name, req.CoverURL, req.PointsPrice, req.Stock, req.Status)
	}
	if err != nil {
		return Goods{}, err
	}
	// 3) 取回自增 id，并返回最新详情（包含 created_at）。
	id, err := res.LastInsertId()
	if err != nil {
		return Goods{}, err
	}
	return s.GetGoods(ctx, uint64(id))
}

// UpdateGoods 更新商品并返回更新后的商品详情。
func (s *service) UpdateGoods(ctx context.Context, id uint64, req UpdateGoodsRequest) (Goods, error) {
	// 1) 基础校验：id/参数合法性。
	if s.db == nil {
		return Goods{}, errors.New("database disabled")
	}
	if id == 0 {
		return Goods{}, errors.New("invalid id")
	}
	if req.Name == "" {
		return Goods{}, errors.New("name is empty")
	}
	if req.PointsPrice < 0 {
		return Goods{}, errors.New("points_price must be >= 0")
	}
	if req.Stock < 0 {
		return Goods{}, errors.New("stock must be >= 0")
	}
	if req.Status == 0 {
		req.Status = 1
	}

	if len(req.ImageURLs) == 0 && req.CoverURL != "" {
		req.ImageURLs = []string{req.CoverURL}
	}
	if len(req.ImageURLs) > 0 {
		req.CoverURL = req.ImageURLs[0]
	}

	imageURLsJSON := ""
	if len(req.ImageURLs) > 0 {
		b, err := json.Marshal(req.ImageURLs)
		if err != nil {
			return Goods{}, errors.New("invalid imageUrls")
		}
		imageURLsJSON = string(b)
	}
	// 2) 更新 goods 表可变字段。
	result, err := s.db.ExecContext(ctx, `
		UPDATE goods
		SET name = ?, cover_url = NULLIF(?, ''), image_urls_json = NULLIF(?, ''), points_price = ?, stock = ?, status = ?
		WHERE id = ?
	`, req.Name, req.CoverURL, imageURLsJSON, req.PointsPrice, req.Stock, req.Status, id)
	if err != nil && isUnknownColumn(err, "image_urls_json") {
		result, err = s.db.ExecContext(ctx, `
			UPDATE goods
			SET name = ?, cover_url = NULLIF(?, ''), points_price = ?, stock = ?, status = ?
			WHERE id = ?
		`, req.Name, req.CoverURL, req.PointsPrice, req.Stock, req.Status, id)
	}
	if err != nil {
		log.Printf("UpdateGoods: id=%d, req=%+v", id, req)
		return Goods{}, err
	}
	// 3) 检查 RowsAffected=0：代表该 id 不存在。
	affected, _ := result.RowsAffected()
	// 如果用户未修改商品信息，RowsAffected=0 也会返回错误。
	if affected == 0 {
		return Goods{}, fmt.Errorf("修改失败，可能是为进行修改或商品不存在")
	}
	// 4) 返回最新详情。
	return s.GetGoods(ctx, id)
}

// DeleteGoods 软删除商品（status=0）。
func (s *service) DeleteGoods(ctx context.Context, id uint64) error {
	// 1) 基础校验。
	if s.db == nil {
		return errors.New("database disabled")
	}
	if id == 0 {
		return errors.New("invalid id")
	}

	// 2) 软删除：把 status 置为 0，避免历史订单/引用数据丢失。
	result, err := s.db.ExecContext(ctx, `
		UPDATE goods
		SET status = 0
		WHERE id = ? AND status <> 0
	`, id)
	if err != nil {
		return err
	}
	// 3) 若没有行被更新，代表商品不存在或已删除。
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("goods not found")
	}
	return nil
}

// GetGoods 获取单个商品详情。
func (s *service) GetGoods(ctx context.Context, id uint64) (Goods, error) {
	// 1) 基础校验。
	if s.db == nil {
		return Goods{}, errors.New("database disabled")
	}
	if id == 0 {
		return Goods{}, errors.New("invalid id")
	}

	// 2) 查询单条记录：cover_url 可空，用 sql.NullString 承接。
	var g Goods
	var cover, imageURLs sql.NullString
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, cover_url, image_urls_json, points_price, stock, status, created_at, updated_at
		FROM goods
		WHERE id = ?
		LIMIT 1
	`, id)
	err := row.Scan(&g.ID, &g.Name, &cover, &imageURLs, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt, &g.UpdatedAt)
	if err != nil && (isUnknownColumn(err, "image_urls_json") || isUnknownColumn(err, "updated_at")) {
		missImage := isUnknownColumn(err, "image_urls_json")
		missUpdated := isUnknownColumn(err, "updated_at")

		switch {
		case missImage && missUpdated:
			row2 := s.db.QueryRowContext(ctx, `
				SELECT id, name, cover_url, points_price, stock, status, created_at
				FROM goods
				WHERE id = ?
				LIMIT 1
			`, id)
			err = row2.Scan(&g.ID, &g.Name, &cover, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt)
			g.UpdatedAt = g.CreatedAt
		case missImage:
			row2 := s.db.QueryRowContext(ctx, `
				SELECT id, name, cover_url, points_price, stock, status, created_at, updated_at
				FROM goods
				WHERE id = ?
				LIMIT 1
			`, id)
			err = row2.Scan(&g.ID, &g.Name, &cover, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt, &g.UpdatedAt)
		case missUpdated:
			row2 := s.db.QueryRowContext(ctx, `
				SELECT id, name, cover_url, image_urls_json, points_price, stock, status, created_at
				FROM goods
				WHERE id = ?
				LIMIT 1
			`, id)
			err = row2.Scan(&g.ID, &g.Name, &cover, &imageURLs, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt)
			g.UpdatedAt = g.CreatedAt
		}
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return Goods{}, fmt.Errorf("goods not found")
		}
		return Goods{}, err
	}
	g.CoverURL = cover.String
	if imageURLs.Valid && strings.TrimSpace(imageURLs.String) != "" {
		var list []string
		if err := json.Unmarshal([]byte(imageURLs.String), &list); err == nil {
			g.ImageURLs = list
		}
	}
	if len(g.ImageURLs) == 0 && g.CoverURL != "" {
		g.ImageURLs = []string{g.CoverURL}
	}
	return g, nil
}

// ListGoods 获取商品列表。
func (s *service) ListGoods(ctx context.Context, req ListGoodsRequest) ([]Goods, error) {
	// 1) 基础校验与分页兜底：limit 默认 20，最大 200。
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

	// 2) 组装筛选条件：status=0 代表不过滤，但默认排除已删除（status=0）。
	statusClause := ""
	args := make([]any, 0, 3)
	if req.Status != 0 {
		statusClause = "WHERE status = ?"
		args = append(args, req.Status)
	} else {
		statusClause = "WHERE status <> 0"
	}
	args = append(args, req.Limit, req.Offset)

	// 3) 查询列表：按 id 倒序，便于后台优先看到最新创建的商品。
	withImageURLsJSON := true
	withUpdatedAt := true
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, IFNULL(cover_url, ''), IFNULL(image_urls_json, ''), points_price, stock, status, created_at, updated_at
		FROM goods
		`+statusClause+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil && (isUnknownColumn(err, "image_urls_json") || isUnknownColumn(err, "updated_at")) {
		if isUnknownColumn(err, "image_urls_json") {
			withImageURLsJSON = false
		}
		if isUnknownColumn(err, "updated_at") {
			withUpdatedAt = false
		}
		selectCols := "SELECT id, name, IFNULL(cover_url, ''), "
		if withImageURLsJSON {
			selectCols += "IFNULL(image_urls_json, ''), "
		}
		selectCols += "points_price, stock, status, created_at"
		if withUpdatedAt {
			selectCols += ", updated_at"
		}
		rows, err = s.db.QueryContext(ctx, `
			`+selectCols+`
			FROM goods
			`+statusClause+`
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 4) 迭代结果集并返回。
	out := make([]Goods, 0, req.Limit)
	for rows.Next() {
		var g Goods
		var imageURLsJSON string
		if withImageURLsJSON {
			if withUpdatedAt {
				if err := rows.Scan(&g.ID, &g.Name, &g.CoverURL, &imageURLsJSON, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt, &g.UpdatedAt); err != nil {
					return nil, err
				}
			} else {
				if err := rows.Scan(&g.ID, &g.Name, &g.CoverURL, &imageURLsJSON, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt); err != nil {
					return nil, err
				}
			}
		} else {
			if withUpdatedAt {
				if err := rows.Scan(&g.ID, &g.Name, &g.CoverURL, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt, &g.UpdatedAt); err != nil {
					return nil, err
				}
			} else {
				if err := rows.Scan(&g.ID, &g.Name, &g.CoverURL, &g.PointsPrice, &g.Stock, &g.Status, &g.CreatedAt); err != nil {
					return nil, err
				}
			}
		}
		if strings.TrimSpace(imageURLsJSON) != "" {
			var list []string
			if err := json.Unmarshal([]byte(imageURLsJSON), &list); err == nil {
				g.ImageURLs = list
			}
		}
		if len(g.ImageURLs) == 0 && g.CoverURL != "" {
			g.ImageURLs = []string{g.CoverURL}
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func isUnknownColumn(err error, column string) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "Error 1054") && strings.Contains(s, "Unknown column") && strings.Contains(s, column)
}
