// auth 模块负责小程序登录、token 签发与鉴权相关的业务能力。
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gamesocial/internal/wechat"
)

type User struct {
	ID        uint64 `json:"id"`
	OpenID    string `json:"openId"`
	UnionID   string `json:"unionId,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Status    int    `json:"status"`
}

type LoginResult struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Service interface {
	WechatLogin(ctx context.Context, code string) (LoginResult, error)
}

type service struct {
	db           *sql.DB
	wechatClient *wechat.Client
	tokenSecret  []byte
	tokenTTL     time.Duration
}

func NewService(db *sql.DB, wechatClient *wechat.Client, tokenSecret string, tokenTTLSeconds int64) Service {
	return &service{
		db:           db,
		wechatClient: wechatClient,
		tokenSecret:  []byte(tokenSecret),
		tokenTTL:     time.Duration(tokenTTLSeconds) * time.Second,
	}
}

func (s *service) WechatLogin(ctx context.Context, code string) (LoginResult, error) {
	if s.wechatClient == nil {
		return LoginResult{}, errors.New("wechat client is nil")
	}
	if s.db == nil {
		return LoginResult{}, errors.New("database disabled")
	}
	if len(s.tokenSecret) == 0 {
		return LoginResult{}, errors.New("AUTH_TOKEN_SECRET is empty")
	}
	if code == "" {
		return LoginResult{}, errors.New("code is empty")
	}

	session, err := s.wechatClient.Code2Session(ctx, code)
	if err != nil {
		return LoginResult{}, err
	}

	u, err := s.ensureUser(ctx, session.OpenID, session.UnionID)
	if err != nil {
		return LoginResult{}, err
	}
	if u.Status == 0 {
		return LoginResult{}, fmt.Errorf("user is banned")
	}

	token, err := MakeTokenV1(u.ID, time.Now().Add(s.tokenTTL), s.tokenSecret)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Token: token,
		User:  u,
	}, nil
}

func (s *service) ensureUser(ctx context.Context, openID, unionID string) (User, error) {
	var u User

	row := s.db.QueryRowContext(ctx, `
		SELECT id, openid, IFNULL(unionid, ''), IFNULL(nickname, ''), IFNULL(avatar_url, ''), status
		FROM user
		WHERE openid = ?
		LIMIT 1
	`, openID)

	switch err := row.Scan(&u.ID, &u.OpenID, &u.UnionID, &u.Nickname, &u.AvatarURL, &u.Status); err {
	case nil:
		if unionID != "" && u.UnionID != unionID {
			if _, err := s.db.ExecContext(ctx, `
				UPDATE user
				SET unionid = NULLIF(?, ''), updated_at = NOW()
				WHERE id = ?
			`, unionID, u.ID); err != nil {
				return User{}, err
			}
			u.UnionID = unionID
		}
		return u, nil
	case sql.ErrNoRows:
		res, err := s.db.ExecContext(ctx, `
			INSERT INTO user (openid, unionid, nickname, avatar_url, status, created_at, updated_at)
			VALUES (?, NULLIF(?, ''), '', '', 1, NOW(), NOW())
		`, openID, unionID)
		if err != nil {
			return User{}, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return User{}, err
		}
		u = User{
			ID:      uint64(id),
			OpenID:  openID,
			UnionID: unionID,
			Status:  1,
		}
		return u, nil
	default:
		return User{}, err
	}
}
