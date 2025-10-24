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

type ProblemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(
	db *gorm.DB,
) *ProblemRepository {
	return &ProblemRepository{
		db: db,
	}
}

func (r *ProblemRepository) ListProblems(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.Problem], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).
		Model(&models.Problem{}).
		Where("problems.deleted = FALSE OR problems.deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var problems []models.Problem
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("problems.created_at DESC NULLS LAST").
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Limit(p.PageSize).
		Offset(offset).
		Find(&problems).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Problem]{
		Items:      problems,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *ProblemRepository) ListProblemsByUser(
	ctx context.Context,
	userID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Problem], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).
		Model(&models.Problem{}).
		Where("(problems.deleted = FALSE OR problems.deleted IS NULL) AND problems.creator_id = ?", userID)

	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var problems []models.Problem
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("problems.created_at DESC NULLS LAST").
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Limit(p.PageSize).
		Offset(offset).
		Find(&problems).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Problem]{
		Items:      problems,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *ProblemRepository) GetProblemByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Problem, error) {
	var problem models.Problem
	err := r.db.WithContext(ctx).
		Model(&models.Problem{}).
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Where("problems.id = ? AND (problems.deleted = FALSE OR problems.deleted IS NULL)", id).
		First(&problem).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &problem, nil
}

func (r *ProblemRepository) CreateProblem(
	ctx context.Context,
	problem *models.Problem,
) error {
	return r.db.WithContext(ctx).Create(problem).Error
}

func (r *ProblemRepository) UpdateProblem(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) (*models.Problem, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Problem{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetProblemByID(ctx, id)
}

func (r *ProblemRepository) SoftDeleteProblem(
	ctx context.Context,
	id uuid.UUID,
) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Problem{}).
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

var _ ports.ProblemRepository = (*ProblemRepository)(nil)
