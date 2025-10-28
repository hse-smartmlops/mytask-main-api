package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

var (
	_ request.CreateStatus
	_ request.UpdateStatus
)

type StatusHandler struct {
	service ports.StatusService
}

func NewStatusHandler(service ports.StatusService) *StatusHandler {
	return &StatusHandler{service: service}
}

// @Summary Register Status Routes
// @Description Register routes for status management
// @Tags Statuses
func RegisterStatusRoutes(group *echo.Group, service ports.StatusService) {
	handler := NewStatusHandler(service)

	sgroup := group.Group("/status")
	{
		sgroup.GET("", handler.ListStatuses)
		sgroup.GET("/:id", handler.GetStatus)
		sgroup.GET("/board/:board_id", handler.ListStatusesByBoard)
		sgroup.POST("", handler.CreateStatus)
		sgroup.PATCH("/:id", handler.UpdateStatus)
		sgroup.DELETE("/:id", handler.DeleteStatus)
	}

}

// @Summary List Statuses
// @Description Retrieve a paginated list of statuses
// @Tags Statuses
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.StatusesListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /statuses [get]
func (h *StatusHandler) ListStatuses(c echo.Context) error {
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
	page, err := h.service.ListStatuses(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("statuses_fetch_failed", "failed to list statuses"))
	}
	pagination, payload := presenter.MapStatusesPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Status
// @Description Retrieve a specific status by its ID
// @Tags Statuses
// @Accept json
// @Produce json
// @Param id path string true "Status ID"
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /statuses/{id} [get]
func (h *StatusHandler) GetStatus(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid status identifier"))
	}

	status, err := h.service.GetStatus(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("status_not_found", "status not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("statuses_fetch_failed", "failed to get status"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToStatusDTO(status))
}

// @Summary List Statuses By Board
// @Description Retrieve a list of statuses associated with a specific board
// @Tags Statuses
// @Accept json
// @Produce json
// @Param board_id path string true "Board ID"
// @Success 200 {object} dto.StatusesListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /boards/{board_id}/statuses [get]
func (h *StatusHandler) ListStatusesByBoard(c echo.Context) error {
	boardIDParam := c.Param("board_id")
	if boardIDParam == "" {
		boardIDParam = c.Param("id")
	}

	boardID, err := parseUUID(boardIDParam)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}
	// TODO add pagination
	statuses, err := h.service.ListStatusesByBoard(c.Request().Context(), boardID, ports.PaginationParams{
		Page:     1,
		PageSize: 0,
	})
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("statuses_fetch_failed", "failed to list statuses"))
	}

	items := make([]response.Status, len(statuses.Items))
	for i := range statuses.Items {
		items[i] = presenter.ToStatusDTO(&statuses.Items[i])
	}

	return respondSuccess(c, http.StatusOK, items)
}

// @Summary Create Status
// @Description Create a new status
// @Tags Statuses
// @Accept json
// @Produce json
// @Param status body request.CreateStatus true "Status creation payload"
// @Success 201 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /statuses [post]
func (h *StatusHandler) CreateStatus(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateStatusPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.CreateStatusInput{
		BoardID:   req.BoardUUID,
		Name:      req.Name,
		Key:       req.Key,
		Color:     req.Color,
		IsDefault: req.IsDefault,
		IsActive:  req.IsActive,
		IsOpen:    req.IsOpen,
		SortOrder: req.SortOrder,
	}

	status, err := h.service.CreateStatus(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("status_create_failed", "failed to create status"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToStatusDTO(status))
}

// @Summary Update Status
// @Description Update an existing status
// @Tags Statuses
// @Accept json
// @Produce json
// @Param id path string true "Status ID"
// @Param status body request.UpdateStatus true "Status data"
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /statuses/{id} [patch]
func (h *StatusHandler) UpdateStatus(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid status identifier"))
	}
	req, err := middleware.BindAndValidate(c, middleware.ValidateUpdateStatusPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateStatusInput{
		Name:      req.Name,
		Key:       req.Key,
		Color:     req.Color,
		IsDefault: req.IsDefault,
		IsActive:  req.IsActive,
		IsOpen:    req.IsOpen,
		SortOrder: req.SortOrder,
	}

	status, err := h.service.UpdateStatus(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("status_not_found", "status not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("status_update_failed", "failed to update status"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToStatusDTO(status))
}

// @Summary Delete Status
// @Description Delete a status by its ID
// @Tags Statuses
// @Accept json
// @Produce json
// @Param id path string true "Status ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /statuses/{id} [delete]
func (h *StatusHandler) DeleteStatus(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid status identifier"))
	}

	if err := h.service.DeleteStatus(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("status_not_found", "status not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("status_delete_failed", "failed to delete status"))
	}

	return c.NoContent(http.StatusNoContent)
}
