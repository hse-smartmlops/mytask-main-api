package main

import (
	"os"
	"os/signal"
	"syscall"

	"reports_llm_ms/internal/config"
	"reports_llm_ms/internal/handler"
	"reports_llm_ms/internal/service"
	"reports_llm_ms/pkg/grpc"
	"reports_llm_ms/pkg/logger"
)

func main() {
    // Инициализация логгера
    log := logger.New()

    log.Info("=== STARTING LLM FOR DESCRIPTION SERVICE ===")
    log.Info("Working directory: %s", getWorkingDirectory())

    // Загрузка конфигурации
    cfg := config.Load()
    logConfig(cfg, log)

    // Инициализация сервисов
    llmService := service.NewWebUILLMService(cfg)
    log.Info("LLM service initialized")

    // Инициализация gRPC handler
    grpcHandler := handler.NewGRPCHandler(llmService, log)
    log.Info("gRPC handler initialized")

    // Настройка и запуск gRPC сервера
    grpcConfig := &grpc.Config{
        Host: cfg.GRPCHost,
        Port: cfg.GRPCPort,
    }

    grpcServer := grpc.NewServer(grpcConfig)
    grpcServer.RegisterService(grpcHandler)

    // Канал для graceful shutdown
    done := make(chan bool, 1)
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Запуск сервера в горутине
    go func() {
        log.Info("Starting gRPC server on %s:%d", cfg.GRPCHost, cfg.GRPCPort)
        
        if err := grpcServer.Start(); err != nil {
            log.Error("Failed to start gRPC server: %v", err)
            done <- true
        }
    }()

    // Ожидание сигнала завершения
    select {
    case <-quit:
        log.Info("Received shutdown signal")
    case <-done:
        log.Info("Server stopped")
    }

    log.Info("Shutting down server...")
    grpcServer.Stop()
    log.Info("Server shutdown complete")
}

// getWorkingDirectory возвращает текущую рабочую директорию
func getWorkingDirectory() string {
    wd, err := os.Getwd()
    if err != nil {
        return "unknown"
    }
    return wd
}

// logConfig логирует конфигурацию (без чувствительных данных)
func logConfig(cfg *config.Config, log *logger.Logger) {
    log.Info("Configuration loaded")
    log.Debug("GRPC Host: %s", cfg.GRPCHost)
    log.Debug("GRPC Port: %d", cfg.GRPCPort)
    log.Debug("WebUI URL: %s", cfg.WebUIURL)
    log.Debug("WebUI Model: %s", cfg.WebUIModel)
    log.Debug("WebUI Insecure: %t", cfg.WebUIInsecure)
    
    // Токен не логируем полностью в целях безопасности
    if cfg.WebUIToken != "" {
        log.Debug("WebUI Token: %s...", cfg.WebUIToken[:min(8, len(cfg.WebUIToken))])
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}