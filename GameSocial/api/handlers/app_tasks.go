package handlers

import (
	"net/http"
	"strconv"

	"gamesocial/modules/task"
)

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
