package service

import (
	"context"
	"io"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	pb "emplacc-api/pkg/pb/v1"
)

type mcpService struct {
	repo   ports.MCPRepository
	config *config.PaginationConfig
}

func NewMCPService(
	repo ports.MCPRepository,
	config *config.PaginationConfig,
) ports.MCPService {
	return &mcpService{
		repo:   repo,
		config: config,
	}
}

func (s *mcpService) ImproveReport(
	ctx context.Context,
	input ports.ImproveReportInput,
) (*ports.ImproveReportResult, error) {
	ctx, cancel := withOptionalTimeout(ctx, input.TimeoutMs)
	defer cancel()

	meta := copyMeta(input.Meta)

	req := &pb.ProcessTaskRequest{
		Description: input.Description,
		Text:        input.UserText,
		TaskId:      input.TaskID,
		Meta:        meta,
		ContentType: input.ContentType,
	}
	if input.TimeoutMs != nil && *input.TimeoutMs > 0 {
		req.TimeoutMs = *input.TimeoutMs
	}

	start := time.Now()
	resp, err := s.repo.ProcessTask(ctx, req)
	if err != nil {
		return nil, err
	}

	return &ports.ImproveReportResult{
		ImprovedText: resp.GetResult(),
		ContentType:  resp.GetContentType(),
		Meta:         meta,
		Duration:     time.Since(start),
	}, nil
}

func (s *mcpService) StreamReport(
	ctx context.Context,
	input ports.ImproveReportInput,
) (*ports.MCPStream, error) {
	streamCtx, cancel := withOptionalTimeout(ctx, input.TimeoutMs)

	meta := copyMeta(input.Meta)

	req := &pb.ProcessTaskRequest{
		Description: input.Description,
		Text:        input.UserText,
		TaskId:      input.TaskID,
		Meta:        meta,
		ContentType: input.ContentType,
	}
	if input.TimeoutMs != nil && *input.TimeoutMs > 0 {
		req.TimeoutMs = *input.TimeoutMs
	}

	client, err := s.repo.StreamProcessTask(streamCtx, req)
	if err != nil {
		cancel()
		return nil, err
	}

	events := make(chan ports.MCPStreamEvent)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)
		defer cancel()

		for {
			msg, err := client.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errs <- err
				return
			}

			events <- convertProcessEvent(msg)

			if msg.GetFinal() != nil {
				return
			}
		}
	}()

	return &ports.MCPStream{
		Events: events,
		Errors: errs,
		Close:  cancel,
	}, nil
}
