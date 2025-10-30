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
	_ request.AuthLogin
	_ request.AuthRefresh
)

type AuthHandler struct {
	service ports.AuthService
}

func NewAuthHandler(service ports.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// @Summary Register Auth Routes
// @Description Register routes for authentication
// @Tags Auth
func RegisterAuthRoutes(group *echo.Group, service ports.AuthService) {
	if service == nil {
		return
	}
	h := NewAuthHandler(service)

	auth := group.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)

		secured := auth.Group("", middleware.KeycloakAuth(service))
		{
			secured.POST("/logout", h.Logout)
			secured.GET("/me", h.Me)
			secured.GET("/validate", h.Validate)
		}
	}
}

// @Summary Login
// @Description Authenticate user and obtain access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body request.AuthLogin true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateAuthLoginPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	result, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_credentials", err.Error()))
		}
		return respondError(c, http.StatusUnauthorized, dto.NewError("invalid_credentials", "invalid credentials"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToAuthResponse(result))
}

// @Summary Logout
// @Description Invalidate the current access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {access_token}"
// @Success 200 {object} dto.LogoutMessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	token, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_authorization_header", err.Error()))
	}

	if err := h.service.Logout(c.Request().Context(), token); err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("logout_failed", "failed to logout"))
	}

	return respondSuccess(c, http.StatusOK, response.LogoutMessage{Message: "sucсessfull logout"})
}

// @Summary Get User Info
// @Description Retrieve information about the authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {access_token}"
// @Success 200 {object} dto.UserInfoResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	token := resolveAuthToken(c)
	if token == "" {
		return respondError(c, http.StatusUnauthorized, dto.NewError("invalid_authorization_header", "authorization header is required"))
	}

	info, err := h.service.GetUserInfo(c.Request().Context(), token)
	if err != nil {
		return respondError(c, http.StatusUnauthorized, dto.NewError("invalid_token", "token validation failed"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToUserInfoResponse(info))
}

// @Summary Refresh Token
// @Description Refresh access and refresh tokens using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body request.AuthRefresh true "Refresh token"
// @Success 200 {object} dto.RefreshResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateAuthRefreshPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	result, err := h.service.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_refresh_token", err.Error()))
		}
		return respondError(c, http.StatusUnauthorized, dto.NewError("invalid_refresh_token", "invalid or expired refresh token"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToRefreshResponse(result))
}

// @Summary Validate Token
// @Description Validate the provided access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {access_token}"
// @Success 200 {object} dto.TokenValidationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/validate [get]
func (h *AuthHandler) Validate(c echo.Context) error {
	token := resolveAuthToken(c)
	if token == "" {
		return respondError(c, http.StatusUnauthorized, dto.NewError("invalid_authorization_header", "authorization header is required"))
	}

	if err := h.service.ValidateToken(c.Request().Context(), token); err != nil {
		return respondError(c, http.StatusUnauthorized, dto.NewError("token_invalid", "token invalid"))
	}

	return respondSuccess(c, http.StatusOK, response.TokenValidation{Message: "Token is valid"})
}
