package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"gamesocial/internal/media"
)

// AdminAuthLogin 管理员登录占位接口。
// POST /admin/auth/login
func AdminAuthLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{
			"token": "admin_token_placeholder",
		})
	}
}

// AdminAuthMe 管理员信息占位接口。
// GET /admin/auth/me
func AdminAuthMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{
			"id":       1,
			"username": "admin",
		})
	}
}

// AdminAuthLogout 管理员登出占位接口。
// POST /admin/auth/logout
func AdminAuthLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"logout": true})
	}
}

// AdminPointsAdjust 积分调整占位接口。
// POST /admin/points/adjust
func AdminPointsAdjust() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"adjusted": true})
	}
}

// AdminUsersDrinksUse 消费饮品占位接口。
// PUT /admin/users/{id}/drinks/use
func AdminUsersDrinksUse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"used": true})
	}
}

// AdminTournamentResultsPublish 发布赛事成绩占位接口。
// POST /admin/tournaments/{id}/results/publish
func AdminTournamentResultsPublish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"published": true})
	}
}

// AdminTournamentAwardsGrant 发放赛事奖励占位接口。
// POST /admin/tournaments/{id}/awards/grant
func AdminTournamentAwardsGrant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"granted": true})
	}
}

func uploadImageToStore(r *http.Request, store media.Store, maxUploadBytes int64) (media.UploadResult, error) {
	if store == nil {
		return media.UploadResult{}, errors.New("media store not configured: set MEDIA_COS_BUCKET_URL (COS bucket domain or CloudBase gateway domain) and restart server; if COS, also set MEDIA_COS_SECRET_ID/MEDIA_COS_SECRET_KEY")
	}
	if maxUploadBytes <= 0 {
		maxUploadBytes = 10 * 1024 * 1024
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return media.UploadResult{}, errors.New("missing form file: file")
	}
	defer file.Close()

	if header == nil || header.Filename == "" {
		return media.UploadResult{}, errors.New("invalid filename")
	}

	if fileSize := header.Size; fileSize > 0 && fileSize > maxUploadBytes {
		return media.UploadResult{}, errors.New("file too large")
	}

	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	if !strings.HasPrefix(ct, "image/") {
		return media.UploadResult{}, errors.New("only image/* is allowed")
	}

	if seeker, ok := file.(interface {
		Seek(offset int64, whence int) (int64, error)
	}); ok {
		end, err := seeker.Seek(0, io.SeekEnd)
		if err == nil {
			if end > maxUploadBytes {
				return media.UploadResult{}, errors.New("file too large")
			}
			_, _ = seeker.Seek(0, io.SeekStart)
		}
	}

	return store.Upload(r.Context(), file, ct, header.Filename)
}
