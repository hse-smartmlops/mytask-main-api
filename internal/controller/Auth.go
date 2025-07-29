package controller

import (
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"github.com/labstack/echo/v4"
	"net/http"
)

func RegisterAuthRoutes(e *echo.Echo) {
	authGroup := e.Group("/auth")
	authGroup.POST("/login", Login)
	authGroup.POST("/logout", Logout)
	authGroup.POST("/register", Register)
	authGroup.GET("/me", Me)
	authGroup.POST("/oauth", OAuth)
	authGroup.POST("/totp", TOTP)
}

func Login(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp := response.AuthResponse{
		UserID:    "user-id",
		Email:     req.Email,
		Token:     "jwt-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	return c.JSON(http.StatusOK, resp)
}

func OAuth(c echo.Context) error {
	var req request.OAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp := response.AuthResponse{
		UserID:    "user-id",
		Email:     "user@example.com",
		Token:     "oauth-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	return c.JSON(http.StatusOK, resp)
}

func TOTP(c echo.Context) error {
	var req request.TOTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp := response.AuthResponse{
		UserID:    "user-id",
		Email:     "user@example.com",
		Token:     "totp-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	return c.JSON(http.StatusOK, resp)
}

func Logout(c echo.Context) error {
	// TODO: Реализовать логику выхода (инвалидация токена/сессии)
	return c.JSON(http.StatusOK, map[string]string{"message": "Выход выполнен"})
}

func Register(c echo.Context) error {
	var req request.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp := response.AuthResponse{
		UserID:    "new-user-id",
		Email:     req.Email,
		Token:     "jwt-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	return c.JSON(http.StatusOK, resp)
}

func Me(c echo.Context) error {
	// TODO: Реализовать получение информации о текущем пользователе
	return c.JSON(http.StatusOK, map[string]string{"message": "Информация о пользователе"})
}
