package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// AdminAuditLogItem 表示管理员审计日志列表项。
type AdminAuditLogItem struct {
	ID         uint64          `json:"id"`
	AdminID    uint64          `json:"adminId"`
	Action     string          `json:"action"`
	BizType    string          `json:"bizType"`
	BizID      string          `json:"bizId"`
	DetailJSON json.RawMessage `json:"detailJson,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

// AdminAuditLogs 查询管理员审计日志。
// GET /admin/audit/logs
func AdminAuditLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}

		if db == nil {
			SendJBizFail(w, "database disabled")
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

		adminID := parseUint64(q.Get("adminId"))
		where := ""
		args := make([]any, 0, 4)
		if adminID != 0 {
			where = "WHERE admin_id = ?"
			args = append(args, adminID)
		}
		args = append(args, limit, offset)

		rows, err := db.QueryContext(r.Context(), `
			SELECT id, admin_id, action, IFNULL(biz_type, ''), IFNULL(biz_id, ''), IFNULL(detail_json, JSON_OBJECT()), created_at
			FROM admin_audit_log
			`+where+`
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`, args...)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		defer rows.Close()

		out := make([]AdminAuditLogItem, 0, limit)
		for rows.Next() {
			var it AdminAuditLogItem
			var detailBytes []byte
			if err := rows.Scan(&it.ID, &it.AdminID, &it.Action, &it.BizType, &it.BizID, &detailBytes, &it.CreatedAt); err != nil {
				SendJBizFail(w, err.Error())
				return
			}
			if len(detailBytes) != 0 {
				it.DetailJSON = json.RawMessage(detailBytes)
			}
			out = append(out, it)
		}
		if err := rows.Err(); err != nil {
			SendJBizFail(w, err.Error())
			return
		}

		SendJSuccess(w, out)
	}
}
