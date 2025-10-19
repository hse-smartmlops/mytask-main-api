package handlers

import (
	"errors"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

type StatusHandler struct {
	service app.StatusService
}

func NewStatusHandler(service app.StatusService) *StatusHandler {
	return &StatusHandler{service: service}
}

func RegisterStatusRoutes(group *echo.Group, service app.StatusService) {
	handler := NewStatusHandler(service)

	group.GET("/statuses", handler.ListStatuses)
	group.GET("/statuses/:id", handler.GetStatus)
	group.GET("/boards/:board_id/statuses", handler.ListStatusesByBoard)
	group.POST("/statuses", handler.CreateStatus)
	group.PATCH("/statuses/:id", handler.UpdateStatus)
	group.DELETE("/statuses/:id", handler.DeleteStatus)
}

func (h *StatusHandler) ListStatuses(c echo.Context) error {
	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListStatuses(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("statuses_fetch_failed", "failed to list statuses"))
	}

	pagination, payload := presenter.MapStatusesPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *StatusHandler) GetStatus(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid status identifier"))
	}

	status, err := h.service.GetStatus(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("status_not_found", "status not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("statuses_fetch_failed", "failed to get status"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToStatusDTO(status)))
}

func (h *StatusHandler) ListStatusesByBoard(c echo.Context) error {
	boardID, err := parseUUID(c.Param("board_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}

	statuses, err := h.service.ListStatusesByBoard(c.Request().Context(), boardID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("statuses_fetch_failed", "failed to list statuses"))
	}

	items := make([]response.Status, len(statuses))
	for i := range statuses {
		items[i] = presenter.ToStatusDTO(&statuses[i])
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(items))
}

func (h *StatusHandler) CreateStatus(c echo.Context) error {
	var req request.CreateStatus
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	boardID, err := parseUUID(req.BoardID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "board_id must be a valid UUID"))
	}

	input := app.CreateStatusInput{
		BoardID:   boardID,
		Name:      req.Name,
		Key:       sanitizeStringPtr(req.Key),
		Color:     sanitizeStringPtr(req.Color),
		IsDefault: req.IsDefault,
		IsActive:  req.IsActive,
		IsOpen:    req.IsOpen,
		SortOrder: req.SortOrder,
	}

	status, err := h.service.CreateStatus(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("status_create_failed", "failed to create status"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToStatusDTO(status)))
}

func (h *StatusHandler) UpdateStatus(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid status identifier"))
	}

	var req request.UpdateStatus
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	input := app.UpdateStatusInput{
		Name:      sanitizeStringPtr(req.Name),
		Key:       sanitizeStringPtr(req.Key),
		Color:     sanitizeStringPtr(req.Color),
		IsDefault: req.IsDefault,
		IsActive:  req.IsActive,
		IsOpen:    req.IsOpen,
		SortOrder: req.SortOrder,
	}

	status, err := h.service.UpdateStatus(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("status_not_found", "status not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("status_update_failed", "failed to update status"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToStatusDTO(status)))
}

func (h *StatusHandler) DeleteStatus(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid status identifier"))
	}

	if err := h.service.DeleteStatus(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("status_not_found", "status not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("status_delete_failed", "failed to delete status"))
	}

	return c.NoContent(http.StatusNoContent)
}

func sanitizeStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	res := trimmed
	return &res
}
