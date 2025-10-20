package minio

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client      *minio.Client
	cfg         config.MinioConfig
	bucketCache map[string]struct{}
	mu          sync.Mutex
}

func NewStorage(cfg config.MinioConfig) (*Storage, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, nil
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	return &Storage{
		client:      client,
		cfg:         cfg,
		bucketCache: make(map[string]struct{}),
	}, nil
}

func (s *Storage) Upload(ctx context.Context, bucket, object string, data []byte, contentType string, metadata map[string]string) error {
	if err := s.ensureBucket(ctx, bucket); err != nil {
		return err
	}

	reader := bytes.NewReader(data)
	options := minio.PutObjectOptions{ContentType: contentType}
	if len(metadata) > 0 {
		options.UserMetadata = metadata
	}

	_, err := s.client.PutObject(ctx, bucket, object, reader, int64(len(data)), options)
	return err
}

func (s *Storage) Delete(ctx context.Context, bucket, object string) error {
	return s.client.RemoveObject(ctx, bucket, object, minio.RemoveObjectOptions{})
}

func (s *Storage) Get(ctx context.Context, bucket, object string) ([]byte, string, error) {
	obj, err := s.client.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		return nil, "", err
	}

	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", err
	}

	return content, strings.TrimSpace(info.ContentType), nil
}

func (s *Storage) ensureBucket(ctx context.Context, bucket string) error {
	s.mu.Lock()
	if _, ok := s.bucketCache[bucket]; ok {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: s.cfg.Region}); err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.bucketCache[bucket] = struct{}{}
	s.mu.Unlock()

	return nil
}

var _ ports.ObjectStorage = (*Storage)(nil)
