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

type problemService struct {
	repo   ports.ProblemRepository
	config *config.PaginationConfig
}

func NewProblemService(
	repo ports.ProblemRepository,
	config *config.PaginationConfig,
) ports.ProblemService {
	return &problemService{
		repo:   repo,
		config: config,
	}
}

func (s *problemService) ListProblems(ctx context.Context, p ports.PaginationParams) (*ports.Page[models.Problem], error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return s.repo.ListProblems(ctx, p)
}

func (s *problemService) ListProblemsByUser(ctx context.Context, userID uuid.UUID, p ports.PaginationParams) (*ports.Page[models.Problem], error) {
	if userID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return s.repo.ListProblemsByUser(ctx, userID, p)
}

func (s *problemService) GetProblem(ctx context.Context, id uuid.UUID) (*models.Problem, error) {
	problem, err := s.repo.GetProblemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if problem == nil {
		return nil, domain.ErrNotFound
	}
	return problem, nil
}

func (s *problemService) CreateProblem(ctx context.Context, input ports.CreateProblemInput) (*models.Problem, error) {
	desc := sanitizeDescription(input.Description)
	if len(desc) == 0 {
		return nil, domain.ErrInvalidInput
	}
	now := time.Now().UTC()
	deleted := false

	problem := &models.Problem{
		ID:          uuid.New(),
		Description: desc,
		CreatorID:   input.CreatorID,
		Name:        cloneStringPtr(input.Name),
		CreatedAt:   &now,
		UpdatedAt:   &now,
		Deleted:     &deleted,
	}

	if err := s.repo.CreateProblem(ctx, problem); err != nil {
		return nil, err
	}

	return s.repo.GetProblemByID(ctx, problem.ID)
}

func (s *problemService) UpdateProblem(ctx context.Context, id uuid.UUID, input ports.UpdateProblemInput) (*models.Problem, error) {
	updates := make(map[string]interface{})

	if input.Description != nil {
		desc := sanitizeDescription(*input.Description)
		if len(desc) == 0 {
			return nil, domain.ErrInvalidInput
		}
		updates["description"] = desc
	}

	if input.Name != nil {
		updates["name"] = cloneStringPtr(input.Name)
	}

	if len(updates) == 0 {
		return s.repo.GetProblemByID(ctx, id)
	}

	return s.repo.UpdateProblem(ctx, id, updates)
}

func (s *problemService) DeleteProblem(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteProblem(ctx, id)
}

var _ ports.ProblemService = (*problemService)(nil)
