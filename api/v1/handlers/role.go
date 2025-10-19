package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	service app.RoleService
}

func NewRoleHandler(service app.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func RegisterRoleRoutes(group *echo.Group, service app.RoleService) {
	handler := NewRoleHandler(service)

	group.GET("/roles", handler.ListRoles)
	group.GET("/roles/:id", handler.GetRole)
	group.POST("/roles", handler.CreateRole)
	group.PATCH("/roles/:id", handler.UpdateRole)
	group.DELETE("/roles/:id", handler.DeleteRole)
}

func (h *RoleHandler) ListRoles(c echo.Context) error {
	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListRoles(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("roles_fetch_failed", "failed to list roles"))
	}

	pagination, payload := presenter.MapRolesPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *RoleHandler) GetRole(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid role identifier"))
	}

	role, err := h.service.GetRole(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("role_not_found", "role not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("roles_fetch_failed", "failed to get role"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToRoleDTO(role)))
}

func (h *RoleHandler) CreateRole(c echo.Context) error {
	var req request.CreateRole
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	role, err := h.service.CreateRole(c.Request().Context(), app.CreateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "role name is required"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("role_create_failed", "failed to create role"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToRoleDTO(role)))
}

func (h *RoleHandler) UpdateRole(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid role identifier"))
	}

	var req request.UpdateRole
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	role, err := h.service.UpdateRole(c.Request().Context(), id, app.UpdateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("role_not_found", "role not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("role_update_failed", "failed to update role"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToRoleDTO(role)))
}

func (h *RoleHandler) DeleteRole(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid role identifier"))
	}

	err = h.service.DeleteRole(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("role_not_found", "role not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("role_delete_failed", "failed to delete role"))
	}

	return c.NoContent(http.StatusNoContent)
}
