// media 放置媒体文件的存储/访问实现。
//
// 这里我们只保留两类能力（对应你要的两个功能）：
//
// 1) 服务端上传（ServerStore）
//   - 后端接收 multipart/form-data 文件流
//   - 使用 COS SDK（SecretId/SecretKey）把文件上传到 COS
//
// 2) 客户端直传（DirectStore）
//   - 后端不接收文件内容，只负责为每个 objectId 计算一次 PUT 签名（短时有效）
//   - 前端拿到 uploadUrl + authorization 后，自己 PUT 上传到 COS 的 temp/ 目录
package media

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type ServerStore interface {
	Upload(ctx context.Context, file io.ReadSeeker, contentType, filename string) (UploadResult, error)
}

type DirectStore interface {
	GetObjectsUploadInfo(ctx context.Context, objectIDs []string) ([]ObjectUploadInfo, error)
}

type UploadResult struct {
	// Key 是对象在存储中的“路径/Key”（例如 picture/20260208/xxx.png 或 temp/.../xxx.png）。
	Key string `json:"key"`
	// URL 是对外可访问的 URL（通常是 publicBase + "/" + Key）。
	URL string `json:"url"`
}

type COSStore struct {
	// client 是腾讯云 COS SDK 的客户端（使用 SecretId/SecretKey 进行签名）。
	client     *cos.Client
	secretID   string
	secretKey  string
	uploadBase string
	// publicBase 是对外访问的 URL 前缀（通常是 bucket 域名或 CDN 域名）。
	publicBase string
	// prefix 是业务目录前缀（例如 picture/），用于把媒体文件归类到一个目录下。
	prefix string

	checkOnce sync.Once
	checkErr  error
}

type ObjectUploadInfo struct {
	// ObjectID 是对象路径（也就是 COS 的 key）。
	// 我们的临时直传接口会生成 temp/<scene>/u<uid>/<sessionId>/<yyyymmdd>/<rand>.<ext>
	ObjectID string `json:"objectId"`
	// UploadURL 是客户端直传用的 PUT URL。
	UploadURL string `json:"uploadUrl"`
	// DownloadURL 是上传后可访问的 URL（兜底拼 publicBase）。
	DownloadURL string `json:"downloadUrl"`
	// Authorization 是 PUT 时需要携带的 Authorization header 值。
	Authorization string `json:"authorization"`
	// Token 是 PUT 时需要携带的 X-Cos-Security-Token header 值（当前用永久密钥签名时为空）。
	Token string `json:"token"`
	// CloudObjectMeta 是 PUT 时需要携带的 X-Cos-Meta-Fileid header 值（当前为空）。
	CloudObjectMeta string `json:"cloudObjectMeta"`
}

// NewCOSStore 构建一个 COSStore（服务端用 COS SDK 直接上传）。
//
// 适用场景：
// - 你希望所有上传都由服务端接收文件并上传到 COS
// - 你的服务端可以安全持有 SecretId/SecretKey（客户端绝对不能持有永久密钥）
func NewCOSStore(bucketURL, secretID, secretKey, publicBase, prefix string) (*COSStore, error) {
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New("invalid bucketURL")
	}
	if secretID == "" || secretKey == "" {
		return nil, errors.New("missing tencent cos secret")
	}
	if publicBase == "" {
		publicBase = bucketURL
	}

	client := cos.NewClient(
		&cos.BaseURL{BucketURL: u},
		&http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  secretID,
				SecretKey: secretKey,
			},
		},
	)

	prefix = strings.Trim(prefix, "/")

	uploadBase := strings.TrimRight(u.Scheme+"://"+u.Host, "/")

	return &COSStore{
		client:     client,
		secretID:   secretID,
		secretKey:  secretKey,
		uploadBase: uploadBase,
		publicBase: strings.TrimRight(publicBase, "/"),
		prefix:     prefix,
	}, nil
}

