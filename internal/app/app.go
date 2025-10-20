package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	v1 "emplacc-api/api/v1"
	"emplacc-api/internal/config"
	appmetrics "emplacc-api/pkg/metrics"
	"emplacc-api/pkg/swagger"
	"emplacc-api/pkg/tracing"

	echoSwagger "github.com/swaggo/echo-swagger"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type HTTPServerDeps struct {
	V1             v1.Deps
	TracerProvider *sdktrace.TracerProvider
	TracerShutdown func(context.Context) error
}

type HTTPServer struct {
	cfg            *config.Config
	logger         *slog.Logger
	echo           *echo.Echo
	tracerProvider *sdktrace.TracerProvider
	tracerShutdown func(context.Context) error
}

func NewHTTPServer(cfg *config.Config, logger *slog.Logger, deps HTTPServerDeps) *HTTPServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(buildCORSConfig(cfg)))
	e.Use(appmetrics.Middleware())

	registerCommonEndpoints(e, cfg)

	if cfg.Tracing.Enabled && deps.TracerProvider != nil {
		tracer := otel.Tracer(cfg.App.Name)
		e.Use(tracing.Middleware(tracer))
	}

	v1Group := e.Group("/v1")
	v1.RegisterRoutes(v1Group, deps.V1)

	tracerShutdown := deps.TracerShutdown
	if tracerShutdown == nil {
		tracerShutdown = func(context.Context) error { return nil }
	}

	return &HTTPServer{
		cfg:            cfg,
		logger:         logger,
		echo:           e,
		tracerProvider: deps.TracerProvider,
		tracerShutdown: tracerShutdown,
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      s.echo,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("http server starting", slog.String("addr", addr))
		if err := s.echo.StartServer(server); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		return err
	}

	if err := s.tracerShutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}

func buildCORSConfig(cfg *config.Config) middleware.CORSConfig {
	c := middleware.CORSConfig{
		AllowOrigins: cfg.Server.AllowedOrigins,
		AllowMethods: cfg.Server.AllowedMethods,
		AllowHeaders: cfg.Server.AllowedHeaders,
	}

	if cfg.Server.AllowCredentials {
		c.AllowCredentials = true
	}

	if len(c.AllowMethods) == 0 {
		c.AllowMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions}
	}
	if len(c.AllowHeaders) == 0 {
		c.AllowHeaders = []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}
	}
	if len(c.AllowOrigins) == 0 {
		c.AllowOrigins = []string{"*"}
	}

	return c
}

func registerCommonEndpoints(e *echo.Echo, cfg *config.Config) {
	e.GET("/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e.GET("/metrics", echo.WrapHandler(appmetrics.Handler()))

	if cfg.Swagger.Enabled {
		swagger.Configure(swagger.Config{
			Title:       cfg.Swagger.Title,
			Description: cfg.Swagger.Description,
			Version:     cfg.Swagger.Version,
			Host:        cfg.Swagger.Host,
			BasePath:    cfg.Swagger.BasePath,
			Schemes:     cfg.Swagger.Schemes,
		})

		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}
