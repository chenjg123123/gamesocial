package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"gamesocial/internal/media"
)

// AdminAuthLogin 管理员登录占位接口。
// POST /admin/auth/login
func AdminAuthLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{
			"token": "admin_token_placeholder",
		})
	}
}

// AdminAuthMe 管理员信息占位接口。
// GET /admin/auth/me
func AdminAuthMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{
			"id":       1,
			"username": "admin",
		})
	}
}

// AdminAuthLogout 管理员登出占位接口。
// POST /admin/auth/logout
func AdminAuthLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"logout": true})
	}
}

// AdminPointsAdjust 积分调整占位接口。
// POST /admin/points/adjust
func AdminPointsAdjust() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		SendJSuccess(w, map[string]any{"adjusted": true})
	}
}

// AdminUsersDrinksUse 消费饮品占位接口。
// PUT /admin/users/{id}/drinks/use
func AdminUsersDrinksUse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"used": true})
	}
}

// AdminTournamentResultsPublish 发布赛事成绩占位接口。
// POST /admin/tournaments/{id}/results/publish
func AdminTournamentResultsPublish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"published": true})
	}
}

// AdminTournamentAwardsGrant 发放赛事奖励占位接口。
// POST /admin/tournaments/{id}/awards/grant
func AdminTournamentAwardsGrant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		_ = parseUint64(r.PathValue("id"))
		SendJSuccess(w, map[string]any{"granted": true})
	}
}

func uploadImageToStore(r *http.Request, store media.ServerStore, maxUploadBytes int64) (media.UploadResult, error) {
	if store == nil {
		return media.UploadResult{}, errors.New("media store not configured: set MEDIA_COS_BUCKET_URL and MEDIA_COS_SECRET_ID/MEDIA_COS_SECRET_KEY, then restart server")
	}
	if maxUploadBytes <= 0 {
		maxUploadBytes = 10 * 1024 * 1024
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return media.UploadResult{}, errors.New("missing form file: file")
	}
	defer file.Close()

	if header == nil || header.Filename == "" {
		return media.UploadResult{}, errors.New("invalid filename")
	}

	if fileSize := header.Size; fileSize > 0 && fileSize > maxUploadBytes {
		return media.UploadResult{}, errors.New("file too large")
	}

	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	if !strings.HasPrefix(ct, "image/") {
		return media.UploadResult{}, errors.New("only image/* is allowed")
	}

	if seeker, ok := file.(interface {
		Seek(offset int64, whence int) (int64, error)
	}); ok {
		end, err := seeker.Seek(0, io.SeekEnd)
		if err == nil {
			if end > maxUploadBytes {
				return media.UploadResult{}, errors.New("file too large")
			}
			_, _ = seeker.Seek(0, io.SeekStart)
		}
	}

	return store.Upload(r.Context(), file, ct, header.Filename)
}

func uploadImagesToStore(r *http.Request, store media.ServerStore, maxUploadBytes int64) ([]media.UploadResult, error) {
	if store == nil {
		return nil, errors.New("media store not configured: set MEDIA_COS_BUCKET_URL and MEDIA_COS_SECRET_ID/MEDIA_COS_SECRET_KEY, then restart server")
	}
	if maxUploadBytes <= 0 {
		maxUploadBytes = 10 * 1024 * 1024
	}
	if r.MultipartForm == nil {
		return nil, errors.New("missing multipart form")
	}

	files := make([]*multipart.FileHeader, 0, 8)
	if v := r.MultipartForm.File["files"]; len(v) > 0 {
		files = append(files, v...)
	}
	if v := r.MultipartForm.File["file"]; len(v) > 0 {
		files = append(files, v...)
	}
	if len(files) == 0 {
		return nil, errors.New("missing form file: file/files")
	}
	if len(files) > 9 {
		return nil, errors.New("too many images")
	}

	out := make([]media.UploadResult, 0, len(files))
	for _, header := range files {
		if header == nil || header.Filename == "" {
			return nil, errors.New("invalid filename")
		}
		if fileSize := header.Size; fileSize > 0 && fileSize > maxUploadBytes {
			return nil, errors.New("file too large")
		}

		file, err := header.Open()
		if err != nil {
			return nil, errors.New("open file failed")
		}

		ct := header.Header.Get("Content-Type")
		if ct == "" {
			var sniff [512]byte
			n, _ := io.ReadFull(file, sniff[:])
			ct = http.DetectContentType(sniff[:n])
			if seeker, ok := file.(interface {
				Seek(offset int64, whence int) (int64, error)
			}); ok {
				_, _ = seeker.Seek(0, io.SeekStart)
			}
		}
		if ct == "" {
			ct = "application/octet-stream"
		}
		if !strings.HasPrefix(ct, "image/") {
			_ = file.Close()
			return nil, errors.New("only image/* is allowed")
		}

		if seeker, ok := file.(interface {
			Seek(offset int64, whence int) (int64, error)
		}); ok {
			end, err := seeker.Seek(0, io.SeekEnd)
			if err == nil {
				if end > maxUploadBytes {
					_ = file.Close()
					return nil, errors.New("file too large")
				}
				_, _ = seeker.Seek(0, io.SeekStart)
			}
		}

		res, err := store.Upload(r.Context(), file, ct, header.Filename)
		_ = file.Close()
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, nil
}

func maybeUploadImageStrings(ctx context.Context, store media.ServerStore, maxUploadBytes int64, list []string) ([]string, error) {
	if len(list) == 0 {
		return nil, nil
	}
	if len(list) > 9 {
		return nil, errors.New("too many images")
	}
	out := make([]string, 0, len(list))
	for _, v := range list {
		url, err := maybeUploadImageString(ctx, store, maxUploadBytes, v)
		if err != nil {
			return nil, err
		}
		if url != "" {
			out = append(out, url)
		}
	}
	return out, nil
}

func maybeUploadImageString(ctx context.Context, store media.ServerStore, maxUploadBytes int64, v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", nil
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return v, nil
	}

	if store == nil {
		return "", errors.New("media store not configured: set MEDIA_COS_BUCKET_URL and restart server")
	}
	if maxUploadBytes <= 0 {
		maxUploadBytes = 10 * 1024 * 1024
	}

	contentType := ""
	raw := v
	if strings.HasPrefix(strings.ToLower(v), "data:") {
		lower := strings.ToLower(v)
		idx := strings.Index(lower, ";base64,")
		if idx <= len("data:") {
			return "", errors.New("invalid image data")
		}
		contentType = strings.TrimSpace(v[len("data:"):idx])
		raw = v[idx+len(";base64,"):]
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("invalid image data")
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(raw)
		if err != nil {
			return "", errors.New("invalid image base64")
		}
	}
	if int64(len(decoded)) > maxUploadBytes {
		return "", errors.New("file too large")
	}

	if contentType == "" {
		contentType = http.DetectContentType(decoded)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return "", errors.New("only image/* is allowed")
	}

	ext := ""
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "image/gif":
		ext = ".gif"
	}
	if ext == "" {
		ext = strings.ToLower(filepath.Ext("x" + contentType))
	}
	if ext == "" {
		ext = ".png"
	}

	out, err := store.Upload(ctx, bytes.NewReader(decoded), contentType, "upload"+ext)
	if err != nil {
		return "", err
	}
	return out.URL, nil
}
