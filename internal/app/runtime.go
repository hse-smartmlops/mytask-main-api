package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	appdb "emplacc-api/internal/app/db"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/migrations"
	mcprepo "emplacc-api/internal/repo/mcp"
	minioRepo "emplacc-api/internal/repo/minio"
	"emplacc-api/internal/service"
	grpcv1 "emplacc-api/internal/transport/grpc/v1"
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

	var storage ports.ObjectStorage

	if strings.TrimSpace(cfg.Minio.Endpoint) != "" {
		minioStorage, err := minioRepo.NewStorage(cfg.Minio)
		if err != nil {
			logger.Warn("failed to initialize object storage", slog.String("error", err.Error()))
		} else if minioStorage != nil {
			storage = minioStorage
		} else {
			logger.Info("object storage disabled", slog.String("reason", "minio storage returned nil"))
		}
	} else {
		logger.Info("object storage disabled", slog.String("reason", "minio endpoint not configured"))
	}

	var (
		mcpClient  *grpcv1.Client
		mcpService ports.MCPService
	)

	if strings.TrimSpace(cfg.MCP.Address) != "" {
		client, err := grpcv1.NewClient(grpcv1.Config{
			Address: cfg.MCP.Address,
			Timeout: cfg.MCP.Timeout,
			UseTLS:  cfg.MCP.UseTLS,
		})
		if err != nil {
			logger.Warn("failed to initialize mcp client", slog.String("error", err.Error()))
		} else {
			mcpClient = client
			repo := mcprepo.New(client.API())
			mcpService = service.NewMCPService(repo, &cfg.Pagination)
		}
	} else {
		logger.Info("mcp disabled", slog.String("reason", "MCP_GRPC_ADDRESS not configured"))
	}

	if mcpClient != nil {
		defer func() {
			if err := mcpClient.Close(); err != nil {
				logger.Warn("mcp client close failed", slog.String("error", err.Error()))
			}
		}()
	}

	container := NewContainer(cfg, db, storage, mcpService)

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
