package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	Server     ServerConfig
	Database   DatabaseConfig
	Logger     LoggerConfig
	Swagger    SwaggerConfig
	Tracing    TracingConfig
	Keycloak   KeycloakConfig
	Minio      MinioConfig
	Pagination PaginationConfig
	RateLimit  RateLimitConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Timezone    string
}

type ServerConfig struct {
	Host             string
	Port             int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	ShutdownTimeout  time.Duration
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	Timezone        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	AutoMigrate     bool
}

type LoggerConfig struct {
	Level      string
	Format     string
	File       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

type SwaggerConfig struct {
	Enabled     bool
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
}

type TracingConfig struct {
	Enabled     bool
	Provider    string
	Endpoint    string
	ServiceName string
	SampleRate  float64
}

type PaginationConfig struct {
	DefaultLimit int
	MaxLimit     int
}

type KeycloakConfig struct {
	URL                  string
	Realm                string
	ClientID             string
	ClientSecret         string
	BackendClientID      string
	BackendClientSecret  string
	TokenExchangeEnabled bool
}

type MinioConfig struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	Region       string
	AvatarBucket string
	ReportBucket string
}

type RateLimitConfig struct {
	User RateLimitRule
	IP   RateLimitRule
}

type RateLimitRule struct {
	Requests int
	Window   time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	dbPort, err := toInt(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_PORT: %w", err)
	}

	serverPort, err := toInt(getEnv("SERVER_PORT", "8081"))
	if err != nil {
		return nil, fmt.Errorf("parse SERVER_PORT: %w", err)
	}

	maxOpenConns, err := toInt(getEnv("DB_MAX_OPEN_CONNS", "50"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConns, err := toInt(getEnv("DB_MAX_IDLE_CONNS", "10"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_MAX_IDLE_CONNS: %w", err)
	}

	connMaxLifetime, err := toDuration(getEnv("DB_CONN_MAX_LIFETIME", "1h"))
	if err != nil {
		return nil, fmt.Errorf("parse DB_CONN_MAX_LIFETIME: %w", err)
	}

	logMaxSize, err := toInt(getEnv("LOG_MAX_SIZE_MB", "50"))
	if err != nil {
		return nil, fmt.Errorf("parse LOG_MAX_SIZE_MB: %w", err)
	}

	logMaxBackups, err := toInt(getEnv("LOG_MAX_BACKUPS", "5"))
	if err != nil {
		return nil, fmt.Errorf("parse LOG_MAX_BACKUPS: %w", err)
	}

	logMaxAge, err := toInt(getEnv("LOG_MAX_AGE_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("parse LOG_MAX_AGE_DAYS: %w", err)
	}

	userRateRequests, err := toInt(getEnv("RATE_LIMIT_USER_REQUESTS", "0"))
	if err != nil {
		return nil, fmt.Errorf("parse RATE_LIMIT_USER_REQUESTS: %w", err)
	}
	userRateWindow, err := toDuration(getEnv("RATE_LIMIT_USER_WINDOW", "1m"))
	if err != nil {
		return nil, fmt.Errorf("parse RATE_LIMIT_USER_WINDOW: %w", err)
	}

	ipRateRequests, err := toInt(getEnv("RATE_LIMIT_IP_REQUESTS", "0"))
	if err != nil {
		return nil, fmt.Errorf("parse RATE_LIMIT_IP_REQUESTS: %w", err)
	}
	ipRateWindow, err := toDuration(getEnv("RATE_LIMIT_IP_WINDOW", "1m"))
	if err != nil {
		return nil, fmt.Errorf("parse RATE_LIMIT_IP_WINDOW: %w", err)
	}

	readTimeout, err := toDuration(getEnv("SERVER_READ_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("parse SERVER_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := toDuration(getEnv("SERVER_WRITE_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("parse SERVER_WRITE_TIMEOUT: %w", err)
	}

	shutdownTimeout, err := toDuration(getEnv("SERVER_SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return nil, fmt.Errorf("parse SERVER_SHUTDOWN_TIMEOUT: %w", err)
	}

	sampleRate, err := strconv.ParseFloat(getEnv("TRACING_SAMPLE_RATE", "1"), 64)
	if err != nil {
		return nil, fmt.Errorf("parse TRACING_SAMPLE_RATE: %w", err)
	}

	defaultLimit, err := toInt(getEnv("PAGINATION_DEFAULT_LIMIT", "10"))
	if err != nil {
		return nil, fmt.Errorf("parse PAGINATION_DEFAULT_LIMIT: %w", err)
	}

	maxLimit, err := toInt(getEnv("PAGINATION_MAX_LIMIT", "100"))
	if err != nil {
		return nil, fmt.Errorf("parse PAGINATION_MAX_LIMIT: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "emplacc-api"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENV", "development"),
			Timezone:    getEnv("APP_TIMEZONE", "Europe/Moscow"),
		},
		Server: ServerConfig{
			Host:             getEnv("SERVER_HOST", "0.0.0.0"),
			Port:             serverPort,
			ReadTimeout:      readTimeout,
			WriteTimeout:     writeTimeout,
			ShutdownTimeout:  shutdownTimeout,
			AllowedOrigins:   splitAndTrim(getEnv("SERVER_ALLOWED_ORIGINS", "*")),
			AllowedMethods:   splitAndTrim(getEnv("SERVER_ALLOWED_METHODS", "GET,POST,PUT,DELETE,PATCH,OPTIONS")),
			AllowedHeaders:   splitAndTrim(getEnv("SERVER_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization")),
			AllowCredentials: toBool(getEnv("SERVER_ALLOW_CREDENTIALS", "true")),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            dbPort,
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASS", ""),
			Name:            getEnv("DB_NAME", "emplacc"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			Timezone:        getEnv("DB_TIMEZONE", "Europe/Moscow"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			AutoMigrate:     toBool(getEnv("DB_AUTO_MIGRATE", "true")),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "text"),
			File:       getEnv("LOG_FILE", ""),
			MaxSizeMB:  logMaxSize,
			MaxBackups: logMaxBackups,
			MaxAgeDays: logMaxAge,
			Compress:   toBool(getEnv("LOG_COMPRESS", "true")),
		},
		Swagger: SwaggerConfig{
			Enabled:     toBool(getEnv("SWAGGER_ENABLED", "true")),
			Title:       getEnv("SWAGGER_TITLE", "Emplacc API"),
			Description: getEnv("SWAGGER_DESCRIPTION", "API для Emplacc."),
			Version:     getEnv("SWAGGER_VERSION", "1.0"),
			Host:        getEnv("SWAGGER_HOST", ""),
			BasePath:    getEnv("SWAGGER_BASE_PATH", "/"),
			Schemes:     splitAndTrim(getEnv("SWAGGER_SCHEMES", "")),
		},
		Tracing: TracingConfig{
			Enabled:     toBool(getEnv("TRACING_ENABLED", "false")),
			Provider:    getEnv("TRACING_PROVIDER", "opentracing"),
			Endpoint:    getEnv("TRACING_ENDPOINT", ""),
			ServiceName: getEnv("TRACING_SERVICE_NAME", "emplacc-api"),
			SampleRate:  sampleRate,
		},
		Keycloak: KeycloakConfig{
			URL:                  getEnv("KEYCLOAK_URL", ""),
			Realm:                getEnv("KEYCLOAK_REALM", ""),
			ClientID:             getEnv("KEYCLOAK_CLIENT_ID", ""),
			ClientSecret:         getEnv("KEYCLOAK_CLIENT_SECRET", ""),
			BackendClientID:      getEnv("KEYCLOAK_BACKEND_CLIENT_ID", ""),
			BackendClientSecret:  getEnv("KEYCLOAK_BACKEND_CLIENT_SECRET", ""),
			TokenExchangeEnabled: toBool(getEnv("KEYCLOAK_TOKEN_EXCHANGE_ENABLED", "false")),
		},
		Minio: MinioConfig{
			Endpoint:     getEnv("MINIO_ENDPOINT", ""),
			AccessKey:    getEnv("MINIO_ACCESS_KEY", ""),
			SecretKey:    getEnv("MINIO_SECRET_KEY", ""),
			UseSSL:       toBool(getEnv("MINIO_USE_SSL", "false")),
			Region:       getEnv("MINIO_REGION", ""),
			AvatarBucket: getEnv("MINIO_AVATAR_BUCKET", "avatars"),
			ReportBucket: getEnv("MINIO_REPORT_BUCKET", "reports"),
		},
		Pagination: PaginationConfig{
			DefaultLimit: defaultLimit,
			MaxLimit:     maxLimit,
		},
		RateLimit: RateLimitConfig{
			User: RateLimitRule{
				Requests: userRateRequests,
				Window:   userRateWindow,
			},
			IP: RateLimitRule{
				Requests: ipRateRequests,
				Window:   ipRateWindow,
			},
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func toInt(value string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(value))
}

func toDuration(value string) (time.Duration, error) {
	return time.ParseDuration(strings.TrimSpace(value))
}

func toBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

func splitAndTrim(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
