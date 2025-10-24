package service

import (
	"context"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type subscriptionService struct {
	repo ports.SubscriptionRepository
	config *config.PaginationConfig
}

func NewSubscriptionService(repo ports.SubscriptionRepository, config *config.PaginationConfig) ports.SubscriptionService {
	return &subscriptionService{repo: repo, config: config}
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.Subscription], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListSubscriptions(ctx, params)
}

func (s *subscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, domain.ErrNotFound
	}
	return sub, nil
}

func (s *subscriptionService) ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID, params ports.PaginationParams) (*ports.Page[models.Subscription], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListSubscriptionsByUser(ctx, userID, params)
}

func (s *subscriptionService) ListSubscriptionsByTarget(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, params ports.PaginationParams) (*ports.Page[models.Subscription], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListSubscriptionsByTarget(ctx, subscriptionID, typeID, params)
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, input ports.CreateSubscriptionInput) (*models.Subscription, error) {
	if input.UserID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}
	if input.SubscriptionID != nil && *input.SubscriptionID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	sub := &models.Subscription{
		ID:             uuid.New(),
		UserID:         input.UserID,
		SubscriptionID: input.SubscriptionID,
		TypeID:         input.TypeID,
		CreatedAt:      &now,
		UpdatedAt:      &now,
		Deleted:        &deleted,
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return s.repo.GetSubscriptionByID(ctx, sub.ID)
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteSubscription(ctx, id)
}

func (s *subscriptionService) ListBySubObject(
	ctx context.Context,
	subobjID uuid.UUID,
	typeID *int8,
	p ports.PaginationParams,
) (*ports.Page[models.Subscription], error) {
	return s.ListSubscriptionsByTarget(ctx, subobjID, typeID, p)
}

var _ ports.SubscriptionService = (*subscriptionService)(nil)
