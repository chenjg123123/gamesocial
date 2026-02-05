// user 模块负责用户资料管理等业务能力。
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

// User 对应数据库 user 表的数据结构。
type User struct {
	ID        uint64    `json:"id,omitempty"`
	OpenID    string    `json:"openId,omitempty"`
	UnionID   string    `json:"unionId,omitempty"`
	Nickname  string    `json:"nickname,omitempty"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	Status    int       `json:"status,omitempty"`
	Level     int       `json:"level,omitempty"`
	Exp       int64     `json:"exp,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// UpdateUserRequest 更新用户资料入参（管理员侧可用）。
type UpdateUserRequest struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	Status    *int   `json:"status"`
}

// ListUserRequest 用户列表入参。
type ListUserRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Status int `json:"status"`
}

// Service 定义 user 模块对外提供的业务接口（用户查询/封禁/更新）。
type Service interface {
	Get(ctx context.Context, id uint64) (User, error)
	List(ctx context.Context, req ListUserRequest) ([]User, error)
	Update(ctx context.Context, id uint64, req UpdateUserRequest) (User, error)
}

type service struct {
	db *sql.DB
}

// NewService 创建 user 模块服务。
func NewService(db *sql.DB) Service {
	return &service{db: db}
}

// Get 获取用户详情。
func (s *service) Get(ctx context.Context, id uint64) (User, error) {
	// 1) 基础校验。
	if s.db == nil {
		return User{}, errors.New("database disabled")
	}
	if id == 0 {
		return User{}, errors.New("invalid id")
	}

	// 2) 读取单条记录：unionid/nickname/avatar_url 允许为空。
	var u User
	row := s.db.QueryRowContext(ctx, `
		SELECT IFNULL(nickname, ''), IFNULL(avatar_url, ''), level, exp, created_at
		FROM user
		WHERE id = ?
		LIMIT 1
	`, id)
	if err := row.Scan(&u.Nickname, &u.AvatarURL, &u.Level, &u.Exp, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, err
	}
	log.Printf("user.Get: %+v", u)
	return u, nil
}

// List 获取用户列表。
func (s *service) List(ctx context.Context, req ListUserRequest) ([]User, error) {
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

	// 2) 组装筛选条件：默认不过滤状态；如果传了 status 则按 status 过滤。
	where := ""
	args := make([]any, 0, 3)
	if req.Status != 0 {
		where = "WHERE status = ?"
		args = append(args, req.Status)
	}
	args = append(args, req.Limit, req.Offset)

	// 3) 查询列表：按 id 倒序，便于后台先看到最近用户。
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, openid, IFNULL(unionid, ''), IFNULL(nickname, ''), IFNULL(avatar_url, ''), status, level, exp, created_at, updated_at
		FROM user
		`+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]User, 0, req.Limit)
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.OpenID, &u.UnionID, &u.Nickname, &u.AvatarURL, &u.Status, &u.Level, &u.Exp, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Update 更新用户资料/状态并返回更新后的用户详情。
func (s *service) Update(ctx context.Context, id uint64, req UpdateUserRequest) (User, error) {
	// 1) 基础校验。
	if s.db == nil {
		return User{}, errors.New("database disabled")
	}
	if id == 0 {
		return User{}, errors.New("invalid id")
	}
	// 2) 更新用户资料：nickname/avatar_url 可为空；status 用于封禁/解封（例如 0=封禁，1=正常）。
	result, err := s.db.ExecContext(ctx, `
		UPDATE user
		SET nickname = NULLIF(?, ''), avatar_url = NULLIF(?, ''), status = IFNULL(?, status), updated_at = NOW()
		WHERE id = ?
	`, req.Nickname, req.AvatarURL, req.Status, id)
	if err != nil {
		return User{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return User{}, fmt.Errorf("user not found")
	}
	return s.Get(ctx, id)
}
