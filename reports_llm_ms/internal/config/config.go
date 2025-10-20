package config

import (
	"os"
	"strconv"
)

// Config содержит все конфигурации приложения
type Config struct {
    GRPCHost     string
    GRPCPort     int
    WebUIURL     string
    WebUIToken   string
    WebUIModel   string
    WebUIInsecure bool
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
    return &Config{
        GRPCHost:     getEnv("GRPC_HOST", "0.0.0.0"),
        GRPCPort:     getEnvAsInt("GRPC_PORT", 50051),
        WebUIURL:     getEnv("WEBUI_URL", "http://10.14.49.32:3000/api/chat/completions"),
        WebUIToken:   getEnv("WEBUI_TOKEN", "sk-18b9c9ead8ed4b51be60416a48e44974"),
        WebUIModel:   getEnv("WEBUI_MODEL", "gemma3:12b"),
        WebUIInsecure: getEnvAsBool("WEBUI_INSECURE", false),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}