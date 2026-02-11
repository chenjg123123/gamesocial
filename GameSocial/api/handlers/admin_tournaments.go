// 管理员侧赛事管理接口（基础增删改查）。
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gamesocial/internal/media"
	"gamesocial/modules/tournament"
)

// AdminTournamentCreate 创建赛事。
// POST /admin/tournaments
func AdminTournamentCreate(svc tournament.Service, store media.ServerStore, maxUploadBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodPost {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		var req tournament.CreateTournamentRequest
		ct := strings.TrimSpace(r.Header.Get("Content-Type"))
		if strings.HasPrefix(strings.ToLower(ct), "multipart/form-data") {
			if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
				SendJBizFail(w, "参数格式错误")
				return
			}
			req.Title = strings.TrimSpace(r.FormValue("title"))
			req.Content = strings.TrimSpace(r.FormValue("content"))
			req.Status = strings.TrimSpace(r.FormValue("status"))
			req.CreatedByAdmin = parseUint64(strings.TrimSpace(r.FormValue("createdByAdminId")))

			if v := strings.TrimSpace(r.FormValue("startAt")); v != "" {
				tm, err := time.Parse(time.RFC3339, v)
				if err != nil {
					SendJBizFail(w, "startAt 格式错误")
					return
				}
				req.StartAt = tm
			}
			if v := strings.TrimSpace(r.FormValue("endAt")); v != "" {
				tm, err := time.Parse(time.RFC3339, v)
				if err != nil {
					SendJBizFail(w, "endAt 格式错误")
					return
				}
				req.EndAt = tm
			}

			if r.MultipartForm != nil && (len(r.MultipartForm.File["file"])+len(r.MultipartForm.File["files"]) > 0) {
				outs, err := uploadImagesToStore(r, store, maxUploadBytes)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				urls := make([]string, 0, len(outs))
				for _, it := range outs {
					if it.URL != "" {
						urls = append(urls, it.URL)
					}
				}
				req.ImageURLs = urls
				if len(urls) > 0 {
					req.CoverURL = urls[0]
				}
			} else if r.MultipartForm == nil {
				SendJBizFail(w, "参数格式错误")
				return
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				SendJBizFail(w, "参数格式错误")
				return
			}
			if req.ImageURLs != nil {
				urls, err := maybeUploadImageStrings(r.Context(), store, maxUploadBytes, req.ImageURLs)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				urls, err = media.MoveTempURLs(r.Context(), store, "tournament", urls)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				req.ImageURLs = urls
				if len(urls) > 0 {
					req.CoverURL = urls[0]
				}
			} else if req.CoverURL != "" {
				url, err := maybeUploadImageString(r.Context(), store, maxUploadBytes, req.CoverURL)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				if url != "" {
					urls, err := media.MoveTempURLs(r.Context(), store, "tournament", []string{url})
					if err != nil {
						SendJBizFail(w, err.Error())
						return
					}
					if len(urls) > 0 {
						req.CoverURL = urls[0]
						req.ImageURLs = urls
					} else {
						req.CoverURL = ""
						req.ImageURLs = nil
					}
				} else {
					req.CoverURL = ""
					req.ImageURLs = nil
				}
			}
		}

		// 4) 调用业务层：写库并返回创建后的赛事详情。
		out, err := svc.Create(r.Context(), req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		// 5) 返回响应。
		SendJSuccess(w, out)
	}
}

