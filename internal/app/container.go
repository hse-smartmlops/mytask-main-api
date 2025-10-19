package app

import (
	v1 "emplacc-api/api/v1"
	"emplacc-api/internal/repo/pg"
	"emplacc-api/internal/service"

	"gorm.io/gorm"
)

type Container struct {
	UserService    UserService
	RoleService    RoleService
	ProjectService ProjectService
	BoardService   BoardService
	TaskService    TaskService
	StatusService  StatusService
}

func NewContainer(db *gorm.DB) *Container {
	userRepo := pg.NewUserRepository(db)
	roleRepo := pg.NewRoleRepository(db)
	projectRepo := pg.NewProjectRepository(db)
	boardRepo := pg.NewBoardRepository(db)
	taskRepo := pg.NewTaskRepository(db)
	statusRepo := pg.NewStatusRepository(db)

	return &Container{
		UserService:    service.NewUserService(userRepo),
		RoleService:    service.NewRoleService(roleRepo),
		ProjectService: service.NewProjectService(projectRepo),
		BoardService:   service.NewBoardService(boardRepo),
		TaskService:    service.NewTaskService(taskRepo),
		StatusService:  service.NewStatusService(statusRepo),
	}
}

func (c *Container) V1Deps() v1.Deps {
	return v1.Deps{
		UserService:    c.UserService,
		RoleService:    c.RoleService,
		ProjectService: c.ProjectService,
		BoardService:   c.BoardService,
		TaskService:    c.TaskService,
		StatusService:  c.StatusService,
	}
}
