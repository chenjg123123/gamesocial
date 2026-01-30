package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gamesocial/api/handlers"
	"gamesocial/api/middleware"
	"gamesocial/internal/config"
	"gamesocial/internal/database"
)

type App struct {
	// Config 是从环境变量加载的运行时配置。
	Config config.Config
	// DB 是可选的 MySQL 连接池（DB 禁用时为 nil）。
	DB *sql.DB
}

func main() {
	// main 加载配置、初始化基础设施、注册路由并启动 HTTP 服务。
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	var db *sql.DB
	if cfg.DBEnabled {
		db, err = database.InitMySQL(database.DBConfig{
			DSN:      cfg.DBDSN,
			Host:     cfg.DBHost,
			Port:     cfg.DBPort,
			User:     cfg.DBUser,
			Password: cfg.DBPassword,
			DBName:   cfg.DBName,
		})
		if err != nil {
			log.Fatalf("init database: %v", err)
		}
		defer db.Close()
	}

	app := App{
		Config: cfg,
		DB:     db,
	}

	mux := http.NewServeMux()
	registerRoutes(mux, app)

	handler := middleware.Chain(
		mux,
		middleware.Recover(),
		middleware.CORS("*"),
		middleware.Logging(),
	)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func registerRoutes(mux *http.ServeMux, app App) {
	// registerRoutes 使用 Go 1.22 的路由模式，将 HTTP 路由绑定到各 handler。
	mux.HandleFunc("GET /health", handlers.Health())
	mux.HandleFunc("POST /api/auth/wechat/login", handlers.WechatLogin())
	mux.HandleFunc("POST /admin/auth/login", handlers.AdminLogin())
	mux.HandleFunc("GET /api/debug/users", handlers.DebugListUsers(app.DB))

	mux.HandleFunc("GET /api/goods", handlers.ListGoods(app.DB))
	mux.HandleFunc("GET /api/goods/{id}", handlers.GetGoods(app.DB))
	mux.HandleFunc("POST /admin/goods", handlers.AdminCreateGoods(app.DB))
	mux.HandleFunc("PUT /admin/goods/{id}", handlers.AdminUpdateGoods(app.DB))
	mux.HandleFunc("DELETE /admin/goods/{id}", handlers.AdminDeleteGoods(app.DB))

	mux.HandleFunc("GET /api/tournaments", handlers.ListTournaments(app.DB))
	mux.HandleFunc("GET /api/tournaments/{id}", handlers.GetTournament(app.DB))
	mux.HandleFunc("POST /admin/tournaments", handlers.AdminCreateTournament(app.DB))
	mux.HandleFunc("PUT /admin/tournaments/{id}", handlers.AdminUpdateTournament(app.DB))
	mux.HandleFunc("DELETE /admin/tournaments/{id}", handlers.AdminDeleteTournament(app.DB))

	mux.HandleFunc("GET /api/tasks", handlers.ListTasks(app.DB))
	mux.HandleFunc("GET /api/tasks/{id}", handlers.GetTask(app.DB))
	mux.HandleFunc("POST /admin/tasks", handlers.AdminCreateTask(app.DB))
	mux.HandleFunc("PUT /admin/tasks/{id}", handlers.AdminUpdateTask(app.DB))
	mux.HandleFunc("DELETE /admin/tasks/{id}", handlers.AdminDeleteTask(app.DB))
}
