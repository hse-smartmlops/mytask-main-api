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

type ForumMessageRepository struct {
	db *gorm.DB
}

func NewForumMessageRepository(
	db *gorm.DB,
) *ForumMessageRepository {
	return &ForumMessageRepository{
		db: db,
	}
}

func (r *ForumMessageRepository) ListMessages(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.ForumMessage], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE OR deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var messages []models.ForumMessage
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("created_at DESC NULLS LAST").
		Preload("Problem").
		Preload("User").
		Limit(p.PageSize).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.ForumMessage]{
		Items:      messages,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *ForumMessageRepository) GetMessageByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.ForumMessage, error) {
	var message models.ForumMessage
	err := r.db.WithContext(ctx).
		Model(&models.ForumMessage{}).
		Preload("Problem").
		Preload("User").
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &message, nil
}

func (r *ForumMessageRepository) ListMessagesByProblem(
	ctx context.Context,
	problemID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.ForumMessage], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).
		Model(&models.ForumMessage{}).
		Where("problem_id = ? AND (deleted = FALSE OR deleted IS NULL)", problemID)
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var messages []models.ForumMessage
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("created_at DESC NULLS LAST").
		Preload("Problem").
		Preload("User").
		Limit(p.PageSize).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.ForumMessage]{
		Items:      messages,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *ForumMessageRepository) CreateMessage(
	ctx context.Context,
	message *models.ForumMessage,
) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *ForumMessageRepository) UpdateMessage(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) (*models.ForumMessage, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.ForumMessage{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetMessageByID(ctx, id)
}

func (r *ForumMessageRepository) SoftDeleteMessage(
	ctx context.Context,
	id uuid.UUID,
) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.ForumMessage{}).
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

var _ ports.ForumMessageRepository = (*ForumMessageRepository)(nil)
