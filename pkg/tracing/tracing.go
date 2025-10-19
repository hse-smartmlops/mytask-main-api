package tracing

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Provider    string
	Endpoint    string
	ServiceName string
	SampleRate  float64
}

func Setup(ctx context.Context, cfg Config, logger *slog.Logger) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	exporter, err := buildExporter(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	sampler := sdktrace.AlwaysSample()
	if cfg.SampleRate > 0 && cfg.SampleRate < 1 {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			attribute.String("exporter", strings.ToLower(cfg.Provider)),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	logger.Info("tracing enabled",
		slog.String("provider", cfg.Provider),
		slog.String("service", cfg.ServiceName),
	)

	return provider, provider.Shutdown, nil
}

func buildExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	switch strings.ToLower(cfg.Provider) {
	case "stdout", "console", "":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlphttp", "otlp":
		client := otlptracehttp.NewClient(otlptracehttp.WithEndpoint(cfg.Endpoint), otlptracehttp.WithInsecure())
		return otlptrace.New(ctx, client)
	default:
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
}

func Middleware(tracer trace.Tracer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx, span := tracer.Start(req.Context(), req.Method+" "+c.Path(), trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			c.SetRequest(req.WithContext(ctx))
			err := next(c)
			if err != nil {
				span.RecordError(err)
				c.Error(err)
			}

			span.SetAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.route", c.Path()),
				attribute.Int("http.status_code", c.Response().Status),
			)

			return err
		}
	}
}
