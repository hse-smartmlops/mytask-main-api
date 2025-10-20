package grpc

import (
	"fmt"
	"net"

	"reports_llm_ms/gen"
	"reports_llm_ms/internal/handler"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server представляет gRPC сервер
type Server struct {
    server *grpc.Server
    config *Config
}

// Config содержит конфигурацию gRPC сервера
type Config struct {
    Host string
    Port int
}

// NewServer создает новый gRPC сервер
func NewServer(cfg *Config) *Server {
    grpcServer := grpc.NewServer(
        grpc.MaxSendMsgSize(64*1024*1024),
        grpc.MaxRecvMsgSize(64*1024*1024),
    )

    return &Server{
        server: grpcServer,
        config: cfg,
    }
}

// RegisterService регистрирует gRPC handler
func (s *Server) RegisterService(handler *handler.GRPCHandler) {
    gen.RegisterEmplaccMcpServer(s.server, handler)
    
    // Включаем reflection для инструментов вроде grpcurl
    reflection.Register(s.server)
}

// Start запускает gRPC сервер
func (s *Server) Start() error {
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    
    lis, err := net.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("failed to listen: %w", err)
    }

    return s.server.Serve(lis)
}

// Stop gracefully останавливает gRPC сервер
func (s *Server) Stop() {
    s.server.GracefulStop()
}