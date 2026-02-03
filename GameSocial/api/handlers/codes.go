// handlers 定义 HTTP 层使用的业务状态码（与 HTTP Status 解耦）。
package handlers

// BizCode 表示接口返回的业务码。
type BizCode int

const (
	// CodeOK 表示业务成功。
	CodeOK BizCode = 200
	// CodeBizNotDone 表示业务失败（但 HTTP 仍返回 200，用业务码区分）。
	CodeBizNotDone BizCode = 201
	// CodeUnauthorized 表示登录态异常或缺失。
	CodeUnauthorized BizCode = 401
	// CodeForbidden 表示无权限访问。
	CodeForbidden BizCode = 403
	// CodeNotFound 表示资源不存在。
	CodeNotFound BizCode = 404
	// CodeInternal 表示服务端内部错误。
	CodeInternal BizCode = 500
)

// DefaultMessage 返回业务码对应的默认提示语。
func (c BizCode) DefaultMessage() string {
	switch c {
	case CodeOK:
		return "ok"
	case CodeBizNotDone:
		return "业务未完成"
	case CodeUnauthorized:
		return "登录异常"
	case CodeForbidden:
		return "无权限"
	case CodeNotFound:
		return "资源不存在"
	case CodeInternal:
		return "服务器异常"
	default:
		return "未知错误"
	}
}
