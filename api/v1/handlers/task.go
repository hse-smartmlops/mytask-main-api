package handlers

import (
	"errors"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

var (
	_ request.CreateTask
	_ request.UpdateTask
	_ request.ImproveTaskReport
)

type TaskHandler struct {
	service      ports.TaskService
	boardService ports.BoardService
	mcpService   ports.MCPService
}

func NewTaskHandler(service ports.TaskService, boardService ports.BoardService, mcpService ports.MCPService) *TaskHandler {
	return &TaskHandler{service: service, boardService: boardService, mcpService: mcpService}
}

// @Summary Register Task Routes
// @Description Register routes for task management
// @Tags Tasks
func RegisterTaskRoutes(group *echo.Group, service ports.TaskService, boardService ports.BoardService, mcpService ports.MCPService) {
	handler := NewTaskHandler(service, boardService, mcpService)

	tgroup := group.Group("/tasks")
	{
		tgroup.GET("", handler.ListTasks)
		tgroup.GET("/:id", handler.GetTask)
		tgroup.GET("/user/:id", handler.ListTasksByUser)
		tgroup.GET("/move", handler.MoveTaskBetweenStatuses)
		tgroup.GET("/user/:user_id/active", handler.ListActiveTasksByUser)
		tgroup.GET("/user/:user_id/project/:project_id", handler.ListTasksByUserAndProject)
		tgroup.POST("", handler.CreateTask)
		tgroup.PATCH("/:id", handler.UpdateTask)
		tgroup.DELETE("/:id", handler.DeleteTask)
		tgroup.POST("/:id/improve-report", handler.ImproveTaskReport)
		tgroup.GET("/:id/improve-report/ws", handler.ImproveTaskReportWS)
		tgroup.GET("/board/:board_id", handler.ListTasksByBoard)
	}
}

// @Summary List Tasks
// @Description Retrieve a paginated list of tasks
// @Tags Tasks
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.TasksListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks [get]
func (h *TaskHandler) ListTasks(c echo.Context) error {
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
	page, err := h.service.ListTasks(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}
	pagination, payload := presenter.MapTasksPage(page)
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

	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
	result, err := h.service.ListTasksByProject(c.Request().Context(), projectID, params)
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
	userIDParam := c.Param("user_id")
	if userIDParam == "" {
		userIDParam = c.Param("id")
	}

	userID, err := parseUUID(userIDParam)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
	result, err := h.service.ListTasksByUser(c.Request().Context(), userID, params)
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

// @Summary Improve Task Report
// @Description Use MCP to improve a user supplied task report text
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body request.ImproveTaskReport true "Improvement payload"
// @Success 200 {object} dto.ImprovedTaskReportResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /tasks/{id}/improve-report [post]
func (h *TaskHandler) ImproveTaskReport(c echo.Context) error {
	if h.mcpService == nil {
		return respondError(c, http.StatusServiceUnavailable, dto.NewError("mcp_unavailable", "mcp service is not configured"))
	}

	taskID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	payload, err := middleware.BindAndValidate(c, middleware.ValidateImproveTaskReportPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	task, err := h.service.GetTask(c.Request().Context(), taskID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("task_not_found", "task not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to get task"))
		}
	}

	description := strings.TrimSpace(getOptionalString(task.Description))
	if description == "" {
		description = strings.TrimSpace(getOptionalString(task.Name))
	}

	contentType := "text/plain"
	if payload.ContentType != nil && *payload.ContentType != "" {
		contentType = *payload.ContentType
	}

	meta := make(map[string]string)
	for k, v := range payload.Meta {
		meta[k] = v
	}
	meta["task_id"] = taskID.String()

	if task.CreatedBy != nil {
		meta["created_by"] = task.CreatedBy.String()
	}
	if task.AssignedTo != nil {
		meta["assigned_to"] = task.AssignedTo.String()
	}

	if task.Status != nil {
		meta["status_id"] = task.StatusID.String()
		boardID := task.Status.BoardID.String()
		if boardID != "" {
			meta["board_id"] = boardID
		}
		if task.Status.Board != nil {
			projectID := task.Status.Board.ProjectID.String()
			if projectID != "" {
				meta["project_id"] = projectID
			}
		}
	}

	input := ports.ImproveReportInput{
		Description: description,
		UserText:    payload.UserText,
		TaskID:      taskID.String(),
		Meta:        meta,
		ContentType: contentType,
		TimeoutMs:   payload.TimeoutMs,
	}

	result, err := h.mcpService.ImproveReport(c.Request().Context(), input)
	if err != nil {
		return respondError(c, http.StatusBadGateway, dto.NewError("mcp_failed", "failed to improve report"))
	}

	respData := response.ImprovedTaskReport{
		TaskID:       taskID.String(),
		OriginalText: payload.UserText,
		ImprovedText: result.ImprovedText,
		ContentType:  result.ContentType,
		Meta:         result.Meta,
	}

	if title := toOptionalString(task.Name); title != nil {
		respData.TaskTitle = title
	}
	if desc := toOptionalString(task.Description); desc != nil {
		respData.TaskDescription = desc
	}

	return respondSuccess(c, http.StatusOK, dto.ImprovedTaskReportResponse{
		SuccessResponse: dto.NewSuccessResponse(respData),
	})
}

// @Summary Stream Task Report Improvement (WebSocket)
// @Description Stream MCP events while improving a task report through WebSocket
// @Tags Tasks
// @Param id path string true "Task ID"
// @Success 101 "Switching Protocols"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /tasks/{id}/improve-report/ws [get]
func (h *TaskHandler) ImproveTaskReportWS(c echo.Context) error {
	if h.mcpService == nil {
		return respondError(c, http.StatusServiceUnavailable, dto.NewError("mcp_unavailable", "mcp service is not configured"))
	}

	taskID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	wsHandler := websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()

		var payload request.ImproveTaskReport
		if err := websocket.JSON.Receive(conn, &payload); err != nil {
			_ = websocket.JSON.Send(conn, map[string]string{"type": "error", "message": "invalid payload"})
			return
		}

		if err := middleware.ValidateImproveTaskReportPayload(&payload); err != nil {
			_ = websocket.JSON.Send(conn, map[string]string{"type": "error", "message": err.Error()})
			return
		}

		task, err := h.service.GetTask(c.Request().Context(), taskID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				_ = websocket.JSON.Send(conn, map[string]string{"type": "error", "message": "task not found"})
				return
			}
			_ = websocket.JSON.Send(conn, map[string]string{"type": "error", "message": "failed to load task"})
			return
		}

		description := strings.TrimSpace(getOptionalString(task.Description))
		if description == "" {
			description = strings.TrimSpace(getOptionalString(task.Name))
		}

		contentType := "text/plain"
		if payload.ContentType != nil && *payload.ContentType != "" {
			contentType = *payload.ContentType
		}

		meta := make(map[string]string)
		for k, v := range payload.Meta {
			meta[k] = v
		}
		meta["task_id"] = taskID.String()

		if task.CreatedBy != nil {
			meta["created_by"] = task.CreatedBy.String()
		}
		if task.AssignedTo != nil {
			meta["assigned_to"] = task.AssignedTo.String()
		}

		if task.Status != nil {
			meta["status_id"] = task.StatusID.String()
			boardID := task.Status.BoardID.String()
			if boardID != "" {
				meta["board_id"] = boardID
			}
			if task.Status.Board != nil {
				projectID := task.Status.Board.ProjectID.String()
				if projectID != "" {
					meta["project_id"] = projectID
				}
			}
		}

		stream, err := h.mcpService.StreamReport(c.Request().Context(), ports.ImproveReportInput{
			Description: description,
			UserText:    payload.UserText,
			TaskID:      taskID.String(),
			Meta:        meta,
			ContentType: contentType,
			TimeoutMs:   payload.TimeoutMs,
		})
		if err != nil {
			_ = websocket.JSON.Send(conn, map[string]string{"type": "error", "message": "failed to start stream"})
			return
		}
		defer stream.Close()

		for {
			select {
			case err, ok := <-stream.Errors:
				if !ok {
					return
				}
				_ = websocket.JSON.Send(conn, map[string]string{"type": "error", "message": err.Error()})
				return
			case ev, ok := <-stream.Events:
				if !ok {
					return
				}

				eventPayload := map[string]interface{}{
					"type": ev.Type,
				}
				if ev.Status != nil {
					eventPayload["status"] = ev.Status
				}
				if ev.Chunk != nil {
					eventPayload["chunk"] = ev.Chunk
				}
				if ev.Final != nil {
					eventPayload["final"] = ev.Final
					eventPayload["task_id"] = taskID.String()
				}
				if ev.Error != nil {
					eventPayload["error"] = ev.Error
				}

				if err := websocket.JSON.Send(conn, eventPayload); err != nil {
					return
				}

				if ev.Final != nil {
					return
				}
			}
		}
	})

	wsHandler.ServeHTTP(c.Response(), c.Request())
	return nil
}

