package service

import (
	"context"
	"io"
	"time"

	"emplacc-api/internal/app/ports"
	pb "emplacc-api/pkg/pb/v1"
)

type mcpService struct {
	repo ports.MCPRepository
}

func NewMCPService(repo ports.MCPRepository) ports.MCPService {
	return &mcpService{repo: repo}
}

func (s *mcpService) ImproveReport(ctx context.Context, input ports.ImproveReportInput) (*ports.ImproveReportResult, error) {
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

func (s *mcpService) StreamReport(ctx context.Context, input ports.ImproveReportInput) (*ports.MCPStream, error) {
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

func copyMeta(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(meta))
	for k, v := range meta {
		out[k] = v
	}
	return out
}

func withOptionalTimeout(ctx context.Context, timeoutMs *int64) (context.Context, context.CancelFunc) {
	if timeoutMs != nil && *timeoutMs > 0 {
		return context.WithTimeout(ctx, time.Duration(*timeoutMs)*time.Millisecond)
	}
	return context.WithCancel(ctx)
}

func convertProcessEvent(ev *pb.ProcessTaskEvent) ports.MCPStreamEvent {
	if status := ev.GetStatus(); status != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventStatus,
			Status: &ports.MCPStatusEvent{
				State:    status.GetState(),
				Message:  status.GetMessage(),
				Progress: status.GetProgress(),
			},
		}
	}

	if chunk := ev.GetChunk(); chunk != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventChunk,
			Chunk: &ports.MCPChunkEvent{
				Data:  chunk.GetData(),
				Index: chunk.GetIndex(),
			},
		}
	}

	if final := ev.GetFinal(); final != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventFinal,
			Final: &ports.MCPFinalEvent{
				Result:      final.GetResult(),
				ContentType: final.GetContentType(),
			},
		}
	}

	if err := ev.GetError(); err != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventError,
			Error: &ports.MCPErrorEvent{
				Code:    err.GetCode(),
				Message: err.GetMessage(),
			},
		}
	}

	return ports.MCPStreamEvent{Type: ports.MCPStreamEventStatus}
}
