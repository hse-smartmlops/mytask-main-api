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
	"context"
	docs "emplacc-api/docs"
	"emplacc-api/internal/controller"
	"emplacc-api/internal/db"
	"emplacc-api/internal/grpc/client"
	"emplacc-api/internal/repository"
	"emplacc-api/internal/service"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var systemUserId uuid.UUID

func main() {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}
	time.Local = loc

	e := echo.New()
	// ── Recovery middleware — поймать панику перед логированием ──
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://emplacc.g-309.ru",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
		},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS, echo.PATCH},
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
	llmHost := os.Getenv("LLM_HOST")
	llmAddr := llmHost + ":50051"

	llmClient, err := client.NewLLMClient(llmAddr)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer llmClient.Close()

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
	apiTokenRepo := repository.NewAPITokenRepository(dbConn)
	apiTokenService := service.NewAPITokenService(apiTokenRepo, userRepo)

	// Redis — сессии без персистентности на диск
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	log.Printf("Redis connected: %s", redisAddr)

	sessionRepo := repository.NewSessionRepository(rdb)
	sessionService := service.NewSessionService(sessionRepo)

	systemUserId, err = userService.CreateSystemUser()
	if err != nil {
		log.Fatalf("service error (create user): %v", err)
	}

	problemService := service.NewProblemService(problemRepo, forumMessageRepo, systemUserId)

	// S3/RustFS storage (опционально — не падаем если не настроен)
	storageService, storageErr := service.NewStorageService()
	if storageErr != nil {
		log.Printf("Warning: S3 storage not configured: %v", storageErr)
	}

	// Swagger: не хардкодим host, оставляем пустым, чтобы UI брал текущий адрес запроса
	docs.SwaggerInfo.Host = ""

	// ── Главный middleware ПЕРЕД маршрутами — критический порядок ──
	// Новый middleware: только сессии (sess_*) и MCP токены (emplacc_*)
	e.Use(controller.AppAuthMiddleware(authService, sessionService, apiTokenService))

	// Role-based middleware — три уровня доступа
	adminMw := controller.RequireRoles(roleRepo, "admin")
	managerMw := controller.RequireRoles(roleRepo, "admin", "manager")
	employeeMw := controller.RequireRoles(roleRepo, "admin", "manager", "employee")

	// freshAvatarURL — хелпер для генерации свежих presigned URL из хранимых путей/URL
	// Если storage не настроен — возвращает значение как есть (деградирует gracefully)
	freshAvatarURL := func(s string) string { return s }
	if storageService != nil {
		freshAvatarURL = storageService.FreshAvatarURL
	}

	// Регистрируем маршруты с зависимостями (ПОСЛЕ глобального middleware)
	controller.RegisterAuthRoutes(e, authService, sessionService, userService)
	controller.RegisterUserRoutes(e, userService, freshAvatarURL, adminMw)
	controller.RegisterRoleRoutes(e, roleService, adminMw)
	controller.RegisterTeamRoutes(e, teamService, freshAvatarURL, managerMw)
	controller.RegisterProjectRoutes(e, projectService, managerMw)
	controller.RegisterBoardRoutes(e, boardService, managerMw)
	controller.RegisterStatusRoutes(e, statusService, managerMw)
	controller.RegisterTaskRoutes(e, taskService, userService, projectService, llmClient, dbConn, freshAvatarURL, employeeMw, managerMw)
	controller.RegisterLLMSettingsRoutes(e, dbConn, adminMw)
	controller.RegisterReportRoutes(e, reportService, freshAvatarURL, employeeMw, managerMw)
	controller.RegisterForumMessagesRoutes(e, forumMessageService, freshAvatarURL, employeeMw, managerMw)
	controller.RegisterProblemRoutes(e, problemService, employeeMw, managerMw)
	controller.RegisterAttendanceRoutes(e, attendanceService, employeeMw, managerMw)
	controller.RegisterSubscriptionRoutes(e, subscriptionService)
	controller.RegisterAPITokenRoutes(e, apiTokenService)
	if storageService != nil {
		controller.RegisterUploadRoutes(e, storageService, userService)
	}

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
