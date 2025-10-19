package app

import (
	"context"
	"math"
	"time"

	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type PaginationParams struct {
	Page     int
	PageSize int
}

type Page[T any] struct {
	Items      []T
	Page       int
	PageSize   int
	TotalCount int64
}

func (p Page[T]) TotalPages() int {
	if p.PageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(p.TotalCount) / float64(p.PageSize)))
}

type CreateUserInput struct {
	Email         string
	IsActive      *bool
	TgID          *string
	TgUserID      *int64
	Profession    *string
	EmailVerified *bool
	FirstName     *string
	LastName      *string
}

type UserRepository interface {
	ListUsers(ctx context.Context, params PaginationParams) (*Page[models.User], error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type UserService interface {
	ListUsers(ctx context.Context, params PaginationParams) (*Page[models.User], error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
}

type CreateRoleInput struct {
	Name        string
	Description *string
}

type UpdateRoleInput struct {
	Name        *string
	Description *string
}

type RoleRepository interface {
	ListRoles(ctx context.Context, params PaginationParams) (*Page[models.Role], error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	CreateRole(ctx context.Context, role *models.Role) error
	UpdateRole(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Role, error)
	SoftDeleteRole(ctx context.Context, id uuid.UUID) error
}

type RoleService interface {
	ListRoles(ctx context.Context, params PaginationParams) (*Page[models.Role], error)
	GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error)
	CreateRole(ctx context.Context, input CreateRoleInput) (*models.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*models.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
}

type CreateProjectInput struct {
	Name            string
	Description     *string
	GitlabProjectID *int
	GitlabURL       *string
	CreatedBy       *uuid.UUID
	Status          *string
}

type UpdateProjectInput struct {
	Name            *string
	Description     *string
	GitlabProjectID *int
	GitlabURL       *string
	Status          *string
}

type ProjectRepository interface {
	ListProjects(ctx context.Context, params PaginationParams) (*Page[models.Project], error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	CreateProject(ctx context.Context, project *models.Project) error
	UpdateProject(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Project, error)
	SoftDeleteProject(ctx context.Context, id uuid.UUID) error
}

type ProjectService interface {
	ListProjects(ctx context.Context, params PaginationParams) (*Page[models.Project], error)
	GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error)
	CreateProject(ctx context.Context, input CreateProjectInput) (*models.Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, input UpdateProjectInput) (*models.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
}

type CreateTaskInput struct {
	StatusID      uuid.UUID
	Priority      *int16
	Name          string
	Description   *string
	CreatedBy     *uuid.UUID
	AssignedTo    *uuid.UUID
	Deadline      *time.Time
	StartDate     *time.Time
	GitlabIssueID *int
	Category      *int8
}

type UpdateTaskInput struct {
	StatusID      *uuid.UUID
	Priority      *int16
	Name          *string
	Description   *string
	CreatedBy     *uuid.UUID
	AssignedTo    *uuid.UUID
	Deadline      *time.Time
	StartDate     *time.Time
	GitlabIssueID *int
	Category      *int8
}

type TaskRepository interface {
	ListTasks(ctx context.Context, params PaginationParams) (*Page[models.Task], error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	ListTasksByProject(ctx context.Context, projectID uuid.UUID, params PaginationParams) (*Page[models.Task], error)
	ListTasksByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Task], error)
	CreateTask(ctx context.Context, task *models.Task) error
	UpdateTask(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Task, error)
	SoftDeleteTask(ctx context.Context, id uuid.UUID) error
}

type TaskService interface {
	ListTasks(ctx context.Context, params PaginationParams) (*Page[models.Task], error)
	GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error)
	ListTasksByProject(ctx context.Context, projectID uuid.UUID, params PaginationParams) (*Page[models.Task], error)
	ListTasksByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Task], error)
	CreateTask(ctx context.Context, input CreateTaskInput) (*models.Task, error)
	UpdateTask(ctx context.Context, id uuid.UUID, input UpdateTaskInput) (*models.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
}

type CreateBoardInput struct {
	ProjectID   uuid.UUID
	Name        string
	Description *string
}

type UpdateBoardInput struct {
	Name        *string
	Description *string
}

type BoardRepository interface {
	ListBoards(ctx context.Context, params PaginationParams) (*Page[models.Board], error)
	GetBoardByID(ctx context.Context, id uuid.UUID) (*models.Board, error)
	ListBoardsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Board, error)
	CreateBoard(ctx context.Context, board *models.Board) error
	UpdateBoard(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Board, error)
	SoftDeleteBoard(ctx context.Context, id uuid.UUID) error
}

type BoardService interface {
	ListBoards(ctx context.Context, params PaginationParams) (*Page[models.Board], error)
	GetBoard(ctx context.Context, id uuid.UUID) (*models.Board, error)
	ListBoardsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Board, error)
	CreateBoard(ctx context.Context, input CreateBoardInput) (*models.Board, error)
	UpdateBoard(ctx context.Context, id uuid.UUID, input UpdateBoardInput) (*models.Board, error)
	DeleteBoard(ctx context.Context, id uuid.UUID) error
}

