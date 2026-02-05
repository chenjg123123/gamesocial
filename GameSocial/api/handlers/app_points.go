package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

// PointsLedgerItem 表示积分流水列表项。
type PointsLedgerItem struct {
	ID           uint64    `json:"id"`
	ChangeAmount int64     `json:"changeAmount"`
	BalanceAfter int64     `json:"balanceAfter"`
	BizType      string    `json:"bizType"`
	BizID        string    `json:"bizId"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
}

// AppPointsBalance 查询用户积分余额。
// GET /api/points/balance
func AppPointsBalance(db *sql.DB) http.HandlerFunc {
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

		var balance int64
		err := db.QueryRowContext(r.Context(), `
			SELECT balance
			FROM points_account
			WHERE user_id = ?
			LIMIT 1
		`, uid).Scan(&balance)
		if err != nil && err != sql.ErrNoRows {
			SendJBizFail(w, err.Error())
			return
		}

		SendJSuccess(w, map[string]any{
			"balance": balance,
		})
	}
}

// AppPointsLedgers 查询用户积分流水。
// GET /api/points/ledgers
func AppPointsLedgers(db *sql.DB) http.HandlerFunc {
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

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit <= 0 {
			limit = 20
		}
		if limit > 200 {
			limit = 200
		}
		if offset < 0 {
			offset = 0
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT id, change_amount, balance_after, biz_type, biz_id, IFNULL(remark, ''), created_at
			FROM points_ledger
			WHERE user_id = ?
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`, uid, limit, offset)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		defer rows.Close()

		out := make([]PointsLedgerItem, 0, limit)
		for rows.Next() {
			var it PointsLedgerItem
			if err := rows.Scan(&it.ID, &it.ChangeAmount, &it.BalanceAfter, &it.BizType, &it.BizID, &it.Remark, &it.CreatedAt); err != nil {
				SendJBizFail(w, err.Error())
				return
			}
			out = append(out, it)
		}
		if err := rows.Err(); err != nil {
			SendJBizFail(w, err.Error())
			return
		}

		SendJSuccess(w, out)
	}
}
