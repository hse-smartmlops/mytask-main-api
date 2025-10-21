package controller

import (
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/service"
	utils "emplacc-api/internal/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func RegisterAuthRoutes(e *echo.Echo, authService service.AuthService) {
	controller := NewAuthController(authService)
	authGroup := e.Group("/auth")
	authGroup.POST("/login", controller.Login)
	authGroup.POST("/logout", controller.Logout)
	authGroup.GET("/me", controller.Me)
	authGroup.GET("/validate", controller.ValidateToken)
	authGroup.POST("/refresh", controller.RefreshToken)
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Аутентифицирует пользователя по email и паролю через Keycloak
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body request.LoginRequest true "Данные для входа"
// @Success 200 {object} response.AuthResponse "Успешная аутентификация"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Router /auth/login [post]
func (ac *AuthController) Login(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	authResp, err := ac.authService.Login(req.Email, req.Password)
	if err != nil {
		log.Printf("Authorization error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	resp := response.AuthResponse{
		UserID:       utils.GetString(authResp.UserID),
		Email:        utils.GetString(authResp.Email),
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    authResp.ExpiresIn,
		RefreshExp:   authResp.RefreshExp,
		TokenType:    authResp.TokenType,
		ExpiresAt:    authResp.ExpiresAt,
	}

	return c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary Выход из системы
// @Description Выполняет выход пользователя из системы, завершая сессию в Keycloak
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Refresh token"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Успешный выход из системы"
// @Failure 400 {object} map[string]string "Отсутствует токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при выходе"
// @Router /auth/logout [post]
func (ac *AuthController) Logout(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing token"})
	}

	err := ac.authService.Logout(auth)
	if err != nil {
		log.Printf("Logout error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "logout failed"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "logout successful"})
}

// Me godoc
// @Summary Получение информации о текущем пользователе
// @Description Возвращает информацию о пользователе на основе переданного токена
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token, например: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
// @Security BearerAuth
// @Success 200 {object} response.UserInfo "Информация о пользователе"
// @Failure 401 {object} map[string]string "Отсутствует или неверный токен"
// @Router /auth/me [get]
func (ac *AuthController) Me(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
	}

	token := auth
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("Bearer "):])
	}

	userInfo, err := ac.authService.GetUserInfo(token)
	if err != nil {
		log.Printf("Me error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
	}

	resp := response.UserInfo{
		Sub:               utils.GetString(userInfo.Sub),
		Name:              utils.GetString(userInfo.Name),
		PreferredUsername: utils.GetString(userInfo.PreferredUsername),
		GivenName:         utils.GetString(userInfo.GivenName),
		FamilyName:        utils.GetString(userInfo.FamilyName),
		Email:             utils.GetString(userInfo.Email),
		EmailVerified:     userInfo.EmailVerified != nil && *userInfo.EmailVerified,
	}

	return c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary Обновление access токена
// @Description Получение нового access_token и refresh_token на основе существующего refresh_token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body request.RefreshRequest true "Refresh token"
// @Success 200 {object} response.RefreshResponse "Новые токены и время жизни"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 401 {object} map[string]string "Неверный или истёкший refresh_token"
// @Router /auth/refresh [post]
func (ac *AuthController) RefreshToken(c echo.Context) error {
	var req request.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	tokenResp, err := ac.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	now := time.Now()

	newTokens := response.RefreshResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		RefreshExp:   tokenResp.RefreshExpiresIn,
		TokenType:    "Bearer",
		ExpiresAt:    now.Add(time.Duration(300) * time.Second),
	}

	return c.JSON(http.StatusOK, newTokens)
}

// ValidateToken godoc
// @Summary Проверка access_token
// @Description Проверяет валидность токена через Keycloak
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token, например: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
// @Success 200 {object} response.TokenValidationResponse "Токен валиден"
// @Failure 401 {object} response.TokenValidationResponse "Невалидный или отсутствующий токен"
// @Router /auth/validate [get]
func (ac *AuthController) ValidateToken(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid Authorization header"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	err := ac.authService.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token invalid"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Token is valid"})
}