// database 提供数据库连接初始化与 DSN 组装等基础设施能力。
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DBConfig struct {
	// DSN: 完整 DSN（优先使用）；为空则用其它字段拼接生成。
	DSN string
	// Host/Port/User/Password/DBName: 用于拼接 DSN 的连接参数。
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// InitMySQL 初始化 MySQL 连接并进行一次 Ping 校验连通性。
// 这一步会设置连接池参数，避免默认值在生产环境下造成资源占用或性能问题。
func InitMySQL(cfg DBConfig) (*sql.DB, error) {
	dsn := cfg.DSN
	if dsn == "" {
		// 未提供完整 DSN 时，按约定的参数生成。
		dsn = buildMySQLDSN(cfg)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 连接池参数：可根据实际负载再调整。
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	// 用超时上下文做一次探活，启动期即可发现连接不可用的问题。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// buildMySQLDSN 将配置拼接为 go-sql-driver/mysql 支持的 DSN。
func buildMySQLDSN(cfg DBConfig) string {
	values := url.Values{}
	// parseTime=true 让 DATETIME/TIMESTAMP 自动解析为 time.Time（避免拿到 []byte/string）。
	values.Set("parseTime", "true")
	values.Set("charset", "utf8mb4")
	values.Set("loc", "Local")
	values.Set("timeout", "10s")
	values.Set("readTimeout", "10s")
	values.Set("writeTimeout", "10s")

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		values.Encode(),
	)
}
