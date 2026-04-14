package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"todotask/backend/pkg/config"
	"todotask/backend/pkg/logger"
	"todotask/backend/pkg/response"
	"todotask/backend/internal/handler"
	"todotask/backend/internal/middleware"
	"todotask/backend/internal/repository"
	"todotask/backend/internal/service"
)

// @title TodoTask API
// @version 1.0
// @description TodoTask Backend API documentation
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /api/v1
func main() {
	// 1. 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	// 3. 连接 MongoDB
	clientOpts := options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetMaxPoolSize(cfg.MongoDB.MaxPoolSize).
		SetMinPoolSize(cfg.MongoDB.MinPoolSize)

	mongoClient, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatal("mongodb connect failed", zap.Error(err))
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Error("mongodb disconnect failed", zap.Error(err))
		}
	}()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Duration(cfg.MongoDB.ConnectTimeoutSeconds)*time.Second)
	defer pingCancel()
	if err := mongoClient.Ping(pingCtx, nil); err != nil {
		log.Fatal("mongodb ping failed", zap.Error(err))
	}
	log.Info("mongodb connected", zap.String("database", cfg.MongoDB.Database))

	// 4. 初始化 Gin
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())

	// 5. 注册路由
	db := mongoClient.Database(cfg.MongoDB.Database)
	registerRoutes(r, log, db, cfg)

	// 6. 启动服务器 + 优雅退出
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: r,
	}

	go func() {
		log.Info("server starting", zap.Int("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server listen failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("server shutting down...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("server shutdown failed", zap.Error(err))
	}
	log.Info("server exited")
}

func registerRoutes(r *gin.Engine, log *zap.Logger, db *mongo.Database, cfg *config.Config) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok", "env": gin.Mode()})
	})

	// 初始化依赖
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authSvc := service.NewAuthService(userRepo, tokenRepo, &cfg.JWT)
	authHandler := handler.NewAuthHandler(authSvc, log)

	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, log)

	taskRepo := repository.NewTaskRepository(db)
	taskSvc := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskSvc, log)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.Refresh)
			authRoutes.POST("/logout", middleware.JWTAuth(&cfg.JWT), authHandler.Logout)
		}

		userRoutes := v1.Group("/users")
		userRoutes.Use(middleware.JWTAuth(&cfg.JWT))
		{
			userRoutes.GET("/me", userHandler.GetMe)
		}

		taskRoutes := v1.Group("/tasks")
		taskRoutes.Use(middleware.JWTAuth(&cfg.JWT))
		{
			taskRoutes.POST("", taskHandler.CreateTask)
			taskRoutes.GET("", taskHandler.ListTasks)
			taskRoutes.GET("/:id", taskHandler.GetTask)
			taskRoutes.PATCH("/:id", taskHandler.UpdateTask)
			taskRoutes.DELETE("/:id", taskHandler.DeleteTask)
		}
	}

	log.Info("routes registered")
}
