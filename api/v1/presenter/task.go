package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

func ToTaskDTO(task *models.Task) response.Task {
	if task == nil {
		return response.Task{}
	}

	var projectID *string
	var status *response.TaskStatus
	if task.Status != nil {
		if task.Status.Board != nil {
			id := task.Status.Board.ProjectID.String()
			projectID = &id
		}
		status = &response.TaskStatus{
			ID:        task.Status.ID.String(),
			Name:      valueOrEmpty(task.Status.Name),
			Key:       task.Status.Key,
			Color:     task.Status.Color,
			IsDefault: task.Status.IsDefault,
			SortOrder: task.Status.SortOrder,
			IsActive:  task.Status.IsActive,
			IsOpen:    task.Status.IsOpen,
		}
	}

	return response.Task{
		ID:             task.ID.String(),
		ProjectID:      projectID,
		StatusID:       task.StatusID.String(),
		Priority:       task.Priority,
		Name:           valueOrEmpty(task.Name),
		Description:    task.Description,
		CreatedBy:      uuidPtrToString(task.CreatedBy),
		AssignedTo:     uuidPtrToString(task.AssignedTo),
		Deadline:       task.Deadline,
		StartDate:      task.StartDate,
		GitlabIssueID:  task.GitlabIssueID,
		Category:       task.Category,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
		Status:         status,
		CreatedByUser:  toUserSummary(task.CreatedByUser),
		AssignedToUser: toUserSummary(task.AssignedToUser),
	}
}

func MapTasksPage(page *app.Page[models.Task]) (dto.Pagination, response.TasksPage) {
	if page == nil {
		return dto.Pagination{}, response.TasksPage{}
	}

	items := make([]response.Task, len(page.Items))
	for i := range page.Items {
		items[i] = ToTaskDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.TasksPage{Tasks: items}
}

func toUserSummary(user *models.User) *response.UserSummary {
	if user == nil {
		return nil
	}
	return &response.UserSummary{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
}

func uuidPtrToString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}