// Upload（COS）：
// - 根据 filename/contentType 推断扩展名
// - buildKey 生成对象 key（路径）
// - 调用 COS SDK 的 Put 上传
func (s *COSStore) Upload(ctx context.Context, file io.ReadSeeker, contentType, filename string) (UploadResult, error) {
	if s == nil || s.client == nil {
		return UploadResult{}, errors.New("media store not configured")
	}
	if file == nil {
		return UploadResult{}, errors.New("file is nil")
	}
	if err := s.ensureAccess(ctx); err != nil {
		return UploadResult{}, err
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
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
	}

	now := time.Now()
	key := buildKey(s.prefix, now, ext)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return UploadResult{}, err
	}

	_, err := s.client.Object.Put(ctx, key, file, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "403") || strings.Contains(strings.ToLower(err.Error()), "accessdenied") {
			return UploadResult{}, fmt.Errorf("cos upload forbidden (403): %w", err)
		}
		return UploadResult{}, err
	}

	return UploadResult{
		Key: key,
		URL: s.publicBase + "/" + key,
	}, nil
}

func (s *COSStore) ensureAccess(ctx context.Context) error {
	if s == nil || s.client == nil {
		return errors.New("media store not configured")
	}
	s.checkOnce.Do(func() {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := s.client.Bucket.Head(ctx2)
		if err != nil {
			if strings.Contains(err.Error(), "403") || strings.Contains(strings.ToLower(err.Error()), "accessdenied") {
				s.checkErr = fmt.Errorf("cos bucket access forbidden (403): %w", err)
				return
			}
			s.checkErr = fmt.Errorf("cos bucket access check failed: %w", err)
		}
	})
	return s.checkErr
}

func (s *COSStore) GetObjectsUploadInfo(ctx context.Context, objectIDs []string) ([]ObjectUploadInfo, error) {
	if s == nil || s.client == nil {
		return nil, errors.New("media store not configured")
	}
	if err := s.ensureAccess(ctx); err != nil {
		return nil, err
	}
	if len(objectIDs) == 0 {
		return nil, nil
	}
	if strings.TrimSpace(s.uploadBase) == "" {
		return nil, errors.New("media store not configured")
	}

	authTime := cos.NewAuthTime(3 * time.Minute)

	out := make([]ObjectUploadInfo, 0, len(objectIDs))
	for _, rawID := range objectIDs {
		id := strings.Trim(rawID, " `\t\r\n")
		id = strings.TrimLeft(id, "/")
		if id == "" {
			return nil, errors.New("invalid objectId")
		}
		if !strings.HasPrefix(id, "temp/") {
			return nil, errors.New("invalid objectId")
		}

		uploadURL := s.uploadBase + "/" + id
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, nil)
		if err != nil {
			return nil, err
		}
		cos.AddAuthorizationHeader(s.secretID, s.secretKey, "", req, authTime)
		auth := strings.TrimSpace(req.Header.Get("Authorization"))
		if auth == "" {
			return nil, errors.New("cos authorization is empty")
		}

		out = append(out, ObjectUploadInfo{
			ObjectID:        id,
			UploadURL:       strings.TrimSpace(uploadURL),
			DownloadURL:     strings.TrimSpace(s.publicBase + "/" + id),
			Authorization:   auth,
			Token:           "",
			CloudObjectMeta: "",
		})
	}
	return out, nil
}

// buildKey 生成对象 key（路径）。
//
// 规则：
// - 基础路径：<yyyymmdd>/<randomHex>
// - 如果 ext 非空，会确保以 "." 开头并追加到末尾
// - 如果 prefix 非空，会变成：<prefix>/<基础路径>
func buildKey(prefix string, now time.Time, ext string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	name := now.Format("20060102") + "/" + hex.EncodeToString(b)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	key := name + ext
	if prefix == "" {
		return key
	}
	return prefix + "/" + key
}
