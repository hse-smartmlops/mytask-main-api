package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BoardHandler struct {
	service app.BoardService
}

func NewBoardHandler(service app.BoardService) *BoardHandler {
	return &BoardHandler{service: service}
}

func RegisterBoardRoutes(group *echo.Group, service app.BoardService) {
	handler := NewBoardHandler(service)

	group.GET("/boards", handler.ListBoards)
	group.GET("/boards/:id", handler.GetBoard)
	group.POST("/boards", handler.CreateBoard)
	group.PATCH("/boards/:id", handler.UpdateBoard)
	group.DELETE("/boards/:id", handler.DeleteBoard)
	group.GET("/projects/:project_id/boards", handler.ListBoardsByProject)
}

func (h *BoardHandler) ListBoards(c echo.Context) error {
	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListBoards(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("boards_fetch_failed", "failed to list boards"))
	}

	pagination, payload := presenter.MapBoardsPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *BoardHandler) GetBoard(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}

	board, err := h.service.GetBoard(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("board_not_found", "board not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("boards_fetch_failed", "failed to get board"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToBoardDTO(board)))
}

func (h *BoardHandler) CreateBoard(c echo.Context) error {
	var req request.CreateBoard
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "project_id must be a valid UUID"))
	}

	board, err := h.service.CreateBoard(c.Request().Context(), app.CreateBoardInput{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("board_create_failed", "failed to create board"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToBoardDTO(board)))
}

func (h *BoardHandler) UpdateBoard(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}

	var req request.UpdateBoard
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	board, err := h.service.UpdateBoard(c.Request().Context(), id, app.UpdateBoardInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("board_not_found", "board not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("board_update_failed", "failed to update board"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToBoardDTO(board)))
}

func (h *BoardHandler) DeleteBoard(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid board identifier"))
	}

	if err := h.service.DeleteBoard(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("board_not_found", "board not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("board_delete_failed", "failed to delete board"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *BoardHandler) ListBoardsByProject(c echo.Context) error {
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	boards, err := h.service.ListBoardsByProject(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("boards_fetch_failed", "failed to list boards"))
	}

	items := make([]response.Board, len(boards))
	for i := range boards {
		items[i] = presenter.ToBoardDTO(&boards[i])
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(items))
}
