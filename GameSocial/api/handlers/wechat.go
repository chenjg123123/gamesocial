package handlers

import (
	"net/http"
)

func WechatLogin() http.HandlerFunc {
	// WechatLogin 是小程序登录接口的占位实现（wx.login code -> openid -> token）。
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		SendError(w, http.StatusNotImplemented, CodeNotImplemented, "not implemented")
	}
}
