package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gamesocial/modules/item"
	"gamesocial/modules/redeem"
	"gamesocial/modules/task"
	"gamesocial/modules/tournament"
	"gamesocial/modules/user"
)

func parseUint64(s string) uint64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func userIDFromRequest(r *http.Request) uint64 {
	if r == nil {
		return 0
	}
	q := r.URL.Query()
	if id := parseUint64(q.Get("userId")); id != 0 {
		return id
	}
	if id := parseUint64(r.Header.Get("X-User-Id")); id != 0 {
		return id
	}
	return 0
}

type PointsLedgerItem struct {
	ID           uint64    `json:"id"`
	ChangeAmount int64     `json:"changeAmount"`
	BalanceAfter int64     `json:"balanceAfter"`
	BizType      string    `json:"bizType"`
	BizID        string    `json:"bizId"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AdminAuditLogItem struct {
	ID         uint64          `json:"id"`
	AdminID    uint64          `json:"adminId"`
	Action     string          `json:"action"`
	BizType    string          `json:"bizType"`
	BizID      string          `json:"bizId"`
	DetailJSON json.RawMessage `json:"detailJson,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

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
			SendJBizFail(w, "userId 不能为空")
			return
		}
		out, err := svc.Get(r.Context(), uid)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

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
			SendJBizFail(w, "userId 不能为空")
			return
		}

		var req user.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		out, err := svc.Update(r.Context(), uid, req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

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
			SendJBizFail(w, "userId 不能为空")
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
			SendJBizFail(w, "userId 不能为空")
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
			SendJBizFail(w, "userId 不能为空")
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

func AppTournamentsList(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status := q.Get("status")

		out, err := svc.List(r.Context(), tournament.ListTournamentRequest{
			Offset: offset,
			Limit:  limit,
			Status: status,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppTournamentsGet(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		id := parseUint64(r.PathValue("id"))
		if id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}
		out, err := svc.Get(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppTournamentsJoin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"joined": true})
	}
}

func AppTournamentsCancel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"canceled": true})
	}
}

func AppTournamentsResults() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, []any{})
	}
}

func AppTasksList(svc task.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}

		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))

		out, err := svc.ListTaskDef(r.Context(), task.ListTaskDefRequest{
			Offset: offset,
			Limit:  limit,
			Status: 1,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppTasksCheckin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"checkedIn": true})
	}
}

func AppTasksClaim() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = r.PathValue("taskCode")
		SendJSuccess(w, map[string]any{"claimed": true})
	}
}

func AppGoodsList(svc item.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))

		out, err := svc.ListGoods(r.Context(), item.ListGoodsRequest{
			Offset: offset,
			Limit:  limit,
			Status: 0,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppGoodsGet(svc item.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		id := parseUint64(r.PathValue("id"))
		if id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}
		out, err := svc.GetGoods(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppRedeemOrderCreate(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		var req redeem.CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		out, err := svc.CreateOrder(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppRedeemOrderList(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status := q.Get("status")
		userID := parseUint64(q.Get("userId"))

		out, err := svc.ListOrders(r.Context(), redeem.ListOrderRequest{
			Offset: offset,
			Limit:  limit,
			Status: status,
			UserID: userID,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppRedeemOrderGet(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		id := parseUint64(r.PathValue("id"))
		if id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}
		out, err := svc.GetOrder(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppRedeemOrderCancel(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		id := parseUint64(r.PathValue("id"))
		if id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}
		out, err := svc.CancelOrder(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

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

func AdminAuthLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"logout": true})
	}
}

func AdminAuditLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}

		if db == nil {
			SendJBizFail(w, "database disabled")
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

		adminID := parseUint64(q.Get("adminId"))
		where := ""
		args := make([]any, 0, 4)
		if adminID != 0 {
			where = "WHERE admin_id = ?"
			args = append(args, adminID)
		}
		args = append(args, limit, offset)

		rows, err := db.QueryContext(r.Context(), `
			SELECT id, admin_id, action, IFNULL(biz_type, ''), IFNULL(biz_id, ''), IFNULL(detail_json, JSON_OBJECT()), created_at
			FROM admin_audit_log
			`+where+`
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`, args...)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		defer rows.Close()

		out := make([]AdminAuditLogItem, 0, limit)
		for rows.Next() {
			var it AdminAuditLogItem
			var detailBytes []byte
			if err := rows.Scan(&it.ID, &it.AdminID, &it.Action, &it.BizType, &it.BizID, &detailBytes, &it.CreatedAt); err != nil {
				SendJBizFail(w, err.Error())
				return
			}
			if len(detailBytes) != 0 {
				it.DetailJSON = json.RawMessage(detailBytes)
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

func AdminPointsAdjust() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"adjusted": true})
	}
}

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

func AdminMediaUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{
			"url":       "",
			"createdAt": time.Now().Format(time.RFC3339),
		})
	}
}
