package handlers

import (
	"errors"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
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
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)

	secured := auth.Group("", middleware.KeycloakAuth(service))
	secured.POST("/logout", h.Logout)
	secured.GET("/me", h.Me)
	secured.GET("/validate", h.Validate)
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
	var req request.AuthLogin
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "email and password are required"))
	}

	result, err := h.service.Login(c.Request().Context(), strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_credentials", err.Error()))
		}
		return c.JSON(http.StatusUnauthorized, dto.NewError("invalid_credentials", "invalid credentials"))
	}

	resp := response.Auth{
		UserID:       result.UserID.String(),
		Email:        result.Email,
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresIn:    result.Tokens.ExpiresIn,
		RefreshExp:   result.Tokens.RefreshExpiresIn,
		TokenType:    result.Tokens.TokenType,
		ExpiresAt:    result.Tokens.ExpiresAt,
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(resp))
}

// @Summary Logout
// @Description Invalidate the current access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {access_token}"
// @Success 200 {object} dto.SuccessResponse[map[string]string]
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	token, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_authorization_header", err.Error()))
	}

	if err := h.service.Logout(c.Request().Context(), token); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("logout_failed", "failed to logout"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(map[string]string{"message": "logout successful"}))
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
	token, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.NewError("invalid_authorization_header", err.Error()))
	}

	info, err := h.service.GetUserInfo(c.Request().Context(), token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.NewError("invalid_token", "token validation failed"))
	}

	resp := response.UserInfo{
		Sub:               info.Subject,
		Name:              info.Name,
		PreferredUsername: info.PreferredUsername,
		GivenName:         info.GivenName,
		FamilyName:        info.FamilyName,
		Email:             info.Email,
		EmailVerified:     info.EmailVerified,
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(resp))
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
	var req request.AuthRefresh
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "refresh_token is required"))
	}

	result, err := h.service.RefreshToken(c.Request().Context(), strings.TrimSpace(req.RefreshToken))
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_refresh_token", err.Error()))
		}
		return c.JSON(http.StatusUnauthorized, dto.NewError("invalid_refresh_token", "invalid or expired refresh token"))
	}

	resp := response.Refresh{
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresIn:    result.Tokens.ExpiresIn,
		RefreshExp:   result.Tokens.RefreshExpiresIn,
		TokenType:    result.Tokens.TokenType,
		ExpiresAt:    result.Tokens.ExpiresAt,
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(resp))
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
	token, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.NewError("invalid_authorization_header", err.Error()))
	}

	if err := h.service.ValidateToken(c.Request().Context(), token); err != nil {
		return c.JSON(http.StatusUnauthorized, dto.NewError("token_invalid", "token invalid"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.TokenValidation{Message: "Token is valid"}))
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("authorization header is required")
	}
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		token := strings.TrimSpace(header[len(prefix):])
		if token == "" {
			return "", errors.New("authorization header contains empty token")
		}
		return token, nil
	}
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return "", errors.New("authorization header contains empty token")
	}
	return trimmed, nil
}
