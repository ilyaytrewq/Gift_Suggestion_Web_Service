package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

func New(cfg config.AppConfig) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(resolveLevel(cfg.LogLevel))

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler).With(
		"service",
		cfg.Name,
		"env",
		cfg.Env,
	)
}

func resolveLevel(raw string) slog.Level {
	switch strings.ToLower(raw) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
