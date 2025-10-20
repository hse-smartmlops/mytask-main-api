package client

import (
	"context"
	"log"
	"time"

	pb "emplacc-api/internal/grpc/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LLMClient struct {
	conn   *grpc.ClientConn
	client pb.EmplaccMcpClient
}

func NewLLMClient(addr string) (*LLMClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewEmplaccMcpClient(conn)
	
	return &LLMClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *LLMClient) Close() error {
	return c.conn.Close()
}

// ProcessTaskWithLLM - основная функция для обработки текста через LLM
func (c *LLMClient) ProcessTaskWithLLM(ctx context.Context, taskDescription, userText string) (string, error) {
	req := &pb.TwoTextRequest{
		Description: taskDescription,
		Text:        userText,
	}

	resp, err := c.client.ProcessTask(ctx, req)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		return "", err
	}

	return resp.Result, nil
}