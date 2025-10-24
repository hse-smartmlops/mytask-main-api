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

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(
	db *gorm.DB,
) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) ListProjectsByTeam(
	ctx context.Context,
	teamID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Project], error) {
	return nil, nil // TODO Add implementation
}

func (r *ProjectRepository) ListProjectsByUser(
	ctx context.Context,
	userID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Project], error) {
	return nil, nil // TODO Add implementation
}

func (r *ProjectRepository) ListProjects(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.Project], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("deleted = FALSE OR deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var projects []models.Project
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("created_at DESC NULLS LAST").
		Limit(p.PageSize).
		Offset(offset).
		Find(&projects).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Project]{
		Items:      projects,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *ProjectRepository) GetProjectByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		First(&project).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) CreateProject(
	ctx context.Context,
	project *models.Project,
) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *ProjectRepository) UpdateProject(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) (*models.Project, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetProjectByID(ctx, id)
}

func (r *ProjectRepository) SoftDeleteProject(
	ctx context.Context,
	id uuid.UUID,
) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Project{}).
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

var _ ports.ProjectRepository = (*ProjectRepository)(nil)
