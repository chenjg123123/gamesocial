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
	// DSN 可选；为空时会根据 host/port/user/password/dbname 组装 DSN。
	DSN      string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// InitMySQL 根据 cfg 创建并校验 MySQL 连接池。
func InitMySQL(cfg DBConfig) (*sql.DB, error) {
	dsn := cfg.DSN
	if dsn == "" {
		dsn = buildMySQLDSN(cfg)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func buildMySQLDSN(cfg DBConfig) string {
	// buildMySQLDSN 组装带安全默认值的 DSN（例如 parseTime 与超时时间）。
	values := url.Values{}
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
