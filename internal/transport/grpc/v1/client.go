package v1

import (
	"context"
	"errors"
	"time"

	pb "emplacc-api/pkg/pb/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrMissingAddress = errors.New("mcp grpc address is required")

type Config struct {
	Address string
	Timeout time.Duration
	UseTLS  bool
	TLS     credentials.TransportCredentials
}

type Client struct {
	conn *grpc.ClientConn
	api  pb.MCPServiceClient
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, ErrMissingAddress
	}

	opts := []grpc.DialOption{}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}

	if cfg.UseTLS {
		creds := cfg.TLS
		if creds == nil {
			creds = credentials.NewClientTLSFromCert(nil, "")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	_, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := grpc.NewClient(cfg.Address, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		api:  pb.NewMCPServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) API() pb.MCPServiceClient {
	return c.api
}

func (c *Client) ProcessTask(ctx context.Context, req *pb.ProcessTaskRequest, opts ...grpc.CallOption) (*pb.ProcessTaskResponse, error) {
	return c.api.ProcessTask(ctx, req, opts...)
}

func (c *Client) StreamProcessTask(ctx context.Context, req *pb.ProcessTaskRequest, opts ...grpc.CallOption) (pb.MCPService_StreamProcessTaskClient, error) {
	return c.api.StreamProcessTask(ctx, req, opts...)
}
