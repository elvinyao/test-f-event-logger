package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/elvinyao/test-f-event-logger/config"
)

// Init 根据配置初始化全局的 slog 日志记录器
func Init(cfg *config.Config) error {
	var level slog.Level
	switch strings.ToUpper(cfg.Log.Level) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var output io.Writer = os.Stdout
	if cfg.Log.FilePath != "" {
		file, err := os.OpenFile(cfg.Log.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = file
	}

	// 使用 JSON Handler 来输出结构化日志
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
	return nil
}
