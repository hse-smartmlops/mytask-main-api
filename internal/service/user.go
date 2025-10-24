package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type userService struct {
	repo    ports.UserRepository
	storage ports.ObjectStorage
	cfg     config.MinioConfig
	config  *config.PaginationConfig
}

func NewUserService(repo ports.UserRepository, storage ports.ObjectStorage, cfg config.MinioConfig, config *config.PaginationConfig) ports.UserService {
	return &userService{repo: repo, storage: storage, cfg: cfg, config: config}
}

func (s *userService) ListUsers(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.User], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListUsers(ctx, params)
}

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (s *userService) CreateUser(ctx context.Context, input ports.CreateUserInput) (*models.User, error) {
	email := strings.TrimSpace(input.Email)
	if email == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()

	user := &models.User{
		ID:            uuid.New(),
		Email:         email,
		IsActive:      boolValue(input.IsActive, true),
		TgID:          stringValue(input.TgID),
		TgUserID:      int64Value(input.TgUserID),
		Profession:    stringValue(input.Profession),
		EmailVerified: boolValue(input.EmailVerified, false),
		FirstName:     stringValue(input.FirstName),
		LastName:      stringValue(input.LastName),
		LastLogin:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) SaveAvatar(ctx context.Context, id uuid.UUID, input ports.SaveUserAvatarInput) (*models.User, error) {
	if s.storage == nil {
		return nil, errors.New("object storage is not configured")
	}

	if len(input.Data) == 0 {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resized, contentType, err := prepareAvatar(input.Data)
	if err != nil {
		return nil, err
	}

	bucket := strings.TrimSpace(s.cfg.AvatarBucket)
	if bucket == "" {
		bucket = "avatars"
	}
	object := fmt.Sprintf("%s.png", id.String())

	if err := s.storage.Upload(ctx, bucket, object, resized, contentType, map[string]string{"user_id": id.String()}); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateUserAvatar(ctx, id, fmt.Sprintf("%s/%s", bucket, object)); err != nil {
		return nil, err
	}

	user.AvatarPath = fmt.Sprintf("%s/%s", bucket, object)

	return user, nil
}

func (s *userService) DeleteAvatar(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.AvatarPath != "" && s.storage != nil {
		parts := strings.SplitN(user.AvatarPath, "/", 2)
		if len(parts) == 2 {
			_ = s.storage.Delete(ctx, parts[0], parts[1])
		}
	}

	if err := s.repo.ClearUserAvatar(ctx, id); err != nil {
		return nil, err
	}

	user.AvatarPath = ""
	return user, nil
}

func (s *userService) GetAvatar(ctx context.Context, id uuid.UUID) (*ports.UserAvatarFile, error) {
	if s.storage == nil {
		return nil, errors.New("object storage is not configured")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(user.AvatarPath) == "" {
		return nil, domain.ErrNotFound
	}

	parts := strings.SplitN(user.AvatarPath, "/", 2)
	if len(parts) != 2 {
		return nil, domain.ErrNotFound
	}

	data, contentType, err := s.storage.Get(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "image/png"
	}

	return &ports.UserAvatarFile{
		FileName:    fmt.Sprintf("%s.png", id.String()),
		ContentType: contentType,
		Data:        data,
	}, nil
}
