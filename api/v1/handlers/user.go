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

type UserHandler struct {
	service app.UserService
}

func NewUserHandler(service app.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func RegisterUserRoutes(group *echo.Group, service app.UserService) {
	handler := NewUserHandler(service)

	group.GET("/users", handler.ListUsers)
	group.GET("/users/:id", handler.GetUser)
	group.POST("/users", handler.CreateUser)
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	page := parsePositiveInt(c.QueryParam("page"), 1)
	pageSize := parsePositiveInt(c.QueryParam("page_size"), 20)

	result, err := h.service.ListUsers(c.Request().Context(), app.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("users_fetch_failed", "failed to list users"))
	}

	pagination, payload := presenter.MapUsersPage(result)
	response := dto.NewPaginatedResponse(payload, pagination)

	return c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	user, err := h.service.GetUser(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("user_not_found", "user not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("users_fetch_failed", "failed to get user"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToUserDTO(user)))
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req request.CreateUser
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	input := app.CreateUserInput{
		Email:         req.Email,
		IsActive:      req.IsActive,
		TgID:          req.TgID,
		TgUserID:      req.TgUserID,
		Profession:    req.Profession,
		EmailVerified: req.EmailVerified,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
	}

	user, err := h.service.CreateUser(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("user_create_failed", "failed to create user"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToUserDTO(user)))
}
