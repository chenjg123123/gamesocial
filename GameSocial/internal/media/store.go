// media 放置媒体文件的存储/访问实现。
package media

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

type CloudBaseStore struct {
	baseURL    string
	publicBase string
	prefix     string
	deviceID   string
	tokenType  string
	token      string

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func NewStore(bucketURL, secretID, secretKey, publicBase, prefix, cloudBaseTokenType, cloudBaseAccessToken, cloudBaseDeviceID string) (Store, error) {
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, err
	}
	host := strings.ToLower(u.Host)
	if strings.Contains(host, "tcloudbasegateway.com") {
		return NewCloudBaseStore(bucketURL, publicBase, prefix, cloudBaseTokenType, cloudBaseAccessToken, cloudBaseDeviceID)
	}
	return NewCOSStore(bucketURL, secretID, secretKey, publicBase, prefix)
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

func NewCloudBaseStore(apiURL, publicBase, prefix, tokenType, accessToken, deviceID string) (*CloudBaseStore, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New("invalid cloudbase apiURL")
	}
	baseURL := strings.TrimRight(u.Scheme+"://"+u.Host, "/")
	if publicBase == "" {
		publicBase = baseURL
	}
	if deviceID == "" {
		var err error
		deviceID, err = randomHex(16)
		if err != nil {
			return nil, err
		}
	}
	prefix = strings.Trim(prefix, "/")
	tokenType = strings.TrimSpace(tokenType)
	accessToken = strings.TrimSpace(accessToken)
	if accessToken != "" && tokenType == "" {
		tokenType = "Bearer"
	}
	return &CloudBaseStore{
		baseURL:     baseURL,
		publicBase:  strings.TrimRight(publicBase, "/"),
		prefix:      prefix,
		deviceID:    deviceID,
		tokenType:   tokenType,
		token:       accessToken,
		accessToken: "",
		expiresAt:   time.Time{},
	}, nil
}

func (s *CloudBaseStore) Upload(ctx context.Context, file io.ReadSeeker, contentType, filename string) (UploadResult, error) {
	if s == nil || s.baseURL == "" {
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
	objectID := buildKey(s.prefix, now, ext)

	tokenType, accessToken, err := s.getToken(ctx)
	if err != nil {
		return UploadResult{}, err
	}

	info, err := s.getUploadInfo(ctx, tokenType, accessToken, objectID)
	if err != nil {
		return UploadResult{}, err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return UploadResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, info.UploadURL, file)
	if err != nil {
		return UploadResult{}, err
	}
	req.Header.Set("Authorization", info.Authorization)
	req.Header.Set("X-Cos-Security-Token", info.Token)
	req.Header.Set("X-Cos-Meta-Fileid", info.CloudObjectMeta)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UploadResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return UploadResult{}, fmt.Errorf("cloudbase upload failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	outURL := info.DownloadURL
	if outURL == "" {
		outURL = s.publicBase + "/" + objectID
	}

	return UploadResult{
		Key: objectID,
		URL: outURL,
	}, nil
}

func (s *CloudBaseStore) getToken(ctx context.Context) (string, string, error) {
	if s.token != "" && s.tokenType != "" {
		return s.tokenType, s.token, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.accessToken != "" && s.tokenType != "" && time.Until(s.expiresAt) > 30*time.Second {
		return s.tokenType, s.accessToken, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/auth/v1/signin/anonymously", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("x-device-id", s.deviceID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", "", fmt.Errorf("cloudbase signin failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", err
	}
	if out.TokenType == "" || out.AccessToken == "" {
		return "", "", errors.New("cloudbase signin returned empty token")
	}

	expiresAt := time.Now().Add(55 * time.Minute)
	if out.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(out.ExpiresIn) * time.Second)
	}

	s.tokenType = out.TokenType
	s.accessToken = out.AccessToken
	s.expiresAt = expiresAt

	return s.tokenType, s.accessToken, nil
}

type cloudBaseUploadInfo struct {
	UploadURL        string `json:"uploadUrl"`
	DownloadURL      string `json:"downloadUrl"`
	Authorization    string `json:"authorization"`
	Token            string `json:"token"`
	CloudObjectMeta  string `json:"cloudObjectMeta"`
	ObjectID         string `json:"objectId"`
	Code             string `json:"code"`
	Message          string `json:"message"`
	DownloadURLEncod string `json:"downloadUrlEncoded"`
}

func (s *CloudBaseStore) getUploadInfo(ctx context.Context, tokenType, accessToken, objectID string) (cloudBaseUploadInfo, error) {
	body, err := json.Marshal([]map[string]any{{"objectId": objectID}})
	if err != nil {
		return cloudBaseUploadInfo{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/storages/get-objects-upload-info", bytes.NewReader(body))
	if err != nil {
		return cloudBaseUploadInfo{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", tokenType+" "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cloudBaseUploadInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return cloudBaseUploadInfo{}, fmt.Errorf("cloudbase get upload info failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	var arr []cloudBaseUploadInfo
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return cloudBaseUploadInfo{}, err
	}
	if len(arr) == 0 {
		return cloudBaseUploadInfo{}, errors.New("cloudbase upload info is empty")
	}
	info := arr[0]
	if info.Code != "" {
		if info.Code == "ACTION_FORBIDDEN" {
			return cloudBaseUploadInfo{}, fmt.Errorf("cloudbase upload info error: %s: %s (gateway permission denied; grant storage upload permission to this role or use a token with permission)", info.Code, info.Message)
		}
		return cloudBaseUploadInfo{}, fmt.Errorf("cloudbase upload info error: %s: %s", info.Code, info.Message)
	}
	if info.UploadURL == "" || info.Authorization == "" || info.Token == "" || info.CloudObjectMeta == "" {
		return cloudBaseUploadInfo{}, errors.New("cloudbase upload info missing fields")
	}
	if info.DownloadURL == "" && info.DownloadURLEncod != "" {
		info.DownloadURL = info.DownloadURLEncod
	}
	return info, nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
