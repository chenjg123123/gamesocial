package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
)

type GoodsRow struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	CoverURL    string `json:"cover_url,omitempty"`
	PointsPrice int64  `json:"points_price"`
	Stock       int    `json:"stock"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type goodsUpsertRequest struct {
	Name        string `json:"name"`
	CoverURL    string `json:"cover_url"`
	PointsPrice int64  `json:"points_price"`
	Stock       int    `json:"stock"`
	Status      *int   `json:"status"`
}

func ListGoods(db *sql.DB) http.HandlerFunc {
	// ListGoods 返回小程序侧展示的商品列表。
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
			SELECT id, name, IFNULL(cover_url, ''), points_price, stock, status, created_at
			FROM goods
		`
		args := make([]any, 0, 1)
		if status != "" {
			query += " WHERE status = ?"
			args = append(args, status)
		}
		query += " ORDER BY id DESC LIMIT 200"

		rows, err := db.QueryContext(r.Context(), query, args...)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		defer rows.Close()

		list := make([]GoodsRow, 0, 16)
		for rows.Next() {
			var row GoodsRow
			var createdAt time.Time
			if err := rows.Scan(&row.ID, &row.Name, &row.CoverURL, &row.PointsPrice, &row.Stock, &row.Status, &createdAt); err != nil {
				SendInternalError(w, err.Error())
				return
			}
			row.CreatedAt = formatTime(createdAt)
			list = append(list, row)
		}
		if err := rows.Err(); err != nil {
			SendInternalError(w, err.Error())
			return
		}

		SendSuccess(w, list)
	}
}

func GetGoods(db *sql.DB) http.HandlerFunc {
	// GetGoods 根据 id 返回单个商品详情。
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

		var row GoodsRow
		var createdAt time.Time
		err = db.QueryRowContext(r.Context(), `
			SELECT id, name, IFNULL(cover_url, ''), points_price, stock, status, created_at
			FROM goods
			WHERE id = ?
		`, id).Scan(&row.ID, &row.Name, &row.CoverURL, &row.PointsPrice, &row.Stock, &row.Status, &createdAt)
		if err == sql.ErrNoRows {
			SendNotFound(w, "not found")
			return
		}
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		row.CreatedAt = formatTime(createdAt)

		SendSuccess(w, row)
	}
}

func AdminCreateGoods(db *sql.DB) http.HandlerFunc {
	// AdminCreateGoods 创建一条商品记录（后台接口）。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		if db == nil {
			SendError(w, http.StatusServiceUnavailable, CodeDBDisabled, "database disabled")
			return
		}

		var req goodsUpsertRequest
		if err := decodeJSON(r, &req); err != nil {
			SendBadRequest(w, "invalid json")
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || req.PointsPrice < 0 || req.Stock < 0 {
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

		res, err := db.ExecContext(r.Context(), `
			INSERT INTO goods (name, cover_url, points_price, stock, status, created_at)
			VALUES (?, NULLIF(?, ''), ?, ?, ?, NOW())
		`, req.Name, strings.TrimSpace(req.CoverURL), req.PointsPrice, req.Stock, status)
		if err != nil {
			SendInternalError(w, err.Error())
			return
		}
		id, _ := res.LastInsertId()

		SendCreated(w, map[string]any{"id": id})
	}
}

func AdminUpdateGoods(db *sql.DB) http.HandlerFunc {
	// AdminUpdateGoods 根据 id 全量更新一条商品记录（后台接口）。
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

		var req goodsUpsertRequest
		if err := decodeJSON(r, &req); err != nil {
			SendBadRequest(w, "invalid json")
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || req.PointsPrice < 0 || req.Stock < 0 || req.Status == nil || (*req.Status != 0 && *req.Status != 1) {
			SendBadRequest(w, "invalid params")
			return
		}

		res, err := db.ExecContext(r.Context(), `
			UPDATE goods
			SET name = ?, cover_url = NULLIF(?, ''), points_price = ?, stock = ?, status = ?
			WHERE id = ?
		`, req.Name, strings.TrimSpace(req.CoverURL), req.PointsPrice, req.Stock, *req.Status, id)
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

func AdminDeleteGoods(db *sql.DB) http.HandlerFunc {
	// AdminDeleteGoods 根据 id 删除一条商品记录（后台接口）。
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

		res, err := db.ExecContext(r.Context(), `DELETE FROM goods WHERE id = ?`, id)
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
