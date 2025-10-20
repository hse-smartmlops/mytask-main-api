package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

var (
	_ request.CreateTask
	_ request.UpdateTask
)

type TaskHandler struct {
	service ports.TaskService
}

func NewTaskHandler(service ports.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

// @Summary Register Task Routes
// @Description Register routes for task management
// @Tags Tasks
func RegisterTaskRoutes(group *echo.Group, service ports.TaskService) {
	handler := NewTaskHandler(service)

	group.GET("/tasks", handler.ListTasks)
	group.GET("/tasks/:id", handler.GetTask)
	group.GET("/projects/:project_id/tasks", handler.ListTasksByProject)
	group.GET("/users/:user_id/tasks", handler.ListTasksByUser)
	group.POST("/tasks", handler.CreateTask)
	group.PATCH("/tasks/:id", handler.UpdateTask)
	group.DELETE("/tasks/:id", handler.DeleteTask)
}

// @Summary List Tasks
// @Description Retrieve a paginated list of tasks
// @Tags Tasks
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.TasksListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks [get]
func (h *TaskHandler) ListTasks(c echo.Context) error {
	page, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	result, err := h.service.ListTasks(c.Request().Context(), ports.PaginationParams{
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}

	pagination, payload := presenter.MapTasksPage(result)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Task
// @Description Retrieve a task by its ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	task, err := h.service.GetTask(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("task_not_found", "task not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to get task"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToTaskDTO(task))
}

// @Summary List Tasks By Project
// @Description Retrieve a paginated list of tasks associated with a specific project
// @Tags Tasks
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.TasksListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{project_id}/tasks [get]
func (h *TaskHandler) ListTasksByProject(c echo.Context) error {
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	page, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	result, err := h.service.ListTasksByProject(c.Request().Context(), projectID, ports.PaginationParams{Page: page, PageSize: size})
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}

	pagination, payload := presenter.MapTasksPage(result)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary List Tasks By User
// @Description Retrieve a paginated list of tasks associated with a specific user
// @Tags Tasks
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.TasksListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{user_id}/tasks [get]
func (h *TaskHandler) ListTasksByUser(c echo.Context) error {
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	page, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	result, err := h.service.ListTasksByUser(c.Request().Context(), userID, ports.PaginationParams{Page: page, PageSize: size})
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}

	pagination, payload := presenter.MapTasksPage(result)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Create Task
// @Description Create a new task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body request.CreateTask true "Task creation payload"
// @Success 201 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateTaskPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.CreateTaskInput{
		StatusID:      req.StatusUUID,
		Priority:      req.Priority,
		Name:          req.Name,
		Description:   req.Description,
		CreatedBy:     req.CreatedByUUID,
		AssignedTo:    req.AssignedToUUID,
		Deadline:      req.DeadlineTime,
		StartDate:     req.StartDateTime,
		GitlabIssueID: req.GitlabIssueID,
		Category:      req.Category,
	}

	task, err := h.service.CreateTask(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("task_create_failed", "failed to create task"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToTaskDTO(task))
}

// @Summary Update Task
// @Description Update an existing task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body request.UpdateTask true "Task data"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/{id} [patch]
func (h *TaskHandler) UpdateTask(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateUpdateTaskPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateTaskInput{
		StatusID:      req.StatusUUID,
		Priority:      req.Priority,
		Name:          req.Name,
		Description:   req.Description,
		CreatedBy:     req.CreatedByUUID,
		AssignedTo:    req.AssignedToUUID,
		Deadline:      req.DeadlineTime,
		StartDate:     req.StartDateTime,
		GitlabIssueID: req.GitlabIssueID,
		Category:      req.Category,
	}

	task, err := h.service.UpdateTask(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("task_not_found", "task not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("task_update_failed", "failed to update task"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToTaskDTO(task))
}

// @Summary Delete Task
// @Description Delete a task by its ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	if err := h.service.DeleteTask(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("task_not_found", "task not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("task_delete_failed", "failed to delete task"))
	}

	return c.NoContent(http.StatusNoContent)
}