// @Summary Get Task Board and Project
// @Description Returns board and project identifiers for the specified task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} dto.TaskBoardProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/{id}/board-project [get]
func (h *TaskHandler) GetTaskBoardAndProject(c echo.Context) error {
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

	var boardIDPtr *string
	var projectIDPtr *string

	if task.Status != nil {
		boardID := task.Status.BoardID.String()
		boardIDPtr = &boardID

		if h.boardService != nil {
			board, err := h.boardService.GetBoard(c.Request().Context(), task.Status.BoardID)
			if err == nil && board != nil {
				projectID := board.ProjectID.String()
				projectIDPtr = &projectID
			}
		} else if task.Status.Board != nil {
			projectID := task.Status.Board.ProjectID.String()
			projectIDPtr = &projectID
		}
	}

	resp := response.TaskBoardProject{
		TaskID:    id.String(),
		BoardID:   boardIDPtr,
		ProjectID: projectIDPtr,
	}

	return respondSuccess(c, http.StatusOK, dto.TaskBoardProjectResponse{
		SuccessResponse: dto.NewSuccessResponse(resp),
	})
}

// @Summary Move Task between statuses
// @Description Move a task from its current status to another status
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task_id query string true "Task ID"
// @Param to_status_id query string true "Destination Status ID"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/move [get]
func (h *TaskHandler) MoveTaskBetweenStatuses(c echo.Context) error {
	taskID, err := parseUUID(c.QueryParam("task_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_task_id", "invalid task identifier"))
	}
	toStatusID, err := parseUUID(c.QueryParam("to_status_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_status_id", "invalid status identifier"))
	}
	task, err := h.service.MoveTaskBetweenStatuses(c.Request().Context(), taskID, toStatusID)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("task_move_failed", "failed to move task"))
	}
	return respondSuccess(c, http.StatusOK, presenter.ToTaskDTO(task))
}

