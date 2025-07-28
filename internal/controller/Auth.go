package controller

import (
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterAuthRoutes(router *gin.Engine) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", Login)
		authGroup.POST("/logout", Logout)
		authGroup.POST("/register", Register)
		authGroup.GET("/me", Me)
		authGroup.POST("/oauth", OAuth)
		authGroup.POST("/totp", TOTP)
	}
}

func Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := response.AuthResponse{
		UserID:    "user-id",
		Email:     req.Email,
		Token:     "jwt-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	c.JSON(http.StatusOK, resp)
}

func OAuth(c *gin.Context) {
	var req request.OAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := response.AuthResponse{
		UserID:    "user-id",
		Email:     "user@example.com",
		Token:     "oauth-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	c.JSON(http.StatusOK, resp)
}

func TOTP(c *gin.Context) {
	var req request.TOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := response.AuthResponse{
		UserID:    "user-id",
		Email:     "user@example.com",
		Token:     "totp-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	c.JSON(http.StatusOK, resp)
}

func Logout(c *gin.Context) {
	// TODO: Реализовать логику выхода (инвалидация токена/сессии)
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен"})
}

func Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := response.AuthResponse{
		UserID:    "new-user-id",
		Email:     req.Email,
		Token:     "jwt-token",
		ExpiresAt: "2025-12-31T23:59:59Z",
	}
	c.JSON(http.StatusCreated, resp)
}

func Me(c *gin.Context) {
	// TODO: Получить информацию о текущем пользователе из контекста/сессии
	resp := response.UserInfoResponse{
		UserID:     "user-id",
		Email:      "user@example.com",
		FirstName:  "Иван",
		LastName:   "Иванов",
		Profession: "Backend",
		IsActive:   true,
		Roles:      []string{"admin", "employee"},
	}
	c.JSON(http.StatusOK, resp)
}
