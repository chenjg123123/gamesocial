// 微信相关接口（小程序登录等，后续扩展）。
package handlers

import (
	"encoding/json"
	"net/http"

	"gamesocial/modules/auth"
)

type wechatLoginRequest struct {
	Code         string `json:"code"`
	OpenID       string `json:"openId"`
	OpenIDLegacy string `json:"openid"`
}

// WechatLogin 处理小程序登录请求：
// - 临时：仅支持直接传 openId/openid 返回用户数据（可选签发 token）
func WechatLogin(svc auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验：小程序登录预期是 POST。
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}

		// 2) 依赖校验：服务启动时未正确注入业务服务则直接返回系统错误。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析请求体：读取 JSON 并取出 wx.login() 的 code。
		var req wechatLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}
		openID := req.OpenID
		if openID == "" {
			openID = req.OpenIDLegacy
		}
		if openID == "" {
			SendJBizFail(w, "openId 不能为空")
			return
		}

		result, err := svc.OpenIDLogin(r.Context(), openID)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}

		// 5) 返回统一响应：data 里包含 token 与 user 信息。
		SendJSuccess(w, result)
	}
}
