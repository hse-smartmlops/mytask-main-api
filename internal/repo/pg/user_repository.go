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

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	return nil // TODO Add implementation
}

func (r *UserRepository) UpdateUser(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.User, error) {
	return nil, nil // TODO Add implementation
}

func (r *UserRepository) ListUsers(ctx context.Context, p ports.PaginationParams) (*ports.Page[models.User], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted = FALSE")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var users []models.User
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("created_at DESC").
		Limit(p.PageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.User]{
		Items:      users,
		Page:       p.Page,
		PageSize:   p.PageSize,
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

func (r *UserRepository) UpdateUserAvatar(ctx context.Context, id uuid.UUID, avatarPath string) error {
	updates := map[string]interface{}{
		"avatar_path": avatarPath,
		"updated_at":  time.Now().UTC(),
	}
	res := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND deleted = FALSE", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) ClearUserAvatar(ctx context.Context, id uuid.UUID) error {
	updates := map[string]interface{}{
		"avatar_path": "",
		"updated_at":  time.Now().UTC(),
	}
	res := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND deleted = FALSE", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var _ ports.UserRepository = (*UserRepository)(nil)
