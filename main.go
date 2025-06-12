package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elvinyao/test-f-event-logger/api"
	"github.com/elvinyao/test-f-event-logger/config"
	"github.com/elvinyao/test-f-event-logger/event"
	"github.com/elvinyao/test-f-event-logger/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	if err := logger.Init(cfg); err != nil {
		slog.Error("Failed to initialize logger", "error", err)
		os.Exit(1)
	}

	slog.Info("Configuration loaded successfully", "logLevel", cfg.Log.Level)
	if cfg.Auth.Token == "a-very-secret-token-you-should-change" || cfg.Auth.Token == "default-secret-token" {
		slog.Warn("Security warning: You are using the default authentication token. Please change it in config.yaml or via environment variables.")
	}

	// 3. 初始化事件存储
	eventStore := event.NewEventStore()

	// 4. 初始化 HTTP 处理器
	webhookHandler := api.NewWebhookHandler(eventStore, cfg.Auth.Token)

	// 5. 设置路由和服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookHandler.HandleWebhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: mux,
	}

	// 6. 启动服务器并实现优雅关停
	go func() {
		slog.Info("Starting server...", "address", cfg.Server.Address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// 创建一个带超时的 context 用于关停
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting")
}
