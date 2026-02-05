// media 放置媒体文件的存储/访问实现。
package media

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// Store 定义媒体存储能力的抽象接口。
type Store interface {
	Upload(ctx context.Context, file io.ReadSeeker, contentType, filename string) (UploadResult, error)
}

type UploadResult struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type COSStore struct {
	client     *cos.Client
	publicBase string
	prefix     string
}

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

	return &COSStore{
		client:     client,
		publicBase: strings.TrimRight(publicBase, "/"),
		prefix:     prefix,
	}, nil
}

func (s *COSStore) Upload(ctx context.Context, file io.ReadSeeker, contentType, filename string) (UploadResult, error) {
	if s == nil || s.client == nil {
		return UploadResult{}, errors.New("media store not configured")
	}
	if file == nil {
		return UploadResult{}, errors.New("file is nil")
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
		return UploadResult{}, err
	}

	return UploadResult{
		Key: key,
		URL: s.publicBase + "/" + key,
	}, nil
}

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
