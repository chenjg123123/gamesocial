package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gamesocial/modules/user"
)

// AppUserMeGet 获取当前用户信息。
// GET /api/users/me
func AppUserMeGet(svc user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}
		out, err := svc.Get(r.Context(), uid)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, struct {
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatarUrl"`
			Level     int    `json:"level"`
			Exp       int64  `json:"exp"`
			CreatedAt string `json:"createdAt"`
		}{
			Nickname:  out.Nickname,
			AvatarURL: out.AvatarURL,
			Level:     out.Level,
			Exp:       out.Exp,
			CreatedAt: out.CreatedAt.Format(time.RFC3339),
		})
	}
}

// AppUserMeUpdate 更新当前用户资料。
// PUT /api/users/me
func AppUserMeUpdate(svc user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		var req struct {
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatarUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		out, err := svc.Update(r.Context(), uid, user.UpdateUserRequest{
			Nickname:  req.Nickname,
			AvatarURL: req.AvatarURL,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, struct {
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatarUrl"`
		}{
			Nickname:  out.Nickname,
			AvatarURL: out.AvatarURL,
		})
	}
}