// @Summary List Active Tasks by User
// @Description Retrieve a paginated list of active tasks assigned to a specific user
// @Tags Tasks
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.TasksListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/user/{user_id}/active [get]
func (h *TaskHandler) ListActiveTasksByUser(c echo.Context) error {
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_user_id", "invalid user identifier"))
	}
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
	page, err := h.service.ListActiveTasksByUser(c.Request().Context(), userID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}
	pagination, payload := presenter.MapTasksPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary List Tasks by User and Project
// @Description Retrieve a paginated list of tasks assigned to a specific user within a specific project
// @Tags Tasks
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param project_id path string true "Project ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.TasksListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/user/{user_id}/project/{project_id} [get]
func (h *TaskHandler) ListTasksByUserAndProject(c echo.Context) error {
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_user_id", "invalid user identifier"))
	}
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_project_id", "invalid project identifier"))
	}
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
	page, err := h.service.ListTasksByUserAndProject(c.Request().Context(), userID, projectID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}
	pagination, payload := presenter.MapTasksPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary List Tasks by Board
// @Description Retrieve a paginated list of tasks associated with a specific board
// @Accept json
// @Produce json
// @Tags Tasks
// @Param board_id path string true "Board ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.TasksListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /tasks/board/{board_id} [get]
func (h *TaskHandler) ListTasksByBoard(c echo.Context) error {
	boardID, err := parseUUID(c.Param("board_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_board_id", "invalid board identifier"))
	}
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
	page, err := h.service.ListTasksByBoard(c.Request().Context(), boardID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("tasks_fetch_failed", "failed to list tasks"))
	}
	pagination, payload := presenter.MapTasksPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}
