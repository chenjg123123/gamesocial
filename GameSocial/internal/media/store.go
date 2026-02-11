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
	// Upload 将文件上传到长期存储，并返回对象 Key 与可访问 URL。
	//
	// 约定：
	// - 这里是“服务端代传”模式：后端接收文件流，然后用 COS SDK 上传
	// - 返回的 URL 一般用于直接展示（可能是 COS 域名，也可能是 CDN 域名）
	Upload(ctx context.Context, file io.ReadSeeker, contentType, filename string) (UploadResult, error)
}

type DirectStore interface {
	// GetObjectsUploadInfo 生成一批对象的直传信息（PUT URL + Authorization 等）。
	//
	// 约定：
	// - 这里是“客户端直传”模式：后端不接收文件内容，只给前端一个短时有效的签名
	// - objectIDs 通常是 temp/ 开头的 key，由后端/前端协商生成
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
	client *cos.Client
	// secretID/secretKey 仅用于服务端签名与访问 COS（绝不能下发给客户端）。
	secretID  string
	secretKey string
	// uploadBase 是 COS bucket 的基础访问地址（用于拼 PUT URL、拼 Copy 源地址等）。
	// 形如：https://<bucket>.cos.<region>.myqcloud.com
	uploadBase string
	// publicBase 是对外访问的 URL 前缀（通常是 bucket 域名或 CDN 域名）。
	publicBase string
	// prefix 是业务目录前缀（例如 picture/），用于把媒体文件归类到一个目录下。
	prefix string

	// checkOnce/checkErr 用于在首次使用时做一次 bucket 权限探测（Head Bucket），
	// 避免每次上传/签名都触发一次探测请求。
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
	// bucketURL 必须是可被 cos-go-sdk 解析的 BucketURL（包含 scheme 与 host）。
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
	// publicBase 允许单独配置为 CDN 域名；未配置时使用 bucketURL。
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

	// uploadBase 只保留 scheme+host，后续拼接 object key 时统一用 “/key” 追加。
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

	// ext 优先取 filename 后缀；如果 filename 没带后缀，则用 contentType 兜底推断。
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

	// key 是最终存储的对象路径：<prefix>/<yyyymmdd>/<randomHex>.<ext>
	now := time.Now()
	key := buildKey(s.prefix, now, ext)

	// Put 之前把读指针归零，确保 SDK 能从头读取内容。
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

// ensureAccess 做一次 bucket 可访问性探测（Head Bucket），并缓存结果。
//
// 为什么需要：
//   - 直传签名/上传都依赖 bucket 正确配置；如果 bucket 不可访问（例如 403），
//     提前给出可读错误比后续各种上传/签名报错更友好。
//
// 注意：
// - 使用 sync.Once 缓存探测结果，避免每个请求都打一次 Head Bucket。
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

// GetObjectsUploadInfo 为一组对象生成客户端直传所需信息。
//
// 返回字段说明：
// - UploadURL：前端用 PUT 上传的地址（带 objectId/key）
// - Authorization：前端 PUT 时必须带的 Authorization header（短时有效）
// - DownloadURL：上传成功后可访问的 URL（通常是 publicBase + "/" + objectId）
//
// 安全约束：
// - 只允许 temp/ 开头的 objectId，避免客户端把文件覆盖到长期目录或其它业务目录。
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

	// 直传签名的有效期：过短会导致网络慢时上传失败；过长会增加泄露窗口。
	authTime := cos.NewAuthTime(3 * time.Minute)

	out := make([]ObjectUploadInfo, 0, len(objectIDs))
	for _, rawID := range objectIDs {
		// 允许前端把 objectId 带一些空白/反引号之类的字符，这里统一清洗。
		id := strings.Trim(rawID, " `\t\r\n")
		id = strings.TrimLeft(id, "/")
		if id == "" {
			return nil, errors.New("invalid objectId")
		}
		// 只允许 temp/ 开头的 key，用于“先直传临时目录，确认后转存长期目录”的工作流。
		if !strings.HasPrefix(id, "temp/") {
			return nil, errors.New("invalid objectId")
		}

		uploadURL := s.uploadBase + "/" + id
		// 这里不需要 body，只是为了复用 SDK 的签名逻辑生成 Authorization header。
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

// moveFile 将对象从 srcKey “移动”到 dstKey（同桶内 Copy + Delete）。
//
// 注意：
//   - COS 没有真正的 rename/move，这里用 Copy 再 Delete 模拟“移动”语义。
//   - 这是网络 IO + COS API 调用，单个文件移动可能会有 100ms~数秒的延迟，
//     多图场景需要配合并发（见 MoveTempURLs）。
//   - srcKey/dstKey 都是 bucket 内的 object key（不带域名），例如：
//   - srcKey: temp/goods/u1/<sessionId>/20260210/xxx.png
//   - dstKey: uploads/goods/20260210/yyy.png
func moveFile(ctx context.Context, s *COSStore, srcKey, dstKey string) error {
	if s == nil || s.client == nil {
		return errors.New("media store not configured")
	}
	srcKey = strings.TrimLeft(strings.TrimSpace(srcKey), "/")
	dstKey = strings.TrimLeft(strings.TrimSpace(dstKey), "/")
	if srcKey == "" || dstKey == "" {
		return errors.New("invalid object key")
	}
	if err := s.ensureAccess(ctx); err != nil {
		return err
	}
	if s.uploadBase == "" {
		return errors.New("media store not configured")
	}
	u, err := url.Parse(s.uploadBase)
	if err != nil || u == nil || u.Host == "" {
		return errors.New("invalid cos base url")
	}

	// MultiCopy 的 sourceURL 需要是 “<bucket-host>/<srcKey>” 形式（不含 scheme）。
	sourceURL := u.Host + "/" + srcKey
	if _, _, err := s.client.Object.MultiCopy(ctx, dstKey, sourceURL, nil); err != nil {
		return err
	}
	if _, err := s.client.Object.Delete(ctx, srcKey); err != nil {
		return err
	}
	return nil
}

// MoveTempURLs 将一组 URL 中的 “temp/...” 对象转存到长期目录，并返回新的 URL 数组。
//
// 设计目标：
// - 解决“前端直传 temp 目录，用户确认提交后再转存到长期目录”的流程需求
// - 支持 goods/tournament 等多图场景，并在后端返回一组新的长期 URL（保持顺序）
//
// 行为说明：
//   - 仅当 URL 对应的对象 key 以 "temp/" 开头时才会触发转存；否则原样返回。
//   - 目标 key 的生成规则：
//     <store.prefix>/<scene>/<yyyymmdd>/<randomHex><ext>
//     例如：uploads/goods/20260210/abcd....png
//   - 转存通过 Copy + Delete 实现（见 moveFile）。
//
// 性能说明：
// - 多张图如果串行移动会把耗时累加，所以这里采用有限并发（maxConcurrent=4）。
// - 返回结果保持与入参 urls 相同的顺序（会跳过空字符串）。
//
// 兼容性：
// - 仅对 *COSStore 生效；如果传入的是其它 store 实现，会直接原样返回 urls（不报错）。
func MoveTempURLs(ctx context.Context, store ServerStore, scene string, urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, nil
	}
	s, ok := store.(*COSStore)
	if !ok || s == nil || s.client == nil {
		out := make([]string, 0, len(urls))
		for _, v := range urls {
			v = strings.TrimSpace(v)
			if v != "" {
				out = append(out, v)
			}
		}
		return out, nil
	}

	if err := s.ensureAccess(ctx); err != nil {
		return nil, err
	}

	// scene 用于做长期目录分类，例如 goods/tournament/user/common。
	// 注意：这里不会校验 scene 白名单，调用方应在 handler 层控制。
	scene = strings.Trim(strings.TrimSpace(scene), "/")
	if scene == "" {
		scene = "common"
	}
	basePrefix := strings.Trim(s.prefix, "/")
	dstPrefix := scene
	if basePrefix != "" {
		dstPrefix = basePrefix + "/" + scene
	}

	// 并发上限：避免一次请求转存太多图导致请求堆积/打爆 COS。
	// 这里是经验值；如果你的图片通常很多，可以考虑做成配置项。
	const maxConcurrent = 20

	// ctx2 用于“一处失败，全体取消”，避免部分移动成功、部分失败导致状态难以预期。
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	// 信号量控制并发数。
	sem := make(chan struct{}, maxConcurrent)
	errCh := make(chan error, 1)
	// res 保留原始下标，用于最终按入参顺序组装返回值。
	res := make([]string, len(urls))

	// 同一批次使用同一个 now，便于归档目录一致（都是同一天目录）。
	now := time.Now()
	var wg sync.WaitGroup

	for i, raw := range urls {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		// 支持传入完整 URL 或纯 key（例如 temp/...），这里统一提取 object key。
		key, ok := extractObjectKey(raw)
		if !ok || !strings.HasPrefix(key, "temp/") {
			res[i] = raw
			continue
		}

		// 保留扩展名：让长期目录里的对象仍然以正确的图片后缀结尾。
		ext := strings.ToLower(filepath.Ext(key))
		dstKey := buildKey(dstPrefix, now, ext)

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, srcKey, dstKey string) {
			defer wg.Done()
			defer func() { <-sem }()

			if ctx2.Err() != nil {
				return
			}
			if err := moveFile(ctx2, s, srcKey, dstKey); err != nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
				return
			}
			// 转存成功后，返回新的长期 URL（由 publicBase + key 拼出）。
			res[idx] = strings.TrimRight(s.publicBase, "/") + "/" + dstKey
		}(i, key, dstKey)
	}

	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	out := make([]string, 0, len(urls))
	for _, v := range res {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out, nil
}

// extractObjectKey 从 URL 或者 key 字符串中提取 COS object key。
//
// 支持两种输入：
// - 完整 URL：https://<bucket>.cos.<region>.myqcloud.com/temp/.../a.png
// - 纯 key：temp/.../a.png
//
// 返回值：
// - key：去掉前导 "/" 后的对象路径
// - ok：是否提取成功
func extractObjectKey(v string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		u, err := url.Parse(v)
		if err != nil || u == nil {
			return "", false
		}
		key := strings.TrimLeft(strings.TrimSpace(u.Path), "/")
		if key == "" {
			return "", false
		}
		return key, true
	}
	key := strings.TrimLeft(strings.TrimSpace(v), "/")
	if key == "" {
		return "", false
	}
	return key, true
}
