// Package app is the COMPOSITION ROOT of the service — the single place that knows
// about every concrete dependency: it constructs adapters, hides them behind ports,
// injects them into services, and wires services into HTTP handlers. Nothing else in
// the codebase imports across layers for construction; that responsibility lives here.
//
// main.go stays thin: load timezone, call Bootstrap, start the server, Close on exit.
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	docs "emplacc-api/docs"
	"emplacc-api/internal/controller"
	"emplacc-api/internal/db"
	"emplacc-api/internal/grpc/client"
	infragit "emplacc-api/internal/infra/git"
	"emplacc-api/internal/infra/storage"
	"emplacc-api/internal/repository"
	"emplacc-api/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// App is the bootstrapped, ready-to-serve application: a configured Echo instance
// plus the cleanup hooks that release infrastructure handles on shutdown.
type App struct {
	Echo    *echo.Echo
	cleanup []func()
}

// Close runs cleanup hooks in reverse construction order.
func (a *App) Close() {
	for i := len(a.cleanup) - 1; i >= 0; i-- {
		a.cleanup[i]()
	}
}

// Bootstrap builds the entire dependency graph and returns a runnable App.
func Bootstrap() (*App, error) {
	app := &App{}

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

	dbConn := db.DB_conn

	llmHost := os.Getenv("LLM_HOST")
	llmAddr := llmHost + ":50051"
	llmClient, err := client.NewLLMClient(llmAddr, os.Getenv("LLM_GRPC_AUTH_TOKEN"))
	if err != nil {
		return nil, fmt.Errorf("create gRPC client: %w", err)
	}
	app.cleanup = append(app.cleanup, func() { llmClient.Close() })

	// ── Repositories + services ──
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
	conveyorRepo := repository.NewConveyorRepository(dbConn)
	conveyorService := service.NewConveyorServiceWithReportLLM(conveyorRepo, service.NewBackendGeneratedReportLLMClient(llmClient))
	pmImportService := service.NewPMImportService(conveyorRepo)
	teamRepo := repository.NewTeamRepository(dbConn)
	teamService := service.NewTeamService(teamRepo)
	userService := service.NewUserService(userRepo)
	apiTokenRepo := repository.NewAPITokenRepository(dbConn)
	apiTokenService := service.NewAPITokenService(apiTokenRepo, userRepo)

	// Git commit-tracker: host-agnostic (порт CommitProvider) + адаптеры GitHub/GitFlic.
	gitRepo := repository.NewGitRepository(dbConn)
	gitService := service.NewGitService(gitRepo, userRepo,
		infragit.NewGitHubProvider(os.Getenv("GITHUB_TOKEN")),
		infragit.NewGitFlicProvider(os.Getenv("GITFLIC_TOKEN"), os.Getenv("GITFLIC_API_URL")),
	)

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
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	log.Printf("Redis connected: %s", redisAddr)
	app.cleanup = append(app.cleanup, func() { _ = rdb.Close() })

	sessionRepo := repository.NewSessionRepository(rdb)
	sessionService := service.NewSessionService(sessionRepo)

	// SSE realtime: in-memory шина событий; conveyor.createEvent публикует в неё через GlobalEventHub.
	eventHub := service.NewEventHub()
	service.GlobalEventHub = eventHub

	systemUserId, err := userService.CreateSystemUser()
	if err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}

	problemService := service.NewProblemService(problemRepo, forumMessageRepo, systemUserId)

	// S3/RustFS storage (опционально — не падаем если не настроен)
	storageService, storageErr := storage.New()
	if storageErr != nil {
		log.Printf("Warning: S3 storage not configured: %v", storageErr)
	}

	// Swagger: не хардкодим host, оставляем пустым, чтобы UI брал текущий адрес запроса
	docs.SwaggerInfo.Host = ""

	// ── Главный middleware ПЕРЕД маршрутами — критический порядок ──
	e.Use(controller.AppAuthMiddleware(authService, sessionService, apiTokenService))

	// Role-based middleware — три уровня доступа
	adminMw := controller.RequireRoles(roleRepo, "admin")
	managerMw := controller.RequireRoles(roleRepo, "admin", "manager")
	employeeMw := controller.RequireRoles(roleRepo, "admin", "manager", "employee")

	// freshAvatarURL — генерация свежих presigned URL; деградирует gracefully без storage
	freshAvatarURL := func(s string) string { return s }
	if storageService != nil {
		freshAvatarURL = storageService.FreshAvatarURL
	}

	// Регистрируем маршруты (ПОСЛЕ глобального middleware)
	controller.RegisterAuthRoutes(e, authService, sessionService, userService)
	controller.RegisterUserRoutes(e, userService, freshAvatarURL, adminMw)
	controller.RegisterRoleRoutes(e, roleService, adminMw)
	controller.RegisterTeamRoutes(e, teamService, freshAvatarURL, managerMw)
	controller.RegisterProjectRoutes(e, projectService, managerMw)
	controller.RegisterBoardRoutes(e, boardService, managerMw)
	controller.RegisterStatusRoutes(e, statusService, managerMw)
	controller.RegisterTaskRoutes(e, taskService, userService, projectService, llmClient, conveyorService, dbConn, freshAvatarURL, employeeMw, managerMw)
	controller.RegisterConveyorRoutes(e, conveyorService, pmImportService, employeeMw, managerMw)
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

	// SSE realtime stream (/v2/stream) — auth по query-токену, см. Stream.go
	controller.RegisterStreamRoutes(e, eventHub, sessionService, apiTokenService)

	// Git commit-tracker (репозитории/коммиты/привязка к задачам)
	controller.RegisterGitRoutes(e, gitService, employeeMw, managerMw)

	// Swagger UI
	e.GET("/swagger", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Preflight OPTIONS для вебхука
	e.OPTIONS("/event", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	app.Echo = e
	return app, nil
}
