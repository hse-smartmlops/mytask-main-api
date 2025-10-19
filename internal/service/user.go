package service

import (
	"context"
	"strings"
	"time"

	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type userService struct {
	repo app.UserRepository
}

func NewUserService(repo app.UserRepository) app.UserService {
	return &userService{repo: repo}
}

func (s *userService) ListUsers(ctx context.Context, params app.PaginationParams) (*app.Page[models.User], error) {
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

func (s *userService) CreateUser(ctx context.Context, input app.CreateUserInput) (*models.User, error) {
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

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
