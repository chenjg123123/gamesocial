package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type TaskDefRow struct {
	ID           uint64          `json:"id"`
	TaskCode     string          `json:"task_code"`
	Name         string          `json:"name"`
	PeriodType   string          `json:"period_type"`
	TargetCount  int             `json:"target_count"`
	RewardPoints int64           `json:"reward_points"`
	Status       int             `json:"status"`
	RuleJSON     json.RawMessage `json:"rule_json,omitempty"`
	CreatedAt    string          `json:"created_at"`
}

type taskUpsertRequest struct {
	TaskCode     string          `json:"task_code"`
	Name         string          `json:"name"`
	PeriodType   string          `json:"period_type"`
	TargetCount  int             `json:"target_count"`
	RewardPoints int64           `json:"reward_points"`
	Status       *int            `json:"status"`
	RuleJSON     json.RawMessage `json:"rule_json"`
}

func ListTasks(db *sql.DB) http.HandlerFunc {
	// ListTasks 返回小程序侧展示的任务定义列表。
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
		periodType := strings.TrimSpace(r.URL.Query().Get("period_type"))

		query := `
			SELECT id, task_code, name, period_type, target_count, reward_points, status, rule_json, created_at
			FROM task_def
		`
		args := make([]any, 0, 2)
		where := make([]string, 0, 2)
		if status != "" {
			where = append(where, "status = ?")
			args = append(args, status)
		}
		if periodType != "" {
			where = append(where, "period_type = ?")
			args = append(args, periodType)
		}
		if len(where) > 0 {
			query += " WHERE " + strings.Join(where, " AND ")
		}
		query += " ORDER BY id DESC LIMIT 200"

		rows, err := db.QueryContext(r.Context(), query, args...)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		defer rows.Close()

		list := make([]TaskDefRow, 0, 16)
		for rows.Next() {
			var row TaskDefRow
			var createdAt time.Time
			var ruleBytes []byte
			if err := rows.Scan(
				&row.ID,
				&row.TaskCode,
				&row.Name,
				&row.PeriodType,
				&row.TargetCount,
				&row.RewardPoints,
				&row.Status,
				&ruleBytes,
				&createdAt,
			); err != nil {
				SendInternalError(w, err.Error())
				return
			}
			row.CreatedAt = formatTime(createdAt)
			if len(ruleBytes) > 0 && string(ruleBytes) != "null" {
				row.RuleJSON = json.RawMessage(ruleBytes)
			}
			list = append(list, row)
		}
		if err := rows.Err(); err != nil {
			SendInternalError(w, err.Error())
			return
		}

		SendSuccess(w, list)
	}
}

func GetTask(db *sql.DB) http.HandlerFunc {
	// GetTask 根据 id 返回单个任务定义详情。
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

		var row TaskDefRow
		var createdAt time.Time
		var ruleBytes []byte
		err = db.QueryRowContext(r.Context(), `
			SELECT id, task_code, name, period_type, target_count, reward_points, status, rule_json, created_at
			FROM task_def
			WHERE id = ?
		`, id).Scan(
			&row.ID,
			&row.TaskCode,
			&row.Name,
			&row.PeriodType,
			&row.TargetCount,
			&row.RewardPoints,
			&row.Status,
			&ruleBytes,
			&createdAt,
		)
		if err == sql.ErrNoRows {
			SendNotFound(w, "not found")
			return
		}
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		row.CreatedAt = formatTime(createdAt)
		if len(ruleBytes) > 0 && string(ruleBytes) != "null" {
			row.RuleJSON = json.RawMessage(ruleBytes)
		}

		SendSuccess(w, row)
	}
}

func AdminCreateTask(db *sql.DB) http.HandlerFunc {
	// AdminCreateTask 创建一条任务定义（task_def，后台接口）。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		var req taskUpsertRequest
		if err := decodeJSON(r, &req); err != nil {
			SendBadRequest(w, "invalid json")
			return
		}

		req.TaskCode = strings.TrimSpace(req.TaskCode)
		req.Name = strings.TrimSpace(req.Name)
		req.PeriodType = strings.TrimSpace(req.PeriodType)
		if req.TaskCode == "" || req.Name == "" || (req.PeriodType != "DAILY" && req.PeriodType != "WEEKLY" && req.PeriodType != "MONTHLY") || req.TargetCount <= 0 || req.RewardPoints < 0 {
			SendBadRequest(w, "invalid params")
			return
		}

		status := 1
		if req.Status != nil {
			if *req.Status != 0 && *req.Status != 1 {
				SendBadRequest(w, "invalid params")
				return
			}
			status = *req.Status
		}

		var ruleValue any = nil
		if len(req.RuleJSON) > 0 {
			if !json.Valid(req.RuleJSON) {
				SendBadRequest(w, "invalid rule_json")
				return
			}
			ruleValue = string(req.RuleJSON)
		}

		res, err := db.ExecContext(r.Context(), `
			INSERT INTO task_def (task_code, name, period_type, target_count, reward_points, status, rule_json, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
		`, req.TaskCode, req.Name, req.PeriodType, req.TargetCount, req.RewardPoints, status, ruleValue)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		id, _ := res.LastInsertId()

		SendCreated(w, map[string]any{"id": id})
	}
}

func AdminUpdateTask(db *sql.DB) http.HandlerFunc {
	// AdminUpdateTask 根据 id 全量更新一条任务定义（后台接口）。
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

		var req taskUpsertRequest
		if err := decodeJSON(r, &req); err != nil {
			SendBadRequest(w, "invalid json")
			return
		}

		req.TaskCode = strings.TrimSpace(req.TaskCode)
		req.Name = strings.TrimSpace(req.Name)
		req.PeriodType = strings.TrimSpace(req.PeriodType)
		if req.TaskCode == "" || req.Name == "" || (req.PeriodType != "DAILY" && req.PeriodType != "WEEKLY" && req.PeriodType != "MONTHLY") || req.TargetCount <= 0 || req.RewardPoints < 0 || req.Status == nil || (*req.Status != 0 && *req.Status != 1) {
			SendBadRequest(w, "invalid params")
			return
		}

		var ruleValue any = nil
		if len(req.RuleJSON) > 0 {
			if !json.Valid(req.RuleJSON) {
				SendBadRequest(w, "invalid rule_json")
				return
			}
			ruleValue = string(req.RuleJSON)
		}

		res, err := db.ExecContext(r.Context(), `
			UPDATE task_def
			SET task_code = ?, name = ?, period_type = ?, target_count = ?, reward_points = ?, status = ?, rule_json = ?
			WHERE id = ?
		`, req.TaskCode, req.Name, req.PeriodType, req.TargetCount, req.RewardPoints, *req.Status, ruleValue, id)
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

func AdminDeleteTask(db *sql.DB) http.HandlerFunc {
	// AdminDeleteTask 根据 id 删除一条任务定义（后台接口）。
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

		res, err := db.ExecContext(r.Context(), `DELETE FROM task_def WHERE id = ?`, id)
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
