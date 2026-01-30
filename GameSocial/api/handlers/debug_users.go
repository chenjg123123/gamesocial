package handlers

import (
	"database/sql"
	"net/http"
)

type DebugUserRow struct {
	ID        uint64 `json:"id"`
	OpenID    string `json:"openid"`
	UnionID   string `json:"unionid,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func DebugListUsers(db *sql.DB) http.HandlerFunc {
	// DebugListUsers 返回用于早期开发快速校验的用户列表（数量有限）。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT id, openid, IFNULL(unionid, ''), IFNULL(nickname, ''), IFNULL(avatar_url, ''), status, created_at, updated_at
			FROM user
			ORDER BY id
			LIMIT 200
		`)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		defer rows.Close()

		users := make([]DebugUserRow, 0, 16)
		for rows.Next() {
			var u DebugUserRow
			if err := rows.Scan(&u.ID, &u.OpenID, &u.UnionID, &u.Nickname, &u.AvatarURL, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
				SendInternalError(w, err.Error())
				return
			}
			users = append(users, u)
		}
		if err := rows.Err(); err != nil {
			SendInternalError(w, err.Error())
			return
		}

		SendSuccess(w, users)
	}
}
