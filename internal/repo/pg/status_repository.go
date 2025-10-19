package pg

import (
	"context"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusRepository struct {
	db *gorm.DB
}

func NewStatusRepository(db *gorm.DB) *StatusRepository {
	return &StatusRepository{db: db}
}

func (r *StatusRepository) ListStatuses(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.Status], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Status{}).Where("statuses.deleted = FALSE OR statuses.deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var statuses []models.Status
	offset := (params.Page - 1) * params.PageSize
	if err := base.Order("statuses.created_at DESC NULLS LAST").
		Preload("Board").
		Limit(params.PageSize).
		Offset(offset).
		Find(&statuses).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Status]{
		Items:      statuses,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *StatusRepository) GetStatusByID(ctx context.Context, id uuid.UUID) (*models.Status, error) {
	var status models.Status
	err := r.db.WithContext(ctx).
		Model(&models.Status{}).
		Preload("Board").
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		First(&status).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &status, nil
}

func (r *StatusRepository) ListStatusesByBoard(ctx context.Context, boardID uuid.UUID) ([]models.Status, error) {
	var statuses []models.Status
	if err := r.db.WithContext(ctx).
		Model(&models.Status{}).
		Where("board_id = ? AND (deleted = FALSE OR deleted IS NULL)", boardID).
		Order("sort_order ASC NULLS LAST").
		Find(&statuses).Error; err != nil {
		return nil, err
	}
	return statuses, nil
}

func (r *StatusRepository) CreateStatus(ctx context.Context, status *models.Status) error {
	return r.db.WithContext(ctx).Create(status).Error
}

func (r *StatusRepository) UpdateStatus(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Status, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Status{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetStatusByID(ctx, id)
}

func (r *StatusRepository) SoftDeleteStatus(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Status{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(map[string]interface{}{
			"deleted":    &deleted,
			"updated_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var _ ports.StatusRepository = (*StatusRepository)(nil)
