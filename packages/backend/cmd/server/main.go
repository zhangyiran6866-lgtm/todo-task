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
)

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
	registerRoutes(r, log)

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

func registerRoutes(r *gin.Engine, log *zap.Logger) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok", "env": gin.Mode()})
	})

	// API v1 路由组（后续 Phase 中挂载各模块 handler）
	v1 := r.Group("/api/v1")
	_ = v1 // 占位，Phase 1 开始使用

	log.Info("routes registered")
}
