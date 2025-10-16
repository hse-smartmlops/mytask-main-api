package service

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

type BoardService interface {
	GetAllBoards(page, pageSize int) ([]models.Board, int64, error)
	GetBoardById(boardID uuid.UUID) (*models.Board, error)
	GetBoardByProjectId(projectID uuid.UUID) ([]models.Board, error)
	CreateBoard(req request.BoardCreateRequest) (uuid.UUID, error)
	UpdateBoard(boardID uuid.UUID, req request.BoardUpdateRequest) error
	DeleteBoard(boardID uuid.UUID) error
}

type boardService struct {
	repo repository.BoardRepository
}

func NewBoardService(repo repository.BoardRepository) BoardService {
	return &boardService{
		repo: repo,
	}
}

func (s *boardService) GetAllBoards(page, pageSize int) ([]models.Board, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAllBoards(pageSize, offset)
}

func (s *boardService) GetBoardById(boardID uuid.UUID) (*models.Board, error) {
	return s.repo.GetBoardById(boardID)
}

func (s *boardService) GetBoardByProjectId(projectID uuid.UUID) ([]models.Board, error) {
	return s.repo.GetBoardByProjectId(projectID)
}

func (s *boardService) CreateBoard(req request.BoardCreateRequest) (uuid.UUID, error) {
	if req.ProjectID == nil {
		return uuid.Nil, errors.New("project_id is required")
	}

	projectID, err := uuid.Parse(*req.ProjectID)
	if err != nil {
		return uuid.Nil, errors.New("invalid project id")
	}

	now := time.Now()
	deleted := false
	boardID := uuid.New()
	tr := true

	board := models.Board{
		ID:          boardID,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Deleted:     &deleted,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// Вспомогательная функция для создания статуса с уникальным Key и порядком
	makeStatus := func(order int, name, color string, isOpen bool) models.Status {
		id := uuid.New()
		key := id.String()[:8] // сокращённый UUID (8 символов)
		return models.Status{
			ID:        id,
			BoardID:   boardID,
			SortOrder:     &order,
			Key:       &key,
			Name:      &name,
			Color:     &color,
			IsDefault: &tr,
			IsActive:  boolPtr(true),
			IsOpen:    &isOpen,
			Deleted:   &deleted,
			CreatedAt: &now,
			UpdatedAt: &now,
		}
	}

	defaultStatuses := []models.Status{
		makeStatus(0, "To Do", "#fc0000ff", true),
		makeStatus(1024, "Done", "#28A745", true),
	}

	err = s.repo.CreateBoardWithStatuses(board, defaultStatuses)
	if err != nil {
		return uuid.Nil, err
	}

	return boardID, nil
}

func (s *boardService) UpdateBoard(boardID uuid.UUID, req request.BoardUpdateRequest) error {
	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	updateData["updated_at"] = time.Now()

	updated, err := s.repo.UpdateBoard(boardID, updateData)
	if err != nil {
		return err
	}

	if !updated {
		return errors.New("board not found")
	}

	return nil
}

func (s *boardService) DeleteBoard(boardID uuid.UUID) error {
	deleted, err := s.repo.DeleteBoard(boardID)
	if err != nil {
		return err
	}

	if !deleted {
		return errors.New("board not found")
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}