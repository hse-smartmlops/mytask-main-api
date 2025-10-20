package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level      string
	Format     string
	File       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)
	handlerOptions := &slog.HandlerOptions{Level: level}

	writer := buildWriter(cfg)

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "json") {
		handler = slog.NewJSONHandler(writer, handlerOptions)
	} else {
		handler = slog.NewTextHandler(writer, handlerOptions)
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

func buildWriter(cfg Config) io.Writer {
	var outputs []io.Writer
	outputs = append(outputs, os.Stdout)

	if strings.TrimSpace(cfg.File) != "" {
		rotator := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    maxInt(cfg.MaxSizeMB, 10),
			MaxBackups: maxInt(cfg.MaxBackups, 5),
			MaxAge:     maxInt(cfg.MaxAgeDays, 30),
			Compress:   cfg.Compress,
		}
		outputs = append(outputs, rotator)
	}

	if len(outputs) == 1 {
		return outputs[0]
	}
	return io.MultiWriter(outputs...)
}

func maxInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
