package handlers

type BizCode int

const (
	CodeOK           BizCode = 200
	CodeBizNotDone   BizCode = 201
	CodeUnauthorized BizCode = 401
	CodeForbidden    BizCode = 403
	CodeNotFound     BizCode = 404
	CodeInternal     BizCode = 500
)

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
