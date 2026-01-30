// 管理员侧任务定义管理接口（基础增删改查）。
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gamesocial/modules/task"
)

// AdminTaskDefCreate 创建任务定义。
// POST /admin/task-defs
func AdminTaskDefCreate(svc task.Service) http.HandlerFunc {
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

		// 3) 解析请求体。
		var req task.CreateTaskDefRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 4) 调用业务层创建并返回详情。
		out, err := svc.CreateTaskDef(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminTaskDefUpdate 更新任务定义。
// PUT /admin/task-defs/{id}
func AdminTaskDefUpdate(svc task.Service) http.HandlerFunc {
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
		var req task.UpdateTaskDefRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 5) 调用业务层更新并返回详情。
		out, err := svc.UpdateTaskDef(r.Context(), id, req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminTaskDefDelete 删除任务定义（软删：status=0）。
// DELETE /admin/task-defs/{id}
func AdminTaskDefDelete(svc task.Service) http.HandlerFunc {
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

		// 4) 调用业务层软删除。
		if err := svc.DeleteTaskDef(r.Context(), id); err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, map[string]any{"deleted": true})
	}
}

// AdminTaskDefGet 获取任务定义详情。
// GET /admin/task-defs/{id}
func AdminTaskDefGet(svc task.Service) http.HandlerFunc {
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
		out, err := svc.GetTaskDef(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminTaskDefList 任务定义列表。
// GET /admin/task-defs?offset=0&limit=20&status=1
func AdminTaskDefList(svc task.Service) http.HandlerFunc {
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
		out, err := svc.ListTaskDef(r.Context(), task.ListTaskDefRequest{
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
