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

type ProblemHandler struct {
	service ports.ProblemService
}

func NewProblemHandler(service ports.ProblemService) *ProblemHandler {
	return &ProblemHandler{service: service}
}

// @Summary Register Problem Routes
// @Description Register routes for problem management
// @Tags Problems
func RegisterProblemRoutes(group *echo.Group, service ports.ProblemService) {
	handler := NewProblemHandler(service)

	group.GET("/problems", handler.ListProblems)
	group.GET("/problems/:id", handler.GetProblem)
	group.POST("/problems", handler.CreateProblem)
	group.PATCH("/problems/:id", handler.UpdateProblem)
	group.DELETE("/problems/:id", handler.DeleteProblem)
}

// @Summary List Problems
// @Description Retrieve a paginated list of problems
// @Tags Problems
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ProblemsListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /problems [get]
func (h *ProblemHandler) ListProblems(c echo.Context) error {
	pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: pageNum, PageSize: size}

	page, err := h.service.ListProblems(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("problems_fetch_failed", "failed to list problems"))
	}

	pagination, payload := presenter.MapProblemsPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Problem
// @Description Retrieve a problem by its ID
// @Tags Problems
// @Accept json
// @Produce json
// @Param id path string true "Problem ID"
// @Success 200 {object} dto.ProblemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /problems/{id} [get]
func (h *ProblemHandler) GetProblem(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid problem identifier"))
	}

	problem, err := h.service.GetProblem(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("problem_not_found", "problem not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("problems_fetch_failed", "failed to get problem"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToProblemDTO(problem))
}

// @Summary Create Problem
// @Description Create a new problem
// @Tags Problems
// @Accept json
// @Produce json
// @Param problem body request.CreateProblem true "Problem data"
// @Success 201 {object} dto.ProblemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /problems [post]
func (h *ProblemHandler) CreateProblem(c echo.Context) error {
	req, err := middleware.BindAndValidate[request.CreateProblem](c, middleware.ValidateCreateProblemPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	creatorID, _ := parseUUIDPointer(req.CreatorID)

	input := ports.CreateProblemInput{
		Description: req.Description,
		CreatorID:   creatorID,
		Name:        req.Name,
	}

	problem, err := h.service.CreateProblem(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("problem_create_failed", "failed to create problem"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToProblemDTO(problem))
}

// @Summary Update Problem
// @Description Update an existing problem
// @Tags Problems
// @Accept json
// @Produce json
// @Param id path string true "Problem ID"
// @Param problem body request.UpdateProblem true "Problem data"
// @Success 200 {object} dto.ProblemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /problems/{id} [patch]
func (h *ProblemHandler) UpdateProblem(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid problem identifier"))
	}

	req, err := middleware.BindAndValidate[request.UpdateProblem](c, middleware.ValidateUpdateProblemPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateProblemInput{
		Description: req.Description,
		Name:        req.Name,
	}

	problem, err := h.service.UpdateProblem(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("problem_not_found", "problem not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("problem_update_failed", "failed to update problem"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToProblemDTO(problem))
}

// @Summary Delete Problem
// @Description Delete a problem by its ID
// @Tags Problems
// @Accept json
// @Produce json
// @Param id path string true "Problem ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /problems/{id} [delete]
func (h *ProblemHandler) DeleteProblem(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid problem identifier"))
	}

	if err := h.service.DeleteProblem(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("problem_not_found", "problem not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("problem_delete_failed", "failed to delete problem"))
	}

	return c.NoContent(http.StatusNoContent)
}
