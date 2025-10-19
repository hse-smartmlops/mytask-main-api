package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service ports.UserService
}

func NewUserHandler(service ports.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// @Summary Register User Routes
// @Description Register routes for user management
// @Tags Users
func RegisterUserRoutes(group *echo.Group, service ports.UserService) {
	handler := NewUserHandler(service)

	group.GET("/users", handler.ListUsers)
	group.GET("/users/:id", handler.GetUser)
	group.POST("/users", handler.CreateUser)
}

// @Summary List Users
// @Description Retrieve a paginated list of users
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.UsersListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	page := parsePositiveInt(c.QueryParam("page"), 1)
	pageSize := parsePositiveInt(c.QueryParam("page_size"), 20)

	result, err := h.service.ListUsers(c.Request().Context(), ports.PaginationParams{
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

// @Summary Get User
// @Description Retrieve a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [get]
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

// @Summary Create User
// @Description Create a new user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body request.CreateUser true "User creation payload"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req request.CreateUser
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	input := ports.CreateUserInput{
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
