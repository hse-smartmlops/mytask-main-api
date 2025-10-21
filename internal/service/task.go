package service

import (
	"context"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type taskService struct {
	repo ports.TaskRepository
}

func NewTaskService(repo ports.TaskRepository) ports.TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) ListTasks(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.Task], error) {
	return s.repo.ListTasks(ctx, params)
}

func (s *taskService) GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	task, err := s.repo.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, domain.ErrNotFound
	}
	return task, nil
}

func (s *taskService) ListTasksByProject(ctx context.Context, projectID uuid.UUID, params ports.PaginationParams) (*ports.Page[models.Task], error) {
	return s.repo.ListTasksByProject(ctx, projectID, params)
}

func (s *taskService) ListTasksByUser(ctx context.Context, userID uuid.UUID, params ports.PaginationParams) (*ports.Page[models.Task], error) {
	return s.repo.ListTasksByUser(ctx, userID, params)
}

func (s *taskService) CreateTask(ctx context.Context, input ports.CreateTaskInput) (*models.Task, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	task := &models.Task{
		ID:            uuid.New(),
		StatusID:      input.StatusID,
		Priority:      input.Priority,
		Name:          stringPtr(name),
		Description:   cloneStringPtr(input.Description),
		CreatedBy:     input.CreatedBy,
		AssignedTo:    input.AssignedTo,
		Deadline:      input.Deadline,
		StartDate:     input.StartDate,
		GitlabIssueID: input.GitlabIssueID,
		Category:      input.Category,
		CreatedAt:     &now,
		UpdatedAt:     &now,
		Deleted:       &deleted,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return s.repo.GetTaskByID(ctx, task.ID)
}

func (s *taskService) UpdateTask(ctx context.Context, id uuid.UUID, input ports.UpdateTaskInput) (*models.Task, error) {
	updates := make(map[string]interface{})

	if input.StatusID != nil {
		updates["status_id"] = *input.StatusID
	}
	if input.Priority != nil {
		updates["priority"] = input.Priority
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.ErrInvalidInput
		}
		updates["name"] = name
	}
	if input.Description != nil {
		updates["description"] = cloneStringPtr(input.Description)
	}
	if input.CreatedBy != nil {
		updates["created_by"] = input.CreatedBy
	}
	if input.AssignedTo != nil {
		updates["assigned_to"] = input.AssignedTo
	}
	if input.Deadline != nil {
		updates["deadline"] = input.Deadline
	}
	if input.StartDate != nil {
		updates["start_date"] = input.StartDate
	}
	if input.GitlabIssueID != nil {
		updates["gitlab_issue_id"] = input.GitlabIssueID
	}
	if input.Category != nil {
		updates["category"] = input.Category
	}

	if len(updates) == 0 {
		return s.repo.GetTaskByID(ctx, id)
	}

	return s.repo.UpdateTask(ctx, id, updates)
}

func (s *taskService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteTask(ctx, id)
}

func (s *taskService) ListActiveTasksByUser(ctx context.Context, userID uuid.UUID, p ports.PaginationParams) (*ports.Page[models.Task], error) {
	return s.repo.ListActiveTasksByUser(ctx, userID, p)
}

func (s *taskService) ListTasksByUserAndProject(ctx context.Context, userID, projectID uuid.UUID, p ports.PaginationParams) (*ports.Page[models.Task], error) {
	return s.repo.ListTasksByUserAndProject(ctx, userID, projectID, p)
}

func (s *taskService) ListTasksByBoard(ctx context.Context, boardID uuid.UUID, p ports.PaginationParams) (*ports.Page[models.Task], error) {
	return s.repo.ListTasksByBoard(ctx, boardID, p)
}

func (s *taskService) MoveTaskBetweenStatuses(ctx context.Context, taskID, toStatusID uuid.UUID) (*models.Task, error) {
	return s.repo.MoveTaskBetweenStatuses(ctx, taskID, toStatusID)
}

var _ ports.TaskService = (*taskService)(nil)
