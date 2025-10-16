package service

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/repository"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ProblemService interface {
	GetAllProblems(page, pageSize int) ([]models.Problem, int64, error)
	GetProblemsByUserId(creatorUUID uuid.UUID, page, pageSize int) ([]models.Problem, int64, error)
	GetProblemByID(problemId uuid.UUID) (*models.Problem, error)
	CreateProblem(req request.ProblemCreateRequest) (uuid.UUID, error)
	UpdateProblem(problemId uuid.UUID, req request.ProblemUpdateRequest) error
	DeleteProblem(problemId uuid.UUID) error
}

type problemService struct {
	repo repository.ProblemRepository
}

func NewProblemService(repo repository.ProblemRepository) ProblemService {
	return &problemService{
		repo: repo,
	}
}

func (s *problemService) GetAllProblems(page, pageSize int) ([]models.Problem, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAllProblems(pageSize, offset)
}

func (s *problemService) GetProblemsByUserId(creatorUUID uuid.UUID, page, pageSize int) ([]models.Problem, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetProblemsByUserId(creatorUUID, pageSize, offset)
}

func (s *problemService) GetProblemByID(problemId uuid.UUID) (*models.Problem, error) {
	return s.repo.GetProblemByID(problemId)
}

func (s *problemService) CreateProblem(req request.ProblemCreateRequest) (uuid.UUID, error) {
	creatorId, err := uuid.Parse(req.CreatorID)
	if err != nil {
		return uuid.Nil, errors.New("invalid creator id")
	}

	now := time.Now()
	del := false

	var description pq.StringArray
	if req.Description != nil {
		description = pq.StringArray(*req.Description)
	} else {
		description = pq.StringArray{}
	}

	p := models.Problem{
		ID:          uuid.New(),
		Description: description,
		CreatorID:   &creatorId,
		Name:        req.Name,
		CreatedAt:   &now,
		Deleted:     &del,
	}

	err = s.repo.CreateProblem(p)
	if err != nil {
		return uuid.Nil, err
	}

	return p.ID, nil
}

func (s *problemService) UpdateProblem(problemId uuid.UUID, req request.ProblemUpdateRequest) error {
	updateData := make(map[string]interface{})
	if req.Description != nil {
		updateData["description"] = pq.StringArray(*req.Description)
	}

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	now := time.Now()
	updateData["updated_at"] = &now

	updated, err := s.repo.UpdateProblem(problemId, updateData)
	if err != nil {
		return err
	}

	if !updated {
		return errors.New("problem not found")
	}

	return nil
}

func (s *problemService) DeleteProblem(problemId uuid.UUID) error {
	deleted, err := s.repo.DeleteProblem(problemId)
	if err != nil {
		return err
	}

	if !deleted {
		return errors.New("problem not found")
	}

	return nil
}