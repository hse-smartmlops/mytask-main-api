package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

type ForumMessageHandler struct {
	service ports.ForumMessageService
}

func NewForumMessageHandler(service ports.ForumMessageService) *ForumMessageHandler {
	return &ForumMessageHandler{service: service}
}

// @Summary Register Forum Message Routes
// @Description Register routes for forum message management
// @Tags ForumMessages
func RegisterForumMessageRoutes(group *echo.Group, service ports.ForumMessageService) {
	handler := NewForumMessageHandler(service)

	group.GET("/forum-messages", handler.ListMessages)
	group.GET("/forum-messages/:id", handler.GetMessage)
	group.GET("/problems/:problem_id/forum-messages", handler.ListMessagesByProblem)
	group.POST("/forum-messages", handler.CreateMessage)
	group.PATCH("/forum-messages/:id", handler.UpdateMessage)
	group.DELETE("/forum-messages/:id", handler.DeleteMessage)
}

// @Summary List Forum Messages
// @Description Retrieve a paginated list of forum messages
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ForumMessagesListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /forum-messages [get]
func (h *ForumMessageHandler) ListMessages(c echo.Context) error {
	params := ports.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListMessages(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("forum_messages_fetch_failed", "failed to list forum messages"))
	}

	pagination, payload := presenter.MapForumMessagesPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

// @Summary Get Forum Message
// @Description Retrieve a specific forum message by its ID
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "Forum Message ID"
// @Success 200 {object} dto.ForumMessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /forum-messages/{id} [get]
func (h *ForumMessageHandler) GetMessage(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid forum message identifier"))
	}

	msg, err := h.service.GetMessage(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("forum_message_not_found", "forum message not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("forum_messages_fetch_failed", "failed to get forum message"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToForumMessageDTO(msg)))
}

// @Summary List Forum Messages by Problem
// @Description Retrieve a paginated list of forum messages for a specific problem
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param problem_id path string true "Problem ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ForumMessagesListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /problems/{problem_id}/forum-messages [get]
func (h *ForumMessageHandler) ListMessagesByProblem(c echo.Context) error {
	problemID, err := parseUUID(c.Param("problem_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid problem identifier"))
	}

	params := ports.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListMessagesByProblem(c.Request().Context(), problemID, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("forum_messages_fetch_failed", "failed to list forum messages"))
	}

	pagination, payload := presenter.MapForumMessagesPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

// @Summary Create Forum Message
// @Description Create a new forum message
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param forum_message body request.CreateForumMessage true "Forum Message data"
// @Success 201 {object} dto.ForumMessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /forum-messages [post]
func (h *ForumMessageHandler) CreateMessage(c echo.Context) error {
	var req request.CreateForumMessage
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	problemID, err := parseUUID(req.ProblemID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "problem_id must be a valid UUID"))
	}
	creatorID, err := parseUUIDPointer(req.CreatorID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "creator_id must be a valid UUID"))
	}

	input := ports.CreateForumMessageInput{
		ProblemID:   problemID,
		Description: req.Description,
		CreatorID:   creatorID,
	}

	msg, err := h.service.CreateMessage(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("forum_message_create_failed", "failed to create forum message"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToForumMessageDTO(msg)))
}

// @Summary Update Forum Message
// @Description Update an existing forum message
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "Forum Message ID"
// @Param forum_message body request.UpdateForumMessage true "Forum Message data"
// @Success 200 {object} dto.ForumMessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /forum-messages/{id} [patch]
func (h *ForumMessageHandler) UpdateMessage(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid forum message identifier"))
	}

	var req request.UpdateForumMessage
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	input := ports.UpdateForumMessageInput{
		Description: req.Description,
	}

	msg, err := h.service.UpdateMessage(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("forum_message_not_found", "forum message not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("forum_message_update_failed", "failed to update forum message"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToForumMessageDTO(msg)))
}

// @Summary Delete Forum Message
// @Description Delete a forum message by its ID
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "Forum Message ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /forum-messages/{id} [delete]
func (h *ForumMessageHandler) DeleteMessage(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid forum message identifier"))
	}

	if err := h.service.DeleteMessage(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("forum_message_not_found", "forum message not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("forum_message_delete_failed", "failed to delete forum message"))
	}

	return c.NoContent(http.StatusNoContent)
}
