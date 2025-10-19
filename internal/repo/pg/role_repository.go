package pg

import (
	"context"
	"time"

	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) ListRoles(ctx context.Context, params app.PaginationParams) (*app.Page[models.Role], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Role{}).Where("deleted = FALSE OR deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var roles []models.Role
	offset := (params.Page - 1) * params.PageSize
	if err := base.Order("created_at DESC NULLS LAST").
		Limit(params.PageSize).
		Offset(offset).
		Find(&roles).Error; err != nil {
		return nil, err
	}

	return &app.Page[models.Role]{
		Items:      roles,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *RoleRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) CreateRole(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RoleRepository) UpdateRole(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Role, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetRoleByID(ctx, id)
}

func (r *RoleRepository) SoftDeleteRole(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Role{}).
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

func timePtr(t time.Time) *time.Time {
	return &t
}

var _ app.RoleRepository = (*RoleRepository)(nil)
