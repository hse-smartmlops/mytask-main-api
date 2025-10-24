package service

import (
	"context"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type statusService struct {
	repo ports.StatusRepository
	config *config.PaginationConfig
}

func NewStatusService(repo ports.StatusRepository, config *config.PaginationConfig) ports.StatusService {
	return &statusService{repo: repo, config: config}
}

func (s *statusService) ListStatuses(ctx context.Context, p ports.PaginationParams) (*ports.Page[models.Status], error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return s.repo.ListStatuses(ctx, p)
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

func (s *statusService) CreateStatus(ctx context.Context, input ports.CreateStatusInput) (*models.Status, error) {
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

func (s *statusService) UpdateStatus(ctx context.Context, id uuid.UUID, input ports.UpdateStatusInput) (*models.Status, error) {
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

var _ ports.StatusService = (*statusService)(nil)
