package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// presignedURLExpiry — максимум по S3 Signature V4: 7 дней (604800 сек).
// Для долгосрочного хранения используй /media/refresh endpoint.
const presignedURLExpiry = 7 * 24 * time.Hour

type StorageService interface {
	UploadFile(file multipart.File, header *multipart.FileHeader, folder string) (string, string, error)
	RefreshURL(objectPath string) (string, error)
}

type storageService struct {
	client    *minio.Client
	bucket    string
	publicURL string
	useSSL    bool
}

func NewStorageService() (StorageService, error) {
	endpoint  := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket    := os.Getenv("S3_BUCKET")
	publicURL := os.Getenv("S3_PUBLIC_URL")
	useSSL    := os.Getenv("S3_USE_SSL") == "true"

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY, S3_BUCKET must be set")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("make bucket: %w", err)
		}
	}

	// Удаляем публичную политику bucket — доступ только через presigned URLs
	if err := client.SetBucketPolicy(ctx, bucket, ""); err != nil {
		// Не фатально, логируем но продолжаем
		_ = err
	}

	if publicURL == "" {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		publicURL = fmt.Sprintf("%s://%s", scheme, endpoint)
	}

	return &storageService{client: client, bucket: bucket, publicURL: strings.TrimRight(publicURL, "/"), useSSL: useSSL}, nil
}

func (s *storageService) RefreshURL(objectPath string) (string, error) {
	presigned, err := s.client.PresignedGetObject(
		context.Background(),
		s.bucket,
		objectPath,
		presignedURLExpiry,
		url.Values{},
	)
	if err != nil {
		return "", fmt.Errorf("presign refresh: %w", err)
	}
	result := presigned.String()
	// Заменяем только scheme+host, без пути bucket (он уже есть в presigned URL)
	internalScheme := "http"
	if s.useSSL {
		internalScheme = "https"
	}
	internalBase := fmt.Sprintf("%s://%s", internalScheme, s.client.EndpointURL().Host)
	if s.publicURL != "" && strings.HasPrefix(result, internalBase) {
		// Извлекаем только scheme+host из publicURL (без пути)
		publicHost := s.publicURL
		if idx := strings.Index(publicHost[8:], "/"); idx >= 0 {
			publicHost = publicHost[:8+idx] // обрезаем путь, оставляем только https://host
		}
		result = strings.Replace(result, internalBase, publicHost, 1)
	}
	return result, nil
}

func (s *storageService) UploadFile(file multipart.File, header *multipart.FileHeader, folder string) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".bin"
	}
	objectName := fmt.Sprintf("%s/%s_%d%s", folder, uuid.New().String(), time.Now().UnixMilli(), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(
		context.Background(),
		s.bucket,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", "", fmt.Errorf("put object: %w", err)
	}

	presignedURL, err := s.RefreshURL(objectName)
	if err != nil {
		return "", "", err
	}

	return presignedURL, objectName, nil
}
