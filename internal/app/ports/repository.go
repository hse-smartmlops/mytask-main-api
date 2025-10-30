package ports

import (
	"context"
	"time"

	"emplacc-api/internal/domain/models"
	pb "emplacc-api/pkg/pb/v1"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	CreateUser(ctx context.Context, user *models.User) error

	UpdateUser(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.User, error)

	SoftDeleteUser(ctx context.Context, id uuid.UUID) error

	UpdateUserAvatar(ctx context.Context, id uuid.UUID, avatarPath string) error

	ClearUserAvatar(ctx context.Context, id uuid.UUID) error

	ListUsers(ctx context.Context, p PaginationParams) (*Page[models.User], error)
}

type RoleRepository interface {
	GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)

	CreateRole(ctx context.Context, role *models.Role) error

	UpdateRole(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Role, error)

	SoftDeleteRole(ctx context.Context, id uuid.UUID) error

	ListRoles(ctx context.Context, p PaginationParams) (*Page[models.Role], error)
}

type AuthRepository interface {
	Login(ctx context.Context, email, password string) (*AuthRepositoryTokens, error)

	RefreshToken(ctx context.Context, refreshToken string) (*AuthRepositoryTokens, error)

	Logout(ctx context.Context, refreshToken string) error

	UserInfo(ctx context.Context, token string) (*AuthUserInfo, error)

	Introspect(ctx context.Context, token string) (bool, error)

	ExchangeToken(ctx context.Context, subjectToken string) (*AuthRepositoryTokens, error)
}

type ProjectRepository interface {
	GetProjectByID(ctx context.Context, id uuid.UUID) (*models.Project, error)

	CreateProject(ctx context.Context, project *models.Project) error

	UpdateProject(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Project, error)

	SoftDeleteProject(ctx context.Context, id uuid.UUID) error

	ListProjects(ctx context.Context, p PaginationParams) (*Page[models.Project], error)
	ListProjectsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Project], error)
	ListProjectsByTeam(ctx context.Context, teamID uuid.UUID, p PaginationParams) (*Page[models.Project], error)
}

type TaskRepository interface {
	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	GetTaskBoardAndProjectIDs(ctx context.Context, taskID uuid.UUID) (boardID uuid.UUID, projectID uuid.UUID, err error)

	CreateTask(ctx context.Context, task *models.Task) error

	UpdateTask(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Task, error)

	SoftDeleteTask(ctx context.Context, id uuid.UUID) error

	MoveTaskBetweenStatuses(ctx context.Context, taskID uuid.UUID, newStatusID uuid.UUID) (*models.Task, error)

	ListTasks(ctx context.Context, p PaginationParams) (*Page[models.Task], error)
	ListTasksByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListTasksByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListTasksByBoard(ctx context.Context, boardID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListActiveTasksByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Task], error)
	ListTasksByUserAndProject(ctx context.Context, userID, projectID uuid.UUID, p PaginationParams) (*Page[models.Task], error)

	StatusExists(ctx context.Context, statusID uuid.UUID) (bool, error)
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
}

type SubscriptionRepository interface {
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)

	CreateSubscription(ctx context.Context, sub *models.Subscription) error

	SoftDeleteSubscription(ctx context.Context, id uuid.UUID) error

	ListSubscriptions(ctx context.Context, p PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByTarget(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, p PaginationParams) (*Page[models.Subscription], error)
}

type BoardRepository interface {
	GetBoardByID(ctx context.Context, id uuid.UUID) (*models.Board, error)

	CreateBoard(ctx context.Context, board *models.Board) error

	UpdateBoard(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Board, error)

	SoftDeleteBoard(ctx context.Context, id uuid.UUID) error

	ListBoards(ctx context.Context, p PaginationParams) (*Page[models.Board], error)
	ListBoardsByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.Board], error)
}

