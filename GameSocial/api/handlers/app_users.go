package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gamesocial/internal/media"
	"gamesocial/modules/user"
)

// AppUserMeGet 获取当前用户信息。
// GET /api/users/me
func AppUserMeGet(svc user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}
		out, err := svc.Get(r.Context(), uid)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, struct {
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatarUrl"`
			Level     int    `json:"level"`
			Exp       int64  `json:"exp"`
			CreatedAt string `json:"createdAt"`
		}{
			Nickname:  out.Nickname,
			AvatarURL: out.AvatarURL,
			Level:     out.Level,
			Exp:       out.Exp,
			CreatedAt: out.CreatedAt.Format(time.RFC3339),
		})
	}
}

// AppUserMeUpdate 更新当前用户资料。
// PUT /api/users/me
func AppUserMeUpdate(svc user.Service, store media.ServerStore, maxUploadBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}

		var nickname *string
		var avatarURL *string

		ct := strings.TrimSpace(r.Header.Get("Content-Type"))
		if strings.HasPrefix(strings.ToLower(ct), "multipart/form-data") {
			if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
				SendJBizFail(w, "参数格式错误")
				return
			}
			if r.MultipartForm != nil {
				if vals, ok := r.MultipartForm.Value["nickname"]; ok && len(vals) != 0 {
					v := vals[0]
					nickname = &v
				}
			}

			f, _, err := r.FormFile("file")
			if err == nil {
				_ = f.Close()
				out, err := uploadImageToStore(r, store, maxUploadBytes)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				v := out.URL
				avatarURL = &v
			} else if !errors.Is(err, http.ErrMissingFile) {
				SendJBizFail(w, "参数格式错误")
				return
			}
		} else {
			var req struct {
				Nickname  *string `json:"nickname"`
				AvatarURL *string `json:"avatarUrl"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				SendJBizFail(w, "参数格式错误")
				return
			}
			nickname = req.Nickname
			if req.AvatarURL != nil {
				url, err := maybeUploadImageString(r.Context(), store, maxUploadBytes, *req.AvatarURL)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				avatarURL = &url
			}
		}

		out, err := svc.Update(r.Context(), uid, user.UpdateUserRequest{
			Nickname:  nickname,
			AvatarURL: avatarURL,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, struct {
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatarUrl"`
			Level     int    `json:"level"`
			Exp       int64  `json:"exp"`
			CreatedAt string `json:"createdAt"`
		}{
			Nickname:  out.Nickname,
			AvatarURL: out.AvatarURL,
			Level:     out.Level,
			Exp:       out.Exp,
			CreatedAt: out.CreatedAt.Format(time.RFC3339),
		})
	}
}

// AppMediaTempUploadInfos 下发临时直传“上传凭证”。
// POST /api/media/temp-upload-infos
//
// 解决的问题（多图上传 + 用户可能取消）：
// - 如果前端一选图就立刻走“正式上传”，用户取消后就会产生垃圾文件，占用存储。
// - 所以改为：先把图片直传到 temp/...；用户最终点“保存/发布”时，再把 URL 写入业务表。
//
// 这个接口只做一件事：根据当前登录用户，生成一批 objectId（全部位于 temp/...），
// 然后为每个 objectId 生成一个短时有效的 PUT 签名（Authorization），返回给前端直传。
//
// 前端拿到 items[] 后，需要对每个 item 执行一次 PUT：
// - URL：item.uploadUrl
// - Headers：
//   - Authorization: item.authorization
//   - Content-Type: 请求时的 contentType
func AppMediaTempUploadInfos(store media.DirectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}

		// 必须登录：这里依赖中间件从 Authorization: Bearer <token> 解析 userId。
		uid := userIDFromRequest(r)
		if uid == 0 {
			SendJError(w, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}
		if store == nil {
			SendJBizFail(w, "媒体直传未配置：请配置 MEDIA_COS_BUCKET_URL + MEDIA_COS_SECRET_ID + MEDIA_COS_SECRET_KEY")
			return
		}

		// 解析请求：count 表示需要申请多少个“上传坑位”（也就是 items 的数量）。
		var req struct {
			Count       int    `json:"count"`
			ContentType string `json:"contentType"`
			Scene       string `json:"scene"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendJBizFail(w, "参数格式错误")
			return
		}

		// 做严格限制，避免一次申请过多，导致滥用或资源浪费。
		count := req.Count
		if count <= 0 {
			SendJBizFail(w, "count 参数错误")
			return
		}
		if count > 10 {
			SendJBizFail(w, "最多申请 10 张图片的上传凭证")
			return
		}

		// contentType 只允许 image/*，防止前端拿直传凭证去上传任意文件。
		// 注意：这只是“接口层面的限制”，最终还应在云开发安全规则里限制 temp 前缀与写权限。
		contentType := strings.TrimSpace(req.ContentType)
		if contentType == "" {
			contentType = "image/png"
		}
		if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
			SendJBizFail(w, "仅支持 image/*")
			return
		}

		// scene 是业务场景目录，用于做图片分类（商品/赛事/头像/通用）。
		// 只允许白名单，防止前端自定义路径突破目录隔离。
		scene := strings.Trim(strings.TrimSpace(req.Scene), "/")
		switch scene {
		case "goods", "tournament", "user", "common":
		default:
			scene = "common"
		}

		// sessionId 用于把“这一批临时图片”归为一次会话：
		// - 前端可以把 sessionId 缓存在表单里
		// - 未来你要做“提交时校验/清理未提交图片”，sessionId 是关键索引
		sessionID, err := randomHexString(16)
		if err != nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 通过 contentType 推断扩展名，主要是为了：
		// - 让对象 key 更直观（.png/.jpg）
		// - 方便部分客户端/平台按扩展名处理
		// 如果遇到未知类型，会返回空扩展名（仍可上传）。
		ext := extFromContentType(contentType)
		date := time.Now().Format("20060102")

		// objectId 统一固定在 temp/... 目录下，防止直接写入正式目录。
		// temp/<scene>/u<uid>/<sessionId>/<yyyymmdd>/<random>.<ext>
		objectPrefix := fmt.Sprintf("temp/%s/u%d/%s/%s", scene, uid, sessionID, date)
		objectIDs := make([]string, 0, count)
		for i := 0; i < count; i++ {
			name, err := randomHexString(16)
			if err != nil {
				SendJError(w, http.StatusInternalServerError, CodeInternal, "")
				return
			}
			objectIDs = append(objectIDs, objectPrefix+"/"+name+ext)
		}

		infos, err := store.GetObjectsUploadInfo(r.Context(), objectIDs)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}

		SendJSuccess(w, struct {
			SessionID string                   `json:"sessionId"`
			Scene     string                   `json:"scene"`
			Items     []media.ObjectUploadInfo `json:"items"`
		}{
			SessionID: sessionID,
			Scene:     scene,
			Items:     infos,
		})
	}
}

// randomHexString 生成随机十六进制字符串，用于 sessionId/文件名等。
// nBytes=16 时，输出长度为 32 个 hex 字符。
func randomHexString(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// extFromContentType 用 Content-Type 推断扩展名。
// 这里不是强依赖：推不出来就返回空字符串，objectId 不带扩展名也能上传。
func extFromContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
