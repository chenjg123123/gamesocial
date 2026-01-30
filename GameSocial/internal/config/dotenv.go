// dotenv 实现一个轻量的 .env 加载器，用于本地开发快速注入环境变量。
package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotEnv 从简单的 KEY=VALUE 文件加载环境变量。
// - 支持空行与 # 开头的注释行
// - 仅设置当前进程里“尚未存在”的环境变量，避免覆盖外部注入的配置
// - 当 path 为空时默认读取 .env
func LoadDotEnv(path string) error {
	if path == "" {
		path = ".env"
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行与注释行，减少解析噪音。
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		// 行内没有 '=' 则忽略（不作为有效配置项）。
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"")
		if key == "" {
			continue
		}
		// 如果已经存在该环境变量则不覆盖，确保“环境变量优先级”高于 .env。
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}

	return scanner.Err()
}
