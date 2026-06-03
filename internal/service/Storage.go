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
	// FreshAvatarURL returns a fresh presigned URL for the stored value.
	// Accepts either an object path ("avatars/uuid.jpg") or a legacy
	// presigned URL containing X-Amz-Signature (auto-extracts the path).
	// Returns "" on error or empty input.
	FreshAvatarURL(storedValue string) string
}

type storageService struct {
	client        *minio.Client // внутренний endpoint для загрузки объектов
	presignClient *minio.Client // публичный endpoint для генерации presigned URL
	bucket        string
	useSSL        bool
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

	// Второй клиент для presigned URLs — использует публичный endpoint
	// чтобы подпись совпадала с hostname в URL (S3 подписывает Host header)
	var presignClient *minio.Client
	if publicURL != "" {
		pubEndpoint := strings.TrimRight(publicURL, "/")
		// Извлекаем host из publicURL (убираем scheme и путь)
		pubEndpoint = strings.TrimPrefix(pubEndpoint, "https://")
		pubEndpoint = strings.TrimPrefix(pubEndpoint, "http://")
		if idx := strings.Index(pubEndpoint, "/"); idx >= 0 {
			pubEndpoint = pubEndpoint[:idx]
		}
		pubSSL := strings.HasPrefix(publicURL, "https://")
		presignClient, _ = minio.New(pubEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: pubSSL,
		})
	}
	if presignClient == nil {
		presignClient = client
	}

	return &storageService{client: client, presignClient: presignClient, bucket: bucket, useSSL: useSSL}, nil
}

func (s *storageService) RefreshURL(objectPath string) (string, error) {
	// Используем presignClient — он настроен на публичный endpoint,
	// поэтому подпись будет валидна для публичного URL
	presigned, err := s.presignClient.PresignedGetObject(
		context.Background(),
		s.bucket,
		objectPath,
		presignedURLExpiry,
		url.Values{},
	)
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}
	return presigned.String(), nil
}

func (s *storageService) FreshAvatarURL(storedValue string) string {
	if storedValue == "" {
		return ""
	}
	objectPath := storedValue
	// Если это старый presigned URL (содержит X-Amz-Signature) — извлекаем путь
	if strings.Contains(storedValue, "X-Amz-Signature") {
		u, err := url.Parse(storedValue)
		if err != nil {
			return ""
		}
		// u.Path выглядит как /bucket/folder/file.ext — убираем bucket (первый сегмент)
		parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
		if len(parts) < 2 {
			return ""
		}
		objectPath = parts[1]
	}
	fresh, err := s.RefreshURL(objectPath)
	if err != nil {
		return ""
	}
	return fresh
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
