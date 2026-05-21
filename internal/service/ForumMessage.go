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

type ForumMessageService interface {
	GetAllForumMessages(page, pageSize int) ([]models.ForumMessage, int64, error)
	GetForumMessagesByProblemId(problemID uuid.UUID, page, pageSize int) ([]models.ForumMessage, int64, error)
	GetForumMessageById(messageID uuid.UUID) (*models.ForumMessage, error)
	CreateForumMessage(req request.CreateForumMessageRequest) (uuid.UUID, error)
	UpdateForumMessage(messageID uuid.UUID, req request.UpdateForumMessageRequest) error
	DeleteForumMessage(messageID uuid.UUID) error
}

type forumMessageService struct {
	repo repository.ForumMessageRepository
}

func NewForumMessageService(repo repository.ForumMessageRepository) ForumMessageService {
	return &forumMessageService{
		repo: repo,
	}
}

func (s *forumMessageService) GetAllForumMessages(page, pageSize int) ([]models.ForumMessage, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAllForumMessages(pageSize, offset)
}

func (s *forumMessageService) GetForumMessagesByProblemId(problemID uuid.UUID, page, pageSize int) ([]models.ForumMessage, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetForumMessagesByProblemId(problemID, pageSize, offset)
}

func (s *forumMessageService) GetForumMessageById(messageID uuid.UUID) (*models.ForumMessage, error) {
	return s.repo.GetForumMessageById(messageID)
}

func (s *forumMessageService) CreateForumMessage(req request.CreateForumMessageRequest) (uuid.UUID, error) {
	problemId, err := uuid.Parse(req.ProblemID)
	if err != nil {
		return uuid.Nil, errors.New("invalid problem id")
	}

	var creatorId *uuid.UUID
	if req.CreatorID != "" {
		v, err := uuid.Parse(req.CreatorID)
		if err != nil {
			return uuid.Nil, errors.New("invalid creator id")
		}
		creatorId = &v
	}

	now := time.Now()
	del := false

	var description pq.StringArray
	if req.Description != nil {
		description = pq.StringArray(*req.Description)
	} else {
		description = pq.StringArray{}
	}

	var replyToID *uuid.UUID
	if req.ReplyToID != nil && *req.ReplyToID != "" {
		v, err := uuid.Parse(*req.ReplyToID)
		if err == nil {
			replyToID = &v
		}
	}

	fm := models.ForumMessage{
		ID:          uuid.New(),
		ProblemID:   problemId,
		Description: description,
		CreatorID:   creatorId,
		ReplyToID:   replyToID,
		CreatedAt:   &now,
		Deleted:     &del,
	}

	err = s.repo.CreateForumMessage(fm)
	if err != nil {
		return uuid.Nil, err
	}

	return fm.ID, nil
}

func (s *forumMessageService) UpdateForumMessage(messageID uuid.UUID, req request.UpdateForumMessageRequest) error {
	updateData := make(map[string]interface{})
	if req.ProblemID != nil {
		pid, err := uuid.Parse(*req.ProblemID)
		if err != nil {
			return errors.New("invalid problem id")
		}
		updateData["problem_id"] = pid
	}
	if req.Description != nil {
		updateData["description"] = pq.StringArray(*req.Description)
	}
	if req.CreatorID != nil {
		cid, err := uuid.Parse(*req.CreatorID)
		if err != nil {
			return errors.New("invalid creator id")
		}
		updateData["creator_id"] = cid
	}

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	now := time.Now()
	updateData["updated_at"] = &now

	updated, err := s.repo.UpdateForumMessage(messageID, updateData)
	if err != nil {
		return err
	}

	if !updated {
		return errors.New("forum message not found")
	}

	return nil
}

func (s *forumMessageService) DeleteForumMessage(messageID uuid.UUID) error {
	deleted, err := s.repo.DeleteForumMessage(messageID)
	if err != nil {
		return err
	}

	if !deleted {
		return errors.New("forum message not found")
	}

	return nil
}