// 健康检查接口：用于部署探活与简单的进程存活验证。
package handlers

import (
	"net/http"
	"time"
)

// Health 返回健康检查信息（状态 + 服务启动时间 + 当前时间）。
func Health() http.HandlerFunc {
	// startedAt 记录服务启动时刻，用于在健康检查里观察进程是否重启过。
	startedAt := time.Now()
	return func(w http.ResponseWriter, r *http.Request) {
		// 统一返回结构：data 放 status/时间信息。
		SendJSuccess(w, map[string]any{
			"status":    "ok",
			"startedAt": startedAt.Format(time.RFC3339),
			"now":       time.Now().Format(time.RFC3339),
		})
	}
}
