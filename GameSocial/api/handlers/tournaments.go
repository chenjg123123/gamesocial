package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
)

type TournamentRow struct {
	ID               uint64 `json:"id"`
	Title            string `json:"title"`
	Content          string `json:"content,omitempty"`
	CoverURL         string `json:"cover_url,omitempty"`
	StartAt          string `json:"start_at"`
	EndAt            string `json:"end_at"`
	Status           string `json:"status"`
	CreatedByAdminID uint64 `json:"created_by_admin_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type tournamentUpsertRequest struct {
	Title            string `json:"title"`
	Content          string `json:"content"`
	CoverURL         string `json:"cover_url"`
	StartAt          string `json:"start_at"`
	EndAt            string `json:"end_at"`
	Status           string `json:"status"`
	CreatedByAdminID uint64 `json:"created_by_admin_id"`
}

func ListTournaments(db *sql.DB) http.HandlerFunc {
	// ListTournaments 返回小程序侧的赛事列表。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		status := strings.TrimSpace(r.URL.Query().Get("status"))
		query := `
			SELECT id, title, IFNULL(content, ''), IFNULL(cover_url, ''), start_at, end_at, status, created_by_admin_id, created_at, updated_at
			FROM tournament
		`
		args := make([]any, 0, 1)
		if status != "" {
			query += " WHERE status = ?"
			args = append(args, status)
		}
		query += " ORDER BY start_at DESC LIMIT 200"

		rows, err := db.QueryContext(r.Context(), query, args...)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		defer rows.Close()

		list := make([]TournamentRow, 0, 16)
		for rows.Next() {
			var row TournamentRow
			var startAt, endAt, createdAt, updatedAt time.Time
			if err := rows.Scan(
				&row.ID,
				&row.Title,
				&row.Content,
				&row.CoverURL,
				&startAt,
				&endAt,
				&row.Status,
				&row.CreatedByAdminID,
				&createdAt,
				&updatedAt,
			); err != nil {
				SendInternalError(w, err.Error())
				return
			}
			row.StartAt = formatTime(startAt)
			row.EndAt = formatTime(endAt)
			row.CreatedAt = formatTime(createdAt)
			row.UpdatedAt = formatTime(updatedAt)
			list = append(list, row)
		}
		if err := rows.Err(); err != nil {
			SendInternalError(w, err.Error())
			return
		}

		SendSuccess(w, list)
	}
}

func GetTournament(db *sql.DB) http.HandlerFunc {
	// GetTournament 根据 id 返回单个赛事详情。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		id, err := parseUint64PathValue(r, "id")
		if err != nil {
			SendBadRequest(w, "invalid id")
			return
		}

		var row TournamentRow
		var startAt, endAt, createdAt, updatedAt time.Time
		err = db.QueryRowContext(r.Context(), `
			SELECT id, title, IFNULL(content, ''), IFNULL(cover_url, ''), start_at, end_at, status, created_by_admin_id, created_at, updated_at
			FROM tournament
			WHERE id = ?
		`, id).Scan(
			&row.ID,
			&row.Title,
			&row.Content,
			&row.CoverURL,
			&startAt,
			&endAt,
			&row.Status,
			&row.CreatedByAdminID,
			&createdAt,
			&updatedAt,
		)
		if err == sql.ErrNoRows {
			SendNotFound(w, "not found")
			return
		}
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}

		row.StartAt = formatTime(startAt)
		row.EndAt = formatTime(endAt)
		row.CreatedAt = formatTime(createdAt)
		row.UpdatedAt = formatTime(updatedAt)

		SendSuccess(w, row)
	}
}

func AdminCreateTournament(db *sql.DB) http.HandlerFunc {
	// AdminCreateTournament 创建一条赛事记录（后台接口）。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		var req tournamentUpsertRequest
		if err := decodeJSON(r, &req); err != nil {
			SendBadRequest(w, "invalid json")
			return
		}

		req.Title = strings.TrimSpace(req.Title)
		req.Status = strings.TrimSpace(req.Status)
		startAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.StartAt))
		if err != nil {
			SendBadRequest(w, "invalid start_at")
			return
		}
		endAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EndAt))
		if err != nil {
			SendBadRequest(w, "invalid end_at")
			return
		}

		if req.Title == "" || req.Status == "" || req.CreatedByAdminID == 0 || endAt.Before(startAt) {
			SendBadRequest(w, "invalid params")
			return
		}

		res, err := db.ExecContext(r.Context(), `
			INSERT INTO tournament (title, content, cover_url, start_at, end_at, status, created_by_admin_id, created_at, updated_at)
			VALUES (?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, NOW(), NOW())
		`, req.Title, strings.TrimSpace(req.Content), strings.TrimSpace(req.CoverURL), startAt, endAt, req.Status, req.CreatedByAdminID)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		id, _ := res.LastInsertId()

		SendCreated(w, map[string]any{"id": id})
	}
}

func AdminUpdateTournament(db *sql.DB) http.HandlerFunc {
	// AdminUpdateTournament 根据 id 全量更新一条赛事记录（后台接口）。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		id, err := parseUint64PathValue(r, "id")
		if err != nil {
			SendBadRequest(w, "invalid id")
			return
		}

		var req tournamentUpsertRequest
		if err := decodeJSON(r, &req); err != nil {
			SendBadRequest(w, "invalid json")
			return
		}

		req.Title = strings.TrimSpace(req.Title)
		req.Status = strings.TrimSpace(req.Status)
		startAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.StartAt))
		if err != nil {
			SendBadRequest(w, "invalid start_at")
			return
		}
		endAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EndAt))
		if err != nil {
			SendBadRequest(w, "invalid end_at")
			return
		}

		if req.Title == "" || req.Status == "" || req.CreatedByAdminID == 0 || endAt.Before(startAt) {
			SendBadRequest(w, "invalid params")
			return
		}

		res, err := db.ExecContext(r.Context(), `
			UPDATE tournament
			SET title = ?, content = NULLIF(?, ''), cover_url = NULLIF(?, ''), start_at = ?, end_at = ?, status = ?, created_by_admin_id = ?, updated_at = NOW()
			WHERE id = ?
		`, req.Title, strings.TrimSpace(req.Content), strings.TrimSpace(req.CoverURL), startAt, endAt, req.Status, req.CreatedByAdminID, id)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			SendNotFound(w, "not found")
			return
		}

		SendSuccess(w, nil)
	}
}

func AdminDeleteTournament(db *sql.DB) http.HandlerFunc {
	// AdminDeleteTournament 根据 id 删除一条赛事记录（后台接口）。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		id, err := parseUint64PathValue(r, "id")
		if err != nil {
			SendBadRequest(w, "invalid id")
			return
		}

		res, err := db.ExecContext(r.Context(), `DELETE FROM tournament WHERE id = ?`, id)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			SendNotFound(w, "not found")
			return
		}

		SendSuccess(w, nil)
	}
}
