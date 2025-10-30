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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	_ request.CreateBoard
	_ request.UpdateBoard
)

type BoardHandler struct {
	service ports.BoardService
}

func NewBoardHandler(service ports.BoardService) *BoardHandler {
	return &BoardHandler{service: service}
}

// @Summary Register Board Routes
// @Description Register routes for board management
// @Tags Boards
func RegisterBoardRoutes(group *echo.Group, service ports.BoardService) {
	handler := NewBoardHandler(service)

	bgroup := group.Group("/board")
	{
		bgroup.GET("", handler.ListBoards)
		bgroup.GET("/:id", handler.GetBoard)
		bgroup.GET("/project/:project_id", handler.ListBoardsByProject)
		bgroup.POST("", handler.CreateBoard)
		bgroup.PATCH("/:id", handler.UpdateBoard)
		bgroup.DELETE("/:id", handler.DeleteBoard)
	}
}

// @Summary List Boards
// @Description Retrieve a paginated list of boards
// @Tags Boards
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.BoardsListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /board [get]
func (h *BoardHandler) ListBoards(c echo.Context) error {
	params, err := paginationParams(c)
	if err != nil {
		code := "invalid_pagination"
		if errors.Is(err, errMissingPagination) {
			code = "pagination_required"
		}
		return respondError(c, http.StatusBadRequest, dto.NewError(code, err.Error()))
	}

	page, err := h.service.ListBoards(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("boards_fetch_failed", "failed to list boards"))
	}

	pagination, payload := presenter.MapBoardsPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Board
// @Description Retrieve a board by its ID
// @Tags Boards
// @Accept json
// @Produce json
// @Param id path string true "Board ID"
// @Success 200 {object} dto.BoardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /boards/{id} [get]
func (h *BoardHandler) GetBoard(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}

	board, err := h.service.GetBoard(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("board_not_found", "board not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("boards_fetch_failed", "failed to get board"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToBoardDTO(board))
}

// @Summary Create Board
// @Description Create a new board
// @Tags Boards
// @Accept json
// @Produce json
// @Param board body request.CreateBoard true "Board data"
// @Success 201 {object} dto.BoardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /boards [post]
func (h *BoardHandler) CreateBoard(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateBoardPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	projectID, _ := uuid.Parse(req.ProjectID)
	board, err := h.service.CreateBoard(c.Request().Context(), ports.CreateBoardInput{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("board_create_failed", "failed to create board"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToBoardDTO(board))
}

// @Summary Update Board
// @Description Update an existing board
// @Tags Boards
// @Accept json
// @Produce json
// @Param id path string true "Board ID"
// @Param board body request.UpdateBoard true "Board data"
// @Success 200 {object} dto.BoardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /boards/{id} [patch]
func (h *BoardHandler) UpdateBoard(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateUpdateBoardPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	board, err := h.service.UpdateBoard(c.Request().Context(), id, ports.UpdateBoardInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("board_not_found", "board not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("board_update_failed", "failed to update board"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToBoardDTO(board))
}

// @Summary Delete Board
// @Description Delete a board by its ID
// @Tags Boards
// @Accept json
// @Produce json
// @Param id path string true "Board ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /boards/{id} [delete]
func (h *BoardHandler) DeleteBoard(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}

	if err := h.service.DeleteBoard(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("board_not_found", "board not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("board_delete_failed", "failed to delete board"))
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary List Boards by Project
// @Description Retrieve a list of boards for a specific project
// @Tags Boards
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.BoardsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /board/project/{project_id} [get]
func (h *BoardHandler) ListBoardsByProject(c echo.Context) error {
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	params, err := paginationParams(c)
	if err != nil {
		code := "invalid_pagination"
		if errors.Is(err, errMissingPagination) {
			code = "pagination_required"
		}
		return respondError(c, http.StatusBadRequest, dto.NewError(code, err.Error()))
	}

	boardsPage, err := h.service.ListBoardsByProject(c.Request().Context(), projectID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("boards_fetch_failed", "failed to list boards"))
	}

	pagination, payload := presenter.MapBoardsPage(boardsPage)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}
