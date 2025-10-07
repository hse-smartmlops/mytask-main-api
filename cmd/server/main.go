// @title Emplacc API
// @version 1.0
// @description API для Emplacc.
// (host не задаём умышленно, чтобы Swagger использовал текущий origin)
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	docs "emplacc-api/docs"
	"emplacc-api/internal/controller"
	"emplacc-api/internal/db"
	"emplacc-api/internal/repository"
	"emplacc-api/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {
	loc, err := time.LoadLocation("Europe/Moscow")
    if err != nil {
        panic(err)
    }
    time.Local = loc
	
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

	// Подключение к БД через ваш существующий файл
	dbConn := db.DB_conn

	// Создаем зависимости
	userRepo := repository.NewUserRepository(dbConn)
	authService := service.NewAuthService(userRepo)
	boardRepo := repository.NewBoardRepository(dbConn)
	boardService := service.NewBoardService(boardRepo)
	reportRepo := repository.NewReportRepository(dbConn)
	reportService := service.NewReportService(reportRepo)
	forumMessageRepo := repository.NewForumMessageRepository(dbConn)
	forumMessageService := service.NewForumMessageService(forumMessageRepo)
	problemRepo := repository.NewProblemRepository(dbConn)
	problemService := service.NewProblemService(problemRepo)
	projectRepo := repository.NewProjectRepository(dbConn)
	projectService := service.NewProjectService(projectRepo)
	attendanceRepo := repository.NewAttendanceRepository(dbConn)
	attendanceService := service.NewAttendanceService(attendanceRepo)
	roleRepo := repository.NewRoleRepository(dbConn)
	roleService := service.NewRoleService(roleRepo)
	statusRepo := repository.NewStatusRepository(dbConn)
	statusService := service.NewStatusService(statusRepo)
	subscriptionRepo := repository.NewSubscriptionRepository(dbConn)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	taskRepo := repository.NewTaskRepository(dbConn)
	taskService := service.NewTaskService(taskRepo)
	teamRepo := repository.NewTeamRepository(dbConn)
	teamService := service.NewTeamService(teamRepo)
	userService := service.NewUserService(userRepo)
	
	// Keycloak auth middleware for protected endpoints
	e.Use(controller.KeycloakAuthMiddleware(authService)) // Передаем authService

	// Swagger: не хардкодим host, оставляем пустым, чтобы UI брал текущий адрес запроса
	docs.SwaggerInfo.Host = ""

	// Регистрируем маршруты с зависимостями
	controller.RegisterAuthRoutes(e, authService)
	controller.RegisterTeamRoutes(e, teamService)
	controller.RegisterProjectRoutes(e, projectService)
	controller.RegisterTaskRoutes(e, taskService)
	controller.RegisterBoardRoutes(e, boardService)
	controller.RegisterUserRoutes(e, userService)
	controller.RegisterReportRoutes(e, reportService)
	controller.RegisterForumMessagesRoutes(e, forumMessageService)
	controller.RegisterProblemRoutes(e, problemService)
	controller.RegisterRoleRoutes(e, roleService)
	controller.RegisterAttendanceRoutes(e, attendanceService)
	controller.RegisterSubscriptionRoutes(e, subscriptionService)
	controller.RegisterStatusRoutes(e, statusService)

	// Swagger UI
	// Редиректим с /swagger на /swagger/index.html, чтобы работало без явного указания файла
	e.GET("/swagger", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Если клиент посылает preflight OPTIONS запрос для вебхука
	e.OPTIONS("/event", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	log.Println("Server started on :8081")
	e.Logger.Fatal(e.Start("0.0.0.0:8081"))
}