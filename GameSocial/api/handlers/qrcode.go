package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gamesocial/modules/qrcode"
)

func AdminQRCodesCreate(svc qrcode.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJBizFail(w, "qrcode service not configured")
			return
		}

		var req qrcode.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		out, err := svc.Create(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

func AppQRCodesVerify(svc qrcode.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJBizFail(w, "qrcode service not configured")
			return
		}
		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		var req struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		p, err := svc.Verify(r.Context(), req.Token)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}

		SendJSuccess(w, map[string]any{
			"uuid":      p.UUID,
			"type":      p.Type,
			"scene":     p.Scene,
			"userId":    p.UserID,
			"issuedAt":  time.Unix(p.IssuedAt, 0).Format(time.RFC3339),
			"expiresAt": time.Unix(p.ExpiresAt, 0).Format(time.RFC3339),
			"data":      json.RawMessage(p.Data),
		})
	}
}

func AppQRCodesUse(svc qrcode.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJBizFail(w, "qrcode service not configured")
			return
		}
		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		var req qrcode.UseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		out, err := svc.Use(r.Context(), uid, req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}
