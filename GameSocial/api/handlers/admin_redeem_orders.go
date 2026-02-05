// 管理员侧兑换订单管理接口（基础增删改查 + 核销）。
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gamesocial/modules/redeem"
)

// AdminRedeemOrderCreate 创建兑换订单（最小 CRUD：不接入积分扣减流水）。
// POST /admin/redeem/orders
func AdminRedeemOrderCreate(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析请求体：需要 user_id 与 items。
		var req redeem.CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 4) 调用业务层：写入 redeem_order 与 redeem_order_item。
		out, err := svc.CreateOrder(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminRedeemOrderGet 获取兑换订单详情（包含 items）。
// GET /admin/redeem/orders/{id}
func AdminRedeemOrderGet(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析 id。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		// 4) 读取订单详情（含 items）。
		out, err := svc.GetOrder(r.Context(), id, 0)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminRedeemOrderList 兑换订单列表（不包含 items）。
// GET /admin/redeem/orders?offset=0&limit=20&status=CREATED&userId=1001
func AdminRedeemOrderList(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析 query：分页 + 状态筛选 + user_id 筛选。
		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status := q.Get("status")
		userIDRaw := q.Get("userId")
		if userIDRaw == "" {
			userIDRaw = q.Get("user_id")
		}
		userID, _ := strconv.ParseUint(userIDRaw, 10, 64)

		// 4) 查询并返回。
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

// AdminRedeemOrderUse 核销兑换订单（CREATED -> USED），避免重复核销。
// PUT /admin/redeem/orders/{id}/use
func AdminRedeemOrderUse(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析 id。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		// 4) 解析请求体（可选：admin_id；未提供则默认 1）。
		var body struct {
			AdminID       uint64 `json:"adminId"`
			AdminIDLegacy uint64 `json:"admin_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		// 5) 条件更新：只有 CREATED 能变更为 USED。
		adminID := body.AdminID
		if adminID == 0 {
			adminID = body.AdminIDLegacy
		}
		out, err := svc.UseOrder(r.Context(), id, adminID)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminRedeemOrderCancel 取消兑换订单（CREATED -> CANCELED）。
// PUT /admin/redeem/orders/{id}/cancel
func AdminRedeemOrderCancel(svc redeem.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析 id。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		// 4) 条件更新：只有 CREATED 能被取消。
		out, err := svc.CancelOrder(r.Context(), id, 0)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}
