// handlers 提供 HTTP Handler 的实现与通用响应封装。
package handlers

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Code BizCode `json:"code"`
	// Data: 业务返回数据（成功时使用；omitempty 避免空值输出）。
	Data any `json:"data,omitempty"`
	// Message: 错误/提示信息（失败时使用；omitempty 避免空值输出）。
	Message string `json:"message,omitempty"`
}

// SendJSON 将统一的 APIResponse 编码为 JSON 并写入响应。
func SendJSON(w http.ResponseWriter, status int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func SendJSuccess(w http.ResponseWriter, data any) {
	SendJSON(w, http.StatusOK, APIResponse{
		Code:    CodeOK,
		Data:    data,
		Message: CodeOK.DefaultMessage(),
	})
}

func SendJBizFail(w http.ResponseWriter, message string) {
	if message == "" {
		message = CodeBizNotDone.DefaultMessage()
	}
	SendJSON(w, http.StatusOK, APIResponse{
		Code:    CodeBizNotDone,
		Message: message,
	})
}

func SendJError(w http.ResponseWriter, httpStatus int, code BizCode, message string) {
	if message == "" {
		message = code.DefaultMessage()
	}
	SendJSON(w, httpStatus, APIResponse{
		Code:    code,
		Message: message,
	})
}