type CreateStatusInput struct {
	BoardID   uuid.UUID
	Name      string
	Key       *string
	Color     *string
	IsDefault *bool
	IsActive  *bool
	IsOpen    *bool
	SortOrder *int
}

type UpdateStatusInput struct {
	Name      *string
	Key       *string
	Color     *string
	IsDefault *bool
	IsActive  *bool
	IsOpen    *bool
	SortOrder *int
}

type StatusRepository interface {
	ListStatuses(ctx context.Context, params PaginationParams) (*Page[models.Status], error)
	GetStatusByID(ctx context.Context, id uuid.UUID) (*models.Status, error)
	ListStatusesByBoard(ctx context.Context, boardID uuid.UUID) ([]models.Status, error)
	CreateStatus(ctx context.Context, status *models.Status) error
	UpdateStatus(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Status, error)
	SoftDeleteStatus(ctx context.Context, id uuid.UUID) error
}

type StatusService interface {
	ListStatuses(ctx context.Context, params PaginationParams) (*Page[models.Status], error)
	GetStatus(ctx context.Context, id uuid.UUID) (*models.Status, error)
	ListStatusesByBoard(ctx context.Context, boardID uuid.UUID) ([]models.Status, error)
	CreateStatus(ctx context.Context, input CreateStatusInput) (*models.Status, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateStatusInput) (*models.Status, error)
	DeleteStatus(ctx context.Context, id uuid.UUID) error
}

type CreateTeamInput struct {
	Name        string
	Description *string
}

type UpdateTeamInput struct {
	Name        *string
	Description *string
}

type TeamMemberInput struct {
	TeamID         uuid.UUID
	UserID         uuid.UUID
	Specialization *string
}

type TeamProjectInput struct {
	TeamID    uuid.UUID
	ProjectID uuid.UUID
}

type TeamRepository interface {
	ListTeams(ctx context.Context, params PaginationParams) (*Page[models.Team], error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error)
	ListTeamsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Team, error)
	CreateTeam(ctx context.Context, team *models.Team) error
	UpdateTeam(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Team, error)
	SoftDeleteTeam(ctx context.Context, id uuid.UUID) error
	AddUserToTeam(ctx context.Context, input TeamMemberInput) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error
	AddTeamToProject(ctx context.Context, input TeamProjectInput) error
	RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error
}

type TeamService interface {
	ListTeams(ctx context.Context, params PaginationParams) (*Page[models.Team], error)
	GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error)
	ListTeamsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Team, error)
	CreateTeam(ctx context.Context, input CreateTeamInput) (*models.Team, error)
	UpdateTeam(ctx context.Context, id uuid.UUID, input UpdateTeamInput) (*models.Team, error)
	DeleteTeam(ctx context.Context, id uuid.UUID) error
	AddUserToTeam(ctx context.Context, input TeamMemberInput) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error
	AddTeamToProject(ctx context.Context, input TeamProjectInput) error
	RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error
}

type CreateSubscriptionInput struct {
	UserID         uuid.UUID
	SubscriptionID *uuid.UUID
	TypeID         *int8
}

type SubscriptionRepository interface {
	ListSubscriptions(ctx context.Context, params PaginationParams) (*Page[models.Subscription], error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByTarget(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, params PaginationParams) (*Page[models.Subscription], error)
	CreateSubscription(ctx context.Context, sub *models.Subscription) error
	SoftDeleteSubscription(ctx context.Context, id uuid.UUID) error
}

type SubscriptionService interface {
	ListSubscriptions(ctx context.Context, params PaginationParams) (*Page[models.Subscription], error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByTarget(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, params PaginationParams) (*Page[models.Subscription], error)
	CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*models.Subscription, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) error
}
