package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	// ServerPort 是 HTTP 服务监听端口。
	ServerPort int

	// DBEnabled 控制是否初始化 MySQL。
	DBEnabled bool
	// DBDSN 设置后将覆盖 DBHost/DBPort/DBUser/DBPassword/DBName 的组合方式。
	DBDSN      string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// MediaDir 是媒体文件本地存储目录（预留）。
	MediaDir string
	// BaseURL 在需要时用于拼接绝对 URL。
	BaseURL string
}

// LoadConfig 从环境变量与可选的 .env 文件加载配置。
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
		ServerPort: mustInt(getenv("SERVER_PORT", "8080")),
		DBDSN:      os.Getenv("DB_DSN"),
		DBEnabled:  mustBool(getenv("DB_ENABLED", "true")),
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     mustInt(getenv("DB_PORT", "3306")),
		DBUser:     getenv("DB_USER", "root"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getenv("DB_NAME", "gamesocial"),
		MediaDir:   getenv("MEDIA_DIR", "./data/media"),
		BaseURL:    getenv("BASE_URL", "http://localhost:8080"),
	}

	if cfg.ServerPort <= 0 {
		return Config{}, fmt.Errorf("invalid SERVER_PORT")
	}
	if cfg.DBPort <= 0 {
		return Config{}, fmt.Errorf("invalid DB_PORT")
	}

	return cfg, nil
}

func findUpwards(filename string) (string, error) {
	// findUpwards 从当前目录向上查找，返回第一个匹配到的文件路径。
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

func getenv(key, fallback string) string {
	// getenv 读取环境变量；未设置或为空时返回 fallback。
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func mustInt(v string) int {
	// mustInt 将字符串转为 int；失败时返回 0。
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func mustBool(v string) bool {
	// mustBool 解析 bool；失败时返回 false。
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return false
	}
	return b
}
