package service

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TaskService interface {
	GetAllTasks(page, pageSize int) ([]models.Task, int64, error)
	GetTaskByID(taskID uuid.UUID) (*models.Task, error)
	GetTasksByProjectID(projectID uuid.UUID, page, pageSize int) ([]models.Task, int64, error)
	CreateTask(req request.TaskCreateRequest) (uuid.UUID, error)
	UpdateTask(taskID uuid.UUID, req request.TaskUpdateRequest) error
	DeleteTask(taskID uuid.UUID) error
	GetTasksByUserId(userID uuid.UUID, page, pageSize int) ([]models.Task, int64, error)
	TaskMoveFunc(taskID, toStatusID uuid.UUID) ([]models.Status, error)
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{
		repo: repo,
	}
}

func (s *taskService) GetAllTasks(page, pageSize int) ([]models.Task, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAllTasks(pageSize, offset)
}

func (s *taskService) GetTaskByID(taskID uuid.UUID) (*models.Task, error) {
	return s.repo.GetTaskByID(taskID)
}

func (s *taskService) GetTasksByProjectID(projectID uuid.UUID, page, pageSize int) ([]models.Task, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetTasksByProjectID(projectID, pageSize, offset)
}

func (s *taskService) CreateTask(req request.TaskCreateRequest) (uuid.UUID, error) {
	// === Валидация и парсинг AssignedTo (исполнитель) ===
	if req.AssignedTo == nil {
		return uuid.Nil, errors.New("invalid assigned_to")
	}
	assigneeID, err := uuid.Parse(*req.AssignedTo)
	if err != nil {
		return uuid.Nil, errors.New("invalid assigned_to")
	}

	// Проверка существования исполнителя
	assigneeExists, err := s.repo.UserExists(assigneeID)
	if err != nil {
		return uuid.Nil, err
	}
	if !assigneeExists {
		return uuid.Nil, errors.New("assignee not found")
	}

	// === Валидация и парсинг CreatorID (поручитель) ===
	if req.CreatorID == nil {
		return uuid.Nil, errors.New("invalid creator_id")
	}
	creatorID, err := uuid.Parse(*req.CreatorID)
	if err != nil {
		return uuid.Nil, errors.New("invalid creator_id")
	}

	// Проверка существования поручителя
	creatorExists, err := s.repo.UserExists(creatorID)
	if err != nil {
		return uuid.Nil, err
	}
	if !creatorExists {
		return uuid.Nil, errors.New("creator not found")
	}

	// === Валидация StatusID ===
	statusID, err := uuid.Parse(req.StatusID)
	if err != nil {
		return uuid.Nil, errors.New("invalid status_id")
	}

	// Проверка существования статуса и его принадлежности к неудалённой доске
	statusExists, err := s.repo.StatusExists(statusID)
	if err != nil {
		return uuid.Nil, err
	}
	if !statusExists {
		return uuid.Nil, errors.New("status not found")
	}

	// === Создание задачи ===
	now := time.Now()
	deleted := false

	task := models.Task{
		ID:            uuid.New(),
		StatusID:      statusID,
		Priority:      req.Priority,
		Name:          req.Name,
		Description:   req.Description,
		CreatedBy:     &creatorID,
		AssignedTo:    &assigneeID,
		Deadline:      req.Deadline,
		StartDate:     req.StartDate,
		GitlabIssueID: req.GitlabIssueID,
		Category:      req.Category,
		Deleted:       &deleted,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}

	err = s.repo.CreateTask(task)
	if err != nil {
		return uuid.Nil, err
	}

	return task.ID, nil
}

func (s *taskService) UpdateTask(taskID uuid.UUID, req request.TaskUpdateRequest) error {
	var updates = make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Deadline != nil {
		updates["deadline"] = *req.Deadline
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.GitlabIssueID != nil {
		updates["gitlab_issue_id"] = *req.GitlabIssueID
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == "" {
			// Очистить назначенного исполнителя
			updates["assigned_to"] = nil
		} else {
			assignedUUID, err := uuid.Parse(*req.AssignedTo)
			if err != nil {
				return errors.New("invalid assigned_to")
			}
			updates["assigned_to"] = assignedUUID
		}
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	updates["updated_at"] = time.Now()

	updated, err := s.repo.UpdateTask(taskID, updates)
	if err != nil {
		return err
	}

	if !updated {
		return errors.New("task not found")
	}

	return nil
}

func (s *taskService) DeleteTask(taskID uuid.UUID) error {
	deleted, err := s.repo.DeleteTask(taskID)
	if err != nil {
		return err
	}

	if !deleted {
		return errors.New("task not found")
	}

	return nil
}

func (s *taskService) GetTasksByUserId(userID uuid.UUID, page, pageSize int) ([]models.Task, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetTasksByUserId(userID, pageSize, offset)
}

func (s *taskService) TaskMoveFunc(taskID, toStatusID uuid.UUID) ([]models.Status, error) {
	// 1. Получаем задачу с её текущим статусом и доской
	task, err := s.repo.GetTaskByID(taskID)
	if err != nil {
		if err.Error() == "task not found" {
			return nil, errors.New("task not found")
		}
		return nil, err
	}

	// 2. Получаем текущий статус задачи — чтобы узнать board_id
	currentStatus, err := s.repo.GetStatusByID(task.StatusID)
	if err != nil {
		return nil, err
	}

	// 3. Проверяем, существует ли целевой статус и принадлежит ли он той же доске
	targetStatus, err := s.repo.GetStatusByID(toStatusID)
	if err != nil {
		if err.Error() == "status not found" {
			return nil, errors.New("status not found")
		}
		return nil, err
	}

	// 4. Запрещаем перемещение между разными досками
	if currentStatus.BoardID != targetStatus.BoardID {
		return nil, errors.New("different board")
	}

	// 5. Обновляем статус задачи
	updatedAt := time.Now()
	err = s.repo.UpdateTaskStatus(taskID, toStatusID, updatedAt)
	if err != nil {
		return nil, err
	}

	// Загружаем статусы доски + задачи к каждому статусу
	statuses, err := s.repo.GetStatusesByBoardID(targetStatus.BoardID)
	if err != nil {
		return nil, err
	}

	return statuses, nil
}