// tournament 模块负责赛事发布、维护与查询等业务能力。
package tournament

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Tournament 对应数据库 tournament 表的数据结构。
type Tournament struct {
	ID             uint64    `json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content,omitempty"`
	CoverURL       string    `json:"coverUrl,omitempty"`
	StartAt        time.Time `json:"startAt"`
	EndAt          time.Time `json:"endAt"`
	Status         string    `json:"status"`
	CreatedByAdmin uint64    `json:"createdByAdminId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateTournamentRequest 创建赛事入参。
type CreateTournamentRequest struct {
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	CoverURL       string    `json:"coverUrl"`
	StartAt        time.Time `json:"startAt"`
	EndAt          time.Time `json:"endAt"`
	Status         string    `json:"status"`
	CreatedByAdmin uint64    `json:"createdByAdminId"`
}

// UpdateTournamentRequest 更新赛事入参。
type UpdateTournamentRequest struct {
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	CoverURL string    `json:"coverUrl"`
	StartAt  time.Time `json:"startAt"`
	EndAt    time.Time `json:"endAt"`
	Status   string    `json:"status"`
}

// ListTournamentRequest 列表查询入参。
type ListTournamentRequest struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Status string `json:"status"`
}

// Service 定义 tournament 模块对外提供的业务接口（赛事 CRUD）。
type Service interface {
	Create(ctx context.Context, req CreateTournamentRequest) (Tournament, error)
	Update(ctx context.Context, id uint64, req UpdateTournamentRequest) (Tournament, error)
	Delete(ctx context.Context, id uint64) error
	Get(ctx context.Context, id uint64) (Tournament, error)
	List(ctx context.Context, req ListTournamentRequest) ([]Tournament, error)
}

type service struct {
	db *sql.DB
}

// NewService 创建 tournament 模块服务。
func NewService(db *sql.DB) Service {
	return &service{db: db}
}

// Create 创建赛事并返回创建后的赛事详情。
func (s *service) Create(ctx context.Context, req CreateTournamentRequest) (Tournament, error) {
	// 1) 基础校验。
	if s.db == nil {
		return Tournament{}, errors.New("database disabled")
	}
	if req.Title == "" {
		return Tournament{}, errors.New("title is empty")
	}
	if req.StartAt.IsZero() || req.EndAt.IsZero() {
		return Tournament{}, errors.New("start_at/end_at is empty")
	}
	if req.EndAt.Before(req.StartAt) {
		return Tournament{}, errors.New("end_at must be >= start_at")
	}
	if req.Status == "" {
		req.Status = "DRAFT"
	}
	if req.CreatedByAdmin == 0 {
		req.CreatedByAdmin = 1
	}

	// 2) 写入 tournament 表，并返回创建后的详情。
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO tournament (title, content, cover_url, start_at, end_at, status, created_by_admin_id, created_at, updated_at)
		VALUES (?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, NOW(), NOW())
	`, req.Title, req.Content, req.CoverURL, req.StartAt, req.EndAt, req.Status, req.CreatedByAdmin)
	if err != nil {
		return Tournament{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Tournament{}, err
	}
	return s.Get(ctx, uint64(id))
}

// Update 更新赛事并返回更新后的赛事详情。
func (s *service) Update(ctx context.Context, id uint64, req UpdateTournamentRequest) (Tournament, error) {
	// 1) 基础校验。
	if s.db == nil {
		return Tournament{}, errors.New("database disabled")
	}
	if id == 0 {
		return Tournament{}, errors.New("invalid id")
	}
	if req.Title == "" {
		return Tournament{}, errors.New("title is empty")
	}
	if req.StartAt.IsZero() || req.EndAt.IsZero() {
		return Tournament{}, errors.New("start_at/end_at is empty")
	}
	if req.EndAt.Before(req.StartAt) {
		return Tournament{}, errors.New("end_at must be >= start_at")
	}
	if req.Status == "" {
		req.Status = "DRAFT"
	}

	// 2) 更新可变字段，并刷新 updated_at。
	result, err := s.db.ExecContext(ctx, `
		UPDATE tournament
		SET title = ?, content = NULLIF(?, ''), cover_url = NULLIF(?, ''), start_at = ?, end_at = ?, status = ?, updated_at = NOW()
		WHERE id = ?
	`, req.Title, req.Content, req.CoverURL, req.StartAt, req.EndAt, req.Status, id)
	if err != nil {
		return Tournament{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return Tournament{}, fmt.Errorf("tournament not found")
	}
	return s.Get(ctx, id)
}

// Delete 软删除赛事（status=CANCELED）。
func (s *service) Delete(ctx context.Context, id uint64) error {
	// 1) 基础校验。
	if s.db == nil {
		return errors.New("database disabled")
	}
	if id == 0 {
		return errors.New("invalid id")
	}

	// 2) 软删除：把 status 标记为 CANCELED，保留历史报名/成绩/发奖的引用。
	result, err := s.db.ExecContext(ctx, `
		UPDATE tournament
		SET status = 'CANCELED', updated_at = NOW()
		WHERE id = ? AND status <> 'CANCELED'
	`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("tournament not found")
	}
	return nil
}

// Get 获取赛事详情。
func (s *service) Get(ctx context.Context, id uint64) (Tournament, error) {
	// 1) 基础校验。
	if s.db == nil {
		return Tournament{}, errors.New("database disabled")
	}
	if id == 0 {
		return Tournament{}, errors.New("invalid id")
	}

	// 2) 查询单条记录：content/cover_url 可空。
	var t Tournament
	var content, cover sql.NullString
	row := s.db.QueryRowContext(ctx, `
		SELECT id, title, content, cover_url, start_at, end_at, status, created_by_admin_id, created_at, updated_at
		FROM tournament
		WHERE id = ?
		LIMIT 1
	`, id)
	if err := row.Scan(&t.ID, &t.Title, &content, &cover, &t.StartAt, &t.EndAt, &t.Status, &t.CreatedByAdmin, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return Tournament{}, fmt.Errorf("tournament not found")
		}
		return Tournament{}, err
	}
	t.Content = content.String
	t.CoverURL = cover.String
	return t, nil
}

// List 获取赛事列表。
func (s *service) List(ctx context.Context, req ListTournamentRequest) ([]Tournament, error) {
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

	// 2) 组装筛选条件：未指定 status 则默认排除 CANCELED。
	where := ""
	args := make([]any, 0, 3)
	if req.Status != "" {
		where = "WHERE status = ?"
		args = append(args, req.Status)
	} else {
		where = "WHERE status <> 'CANCELED'"
	}
	args = append(args, req.Limit, req.Offset)

	// 3) 查询列表：按 start_at 倒序，便于后台优先看到最近赛事。
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, IFNULL(content, ''), IFNULL(cover_url, ''), start_at, end_at, status, created_by_admin_id, created_at, updated_at
		FROM tournament
		`+where+`
		ORDER BY start_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Tournament, 0, req.Limit)
	for rows.Next() {
		var t Tournament
		if err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CoverURL, &t.StartAt, &t.EndAt, &t.Status, &t.CreatedByAdmin, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
