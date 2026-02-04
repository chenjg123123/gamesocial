// handlers 实现面向小程序端与管理端的聚合接口（部分为占位实现）。
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
	if id := parseUint64(r.Header.Get("X-User-Id")); id != 0 {
		return id
	}
	q := r.URL.Query()
	if id := parseUint64(q.Get("userId")); id != 0 {
		return id
	}
	return 0
}

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

// AdminAuditLogItem 表示管理员审计日志列表项。
type AdminAuditLogItem struct {
	ID         uint64          `json:"id"`
	AdminID    uint64          `json:"adminId"`
	Action     string          `json:"action"`
	BizType    string          `json:"bizType"`
	BizID      string          `json:"bizId"`
	DetailJSON json.RawMessage `json:"detailJson,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

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
		SendJSuccess(w, out)
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

// AppTournamentsList 获取赛事列表。
// GET /api/tournaments
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

// AppTournamentsGet 获取赛事详情。
// GET /api/tournaments/{id}
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

// AppTournamentsJoin 赛事报名占位接口。
// POST /api/tournaments/{id}/join
func AppTournamentsJoin(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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

		id := parseUint64(r.PathValue("id"))
		if id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		if err := svc.Join(r.Context(), id, uid); err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, map[string]any{"joined": true})
	}
}

// AppTournamentsCancel 赛事取消报名占位接口。
// PUT /api/tournaments/{id}/cancel
func AppTournamentsCancel(svc tournament.Service) http.HandlerFunc {
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

		id := parseUint64(r.PathValue("id"))
		if id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		if err := svc.Cancel(r.Context(), id, uid); err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, map[string]any{"canceled": true})
	}
}

// AppTournamentsResults 赛事成绩占位接口。
// GET /api/tournaments/{id}/results
func AppTournamentsResults(svc tournament.Service) http.HandlerFunc {
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

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))

		uid := userIDFromRequest(r)
		out, err := svc.GetResults(r.Context(), id, uid, offset, limit)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AppTasksList 获取任务定义列表（当前仅返回有效任务定义）。
// GET /api/tasks
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

// AppTasksCheckin 任务打卡占位接口。
// POST /api/tasks/checkin
func AppTasksCheckin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"checkedIn": true})
	}
}

// AppTasksClaim 任务领奖占位接口。
// POST /api/tasks/{taskCode}/claim
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

// AppGoodsList 获取商品列表。
// GET /api/goods
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

// AppGoodsGet 获取商品详情。
// GET /api/goods/{id}
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

// AppRedeemOrderCreate 创建兑换订单。
// POST /api/redeem/orders
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

		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		var req redeem.CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}
		req.UserID = uid

		out, err := svc.CreateOrder(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AppRedeemOrderList 查询兑换订单列表。
// GET /api/redeem/orders
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

		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status := q.Get("status")

		out, err := svc.ListOrders(r.Context(), redeem.ListOrderRequest{
			Offset: offset,
			Limit:  limit,
			Status: status,
			UserID: uid,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AppRedeemOrderGet 查询兑换订单详情。
// GET /api/redeem/orders/{id}
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

		if userIDFromRequest(r) == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
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

// AppRedeemOrderCancel 取消兑换订单（CREATED -> CANCELED）。
// PUT /api/redeem/orders/{id}/cancel
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

		if userIDFromRequest(r) == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
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

// AdminAuditLogs 查询管理员审计日志。
// GET /admin/audit/logs
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

// AdminMediaUpload 上传媒体占位接口。
// POST /admin/media/upload
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
