package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	service ports.RoleService
}

func NewRoleHandler(service ports.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

// @Summary Register Role Routes
// @Description Register routes for role management
// @Tags Roles
func RegisterRoleRoutes(group *echo.Group, service ports.RoleService) {
	handler := NewRoleHandler(service)

	group.GET("/roles", handler.ListRoles)
	group.GET("/roles/:id", handler.GetRole)
	group.POST("/roles", handler.CreateRole)
	group.PATCH("/roles/:id", handler.UpdateRole)
	group.DELETE("/roles/:id", handler.DeleteRole)
}

// @Summary List Roles
// @Description Retrieve a paginated list of roles
// @Tags Roles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.RolesListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles [get]
func (h *RoleHandler) ListRoles(c echo.Context) error {
	pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: pageNum, PageSize: size}

	page, err := h.service.ListRoles(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("roles_fetch_failed", "failed to list roles"))
	}

	pagination, payload := presenter.MapRolesPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Role
// @Description Retrieve a role by its ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} dto.RoleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRole(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid role identifier"))
	}

	role, err := h.service.GetRole(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("role_not_found", "role not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("roles_fetch_failed", "failed to get role"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToRoleDTO(role))
}

// @Summary Create Role
// @Description Create a new role
// @Tags Roles
// @Accept json
// @Produce json
// @Param role body request.CreateRole true "Role creation payload"
// @Success 201 {object} dto.RoleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateRolePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	role, err := h.service.CreateRole(c.Request().Context(), ports.CreateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("role_create_failed", "failed to create role"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToRoleDTO(role))
}

// @Summary Update Role
// @Description Update an existing role
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body request.UpdateRole true "Role data"
// @Success 200 {object} dto.RoleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles/{id} [patch]
func (h *RoleHandler) UpdateRole(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid role identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateUpdateRolePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	role, err := h.service.UpdateRole(c.Request().Context(), id, ports.UpdateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("role_not_found", "role not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("role_update_failed", "failed to update role"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToRoleDTO(role))
}

// @Summary Delete Role
// @Description Delete a role by its ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid role identifier"))
	}

	err = h.service.DeleteRole(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("role_not_found", "role not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("role_delete_failed", "failed to delete role"))
	}

	return c.NoContent(http.StatusNoContent)
}
