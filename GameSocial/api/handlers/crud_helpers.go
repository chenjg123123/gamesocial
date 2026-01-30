package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func decodeJSON(r *http.Request, dst any) error {
	// decodeJSON 将请求体 JSON 解析到 dst，并拒绝未知字段，便于接口安全演进。
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("invalid json")
	}
	return nil
}

func parseUint64PathValue(r *http.Request, key string) (uint64, error) {
	// parseUint64PathValue 读取路径参数（Go 1.22 路由模式），并解析为 uint64。
	raw := strings.TrimSpace(r.PathValue(key))
	return strconv.ParseUint(raw, 10, 64)
}

func formatTime(t time.Time) string {
	// formatTime 将时间转为 RFC3339 字符串；当为零值时间时返回空字符串。
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
