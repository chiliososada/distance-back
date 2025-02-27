package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chiliososada/distance-back/config"
	"github.com/chiliososada/distance-back/internal/api/handler"
	"github.com/chiliososada/distance-back/internal/api/router"
	"github.com/chiliososada/distance-back/internal/repository/mysql"
	"github.com/chiliososada/distance-back/internal/service"
	"github.com/chiliososada/distance-back/pkg/auth"
	"github.com/chiliososada/distance-back/pkg/cache"
	"github.com/chiliososada/distance-back/pkg/database"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/chiliososada/distance-back/pkg/storage"

	"github.com/gin-gonic/gin"
)

var (
	configFile string
	env        string
)

func init() {
	flag.StringVar(&configFile, "config", "", "path to config file")
	flag.StringVar(&env, "env", "development", "running environment (development/production)")
	flag.Parse()

	// 设置gin模式
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	// 1. 加载配置
	if configFile == "" {
		configFile = config.GetConfigPath(env)
	}
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger.InitLogger(&logger.Options{
		LogFileDir:  "./logs",
		AppName:     cfg.App.Name,
		Development: cfg.App.Mode == "development",
		MaxSize:     100, // MB
		MaxBackups:  60,
		MaxAge:      30, // days
	})
	defer logger.Sync()

	if env == "production" {
		logger.Warn("生产环境下请检查session cookie设定")
	}

	// 3. 初始化数据库
	db, err := database.InitMySQL(&cfg.MySQL)
	if err != nil {
		logger.Error("Failed to init MySQL", logger.Any("error", err))
		os.Exit(1)
	}
	defer database.Close()

	// 4. 初始化Redis
	if err := cache.InitRedis(&cfg.Redis); err != nil {
		logger.Error("Failed to init Redis", logger.Any("error", err))
		os.Exit(1)
	}
	defer cache.Close()

	// 5. 初始化Firebase
	if err := auth.InitFirebase(&cfg.Firebase); err != nil {
		logger.Error("Failed to init Firebase", logger.Any("error", err))
		os.Exit(1)
	}

	// 6. 初始化存储服务
	if err := storage.InitStorage(&cfg.Firebase); err != nil {
		logger.Error("Failed to init Storage", logger.Any("error", err))
		os.Exit(1)
	}

	// 7. 初始化仓储层
	userRepo := mysql.NewUserRepository(db)
	topicRepo := mysql.NewTopicRepository(db)
	chatRepo := mysql.NewChatRepository(db)
	relationshipRepo := mysql.NewRelationshipRepository(db)

	// 8. 初始化服务层
	storageService := storage.GetStorage()
	userService := service.NewUserService(userRepo, storageService)
	chatService := service.NewChatService(chatRepo, userRepo, relationshipRepo, storageService)
	relationshipService := service.NewRelationshipService(relationshipRepo, userRepo, chatService)
	topicService := service.NewTopicService(topicRepo, userRepo, relationshipRepo, storageService)

	// 9. 初始化处理器
	h := handler.NewHandler(
		userService,
		topicService,
		chatService,
		relationshipService,
	)

	// 10. 初始化路由
	r := router.SetupRouter(h)

	// 11. 创建HTTP服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", cfg.App.Port),
		Handler:        r,
		ReadTimeout:    cfg.App.ReadTimeout,
		WriteTimeout:   cfg.App.WriteTimeout,
		MaxHeaderBytes: cfg.App.MaxHeaderBytes,
	}

	// 12. 启动服务器

	go func() {

		logger.Info("Server is starting",
			logger.String("addr", srv.Addr),
			logger.String("mode", cfg.App.Mode))

		if err := srv.ListenAndServeTLS("fullchain.pem", "dev.key"); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", logger.Any("error", err))
			os.Exit(1)
		}

	}()

	// 13. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// 14. 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 15. 关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.Any("error", err))
	}

	logger.Info("Server exited")
}
