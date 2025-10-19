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

type statusService struct {
	repo app.StatusRepository
}

func NewStatusService(repo app.StatusRepository) app.StatusService {
	return &statusService{repo: repo}
}

func (s *statusService) ListStatuses(ctx context.Context, params app.PaginationParams) (*app.Page[models.Status], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListStatuses(ctx, params)
}

func (s *statusService) GetStatus(ctx context.Context, id uuid.UUID) (*models.Status, error) {
	status, err := s.repo.GetStatusByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, domain.ErrNotFound
	}
	return status, nil
}

func (s *statusService) ListStatusesByBoard(ctx context.Context, boardID uuid.UUID) ([]models.Status, error) {
	return s.repo.ListStatusesByBoard(ctx, boardID)
}

func (s *statusService) CreateStatus(ctx context.Context, input app.CreateStatusInput) (*models.Status, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	status := &models.Status{
		ID:        uuid.New(),
		BoardID:   input.BoardID,
		Name:      stringPtr(name),
		Key:       cloneStringPtr(input.Key),
		Color:     cloneStringPtr(input.Color),
		IsDefault: input.IsDefault,
		IsActive:  input.IsActive,
		IsOpen:    input.IsOpen,
		SortOrder: input.SortOrder,
		CreatedAt: &now,
		UpdatedAt: &now,
		Deleted:   &deleted,
	}

	if err := s.repo.CreateStatus(ctx, status); err != nil {
		return nil, err
	}

	return s.repo.GetStatusByID(ctx, status.ID)
}

func (s *statusService) UpdateStatus(ctx context.Context, id uuid.UUID, input app.UpdateStatusInput) (*models.Status, error) {
	updates := make(map[string]interface{})

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.ErrInvalidInput
		}
		updates["name"] = name
	}
	if input.Key != nil {
		updates["key"] = cloneStringPtr(input.Key)
	}
	if input.Color != nil {
		updates["color"] = cloneStringPtr(input.Color)
	}
	if input.IsDefault != nil {
		updates["is_default"] = input.IsDefault
	}
	if input.IsActive != nil {
		updates["is_active"] = input.IsActive
	}
	if input.IsOpen != nil {
		updates["is_open"] = input.IsOpen
	}
	if input.SortOrder != nil {
		updates["sort_order"] = input.SortOrder
	}

	if len(updates) == 0 {
		return s.repo.GetStatusByID(ctx, id)
	}

	return s.repo.UpdateStatus(ctx, id, updates)
}

func (s *statusService) DeleteStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteStatus(ctx, id)
}

var _ app.StatusService = (*statusService)(nil)
