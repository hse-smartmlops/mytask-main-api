package app

import (
	v1 "emplacc-api/api/v1"
	v1middleware "emplacc-api/api/v1/middleware"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/repo/pg"
	"emplacc-api/internal/service"

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
}

func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
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
	authService := service.NewAuthService(cfg.Keycloak, userRepo)

	return &Container{
		AttendanceService:   service.NewAttendanceService(attendanceRepo),
		AuthService:         authService,
		UserService:         service.NewUserService(userRepo),
		RoleService:         service.NewRoleService(roleRepo),
		ProjectService:      service.NewProjectService(projectRepo),
		BoardService:        service.NewBoardService(boardRepo),
		TaskService:         service.NewTaskService(taskRepo),
		StatusService:       service.NewStatusService(statusRepo),
		TeamService:         service.NewTeamService(teamRepo),
		SubscriptionService: service.NewSubscriptionService(subRepo),
		ProblemService:      service.NewProblemService(problemRepo),
		DailyReportService:  service.NewDailyReportService(reportRepo),
		ForumMessageService: service.NewForumMessageService(forumRepo),
		tokenValidator:      authService,
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
	}
}