type DailyReportRepository interface {
	GetReportByID(ctx context.Context, id uuid.UUID) (*models.DailyReport, error)

	CreateReport(ctx context.Context, input CreateDailyReportInput) (*models.DailyReport, error)

	UpdateReport(ctx context.Context, id uuid.UUID, input UpdateDailyReportInput) (*models.DailyReport, error)

	SoftDeleteReport(ctx context.Context, id uuid.UUID) error

	UpdateReportStorage(ctx context.Context, id uuid.UUID, storageObject string) error

	UpdateCompletedWork(ctx context.Context, id uuid.UUID, input UpdateCompletedWorkInput) (*models.CompletedWork, error)

	UpdateHelpRequest(ctx context.Context, id uuid.UUID, input UpdateHelpRequestInput) (*models.HelpRequest, error)
	SoftDeleteHelpRequest(ctx context.Context, id uuid.UUID) error

	UpdateTomorrowPlan(ctx context.Context, id uuid.UUID, input UpdateTomorrowPlanInput) (*models.TomorrowPlans, error)

	ListReports(ctx context.Context, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByTask(ctx context.Context, taskID uuid.UUID, p PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByDateRange(ctx context.Context, startDate, endDate time.Time, p PaginationParams) (*Page[models.DailyReport], error)
	ListHelpRequestsByHelper(ctx context.Context, helperID uuid.UUID, p PaginationParams) (*Page[models.HelpRequest], error)
	ListReportsByDateRangeWithoutPagination(ctx context.Context, startDate, endDate time.Time) (*[]models.DailyReport, error)
}

type TeamRepository interface {
	GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error)

	CreateTeam(ctx context.Context, team *models.Team) error

	UpdateTeam(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Team, error)

	SoftDeleteTeam(ctx context.Context, id uuid.UUID) error

	AddUserToTeam(ctx context.Context, input TeamMemberInput) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error

	AddTeamToProject(ctx context.Context, input TeamProjectInput) error
	RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error

	ListTeams(ctx context.Context, p PaginationParams) (*Page[models.Team], error)
	ListTeamsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Team], error)
	ListTeamsByProject(ctx context.Context, projectID uuid.UUID, p PaginationParams) (*Page[models.Team], error)
}

type AttendanceRepository interface {
	GetAttendanceByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error)

	CreateAttendance(ctx context.Context, attendance *models.Attendance) error

	UpdateAttendance(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Attendance, error)

	SoftDeleteAttendance(ctx context.Context, id uuid.UUID) error

	ListAttendances(ctx context.Context, p PaginationParams) (*Page[models.Attendance], error)
	ListAttendancesByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Attendance], error)
}

type ProblemRepository interface {
	GetProblemByID(ctx context.Context, id uuid.UUID) (*models.Problem, error)

	CreateProblem(ctx context.Context, problem *models.Problem) error

	UpdateProblem(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Problem, error)

	SoftDeleteProblem(ctx context.Context, id uuid.UUID) error

	ListProblems(ctx context.Context, p PaginationParams) (*Page[models.Problem], error)
	ListProblemsByUser(ctx context.Context, userID uuid.UUID, p PaginationParams) (*Page[models.Problem], error)
}

type StatusRepository interface {
	GetStatusByID(ctx context.Context, id uuid.UUID) (*models.Status, error)

	CreateStatus(ctx context.Context, status *models.Status) error

	UpdateStatus(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Status, error)

	SoftDeleteStatus(ctx context.Context, id uuid.UUID) error

	ListStatuses(ctx context.Context, p PaginationParams) (*Page[models.Status], error)
	ListStatusesByBoard(ctx context.Context, boardID uuid.UUID, p PaginationParams) (*Page[models.Status], error)
}

type ForumMessageRepository interface {
	GetMessageByID(ctx context.Context, id uuid.UUID) (*models.ForumMessage, error)

	CreateMessage(ctx context.Context, message *models.ForumMessage) error

	UpdateMessage(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.ForumMessage, error)

	SoftDeleteMessage(ctx context.Context, id uuid.UUID) error

	ListMessages(ctx context.Context, p PaginationParams) (*Page[models.ForumMessage], error)
	ListMessagesByProblem(ctx context.Context, problemID uuid.UUID, p PaginationParams) (*Page[models.ForumMessage], error)
}

type MCPRepository interface {
	ProcessTask(ctx context.Context, req *pb.ProcessTaskRequest) (*pb.ProcessTaskResponse, error)
	StreamProcessTask(ctx context.Context, req *pb.ProcessTaskRequest) (pb.MCPService_StreamProcessTaskClient, error)
}
