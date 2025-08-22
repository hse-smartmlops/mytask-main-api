// @title Emplacc API
// @version 1.0
// @description API для Emplacc.
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	_ "emplacc-api/docs"
	"emplacc-api/internal/controller"
	"log"
	"net/http"

	"github.com/labstack/echo/v4/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New() 
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"*"}, // или конкретный фронтенд, например "http://localhost:3000"
    AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
    AllowHeaders: []string{
        echo.HeaderOrigin,
        echo.HeaderContentType,
        echo.HeaderAccept,
        echo.HeaderAuthorization,
    },
    AllowCredentials: true,
	}))

	// Регистрируем маршруты
	controller.RegisterAuthRoutes(e)
	controller.RegisterTeamRoutes(e)
	controller.RegisterProjectRoutes(e)
	controller.RegisterTaskRoutes(e)
	controller.RegisterBoardRoutes(e)
	controller.RegisterUserRoutes(e)
	controller.RegisterReportRoutes(e)
	controller.RegisterForumMessagesRoutes(e)
	controller.RegisterProblemRoutes(e)
	controller.RegisterRoleRoutes(e)
	controller.RegisterWebhookRoutes(e)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Если клиент посылает preflight OPTIONS запрос для вебхука
	e.OPTIONS("/event", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	log.Println("Server started on :8081")
	e.Logger.Fatal(e.Start("0.0.0.0:8081"))
}