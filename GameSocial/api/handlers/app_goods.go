package handlers

import (
	"net/http"
	"strconv"

	"gamesocial/modules/item"
)

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
		if limit <= 0 {
			limit = 20
		}
		if limit > 200 {
			limit = 200
		}
		if offset < 0 {
			offset = 0
		}

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
