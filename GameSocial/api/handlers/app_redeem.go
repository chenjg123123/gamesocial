package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gamesocial/modules/redeem"
)

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
		if limit <= 0 {
			limit = 20
		}
		if limit > 200 {
			limit = 200
		}
		if offset < 0 {
			offset = 0
		}

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
		out, err := svc.GetOrder(r.Context(), id, uid)
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
		out, err := svc.CancelOrder(r.Context(), id, uid)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}
