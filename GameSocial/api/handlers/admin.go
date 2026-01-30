package handlers

import (
	"net/http"
)

func AdminLogin() http.HandlerFunc {
	// AdminLogin 是后台登录接口的占位实现。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		SendError(w, http.StatusNotImplemented, CodeNotImplemented, "not implemented")
	}
}
