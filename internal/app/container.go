package app

import (
	v1 "emplacc-api/api/v1"
	v1middleware "emplacc-api/api/v1/middleware"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	authrepo "emplacc-api/internal/repo/auth"
	minioRepo "emplacc-api/internal/repo/minio"
	"emplacc-api/internal/repo/pg"
	"emplacc-api/internal/service"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Container struct {
	AttendanceService   ports.AttendanceService
	AuthService         ports.AuthService
	UserService         ports.UserService
	RoleService         ports.RoleService
	ProjectService      ports.ProjectService
	BoardService        ports.BoardService
	TaskService         ports.TaskService
	StatusService       ports.StatusService
	TeamService         ports.TeamService
	SubscriptionService ports.SubscriptionService
	ProblemService      ports.ProblemService
	DailyReportService  ports.DailyReportService
	ForumMessageService ports.ForumMessageService
	tokenValidator      ports.TokenValidator
	rateLimiter         echo.MiddlewareFunc
	traceMetaEnabled    bool
	mcpService          ports.MCPService
}

func NewContainer(cfg *config.Config, db *gorm.DB, storage ports.ObjectStorage, mcpService ports.MCPService) *Container {
	attendanceRepo := pg.NewAttendanceRepository(db)
	userRepo := pg.NewUserRepository(db)
	roleRepo := pg.NewRoleRepository(db)
	projectRepo := pg.NewProjectRepository(db)
	boardRepo := pg.NewBoardRepository(db)
	taskRepo := pg.NewTaskRepository(db)
	statusRepo := pg.NewStatusRepository(db)
	teamRepo := pg.NewTeamRepository(db)
	subRepo := pg.NewSubscriptionRepository(db)
	problemRepo := pg.NewProblemRepository(db)
	forumRepo := pg.NewForumMessageRepository(db)
	reportRepo := pg.NewDailyReportRepository(db)
	authRepo := authrepo.NewKeycloakRepository(cfg.Keycloak)
	minioStorage := storage
	if minioStorage == nil {
		minioStorage, _ = minioRepo.NewStorage(cfg.Minio)
	}

	authService := service.NewAuthService(cfg.Keycloak, authRepo, userRepo)

	rateLimiter := v1middleware.NewRateLimiter(cfg.RateLimit)

	return &Container{
		AttendanceService:   service.NewAttendanceService(attendanceRepo),
		AuthService:         authService,
		UserService:         service.NewUserService(userRepo, minioStorage, cfg.Minio),
		RoleService:         service.NewRoleService(roleRepo),
		ProjectService:      service.NewProjectService(projectRepo),
		BoardService:        service.NewBoardService(boardRepo),
		TaskService:         service.NewTaskService(taskRepo),
		StatusService:       service.NewStatusService(statusRepo),
		TeamService:         service.NewTeamService(teamRepo),
		SubscriptionService: service.NewSubscriptionService(subRepo),
		ProblemService:      service.NewProblemService(problemRepo),
		DailyReportService:  service.NewDailyReportService(reportRepo, minioStorage, cfg.Minio),
		ForumMessageService: service.NewForumMessageService(forumRepo),
		tokenValidator:      authService,
		rateLimiter:         rateLimiter,
		traceMetaEnabled:    cfg.Tracing.Enabled,
		mcpService:          mcpService,
	}
}

func (c *Container) V1Deps() v1.Deps {
	return v1.Deps{
		AttendanceService:   c.AttendanceService,
		AuthService:         c.AuthService,
		UserService:         c.UserService,
		RoleService:         c.RoleService,
		ProjectService:      c.ProjectService,
		BoardService:        c.BoardService,
		TaskService:         c.TaskService,
		StatusService:       c.StatusService,
		TeamService:         c.TeamService,
		SubscriptionService: c.SubscriptionService,
		ProblemService:      c.ProblemService,
		DailyReportService:  c.DailyReportService,
		ForumMessageService: c.ForumMessageService,
		AuthMiddleware:      v1middleware.KeycloakAuth(c.tokenValidator),
		RateLimiter:         c.rateLimiter,
		TraceMetaEnabled:    c.traceMetaEnabled,
		MCPService:          c.mcpService,
	}
}
