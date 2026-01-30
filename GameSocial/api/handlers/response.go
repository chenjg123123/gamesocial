package handlers

import (
	"encoding/json"
	"net/http"
)

type ResultCode int

const (
	// CodeOK 表示业务层请求成功。
	CodeOK ResultCode = 200
	// CodeWarning 表示请求成功，但需要客户端关注（非错误）。
	CodeWarning ResultCode = 300
	// CodeBadRequest 表示客户端参数不合法。
	CodeBadRequest ResultCode = 400
	// CodeUnauthorized 表示未登录或 token 无效。
	CodeUnauthorized ResultCode = 401
	// CodeForbidden 表示无权限执行该操作。
	CodeForbidden ResultCode = 403
	// CodeNotFound 表示请求的资源不存在。
	CodeNotFound ResultCode = 404
	// CodeMethodNotAllowed 表示该路由不支持当前 HTTP 方法。
	CodeMethodNotAllowed ResultCode = 405
	// CodeInternalError 表示服务端发生未预期的错误。
	CodeInternalError ResultCode = 500
	// CodeNotImplemented 表示 API 已定义但尚未实现。
	CodeNotImplemented ResultCode = 501

	// CodeLoginError 用于登录/鉴权相关的业务失败（例如微信登录）。
	CodeLoginError ResultCode = 1000
	// CodeDBDisabled 表示服务以禁用数据库模式启动，无法提供依赖数据库的接口。
	CodeDBDisabled ResultCode = 1100
)

// APIResponse 是所有 HTTP API 统一的 JSON 返回结构。
type APIResponse struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// SendJSON 使用指定的 HTTP 状态码写入 JSON 响应。
func SendJSON(w http.ResponseWriter, status int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// SendSuccess 返回 HTTP 200，并设置业务码为 CodeOK。
func SendSuccess(w http.ResponseWriter, data any) {
	SendJSON(w, http.StatusOK, APIResponse{Code: int(CodeOK), Data: data})
}

// SendCreated 返回 HTTP 201，并设置业务码为 CodeOK。
func SendCreated(w http.ResponseWriter, data any) {
	SendJSON(w, http.StatusCreated, APIResponse{Code: int(CodeOK), Data: data})
}

// SendWarning 返回 HTTP 200，但业务码为非 CodeOK，用于非错误但需要提示的场景。
func SendWarning(w http.ResponseWriter, code ResultCode, message string, data any) {
	SendJSON(w, http.StatusOK, APIResponse{Code: int(code), Message: message, Data: data})
}

// SendError 同时设置 HTTP 状态码与业务码返回错误信息。
func SendError(w http.ResponseWriter, status int, code ResultCode, message string) {
	SendJSON(w, status, APIResponse{Code: int(code), Message: message})
}

// SendBadRequest 是参数不合法场景的便捷封装。
func SendBadRequest(w http.ResponseWriter, message string) {
	SendError(w, http.StatusBadRequest, CodeBadRequest, message)
}

// SendUnauthorized 是未登录/鉴权失败场景的便捷封装。
func SendUnauthorized(w http.ResponseWriter, message string) {
	SendError(w, http.StatusUnauthorized, CodeUnauthorized, message)
}

// SendNotFound 是资源不存在场景的便捷封装。
func SendNotFound(w http.ResponseWriter, message string) {
	SendError(w, http.StatusNotFound, CodeNotFound, message)
}

// SendInternalError 是服务端内部错误场景的便捷封装。
func SendInternalError(w http.ResponseWriter, message string) {
	SendError(w, http.StatusInternalServerError, CodeInternalError, message)
}
