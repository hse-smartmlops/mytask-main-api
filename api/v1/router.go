package v1

import (
	"emplacc-api/api/v1/handlers"
	"emplacc-api/internal/app/ports"

	"github.com/labstack/echo/v4"
)

type Deps struct {
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
	AuthMiddleware      echo.MiddlewareFunc
}

func RegisterRoutes(group *echo.Group, deps Deps) {
	if deps.AuthService != nil {
		handlers.RegisterAuthRoutes(group, deps.AuthService)
	}
	if deps.AuthMiddleware != nil {
		group.Use(deps.AuthMiddleware)
	}
	if deps.AttendanceService != nil {
		handlers.RegisterAttendanceRoutes(group, deps.AttendanceService)
	}
	if deps.DailyReportService != nil {
		handlers.RegisterDailyReportRoutes(group, deps.DailyReportService)
	}
	if deps.UserService != nil {
		handlers.RegisterUserRoutes(group, deps.UserService)
	}
	if deps.RoleService != nil {
		handlers.RegisterRoleRoutes(group, deps.RoleService)
	}
	if deps.ProjectService != nil {
		handlers.RegisterProjectRoutes(group, deps.ProjectService)
	}
	if deps.BoardService != nil {
		handlers.RegisterBoardRoutes(group, deps.BoardService)
	}
	if deps.TaskService != nil {
		handlers.RegisterTaskRoutes(group, deps.TaskService)
	}
	if deps.StatusService != nil {
		handlers.RegisterStatusRoutes(group, deps.StatusService)
	}
	if deps.TeamService != nil {
		handlers.RegisterTeamRoutes(group, deps.TeamService)
	}
	if deps.SubscriptionService != nil {
		handlers.RegisterSubscriptionRoutes(group, deps.SubscriptionService)
	}
	if deps.ProblemService != nil {
		handlers.RegisterProblemRoutes(group, deps.ProblemService)
	}
	if deps.ForumMessageService != nil {
		handlers.RegisterForumMessageRoutes(group, deps.ForumMessageService)
	}
}
