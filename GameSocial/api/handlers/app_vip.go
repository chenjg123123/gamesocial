package handlers

import (
	"database/sql"
	"net/http"
	"time"
)

// AppVipStatus 查询用户会员订阅状态（基于 vip_subscription 表）。
// GET /api/vip/status
func AppVipStatus(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}

		if db == nil {
			SendJBizFail(w, "database disabled")
			return
		}

		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		var endAt time.Time
		var status string
		err := db.QueryRowContext(r.Context(), `
			SELECT end_at, status
			FROM vip_subscription
			WHERE user_id = ?
			ORDER BY end_at DESC
			LIMIT 1
		`, uid).Scan(&endAt, &status)
		if err != nil && err != sql.ErrNoRows {
			SendJBizFail(w, err.Error())
			return
		}

		isVip := status == "ACTIVE" && endAt.After(time.Now())
		expireAt := ""
		if !endAt.IsZero() {
			expireAt = endAt.Format(time.RFC3339)
		}

		SendJSuccess(w, map[string]any{
			"isVip":    isVip,
			"expireAt": expireAt,
		})
	}
}
