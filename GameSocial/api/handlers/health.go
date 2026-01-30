package handlers

import (
	"net/http"
	"time"
)

func Health() http.HandlerFunc {
	// Health 是用于监控/部署探活的简单健康检查接口。
	startedAt := time.Now()
	return func(w http.ResponseWriter, r *http.Request) {
		SendSuccess(w, map[string]any{
			"status":     "ok",
			"started_at": startedAt.Format(time.RFC3339),
			"now":        time.Now().Format(time.RFC3339),
		})
	}
}
