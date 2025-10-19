package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ProjectHandler struct {
	service app.ProjectService
}

func NewProjectHandler(service app.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func RegisterProjectRoutes(group *echo.Group, service app.ProjectService) {
	handler := NewProjectHandler(service)

	group.GET("/projects", handler.ListProjects)
	group.GET("/projects/:id", handler.GetProject)
	group.POST("/projects", handler.CreateProject)
	group.PATCH("/projects/:id", handler.UpdateProject)
	group.DELETE("/projects/:id", handler.DeleteProject)
}

func (h *ProjectHandler) ListProjects(c echo.Context) error {
	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListProjects(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("projects_fetch_failed", "failed to list projects"))
	}

	pagination, payload := presenter.MapProjectsPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *ProjectHandler) GetProject(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	project, err := h.service.GetProject(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("project_not_found", "project not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("projects_fetch_failed", "failed to get project"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToProjectDTO(project)))
}

func (h *ProjectHandler) CreateProject(c echo.Context) error {
	var req request.CreateProject
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	var createdBy *uuid.UUID
	if req.CreatedBy != nil {
		id, err := uuid.Parse(*req.CreatedBy)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "created_by must be a valid UUID"))
		}
		createdBy = &id
	}

	project, err := h.service.CreateProject(c.Request().Context(), app.CreateProjectInput{
		Name:            req.Name,
		Description:     req.Description,
		GitlabProjectID: req.GitlabProjectID,
		GitlabURL:       req.GitlabURL,
		CreatedBy:       createdBy,
		Status:          req.Status,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("project_create_failed", "failed to create project"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToProjectDTO(project)))
}

func (h *ProjectHandler) UpdateProject(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	var req request.UpdateProject
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	project, err := h.service.UpdateProject(c.Request().Context(), id, app.UpdateProjectInput{
		Name:            req.Name,
		Description:     req.Description,
		GitlabProjectID: req.GitlabProjectID,
		GitlabURL:       req.GitlabURL,
		Status:          req.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("project_not_found", "project not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("project_update_failed", "failed to update project"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToProjectDTO(project)))
}

func (h *ProjectHandler) DeleteProject(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	if err := h.service.DeleteProject(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("project_not_found", "project not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("project_delete_failed", "failed to delete project"))
	}

	return c.NoContent(http.StatusNoContent)
}
