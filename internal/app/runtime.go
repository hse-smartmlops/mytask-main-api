package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	appdb "emplacc-api/internal/app/db"
	"emplacc-api/internal/config"
	"emplacc-api/internal/migrations"
	pkglogger "emplacc-api/pkg/logger"
	"emplacc-api/pkg/tracing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"gorm.io/gorm"
)

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := pkglogger.New(pkglogger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})

	applyTimezone(cfg, logger)

	db, err := appdb.Connect(cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer closeDB(db, logger)

	if cfg.Database.AutoMigrate {
		if err := migrations.Migrate(db, logger); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}
	}

	container := NewContainer(cfg, db)

	tracerProvider, tracerShutdown, err := setupTracing(ctx, cfg, logger)
	if err != nil {
		return err
	}

	server := NewHTTPServer(cfg, logger, HTTPServerDeps{
		V1:             container.V1Deps(),
		TracerProvider: tracerProvider,
		TracerShutdown: tracerShutdown,
	})

	return server.Start(ctx)
}

func setupTracing(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	if !cfg.Tracing.Enabled {
		return nil, func(context.Context) error { return nil }, nil
	}

	provider, shutdown, err := tracing.Setup(ctx, tracing.Config{
		Provider:    cfg.Tracing.Provider,
		Endpoint:    cfg.Tracing.Endpoint,
		ServiceName: cfg.Tracing.ServiceName,
		SampleRate:  cfg.Tracing.SampleRate,
	}, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("setup tracing: %w", err)
	}
	return provider, shutdown, nil
}

func applyTimezone(cfg *config.Config, logger *slog.Logger) {
	if cfg.App.Timezone == "" {
		return
	}

	location, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		logger.Warn("failed to load timezone, falling back to system default",
			slog.String("timezone", cfg.App.Timezone),
			slog.String("error", err.Error()))
		return
	}
	time.Local = location
}

func closeDB(db *gorm.DB, logger *slog.Logger) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Warn("database close failed", slog.String("error", err.Error()))
		return
	}
	if err := sqlDB.Close(); err != nil {
		logger.Warn("database close failed", slog.String("error", err.Error()))
	}
}
