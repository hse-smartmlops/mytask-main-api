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
	_ request.CreateForumMessage
	_ request.UpdateForumMessage
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

	fgroup := group.Group("/forum-message")
	{
		fgroup.GET("", handler.ListMessages)
		fgroup.GET("/:id", handler.GetMessage)
		fgroup.GET("/problem/:problem_id", handler.ListMessagesByProblem)
		fgroup.POST("", handler.CreateMessage)
		fgroup.PATCH("/:id", handler.UpdateMessage)
		fgroup.DELETE("/:id", handler.DeleteMessage)
	}
}

// @Summary List Forum Messages
// @Description Retrieve a paginated list of forum messages
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.ForumMessagesListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /forum-messages [get]
func (h *ForumMessageHandler) ListMessages(c echo.Context) error {
	params, err := paginationParams(c)
	if err != nil {
		code := "invalid_pagination"
		if errors.Is(err, errMissingPagination) {
			code = "pagination_required"
		}
		return respondError(c, http.StatusBadRequest, dto.NewError(code, err.Error()))
	}

	page, err := h.service.ListMessages(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("forum_messages_fetch_failed", "failed to list forum messages"))
	}

	pagination, payload := presenter.MapForumMessagesPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
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
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid forum message identifier"))
	}

	msg, err := h.service.GetMessage(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("forum_message_not_found", "forum message not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("forum_messages_fetch_failed", "failed to get forum message"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToForumMessageDTO(msg))
}

// @Summary List Forum Messages by Problem
// @Description Retrieve a paginated list of forum messages for a specific problem
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param problem_id path string true "Problem ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.ForumMessagesListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /problems/{problem_id}/forum-messages [get]
func (h *ForumMessageHandler) ListMessagesByProblem(c echo.Context) error {
	problemID, err := parseUUID(c.Param("problem_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid problem identifier"))
	}

	params, err := paginationParams(c)
	if err != nil {
		code := "invalid_pagination"
		if errors.Is(err, errMissingPagination) {
			code = "pagination_required"
		}
		return respondError(c, http.StatusBadRequest, dto.NewError(code, err.Error()))
	}

	page, err := h.service.ListMessagesByProblem(c.Request().Context(), problemID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("forum_messages_fetch_failed", "failed to list forum messages"))
	}

	pagination, payload := presenter.MapForumMessagesPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
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
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateForumMessagePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.CreateForumMessageInput{
		ProblemID:   req.ProblemUUID,
		Description: req.CleanDesc,
		CreatorID:   req.CreatorUUID,
	}

	msg, err := h.service.CreateMessage(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("forum_message_create_failed", "failed to create forum message"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToForumMessageDTO(msg))
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
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid forum message identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateUpdateForumMessagePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateForumMessageInput{
		Description: req.CleanDesc,
	}

	msg, err := h.service.UpdateMessage(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("forum_message_not_found", "forum message not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("forum_message_update_failed", "failed to update forum message"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToForumMessageDTO(msg))
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
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid forum message identifier"))
	}

	if err := h.service.DeleteMessage(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("forum_message_not_found", "forum message not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("forum_message_delete_failed", "failed to delete forum message"))
	}

	return c.NoContent(http.StatusNoContent)
}
