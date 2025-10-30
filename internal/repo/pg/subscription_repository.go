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

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(
	db *gorm.DB,
) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (r *SubscriptionRepository) ListSubscriptions(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.Subscription], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Where("subscriptions.deleted = FALSE OR subscriptions.deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var subs []models.Subscription
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("subscriptions.created_at DESC NULLS LAST").
		Limit(p.PageSize).
		Offset(offset).
		Find(&subs).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Subscription]{
		Items:      subs,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *SubscriptionRepository) GetSubscriptionByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Where("subscriptions.id = ? AND (subscriptions.deleted = FALSE OR subscriptions.deleted IS NULL)", id).
		First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepository) ListSubscriptionsByUser(
	ctx context.Context,
	userID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Subscription], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Where("subscriptions.user_id = ? AND (subscriptions.deleted = FALSE OR subscriptions.deleted IS NULL)", userID)
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var subs []models.Subscription
	offset := (p.Page - 1) * p.PageSize
	if err := base.
		Order("subscriptions.created_at DESC NULLS LAST").
		Limit(p.PageSize).
		Offset(offset).
		Find(&subs).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Subscription]{
		Items:      subs,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *SubscriptionRepository) ListSubscriptionsByTarget(
	ctx context.Context,
	subscriptionID uuid.UUID,
	typeID *int8,
	p ports.PaginationParams,
) (*ports.Page[models.Subscription], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Where("subscriptions.subscription_id = ? AND (subscriptions.deleted = FALSE OR subscriptions.deleted IS NULL)", subscriptionID)
	if typeID != nil {
		base = base.Where("subscriptions.type_id = ?", typeID)
	}
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var subs []models.Subscription
	offset := (p.Page - 1) * p.PageSize
	if err := base.
		Order("subscriptions.created_at DESC NULLS LAST").
		Limit(p.PageSize).
		Offset(offset).
		Find(&subs).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Subscription]{
		Items:      subs,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *SubscriptionRepository) CreateSubscription(
	ctx context.Context,
	sub *models.Subscription,
) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepository) SoftDeleteSubscription(
	ctx context.Context,
	id uuid.UUID,
) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
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

var _ ports.SubscriptionRepository = (*SubscriptionRepository)(nil)
