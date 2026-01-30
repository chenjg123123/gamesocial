// 管理员侧赛事管理接口（基础增删改查）。
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gamesocial/modules/tournament"
)

// AdminTournamentCreate 创建赛事。
// POST /admin/tournaments
func AdminTournamentCreate(svc tournament.Service) http.HandlerFunc {
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

		// 3) 解析请求体：创建赛事需要 start_at/end_at 等字段。
		var req tournament.CreateTournamentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 4) 调用业务层：写库并返回创建后的赛事详情。
		out, err := svc.Create(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		// 5) 返回响应。
		SendJSuccess(w, out)
	}
}

// AdminTournamentUpdate 更新赛事。
// PUT /admin/tournaments/{id}
func AdminTournamentUpdate(svc tournament.Service) http.HandlerFunc {
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

		// 3) 解析路径参数 id。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		// 4) 解析请求体。
		var req tournament.UpdateTournamentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 5) 调用业务层：更新并返回最新详情。
		out, err := svc.Update(r.Context(), id, req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminTournamentDelete 删除赛事（软删：status=CANCELED）。
// DELETE /admin/tournaments/{id}
func AdminTournamentDelete(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
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

		// 4) 调用业务层：软删除，保留历史引用。
		if err := svc.Delete(r.Context(), id); err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, map[string]any{"deleted": true})
	}
}

// AdminTournamentGet 获取赛事详情。
// GET /admin/tournaments/{id}
func AdminTournamentGet(svc tournament.Service) http.HandlerFunc {
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

		// 4) 调用业务层读取单条数据。
		out, err := svc.Get(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminTournamentList 赛事列表。
// GET /admin/tournaments?offset=0&limit=20&status=PUBLISHED
func AdminTournamentList(svc tournament.Service) http.HandlerFunc {
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

		// 3) 解析 query：分页 + 状态筛选。
		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status := q.Get("status")

		// 4) 调用业务层并返回列表。
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
