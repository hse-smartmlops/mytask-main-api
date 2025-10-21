package ports

import (
	"context"
	"time"

	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type UserService interface {
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)

	CreateUser(ctx context.Context, input UserInput) (*models.User, error)

	UpdateUser(ctx context.Context, id uuid.UUID, input UserInput) (*models.User, error)

	DeleteUser(ctx context.Context, id uuid.UUID) error

	SaveAvatar(ctx context.Context, id uuid.UUID, input UserAvatarInput) (*models.User, error)

	DeleteAvatar(ctx context.Context, id uuid.UUID) (*models.User, error)

	GetAvatar(ctx context.Context, id uuid.UUID) (*UserAvatarFile, error)

	UpdateAvatar(ctx context.Context, id uuid.UUID, input UserAvatarInput) (*models.User, error)

	ListUsers(ctx context.Context, p PaginationParams) (*Page[models.User], error)
}

type AuthService interface {
	TokenValidator

	Login(ctx context.Context, email, password string) (*AuthLoginResult, error)

	Logout(ctx context.Context, refreshToken string) error

	GetUserInfo(ctx context.Context, token string) (*AuthUserInfo, error)

	RefreshToken(ctx context.Context, refreshToken string) (*AuthRefreshResult, error)
}

type RoleService interface {
	GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error)

	CreateRole(ctx context.Context, input CreateRoleInput) (*models.Role, error)

	UpdateRole(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*models.Role, error)

	DeleteRole(ctx context.Context, id uuid.UUID) error

	ListRoles(ctx context.Context, p PaginationParams) (*Page[models.Role], error)
}

type ProjectService interface {
	GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error)

	CreateProject(ctx context.Context, input CreateProjectInput) (*models.Project, error)

	UpdateProject(ctx context.Context, id uuid.UUID, input UpdateProjectInput) (*models.Project, error)

	DeleteProject(ctx context.Context, id uuid.UUID) error

	ListProjects(ctx context.Context, p PaginationParams) (*Page[models.Project], error)
	ListProjectsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Project], error)
	ListTeamsByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.Team], error)
	ListProjectsByTeam(ctx context.Context, teamID uuid.UUID, p PaginationParams) (*Page[models.Project], error)
}

type TaskService interface {
	GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error)

	CreateTask(ctx context.Context, input CreateTaskInput) (*models.Task, error)

	UpdateTask(ctx context.Context, id uuid.UUID, input UpdateTaskInput) (*models.Task, error)

	DeleteTask(ctx context.Context, id uuid.UUID) error

	MoveTaskBetweenStatuses(ctx context.Context, taskID, toStatusID uuid.UUID) (*models.Task, error)

	ListTasks(ctx context.Context, p PaginationParams) (*Page[models.Task], error)
	ListTasksByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListTasksByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListActiveTasksByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListTasksByUserAndProject(ctx context.Context, userID, projectID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListTasksByBoard(ctx context.Context, boardID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
}

type BoardService interface {
	GetBoard(ctx context.Context, id uuid.UUID) (*models.Board, error)

	CreateBoard(ctx context.Context, input CreateBoardInput) (*models.Board, error)

	UpdateBoard(ctx context.Context, id uuid.UUID, input UpdateBoardInput) (*models.Board, error)

	DeleteBoard(ctx context.Context, id uuid.UUID) error

	ListBoards(ctx context.Context, p PaginationParams) (*Page[models.Board], error)
	ListBoardsByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.Board], error)
}

type TeamService interface {
	GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error)

	CreateTeam(ctx context.Context, input CreateTeamInput) (*models.Team, error)

	UpdateTeam(ctx context.Context, id uuid.UUID, input UpdateTeamInput) (*models.Team, error)

	DeleteTeam(ctx context.Context, id uuid.UUID) error

	AddUserToTeam(ctx context.Context, input TeamMemberInput) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error

	AddTeamToProject(ctx context.Context, input TeamProjectInput) error
	RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error

	ListTeams(ctx context.Context, p PaginationParams) (*Page[models.Team], error)
	ListTeamsByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.Team], error)
	ListTeamsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Team], error)
}

