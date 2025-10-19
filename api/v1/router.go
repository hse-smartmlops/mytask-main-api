package v1

import (
	"emplacc-api/api/v1/handlers"
	"emplacc-api/internal/app"

	"github.com/labstack/echo/v4"
)

type Deps struct {
	UserService    app.UserService
	RoleService    app.RoleService
	ProjectService app.ProjectService
	BoardService   app.BoardService
	TaskService    app.TaskService
	StatusService  app.StatusService
	AuthMiddleware echo.MiddlewareFunc
}

func RegisterRoutes(group *echo.Group, deps Deps) {
	if deps.AuthMiddleware != nil {
		group.Use(deps.AuthMiddleware)
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
}
