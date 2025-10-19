package pg

import (
	"context"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) ListUsers(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.User], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted = FALSE")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var users []models.User
	offset := (params.Page - 1) * params.PageSize
	if err := base.Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.User]{
		Items:      users,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND deleted = FALSE", id).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

var _ ports.UserRepository = (*UserRepository)(nil)
