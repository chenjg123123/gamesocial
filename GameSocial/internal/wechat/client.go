// wechat 封装与微信开放接口的交互（此处先实现小程序 code2session）。
package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	appID      string
	appSecret  string
}

type Code2SessionResult struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// NewClient 创建微信客户端。
// appid/secret 来自微信小程序后台；用于 code2session 换取 openid。
func NewClient(appID, appSecret string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		appID:      appID,
		appSecret:  appSecret,
	}
}

// Code2Session 使用小程序 wx.login() 返回的 code 换取用户 openid/unionid 与 session_key。
func (c *Client) Code2Session(ctx context.Context, code string) (Code2SessionResult, error) {
	values := url.Values{}
	values.Set("appid", c.appID)
	values.Set("secret", c.appSecret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.weixin.qq.com/sns/jscode2session?"+values.Encode(), nil)
	if err != nil {
		return Code2SessionResult{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Code2SessionResult{}, err
	}
	defer resp.Body.Close()

	var out Code2SessionResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Code2SessionResult{}, err
	}
	if out.ErrCode != 0 {
		return Code2SessionResult{}, fmt.Errorf("wechat code2session failed: %d %s", out.ErrCode, out.ErrMsg)
	}
	if out.OpenID == "" {
		return Code2SessionResult{}, fmt.Errorf("wechat code2session missing openid")
	}
	return out, nil
}