type SubscriptionService interface {
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)

	CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*models.Subscription, error)

	DeleteSubscription(ctx context.Context, id uuid.UUID) error

	ListBySubObject(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, p PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptions(ctx context.Context, p PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByTarget(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, p PaginationParams) (*Page[models.Subscription], error)
}

type AttendanceService interface {
	GetAttendance(ctx context.Context, id uuid.UUID) (*models.Attendance, error)

	CreateAttendance(ctx context.Context, input CreateAttendanceInput) (*models.Attendance, error)

	UpdateAttendance(ctx context.Context, id uuid.UUID, input UpdateAttendanceInput) (*models.Attendance, error)

	DeleteAttendance(ctx context.Context, id uuid.UUID) error

	ListAttendances(ctx context.Context, p PaginationParams) (*Page[models.Attendance], error)
	ListAttendancesByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Attendance], error)
}

type StatusService interface {
	GetStatus(ctx context.Context, id uuid.UUID) (*models.Status, error)

	CreateStatus(ctx context.Context, input CreateStatusInput) (*models.Status, error)

	UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateStatusInput) (*models.Status, error)

	DeleteStatus(ctx context.Context, id uuid.UUID) error

	ListStatuses(ctx context.Context, p PaginationParams) (*Page[models.Status], error)
	ListStatusesByBoard(ctx context.Context, boardID uuid.UUID, p PaginationParams) (*Page[models.Status], error)
}

type DailyReportService interface {
	GetReport(ctx context.Context, id uuid.UUID) (*models.DailyReport, error)

	CreateReport(ctx context.Context, input CreateDailyReportInput) (*models.DailyReport, error)

	UpdateReport(ctx context.Context, id uuid.UUID, input UpdateDailyReportInput) (*models.DailyReport, error)

	DeleteReport(ctx context.Context, id uuid.UUID) error

	UpdateCompletedWork(ctx context.Context, id uuid.UUID, input UpdateCompletedWorkInput) (*models.CompletedWork, error)

	UpdateHelpRequest(ctx context.Context, id uuid.UUID, input UpdateHelpRequestInput) (*models.HelpRequest, error)
	DeleteHelpRequest(ctx context.Context, id uuid.UUID) error

	UpdateTomorrowPlan(ctx context.Context, id uuid.UUID, input UpdateTomorrowPlanInput) (*models.TomorrowPlans, error)

	ExportReportsToXLSX(ctx context.Context, input ReportsByDateInput, fileName string) ([]byte, error)

	ListReports(ctx context.Context, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByTask(ctx context.Context, taskID uuid.UUID, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByDateRange(ctx context.Context, startDate, endDate time.Time, p PaginationParams) (*Page[models.DailyReport], error)
	ListHelpRequestsByHelper(ctx context.Context, helperID uuid.UUID, p PaginationParams) (*Page[models.HelpRequest], error)
}

type ProblemService interface {
	GetProblem(ctx context.Context, id uuid.UUID) (*models.Problem, error)

	CreateProblem(ctx context.Context, input CreateProblemInput) (*models.Problem, error)

	UpdateProblem(ctx context.Context, id uuid.UUID, input UpdateProblemInput) (*models.Problem, error)

	DeleteProblem(ctx context.Context, id uuid.UUID) error

	ListProblems(ctx context.Context, p PaginationParams) (*Page[models.Problem], error)
	ListProblemsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Problem], error)
}

type ForumMessageService interface {
	GetMessage(ctx context.Context, id uuid.UUID) (*models.ForumMessage, error)

	CreateMessage(ctx context.Context, input CreateForumMessageInput) (*models.ForumMessage, error)

	UpdateMessage(ctx context.Context, id uuid.UUID, input UpdateForumMessageInput) (*models.ForumMessage, error)

	DeleteMessage(ctx context.Context, id uuid.UUID) error

	ListMessages(ctx context.Context, p PaginationParams) (*Page[models.ForumMessage], error)
	ListMessagesByProblem(ctx context.Context, problemID uuid.UUID, p PaginationParams) (*Page[models.ForumMessage], error)
}

type MCPService interface {
	ImproveReport(ctx context.Context, input ImproveReportInput) (*ImproveReportResult, error)
	StreamReport(ctx context.Context, input ImproveReportInput) (*MCPStream, error)
}
