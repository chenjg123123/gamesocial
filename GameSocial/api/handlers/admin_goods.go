// 管理员侧商品管理接口（基础增删改查）。
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gamesocial/modules/item"
)

// AdminGoodsCreate 创建商品。
// POST /admin/goods
func AdminGoodsCreate(svc item.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验：创建商品必须是 POST。
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验：服务未注入代表启动阶段组装失败。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析请求体：读取 JSON 到 CreateGoodsRequest。
		var req item.CreateGoodsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 4) 调用业务层：落库并返回创建后的商品详情。
		g, err := svc.CreateGoods(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		// 5) 返回响应。
		SendJSuccess(w, g)
	}
}

// AdminGoodsUpdate 更新商品。
// PUT /admin/goods/{id}
func AdminGoodsUpdate(svc item.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验：更新商品必须是 PUT。
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 路径参数解析：从 /admin/goods/{id} 取出 id 并转为 uint64。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		// 4) 解析请求体。
		var req item.UpdateGoodsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 5) 调用业务层：更新成功后返回最新详情。
		g, err := svc.UpdateGoods(r.Context(), id, req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		// 6) 返回响应。
		SendJSuccess(w, g)
	}
}

// AdminGoodsDelete 删除商品（软删除：status=0）。
// DELETE /admin/goods/{id}
func AdminGoodsDelete(svc item.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验：删除商品必须是 DELETE。
		if r.Method != http.MethodDelete {
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

		// 4) 调用业务层：这里采用软删除（status=0），便于保留历史订单引用。
		if err := svc.DeleteGoods(r.Context(), id); err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		// 5) 返回响应。
		SendJSuccess(w, map[string]any{"deleted": true})
	}
}

// AdminGoodsGet 获取单个商品。
// GET /admin/goods/{id}
func AdminGoodsGet(svc item.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验：查询详情必须是 GET。
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

		// 4) 调用业务层：从 goods 表读取一条记录。
		g, err := svc.GetGoods(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		// 5) 返回响应。
		SendJSuccess(w, g)
	}
}

// AdminGoodsList 商品列表（默认返回 status!=0 的数据）。
// GET /admin/goods?offset=0&limit=20&status=1
func AdminGoodsList(svc item.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验：列表查询必须是 GET。
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 读取 query：分页 offset/limit，状态筛选 status（0 表示不过滤，仅排除已删除）。
		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status, _ := strconv.Atoi(q.Get("status"))

		// 4) 调用业务层：返回商品列表。
		list, err := svc.ListGoods(r.Context(), item.ListGoodsRequest{
			Offset: offset,
			Limit:  limit,
			Status: status,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		// 5) 返回响应。
		SendJSuccess(w, list)
	}
}
