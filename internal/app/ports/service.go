package ports

import (
	"context"
	"time"

	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type UserService interface {
	ListUsers(ctx context.Context, params PaginationParams) (*Page[models.User], error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
	SaveAvatar(ctx context.Context, id uuid.UUID, input SaveUserAvatarInput) (*models.User, error)
	DeleteAvatar(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetAvatar(ctx context.Context, id uuid.UUID) (*UserAvatarFile, error)
}

type AuthService interface {
	TokenValidator
	Login(ctx context.Context, email, password string) (*AuthLoginResult, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserInfo(ctx context.Context, token string) (*AuthUserInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthRefreshResult, error)
}

type RoleService interface {
	ListRoles(ctx context.Context, params PaginationParams) (*Page[models.Role], error)
	GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error)
	CreateRole(ctx context.Context, input CreateRoleInput) (*models.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*models.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
}

type ProjectService interface {
	ListProjects(ctx context.Context, params PaginationParams) (*Page[models.Project], error)
	GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error)
	CreateProject(ctx context.Context, input CreateProjectInput) (*models.Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, input UpdateProjectInput) (*models.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
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

type BoardService interface {
	ListBoards(ctx context.Context, params PaginationParams) (*Page[models.Board], error)
	GetBoard(ctx context.Context, id uuid.UUID) (*models.Board, error)
	ListBoardsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Board, error)
	CreateBoard(ctx context.Context, input CreateBoardInput) (*models.Board, error)
	UpdateBoard(ctx context.Context, id uuid.UUID, input UpdateBoardInput) (*models.Board, error)
	DeleteBoard(ctx context.Context, id uuid.UUID) error
}

type TeamService interface {
	ListAllTeams(ctx context.Context) ([]models.Team, error)
	ListTeams(ctx context.Context, params PaginationParams) (*Page[models.Team], error)
	GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error)
	ListTeamsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Team, error)
	ListTeamsByUser(ctx context.Context, userID uuid.UUID) ([]models.Team, error)
	CreateTeam(ctx context.Context, input CreateTeamInput) (*models.Team, error)
	UpdateTeam(ctx context.Context, id uuid.UUID, input UpdateTeamInput) (*models.Team, error)
	DeleteTeam(ctx context.Context, id uuid.UUID) error
	AddUserToTeam(ctx context.Context, input TeamMemberInput) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error
	AddTeamToProject(ctx context.Context, input TeamProjectInput) error
	RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error
}

type SubscriptionService interface {
	ListSubscriptions(ctx context.Context, params PaginationParams) (*Page[models.Subscription], error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByTarget(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, params PaginationParams) (*Page[models.Subscription], error)
	CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*models.Subscription, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
}

type AttendanceService interface {
	ListAttendances(ctx context.Context, params PaginationParams) (*Page[models.Attendance], error)
	ListAttendancesByUser(ctx context.Context, userID uuid.UUID) ([]models.Attendance, error)
	GetAttendance(ctx context.Context, id uuid.UUID) (*models.Attendance, error)
	CreateAttendance(ctx context.Context, input CreateAttendanceInput) (*models.Attendance, error)
	UpdateAttendance(ctx context.Context, id uuid.UUID, input UpdateAttendanceInput) (*models.Attendance, error)
	DeleteAttendance(ctx context.Context, id uuid.UUID) error
}

type StatusService interface {
	ListStatuses(ctx context.Context, params PaginationParams) (*Page[models.Status], error)
	GetStatus(ctx context.Context, id uuid.UUID) (*models.Status, error)
	ListStatusesByBoard(ctx context.Context, boardID uuid.UUID) ([]models.Status, error)
	CreateStatus(ctx context.Context, input CreateStatusInput) (*models.Status, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateStatusInput) (*models.Status, error)
	DeleteStatus(ctx context.Context, id uuid.UUID) error
}

type DailyReportService interface {
	ListReports(ctx context.Context, params PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByProject(ctx context.Context, projectID uuid.UUID) ([]models.DailyReport, error)
	ListReportsByTask(ctx context.Context, taskID uuid.UUID) ([]models.DailyReport, error)
	ListReportsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.DailyReport, error)
	ListHelpRequestsByHelper(ctx context.Context, helperID uuid.UUID) ([]models.HelpRequest, error)
	GetReport(ctx context.Context, id uuid.UUID) (*models.DailyReport, error)
	CreateReport(ctx context.Context, input CreateDailyReportInput) (*models.DailyReport, error)
	UpdateReport(ctx context.Context, id uuid.UUID, input UpdateDailyReportInput) (*models.DailyReport, error)
	DeleteReport(ctx context.Context, id uuid.UUID) error
	UpdateCompletedWork(ctx context.Context, id uuid.UUID, input UpdateCompletedWorkInput) (*models.CompletedWork, error)
	UpdateHelpRequest(ctx context.Context, id uuid.UUID, input UpdateHelpRequestInput) (*models.HelpRequest, error)
	DeleteHelpRequest(ctx context.Context, id uuid.UUID) error
	UpdateTomorrowPlan(ctx context.Context, id uuid.UUID, input UpdateTomorrowPlanInput) (*models.TomorrowPlans, error)
	ExportReportsToXLSX(ctx context.Context, input ReportsByDateInput, fileName string) ([]byte, error)
}

type ProblemService interface {
	ListProblems(ctx context.Context, params PaginationParams) (*Page[models.Problem], error)
	ListProblemsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Problem], error)
	GetProblem(ctx context.Context, id uuid.UUID) (*models.Problem, error)
	CreateProblem(ctx context.Context, input CreateProblemInput) (*models.Problem, error)
	UpdateProblem(ctx context.Context, id uuid.UUID, input UpdateProblemInput) (*models.Problem, error)
	DeleteProblem(ctx context.Context, id uuid.UUID) error
}

type ForumMessageService interface {
	ListMessages(ctx context.Context, params PaginationParams) (*Page[models.ForumMessage], error)
	GetMessage(ctx context.Context, id uuid.UUID) (*models.ForumMessage, error)
	ListMessagesByProblem(ctx context.Context, problemID uuid.UUID, params PaginationParams) (*Page[models.ForumMessage], error)
	CreateMessage(ctx context.Context, input CreateForumMessageInput) (*models.ForumMessage, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, input UpdateForumMessageInput) (*models.ForumMessage, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error
}

type MCPService interface {
	ImproveReport(ctx context.Context, input ImproveReportInput) (*ImproveReportResult, error)
	StreamReport(ctx context.Context, input ImproveReportInput) (*MCPStream, error)
}
