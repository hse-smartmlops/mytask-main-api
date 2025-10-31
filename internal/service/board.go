package service

import (
	"context"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type boardService struct {
	repo   ports.BoardRepository
	config *config.PaginationConfig
}

func NewBoardService(
	repo ports.BoardRepository,
	config *config.PaginationConfig,
) ports.BoardService {
	return &boardService{
		repo:   repo,
		config: config,
	}
}

func (s *boardService) ListBoards(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.Board], error) {
	checkPagination(&p, s.config)
	return s.repo.ListBoards(ctx, p)
}

func (s *boardService) GetBoard(
	ctx context.Context,
	id uuid.UUID,
) (*models.Board, error) {
	board, err := s.repo.GetBoardByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if board == nil {
		return nil, domain.ErrNotFound
	}
	return board, nil
}

func (s *boardService) ListBoardsByProject(
	ctx context.Context,
	projectID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Board], error) {
	checkPagination(&p, s.config)
	return s.repo.ListBoardsByProject(ctx, projectID, p)
}

func (s *boardService) CreateBoard(
	ctx context.Context,
	input ports.CreateBoardInput,
) (*models.Board, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	board := &models.Board{
		ID:          uuid.New(),
		ProjectID:   input.ProjectID,
		Name:        stringPtr(name),
		Description: cloneStringPtr(input.Description),
		CreatedAt:   &now,
		UpdatedAt:   &now,
		Deleted:     &deleted,
	}

	if err := s.repo.CreateBoard(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *boardService) UpdateBoard(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateBoardInput,
) (*models.Board, error) {
	updates := make(map[string]interface{})

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.ErrInvalidInput
		}
		updates["name"] = &name
	}

	if input.Description != nil {
		updates["description"] = cloneStringPtr(input.Description)
	}

	if len(updates) == 0 {
		return s.repo.GetBoardByID(ctx, id)
	}

	return s.repo.UpdateBoard(ctx, id, updates)
}

func (s *boardService) DeleteBoard(
	ctx context.Context, 
	id uuid.UUID,
) error {
	return s.repo.SoftDeleteBoard(ctx, id)
}

var _ ports.BoardService = (*boardService)(nil)