// AdminTournamentUpdate 更新赛事。
// PUT /admin/tournaments/{id}
func AdminTournamentUpdate(svc tournament.Service, store media.ServerStore, maxUploadBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodPut {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析路径参数 id。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		var req tournament.UpdateTournamentRequest
		ct := strings.TrimSpace(r.Header.Get("Content-Type"))
		if strings.HasPrefix(strings.ToLower(ct), "multipart/form-data") {
			if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
				SendJBizFail(w, "参数格式错误")
				return
			}
			req.Title = strings.TrimSpace(r.FormValue("title"))
			req.Content = strings.TrimSpace(r.FormValue("content"))
			req.Status = strings.TrimSpace(r.FormValue("status"))

			if v := strings.TrimSpace(r.FormValue("startAt")); v != "" {
				tm, err := time.Parse(time.RFC3339, v)
				if err != nil {
					SendJBizFail(w, "startAt 格式错误")
					return
				}
				req.StartAt = tm
			}
			if v := strings.TrimSpace(r.FormValue("endAt")); v != "" {
				tm, err := time.Parse(time.RFC3339, v)
				if err != nil {
					SendJBizFail(w, "endAt 格式错误")
					return
				}
				req.EndAt = tm
			}

			if r.MultipartForm != nil && (len(r.MultipartForm.File["file"])+len(r.MultipartForm.File["files"]) > 0) {
				outs, err := uploadImagesToStore(r, store, maxUploadBytes)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				urls := make([]string, 0, len(outs))
				for _, it := range outs {
					if it.URL != "" {
						urls = append(urls, it.URL)
					}
				}
				req.ImageURLs = urls
				if len(urls) > 0 {
					req.CoverURL = urls[0]
				}
			} else {
				cur, err := svc.Get(r.Context(), id)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				req.CoverURL = cur.CoverURL
				req.ImageURLs = cur.ImageURLs
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				SendJBizFail(w, "参数格式错误")
				return
			}
			if req.ImageURLs != nil {
				urls, err := maybeUploadImageStrings(r.Context(), store, maxUploadBytes, req.ImageURLs)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				urls, err = media.MoveTempURLs(r.Context(), store, "tournament", urls)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				req.ImageURLs = urls
				if len(urls) > 0 {
					req.CoverURL = urls[0]
				} else {
					req.CoverURL = ""
				}
			} else if req.CoverURL != "" {
				url, err := maybeUploadImageString(r.Context(), store, maxUploadBytes, req.CoverURL)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				if url != "" {
					urls, err := media.MoveTempURLs(r.Context(), store, "tournament", []string{url})
					if err != nil {
						SendJBizFail(w, err.Error())
						return
					}
					if len(urls) > 0 {
						req.CoverURL = urls[0]
						req.ImageURLs = urls
					} else {
						req.CoverURL = ""
						req.ImageURLs = nil
					}
				} else {
					req.CoverURL = ""
					req.ImageURLs = nil
				}
			} else {
				cur, err := svc.Get(r.Context(), id)
				if err != nil {
					SendJBizFail(w, err.Error())
					return
				}
				req.CoverURL = cur.CoverURL
				req.ImageURLs = cur.ImageURLs
			}
		}

		// 5) 调用业务层：更新并返回最新详情。
		out, err := svc.Update(r.Context(), id, req)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminTournamentDelete 删除赛事（软删：status=CANCELED）。
// DELETE /admin/tournaments/{id}
func AdminTournamentDelete(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodDelete {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析 id。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		// 4) 调用业务层：软删除，保留历史引用。
		if err := svc.Delete(r.Context(), id); err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, map[string]any{"deleted": true})
	}
}

// AdminTournamentGet 获取赛事详情。
// GET /admin/tournaments/{id}
func AdminTournamentGet(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析 id。
		idRaw := r.PathValue("id")
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil || id == 0 {
			SendJBizFail(w, "id 不合法")
			return
		}

		// 4) 调用业务层读取单条数据。
		out, err := svc.Get(r.Context(), id)
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}

// AdminTournamentList 赛事列表。
// GET /admin/tournaments?offset=0&limit=20&status=PUBLISHED
func AdminTournamentList(svc tournament.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 方法校验。
		if r.Method != http.MethodGet {
			SendJError(w, http.StatusMethodNotAllowed, CodeBizNotDone, "method not allowed")
			return
		}
		// 2) 依赖校验。
		if svc == nil {
			SendJError(w, http.StatusInternalServerError, CodeInternal, "")
			return
		}

		// 3) 解析 query：分页 + 状态筛选。
		q := r.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		limit, _ := strconv.Atoi(q.Get("limit"))
		status := q.Get("status")

		// 4) 调用业务层并返回列表。
		out, err := svc.List(r.Context(), tournament.ListTournamentRequest{
			Offset: offset,
			Limit:  limit,
			Status: status,
		})
		if err != nil {
			SendJBizFail(w, err.Error())
			return
		}
		SendJSuccess(w, out)
	}
}
