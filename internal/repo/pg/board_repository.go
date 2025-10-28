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

type BoardRepository struct {
	db *gorm.DB
}

func NewBoardRepository(
	db *gorm.DB,
) *BoardRepository {
	return &BoardRepository{
		db: db,
	}
}

func (r *BoardRepository) ListBoards(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.Board], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Board{}).Where("deleted = FALSE OR deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var boards []models.Board
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("created_at DESC NULLS LAST").
		Preload("Statuses", "deleted = FALSE OR deleted IS NULL").
		Preload("Statuses.Tasks", "tasks.deleted = FALSE OR tasks.deleted IS NULL").
		Limit(p.PageSize).
		Offset(offset).
		Find(&boards).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Board]{
		Items:      boards,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *BoardRepository) GetBoardByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Board, error) {
	var board models.Board
	err := r.db.WithContext(ctx).
		Model(&models.Board{}).
		Preload("Statuses", "deleted = FALSE OR deleted IS NULL").
		Preload("Statuses.Tasks", "tasks.deleted = FALSE OR tasks.deleted IS NULL").
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		First(&board).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &board, nil
}

func (r *BoardRepository) ListBoardsByProject(
	ctx context.Context,
	projectID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Board], error) {
	var boards []models.Board
	err := r.db.WithContext(ctx).
		Model(&models.Board{}).
		Preload("Statuses", "deleted = FALSE OR deleted IS NULL").
		Preload("Statuses.Tasks", "tasks.deleted = FALSE OR tasks.deleted IS NULL").
		Where("project_id = ? AND (deleted = FALSE OR deleted IS NULL)", projectID).
		Order("created_at DESC NULLS LAST").
		Find(&boards).Error
	if err != nil {
		return nil, err
	}
	return &ports.Page[models.Board]{ // TODO Add pagination on db layer
		Items:      boards,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: 0,
	}, nil
}

func (r *BoardRepository) CreateBoard(
	ctx context.Context,
	board *models.Board,
) error {
	return r.db.WithContext(ctx).Create(board).Error
}

func (r *BoardRepository) UpdateBoard(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) (*models.Board, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Board{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetBoardByID(ctx, id)
}

func (r *BoardRepository) SoftDeleteBoard(
	ctx context.Context,
	id uuid.UUID,
) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Board{}).
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

var _ ports.BoardRepository = (*BoardRepository)(nil)
