package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TaskHandler struct {
	service app.TaskService
}

func NewTaskHandler(service app.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func RegisterTaskRoutes(group *echo.Group, service app.TaskService) {
	handler := NewTaskHandler(service)

	group.GET("/tasks", handler.ListTasks)
	group.GET("/tasks/:id", handler.GetTask)
	group.GET("/projects/:project_id/tasks", handler.ListTasksByProject)
	group.GET("/users/:user_id/tasks", handler.ListTasksByUser)
	group.POST("/tasks", handler.CreateTask)
	group.PATCH("/tasks/:id", handler.UpdateTask)
	group.DELETE("/tasks/:id", handler.DeleteTask)
}

func (h *TaskHandler) ListTasks(c echo.Context) error {
	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListTasks(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}

	pagination, payload := presenter.MapTasksPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *TaskHandler) GetTask(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	task, err := h.service.GetTask(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("task_not_found", "task not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to get task"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToTaskDTO(task)))
}

func (h *TaskHandler) ListTasksByProject(c echo.Context) error {
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListTasksByProject(c.Request().Context(), projectID, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}

	pagination, payload := presenter.MapTasksPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *TaskHandler) ListTasksByUser(c echo.Context) error {
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListTasksByUser(c.Request().Context(), userID, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}

	pagination, payload := presenter.MapTasksPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *TaskHandler) CreateTask(c echo.Context) error {
	var req request.CreateTask
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	statusID, err := parseUUID(req.StatusID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "status_id must be a valid UUID"))
	}
	createdBy, err := parseUUIDPointer(req.CreatedBy)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "created_by must be a valid UUID"))
	}
	assignedTo, err := parseUUIDPointer(req.AssignedTo)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "assigned_to must be a valid UUID"))
	}
	deadline, err := parseTimePointer(req.Deadline)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "deadline must be RFC3339 timestamp"))
	}
	startDate, err := parseTimePointer(req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "start_date must be RFC3339 timestamp"))
	}

	input := app.CreateTaskInput{
		StatusID:      statusID,
		Priority:      req.Priority,
		Name:          req.Name,
		Description:   req.Description,
		CreatedBy:     createdBy,
		AssignedTo:    assignedTo,
		Deadline:      deadline,
		StartDate:     startDate,
		GitlabIssueID: req.GitlabIssueID,
		Category:      req.Category,
	}

	task, err := h.service.CreateTask(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("task_create_failed", "failed to create task"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToTaskDTO(task)))
}

func (h *TaskHandler) UpdateTask(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	var req request.UpdateTask
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	statusID, err := parseUUIDPointer(req.StatusID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "status_id must be a valid UUID"))
	}
	createdBy, err := parseUUIDPointer(req.CreatedBy)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "created_by must be a valid UUID"))
	}
	assignedTo, err := parseUUIDPointer(req.AssignedTo)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "assigned_to must be a valid UUID"))
	}
	deadline, err := parseTimePointer(req.Deadline)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "deadline must be RFC3339 timestamp"))
	}
	startDate, err := parseTimePointer(req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "start_date must be RFC3339 timestamp"))
	}

	input := app.UpdateTaskInput{
		StatusID:      statusID,
		Priority:      req.Priority,
		Name:          req.Name,
		Description:   req.Description,
		CreatedBy:     createdBy,
		AssignedTo:    assignedTo,
		Deadline:      deadline,
		StartDate:     startDate,
		GitlabIssueID: req.GitlabIssueID,
		Category:      req.Category,
	}

	task, err := h.service.UpdateTask(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("task_not_found", "task not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("task_update_failed", "failed to update task"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToTaskDTO(task)))
}

func (h *TaskHandler) DeleteTask(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	if err := h.service.DeleteTask(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("task_not_found", "task not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("task_delete_failed", "failed to delete task"))
	}

	return c.NoContent(http.StatusNoContent)
}

func parseUUIDPointer(value *string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseTimePointer(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
