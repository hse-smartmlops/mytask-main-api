package client

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "emplacc-api/internal/grpc/gen/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LLMOverrides — параметры, которые переопределяют дефолты LLM-сервиса
type LLMOverrides struct {
	Model        string
	URL          string
	SystemPrompt string
}

type LLMClient struct {
	conn   *grpc.ClientConn
	client v1.MCPServiceClient
}

func NewLLMClient(addr string) (*LLMClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return &LLMClient{conn: conn, client: v1.NewMCPServiceClient(conn)}, nil
}

func (c *LLMClient) Close() error { return c.conn.Close() }

// buildMeta строит map для передачи оверрайдов в LLM-сервис
func buildMeta(overrides *LLMOverrides) map[string]string {
	meta := map[string]string{}
	if overrides == nil {
		return meta
	}
	if overrides.Model != "" {
		meta["model"] = overrides.Model
	}
	if overrides.URL != "" {
		meta["url"] = overrides.URL
	}
	if overrides.SystemPrompt != "" {
		meta["system_prompt"] = overrides.SystemPrompt
	}
	return meta
}

// ProcessTaskWithLLM — синхронный вызов с опциональными оверрайдами настроек
func (c *LLMClient) ProcessTaskWithLLM(ctx context.Context, taskDescription, userText, taskId string, overrides *LLMOverrides) (string, error) {
	req := &v1.ProcessTaskRequest{
		Description: taskDescription,
		Text:        userText,
		TaskId:      taskId,
		ContentType: "text/plain",
		TimeoutMs:   120000,
		Meta:        buildMeta(overrides),
	}

	resp, err := c.client.ProcessTask(ctx, req)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		return "", err
	}
	return resp.Result, nil
}

// ProcessTaskWithLLMStream — потоковый вызов с опциональными оверрайдами
func (c *LLMClient) ProcessTaskWithLLMStream(ctx context.Context, taskDescription, userText, taskId string, overrides *LLMOverrides) (string, error) {
	req := &v1.ProcessTaskRequest{
		Description: taskDescription,
		Text:        userText,
		TaskId:      taskId,
		ContentType: "text/plain",
		TimeoutMs:   120000,
		Meta:        buildMeta(overrides),
	}

	stream, err := c.client.StreamProcessTask(ctx, req)
	if err != nil {
		log.Printf("gRPC stream call failed: %v", err)
		return "", err
	}

	var result string
	for {
		event, err := stream.Recv()
		if err != nil {
			return "", err
		}
		switch e := event.Event.(type) {
		case *v1.ProcessTaskEvent_Chunk:
			result += e.Chunk.Data
		case *v1.ProcessTaskEvent_Final:
			return e.Final.Result, nil
		case *v1.ProcessTaskEvent_Status:
			log.Printf("LLM status: %s %s", e.Status.State, e.Status.Message)
		case *v1.ProcessTaskEvent_Error:
			return "", fmt.Errorf("LLM error %d: %s", e.Error.Code, e.Error.Message)
		}
	}
}
