package handlers

import (
	"net/http"
	"strconv"

	"gamesocial/modules/tournament"
)

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

// AppTournamentsJoined 获取当前登录用户已报名（JOINED）的赛事列表。
// GET /api/tournaments/joined
// Query：
// - offset: 默认 0
// - limit: 默认 20，最大 200
// - status: 可选，按赛事状态过滤（如 PUBLISHED/FINISHED）；为空则默认排除 CANCELED
// - q: 可选，按赛事标题模糊搜索（LIKE %q%）
func AppTournamentsJoined(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// userId 仅从 JWT 中间件注入的请求头获取，禁止由前端传入。
		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status := q.Get("status")
		keyword := q.Get("q")

		out, err := svc.ListJoined(r.Context(), uid, tournament.ListJoinedTournamentRequest{
			Offset:  offset,
			Limit:   limit,
			Status:  status,
			Keyword: keyword,
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
