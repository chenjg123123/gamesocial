// handlers 实现面向小程序端与管理端的聚合接口（部分为占位实现）。
package handlers

import (
	"net/http"
	"strconv"
)

func parseUint64(s string) uint64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func userIDFromRequest(r *http.Request) uint64 {
	if r == nil {
		return 0
	}
	if id := parseUint64(r.Header.Get("X-User-Id")); id != 0 {
		return id
	}
	return 0
}
