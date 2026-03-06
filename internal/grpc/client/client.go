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

	client := v1.NewMCPServiceClient(conn)
	
	return &LLMClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *LLMClient) Close() error {
	return c.conn.Close()
}

func (c *LLMClient) ProcessTaskWithLLM(ctx context.Context, taskDescription, userText string, taskId string) (string, error) {
	req := &v1.ProcessTaskRequest{
		Description: taskDescription,
		Text:        userText,
		TaskId:      taskId,  
		ContentType: "text/plain",      
		TimeoutMs:   120000,           
	}

	resp, err := c.client.ProcessTask(ctx, req)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		return "", err
	}

	return resp.Result, nil
}

// ProcessTaskWithLLMStream - для потоковой обработки (опционально)
func (c *LLMClient) ProcessTaskWithLLMStream(ctx context.Context, taskDescription, userText string, taskId string) (string, error) {
	req := &v1.ProcessTaskRequest{
		Description: taskDescription,
		Text:        userText,
		TaskId:      taskId,
		ContentType: "text/plain",
		TimeoutMs:   120000,
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
			log.Printf("Stream receive error: %v", err)
			return "", err
		}

		switch e := event.Event.(type) {
		case *v1.ProcessTaskEvent_Chunk:
			// Собираем чанки
			result += e.Chunk.Data
			log.Printf("Received chunk %d: %s", e.Chunk.Index, e.Chunk.Data[:min(50, len(e.Chunk.Data))])
		
		case *v1.ProcessTaskEvent_Final:
			// Финальный результат
			result = e.Final.Result
			log.Printf("Received final result, length: %d", len(result))
			return result, nil
		
		case *v1.ProcessTaskEvent_Status:
			// Логируем статус
			log.Printf("Status: %s - %s (progress: %d%%)", 
				e.Status.State, e.Status.Message, e.Status.Progress)
		
		case *v1.ProcessTaskEvent_Error:
			// Обрабатываем ошибку
			log.Printf("Error from LLM service: %d - %s", e.Error.Code, e.Error.Message)
			return "", fmt.Errorf("LLM error %d: %s", e.Error.Code, e.Error.Message)
		}
	}
}


func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}