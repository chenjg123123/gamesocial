// 管理员侧用户管理接口（基础查询/更新/封禁）。
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gamesocial/modules/user"
)

// AdminUserGet 获取用户详情。
// GET /admin/users/{id}
func AdminUserGet(svc user.Service) http.HandlerFunc {
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

		// 4) 读取并返回。
		out, err := svc.Get(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminUserList 用户列表。
// GET /admin/users?offset=0&limit=20&status=1
func AdminUserList(svc user.Service) http.HandlerFunc {
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

		// 3) 解析 query。
		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status, _ := strconv.Atoi(q.Get("status"))

		// 4) 调用业务层查询列表。
		out, err := svc.List(r.Context(), user.ListUserRequest{
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

// AdminUserUpdate 更新用户资料/状态（可用于封禁或解封）。
// PUT /admin/users/{id}
func AdminUserUpdate(svc user.Service) http.HandlerFunc {
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

		// 4) 解析请求体。
		var req user.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 5) 调用业务层更新并返回最新详情。
		out, err := svc.Update(r.Context(), id, req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}
