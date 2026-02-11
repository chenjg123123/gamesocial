// config 负责加载与校验应用运行配置（来自环境变量与 .env 文件）。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config 表示应用运行时配置（来自环境变量与 .env 文件）。
type Config struct {
	// ServerPort: HTTP 服务监听端口。
	ServerPort int

	// DBEnabled: 是否启用数据库（false 时应用会跳过 DB 初始化）。
	DBEnabled bool
	// DBDSN: 完整 DSN；为空时会由 DBHost/DBPort/DBUser/DBPassword/DBName 组合生成。
	DBDSN string
	// DBHost/DBPort/DBUser/DBPassword/DBName: 生成 DSN 所需的各项参数。
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// MediaDir: 本地媒体文件保存目录（后续可用于赛事封面/商品封面上传）。
	MediaDir string
	// BaseURL: 对外访问的基础地址（后续可用于拼接媒体访问 URL 等）。
	BaseURL string

	// MediaMaxUploadMB: 上传文件大小限制（MB）。
	MediaMaxUploadMB int64

	// 服务端上传（COS SDK）配置：
	MediaCOSBucketURL string
	MediaCOSSecretID  string
	MediaCOSSecretKey string

	// WechatAppID/WechatAppSecret: 微信小程序的 appid/secret，用于 code2session 换取 openid。
	WechatAppID     string
	WechatAppSecret string

	// AuthTokenSecret: 用户 token 签名密钥（自定义 HMAC token）；需要自行设置为随机长字符串。
	AuthTokenSecret string
	// AuthTokenTTLSeconds: token 有效期（秒）。
	AuthTokenTTLSeconds int64

	// QRCodePublicKeyPEMBase64 / QRCodePrivateKeyPEMBase64：
	// - 二维码 token 使用“非对称加密（RSA）”
	// - 环境变量建议用 base64 存 PEM，避免换行问题（也支持直接放 PEM）
	QRCodePublicKeyPEMBase64  string
	QRCodePrivateKeyPEMBase64 string

	// QRCodeDefaultTTLSeconds: 生成二维码时默认有效期（秒）。
	QRCodeDefaultTTLSeconds int64
	// QRCodePNGSize: 生成二维码 PNG 默认边长（像素）。
	QRCodePNGSize int
}

// LoadConfig 加载应用配置。
// 加载顺序：
// 1) 如果设置了 ENV_FILE，则从该文件加载（仅对未设置的环境变量做补充）。
// 2) 否则向上查找 .env 并加载（同样只补充未设置的环境变量）。
// 3) 最终从环境变量读取配置并做基础校验。
func LoadConfig() (Config, error) {
	envFile := os.Getenv("ENV_FILE")
	if envFile != "" {
		if err := LoadDotEnv(envFile); err != nil {
			return Config{}, fmt.Errorf("load ENV_FILE %s: %w", envFile, err)
		}
	} else {
		if p, err := findUpwards(".env"); err == nil {
			_ = LoadDotEnv(p)
		}
	}

	cfg := Config{
		ServerPort:          mustInt(getenv("SERVER_PORT", "8080")),
		DBDSN:               os.Getenv("DB_DSN"),
		DBEnabled:           mustBool(getenv("DB_ENABLED", "true")),
		DBHost:              getenv("DB_HOST", "127.0.0.1"),
		DBPort:              mustInt(getenv("DB_PORT", "3306")),
		DBUser:              getenv("DB_USER", "root"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              getenv("DB_NAME", "gamesocial"),
		MediaDir:            getenv("MEDIA_DIR", "./data/media"),
		BaseURL:             getenv("BASE_URL", "http://localhost:8080"),
		MediaMaxUploadMB:    mustInt64(getenv("MEDIA_MAX_UPLOAD_MB", "10")),
		MediaCOSBucketURL:   os.Getenv("MEDIA_COS_BUCKET_URL"),
		MediaCOSSecretID:    os.Getenv("MEDIA_COS_SECRET_ID"),
		MediaCOSSecretKey:   os.Getenv("MEDIA_COS_SECRET_KEY"),
		WechatAppID:         os.Getenv("WECHAT_APP_ID"),
		WechatAppSecret:     os.Getenv("WECHAT_APP_SECRET"),
		AuthTokenSecret:     os.Getenv("AUTH_TOKEN_SECRET"),
		AuthTokenTTLSeconds: mustInt64(getenv("AUTH_TOKEN_TTL_SECONDS", "604800")),

		QRCodePublicKeyPEMBase64:  os.Getenv("QRCODE_PUBLIC_KEY_PEM"),
		QRCodePrivateKeyPEMBase64: os.Getenv("QRCODE_PRIVATE_KEY_PEM"),
		QRCodeDefaultTTLSeconds:   mustInt64(getenv("QRCODE_DEFAULT_TTL_SECONDS", "300")),
		QRCodePNGSize:             mustInt(getenv("QRCODE_PNG_SIZE", "320")),
	}

	if cfg.ServerPort <= 0 {
		return Config{}, fmt.Errorf("invalid SERVER_PORT")
	}
	if cfg.DBPort <= 0 {
		return Config{}, fmt.Errorf("invalid DB_PORT")
	}
	if cfg.AuthTokenTTLSeconds <= 0 {
		return Config{}, fmt.Errorf("invalid AUTH_TOKEN_TTL_SECONDS")
	}
	if cfg.MediaMaxUploadMB <= 0 {
		return Config{}, fmt.Errorf("invalid MEDIA_MAX_UPLOAD_MB")
	}
	if cfg.QRCodeDefaultTTLSeconds <= 0 {
		return Config{}, fmt.Errorf("invalid QRCODE_DEFAULT_TTL_SECONDS")
	}
	if cfg.QRCodePNGSize <= 0 {
		return Config{}, fmt.Errorf("invalid QRCODE_PNG_SIZE")
	}

	return cfg, nil
}

// findUpwards 从当前工作目录开始向上查找指定文件（最多 8 层），用于在本地开发时自动发现 .env。
func findUpwards(filename string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 8; i++ {
		p := filepath.Join(dir, filename)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("%s not found", filename)
}

// getenv 读取环境变量；为空时返回 fallback。
func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

// mustInt 将字符串转为 int；解析失败时返回 0（上层再进行范围校验）。
func mustInt(v string) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

// mustInt64 将字符串转为 int64；解析失败时返回 0。
func mustInt64(v string) int64 {
	n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// mustBool 将字符串转为 bool；解析失败时返回 false。
func mustBool(v string) bool {
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return false
	}
	return b
}
