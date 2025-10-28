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

type forumMessageService struct {
	repo   ports.ForumMessageRepository
	config *config.PaginationConfig
}

func NewForumMessageService(
	repo ports.ForumMessageRepository,
	config *config.PaginationConfig,
) ports.ForumMessageService {
	return &forumMessageService{
		repo:   repo,
		config: config,
	}
}

func (s *forumMessageService) ListMessages(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.ForumMessage], error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return s.repo.ListMessages(ctx, p)
}

func (s *forumMessageService) GetMessage(
	ctx context.Context,
	id uuid.UUID,
) (*models.ForumMessage, error) {
	msg, err := s.repo.GetMessageByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, domain.ErrNotFound
	}
	return msg, nil
}

func (s *forumMessageService) ListMessagesByProblem(
	ctx context.Context,
	problemID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.ForumMessage], error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return s.repo.ListMessagesByProblem(ctx, problemID, p)
}

func (s *forumMessageService) CreateMessage(
	ctx context.Context,
	input ports.CreateForumMessageInput,
) (*models.ForumMessage, error) {
	if input.ProblemID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}
	if len(input.Description) == 0 {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	msg := &models.ForumMessage{
		ID:          uuid.New(),
		ProblemID:   input.ProblemID,
		Description: sanitizeDescription(input.Description),
		CreatorID:   input.CreatorID,
		CreatedAt:   &now,
		UpdatedAt:   &now,
		Deleted:     &deleted,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	return s.repo.GetMessageByID(ctx, msg.ID)
}

func (s *forumMessageService) UpdateMessage(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateForumMessageInput,
) (*models.ForumMessage, error) {
	updates := make(map[string]interface{})

	if input.Description != nil {
		desc := sanitizeDescription(*input.Description)
		if len(desc) == 0 {
			return nil, domain.ErrInvalidInput
		}
		updates["description"] = desc
	}

	if len(updates) == 0 {
		return s.repo.GetMessageByID(ctx, id)
	}

	return s.repo.UpdateMessage(ctx, id, updates)
}

func (s *forumMessageService) DeleteMessage(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.SoftDeleteMessage(ctx, id)
}

type reportFilePayload struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	Checked       *int8                  `json:"checked,omitempty"`
	ReportDate    *time.Time             `json:"report_date,omitempty"`
	CreatedAt     *time.Time             `json:"created_at,omitempty"`
	UpdatedAt     *time.Time             `json:"updated_at,omitempty"`
	CompletedWork []models.CompletedWork `json:"completed_work,omitempty"`
	HelpRequests  []models.HelpRequest   `json:"help_requests,omitempty"`
	TomorrowPlans []models.TomorrowPlans `json:"tomorrow_plans,omitempty"`
	Problems      []models.ReportProblem `json:"problems,omitempty"`
}

var _ ports.ForumMessageService = (*forumMessageService)(nil)
