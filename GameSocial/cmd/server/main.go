// GameSocial 服务入口：加载配置 -> 初始化基础设施（DB 等）-> 注册路由/中间件 -> 启动 HTTP Server。
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
	"gamesocial/internal/media"
	"gamesocial/internal/wechat"
	"gamesocial/modules/auth"
	"gamesocial/modules/item"
	"gamesocial/modules/redeem"
	"gamesocial/modules/task"
	"gamesocial/modules/tournament"
	"gamesocial/modules/user"
)

// App 聚合服务运行所需的配置与各业务模块依赖。
type App struct {
	// Config: 运行时配置（端口、DB 开关等）。
	Config config.Config
	// DB: 数据库连接；当 DBEnabled=false 时为 nil。
	DB *sql.DB
	// AuthSvc: 登录与 token 签发服务。
	AuthSvc auth.Service
	// ItemSvc: 商品/饮品业务服务。
	ItemSvc item.Service
	// TournamentSvc: 赛事业务服务。
	TournamentSvc tournament.Service
	// TaskSvc: 任务定义业务服务。
	TaskSvc task.Service
	// UserSvc: 用户管理业务服务。
	UserSvc user.Service
	// RedeemSvc: 兑换订单业务服务。
	RedeemSvc redeem.Service

	// MediaStore: 媒体上传存储（如腾讯云 COS）。
	MediaStore media.Store
	// MediaMaxUploadBytes: 上传文件大小限制（字节）。
	MediaMaxUploadBytes int64
}

