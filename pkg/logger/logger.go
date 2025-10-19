package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level  string
	Format string
}

func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)
	handlerOptions := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "json") {
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
