package handler

import (
	"context"
	"fmt"

	"reports_llm_ms/gen"
	"reports_llm_ms/internal/domain"
	"reports_llm_ms/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler обрабатывает gRPC запросы
type GRPCHandler struct {
    gen.UnimplementedEmplaccMcpServer
    llmService service.LLMService
    logger     Logger
}

// Logger интерфейс для логирования
type Logger interface {
    Info(msg string, fields ...any)
    Error(msg string, fields ...any)
    Debug(msg string, fields ...any)
}

// NewGRPCHandler создает новый экземпляр gRPC handler
func NewGRPCHandler(llmService service.LLMService, logger Logger) *GRPCHandler {
    return &GRPCHandler{
        llmService: llmService,
        logger:     logger,
    }
}

// ProcessTask обрабатывает синхронный gRPC запрос
func (h *GRPCHandler) ProcessTask(ctx context.Context, req *gen.TwoTextRequest) (*gen.TwoTextReply, error) {
    h.logger.Info("ProcessTask called", 
        "description_length", len(req.GetDescription()),
        "text_length", len(req.GetText()))

    if err := validateRequest(req); err != nil {
        h.logger.Error("Validation failed", "error", err)
        return nil, err
    }

    llmReq := &domain.LLMRequest{
        Description: req.GetDescription(),
        Text:        req.GetText(),
    }

    h.logger.Debug("Calling LLM service")
    response, err := h.llmService.Process(ctx, llmReq)
    if err != nil {
        h.logger.Error("LLM service failed", "error", err)
        return nil, status.Error(codes.Internal, fmt.Sprintf("LLM error: %v", err))
    }

    h.logger.Info("LLM service completed", 
        "result_length", len(response.Result))

    return &gen.TwoTextReply{Result: response.Result}, nil
}

// StreamProcessTask обрабатывает потоковый gRPC запрос
func (h *GRPCHandler) StreamProcessTask(req *gen.TwoTextRequest, stream gen.EmplaccMcp_StreamProcessTaskServer) error {
    h.logger.Info("StreamProcessTask called",
        "description_length", len(req.GetDescription()),
        "text_length", len(req.GetText()))

    if err := validateRequest(req); err != nil {
        h.logger.Error("Validation failed", "error", err)
        return err
    }

    llmReq := &domain.LLMRequest{
        Description: req.GetDescription(),
        Text:        req.GetText(),
    }

    h.logger.Debug("Starting LLM stream")
    resultChan, errChan := h.llmService.ProcessStream(stream.Context(), llmReq)
    chunkCount := 0

    for {
        select {
        case result, ok := <-resultChan:
            if !ok {
                h.logger.Info("Stream completed", "total_chunks", chunkCount)
                return nil
            }
            
            chunkCount++
            if chunkCount <= 3 {
                h.logger.Debug("Stream chunk", 
                    "chunk_number", chunkCount,
                    "preview", previewString(result.Result, 50))
            }

            if err := stream.Send(&gen.TwoTextReply{Result: result.Result}); err != nil {
                h.logger.Error("Failed to send chunk", "error", err)
                return err
            }

        case err := <-errChan:
            if err != nil {
                h.logger.Error("LLM stream failed", "error", err)
                return status.Error(codes.Internal, fmt.Sprintf("LLM stream error: %v", err))
            }
            return nil
            
        case <-stream.Context().Done():
            h.logger.Info("Stream canceled by client")
            return nil
        }
    }
}

// validateRequest проверяет валидность входящего запроса
func validateRequest(req *gen.TwoTextRequest) error {
    if req.GetDescription() == "" {
        return status.Error(codes.InvalidArgument, "description is empty")
    }
    if req.GetText() == "" {
        return status.Error(codes.InvalidArgument, "text is empty")
    }
    return nil
}

// previewString обрезает строку для логов
func previewString(s string, length int) string {
    if len(s) <= length {
        return s
    }
    return s[:length] + "..."
}