func main() {
	// 加载 .env 与环境变量配置（.env 会被环境变量覆盖）。
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	var db *sql.DB
	if cfg.DBEnabled {
		// DBEnabled=true 时才初始化数据库连接；否则服务仍可启动（例如仅用于健康检查）。
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
		AuthSvc: auth.NewService(
			db,
			wechat.NewClient(cfg.WechatAppID, cfg.WechatAppSecret),
			cfg.AuthTokenSecret,
			cfg.AuthTokenTTLSeconds,
		),
		ItemSvc:       item.NewService(db),
		TournamentSvc: tournament.NewService(db),
		TaskSvc:       task.NewService(db),
		UserSvc:       user.NewService(db),
		RedeemSvc:     redeem.NewService(db),
	}

	app.MediaMaxUploadBytes = cfg.MediaMaxUploadMB * 1024 * 1024
	if cfg.MediaCOSBucketURL != "" {
		store, err := media.NewStore(
			cfg.MediaCOSBucketURL,
			cfg.MediaCOSSecretID,
			cfg.MediaCOSSecretKey,
			cfg.MediaCOSPublicBaseURL,
			cfg.MediaCOSKeyPathPrefix,
			cfg.MediaCloudBaseTokenType,
			cfg.MediaCloudBaseAccessToken,
			cfg.MediaCloudBaseDeviceID,
		)
		if err != nil {
			log.Fatalf("init media store: %v", err)
		}
		app.MediaStore = store
	}

	// 使用 net/http 的 ServeMux 进行路由分发（Go 1.22+ 支持 "METHOD /path" 形式的模式）。
	mux := http.NewServeMux()
	registerRoutes(mux, app)

	// 将中间件包裹在路由处理器外层：Recover(防崩溃) -> CORS -> Logging。
	handler := middleware.Chain(
		mux,
		middleware.Recover(),
		middleware.InjectUserIDFromToken(cfg.AuthTokenSecret),
		middleware.CORS("*"),
		middleware.Logging(),
	)

	// 配置 HTTP Server 的超时，避免慢请求占用连接资源。
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 等待退出信号，优雅关闭服务（让 in-flight 请求有机会完成）。
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
	// 路由只负责 HTTP 语义（方法/路径/参数），具体业务逻辑由 handlers 层实现。
	mux.HandleFunc("GET /health", handlers.Health())
	mux.HandleFunc("POST /api/auth/wechat/login", handlers.WechatLogin(app.AuthSvc))

	mux.HandleFunc("GET /api/users/me", handlers.AppUserMeGet(app.UserSvc))
	mux.HandleFunc("PUT /api/users/me", handlers.AppUserMeUpdate(app.UserSvc))
	mux.HandleFunc("GET /api/goods", handlers.AppGoodsList(app.ItemSvc))
	mux.HandleFunc("GET /api/goods/{id}", handlers.AppGoodsGet(app.ItemSvc))
	mux.HandleFunc("GET /api/tournaments", handlers.AppTournamentsList(app.TournamentSvc))
	mux.HandleFunc("GET /api/tournaments/joined", handlers.AppTournamentsJoined(app.TournamentSvc))
	mux.HandleFunc("GET /api/tournaments/{id}", handlers.AppTournamentsGet(app.TournamentSvc))
	mux.HandleFunc("POST /api/tournaments/{id}/join", handlers.AppTournamentsJoin(app.TournamentSvc))
	mux.HandleFunc("PUT /api/tournaments/{id}/cancel", handlers.AppTournamentsCancel(app.TournamentSvc))
	mux.HandleFunc("GET /api/tournaments/{id}/results", handlers.AppTournamentsResults(app.TournamentSvc))
	mux.HandleFunc("GET /api/redeem/orders", handlers.AppRedeemOrderList(app.RedeemSvc))
	mux.HandleFunc("POST /api/redeem/orders", handlers.AppRedeemOrderCreate(app.RedeemSvc))
	mux.HandleFunc("GET /api/redeem/orders/{id}", handlers.AppRedeemOrderGet(app.RedeemSvc))
	mux.HandleFunc("PUT /api/redeem/orders/{id}/cancel", handlers.AppRedeemOrderCancel(app.RedeemSvc))
	mux.HandleFunc("GET /api/points/balance", handlers.AppPointsBalance(app.DB))
	mux.HandleFunc("GET /api/points/ledgers", handlers.AppPointsLedgers(app.DB))
	mux.HandleFunc("GET /api/vip/status", handlers.AppVipStatus(app.DB))
	mux.HandleFunc("GET /api/tasks", handlers.AppTasksList(app.TaskSvc))
	mux.HandleFunc("POST /api/tasks/checkin", handlers.AppTasksCheckin())
	mux.HandleFunc("POST /api/tasks/{taskCode}/claim", handlers.AppTasksClaim())
	mux.HandleFunc("POST /api/media/upload", handlers.AppMediaUpload(app.MediaStore, app.MediaMaxUploadBytes))

	// 管理员侧：商品管理 CRUD（暂未接入管理员鉴权中间件）。
	mux.HandleFunc("POST /admin/goods", handlers.AdminGoodsCreate(app.ItemSvc))
	mux.HandleFunc("GET /admin/goods", handlers.AdminGoodsList(app.ItemSvc))
	mux.HandleFunc("GET /admin/goods/{id}", handlers.AdminGoodsGet(app.ItemSvc))
	mux.HandleFunc("PUT /admin/goods/{id}", handlers.AdminGoodsUpdate(app.ItemSvc))
	mux.HandleFunc("DELETE /admin/goods/{id}", handlers.AdminGoodsDelete(app.ItemSvc))

	// 管理员侧：赛事管理 CRUD。
	mux.HandleFunc("POST /admin/tournaments", handlers.AdminTournamentCreate(app.TournamentSvc))
	mux.HandleFunc("GET /admin/tournaments", handlers.AdminTournamentList(app.TournamentSvc))
	mux.HandleFunc("GET /admin/tournaments/{id}", handlers.AdminTournamentGet(app.TournamentSvc))
	mux.HandleFunc("PUT /admin/tournaments/{id}", handlers.AdminTournamentUpdate(app.TournamentSvc))
	mux.HandleFunc("DELETE /admin/tournaments/{id}", handlers.AdminTournamentDelete(app.TournamentSvc))

	// 管理员侧：任务定义管理 CRUD。
	mux.HandleFunc("POST /admin/task-defs", handlers.AdminTaskDefCreate(app.TaskSvc))
	mux.HandleFunc("GET /admin/task-defs", handlers.AdminTaskDefList(app.TaskSvc))
	mux.HandleFunc("GET /admin/task-defs/{id}", handlers.AdminTaskDefGet(app.TaskSvc))
	mux.HandleFunc("PUT /admin/task-defs/{id}", handlers.AdminTaskDefUpdate(app.TaskSvc))
	mux.HandleFunc("DELETE /admin/task-defs/{id}", handlers.AdminTaskDefDelete(app.TaskSvc))

	// 管理员侧：用户查询/更新/封禁。
	mux.HandleFunc("GET /admin/users", handlers.AdminUserList(app.UserSvc))
	mux.HandleFunc("GET /admin/users/{id}", handlers.AdminUserGet(app.UserSvc))
	mux.HandleFunc("PUT /admin/users/{id}", handlers.AdminUserUpdate(app.UserSvc))

	// 管理员侧：兑换订单 CRUD + 核销。
	mux.HandleFunc("POST /admin/redeem/orders", handlers.AdminRedeemOrderCreate(app.RedeemSvc))
	mux.HandleFunc("GET /admin/redeem/orders", handlers.AdminRedeemOrderList(app.RedeemSvc))
	mux.HandleFunc("GET /admin/redeem/orders/{id}", handlers.AdminRedeemOrderGet(app.RedeemSvc))
	mux.HandleFunc("PUT /admin/redeem/orders/{id}/use", handlers.AdminRedeemOrderUse(app.RedeemSvc))
	mux.HandleFunc("PUT /admin/redeem/orders/{id}/cancel", handlers.AdminRedeemOrderCancel(app.RedeemSvc))

	mux.HandleFunc("POST /admin/auth/login", handlers.AdminAuthLogin())
	mux.HandleFunc("GET /admin/auth/me", handlers.AdminAuthMe())
	mux.HandleFunc("POST /admin/auth/logout", handlers.AdminAuthLogout())
	mux.HandleFunc("GET /admin/audit/logs", handlers.AdminAuditLogs(app.DB))
	mux.HandleFunc("POST /admin/points/adjust", handlers.AdminPointsAdjust())
	mux.HandleFunc("PUT /admin/users/{id}/drinks/use", handlers.AdminUsersDrinksUse())
	mux.HandleFunc("POST /admin/tournaments/{id}/results/publish", handlers.AdminTournamentResultsPublish())
	mux.HandleFunc("POST /admin/tournaments/{id}/awards/grant", handlers.AdminTournamentAwardsGrant())
	mux.HandleFunc("POST /admin/media/upload", handlers.AdminMediaUpload(app.MediaStore, app.MediaMaxUploadBytes))
}
