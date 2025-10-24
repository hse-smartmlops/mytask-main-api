package mcp

import (
	"context"

	"emplacc-api/internal/app/ports"
	pb "emplacc-api/pkg/pb/v1"
)

type repository struct {
	client pb.MCPServiceClient
}

func New(
	client pb.MCPServiceClient,
) ports.MCPRepository {
	return &repository{
		client: client,
	}
}

func (r *repository) ProcessTask(ctx context.Context, req *pb.ProcessTaskRequest) (*pb.ProcessTaskResponse, error) {
	return r.client.ProcessTask(ctx, req)
}

func (r *repository) StreamProcessTask(ctx context.Context, req *pb.ProcessTaskRequest) (pb.MCPService_StreamProcessTaskClient, error) {
	return r.client.StreamProcessTask(ctx, req)
}
