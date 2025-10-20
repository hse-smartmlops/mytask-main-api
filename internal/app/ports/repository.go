package ports

import (
	"context"
	"time"

	"emplacc-api/internal/domain/models"
	pb "emplacc-api/pkg/pb/v1"

	"github.com/google/uuid"
)

type UserRepository interface {
	ListUsers(ctx context.Context, params PaginationParams) (*Page[models.User], error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUserAvatar(ctx context.Context, id uuid.UUID, avatarPath string) error
	ClearUserAvatar(ctx context.Context, id uuid.UUID) error
}

type RoleRepository interface {
	ListRoles(ctx context.Context, params PaginationParams) (*Page[models.Role], error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	CreateRole(ctx context.Context, role *models.Role) error
	UpdateRole(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Role, error)
	SoftDeleteRole(ctx context.Context, id uuid.UUID) error
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
	ListProjects(ctx context.Context, params PaginationParams) (*Page[models.Project], error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	CreateProject(ctx context.Context, project *models.Project) error
	UpdateProject(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Project, error)
	SoftDeleteProject(ctx context.Context, id uuid.UUID) error
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

type SubscriptionRepository interface {
	ListSubscriptions(ctx context.Context, params PaginationParams) (*Page[models.Subscription], error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Subscription], error)
	ListSubscriptionsByTarget(ctx context.Context, subscriptionID uuid.UUID, typeID *int8, params PaginationParams) (*Page[models.Subscription], error)
	CreateSubscription(ctx context.Context, sub *models.Subscription) error
	SoftDeleteSubscription(ctx context.Context, id uuid.UUID) error
}

type BoardRepository interface {
	ListBoards(ctx context.Context, params PaginationParams) (*Page[models.Board], error)
	GetBoardByID(ctx context.Context, id uuid.UUID) (*models.Board, error)
	ListBoardsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Board, error)
	CreateBoard(ctx context.Context, board *models.Board) error
	UpdateBoard(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Board, error)
	SoftDeleteBoard(ctx context.Context, id uuid.UUID) error
}

type DailyReportRepository interface {
	ListReports(ctx context.Context, params PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.DailyReport], error)
	ListReportsByProject(ctx context.Context, projectID uuid.UUID) ([]models.DailyReport, error)
	ListReportsByTask(ctx context.Context, taskID uuid.UUID) ([]models.DailyReport, error)
	ListReportsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.DailyReport, error)
	ListHelpRequestsByHelper(ctx context.Context, helperID uuid.UUID) ([]models.HelpRequest, error)
	GetReportByID(ctx context.Context, id uuid.UUID) (*models.DailyReport, error)
	CreateReport(ctx context.Context, input CreateDailyReportInput) (*models.DailyReport, error)
	UpdateReport(ctx context.Context, id uuid.UUID, input UpdateDailyReportInput) (*models.DailyReport, error)
	SoftDeleteReport(ctx context.Context, id uuid.UUID) error
	UpdateReportStorage(ctx context.Context, id uuid.UUID, storageObject string) error
	UpdateCompletedWork(ctx context.Context, id uuid.UUID, input UpdateCompletedWorkInput) (*models.CompletedWork, error)
	UpdateHelpRequest(ctx context.Context, id uuid.UUID, input UpdateHelpRequestInput) (*models.HelpRequest, error)
	SoftDeleteHelpRequest(ctx context.Context, id uuid.UUID) error
	UpdateTomorrowPlan(ctx context.Context, id uuid.UUID, input UpdateTomorrowPlanInput) (*models.TomorrowPlans, error)
}

type TeamRepository interface {
	ListAllTeams(ctx context.Context) ([]models.Team, error)
	ListTeams(ctx context.Context, params PaginationParams) (*Page[models.Team], error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error)
	ListTeamsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Team, error)
	ListTeamsByUser(ctx context.Context, userID uuid.UUID) ([]models.Team, error)
	CreateTeam(ctx context.Context, team *models.Team) error
	UpdateTeam(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Team, error)
	SoftDeleteTeam(ctx context.Context, id uuid.UUID) error
	AddUserToTeam(ctx context.Context, input TeamMemberInput) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error
	AddTeamToProject(ctx context.Context, input TeamProjectInput) error
	RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error
}

type AttendanceRepository interface {
	ListAttendances(ctx context.Context, params PaginationParams) (*Page[models.Attendance], error)
	ListAttendancesByUser(ctx context.Context, userID uuid.UUID) ([]models.Attendance, error)
	GetAttendanceByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error)
	CreateAttendance(ctx context.Context, attendance *models.Attendance) error
	UpdateAttendance(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Attendance, error)
	SoftDeleteAttendance(ctx context.Context, id uuid.UUID) error
}

type ProblemRepository interface {
	ListProblems(ctx context.Context, params PaginationParams) (*Page[models.Problem], error)
	ListProblemsByUser(ctx context.Context, userID uuid.UUID, params PaginationParams) (*Page[models.Problem], error)
	GetProblemByID(ctx context.Context, id uuid.UUID) (*models.Problem, error)
	CreateProblem(ctx context.Context, problem *models.Problem) error
	UpdateProblem(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Problem, error)
	SoftDeleteProblem(ctx context.Context, id uuid.UUID) error
}

type StatusRepository interface {
	ListStatuses(ctx context.Context, params PaginationParams) (*Page[models.Status], error)
	GetStatusByID(ctx context.Context, id uuid.UUID) (*models.Status, error)
	ListStatusesByBoard(ctx context.Context, boardID uuid.UUID) ([]models.Status, error)
	CreateStatus(ctx context.Context, status *models.Status) error
	UpdateStatus(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Status, error)
	SoftDeleteStatus(ctx context.Context, id uuid.UUID) error
}

type ForumMessageRepository interface {
	ListMessages(ctx context.Context, params PaginationParams) (*Page[models.ForumMessage], error)
	GetMessageByID(ctx context.Context, id uuid.UUID) (*models.ForumMessage, error)
	ListMessagesByProblem(ctx context.Context, problemID uuid.UUID, params PaginationParams) (*Page[models.ForumMessage], error)
	CreateMessage(ctx context.Context, message *models.ForumMessage) error
	UpdateMessage(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.ForumMessage, error)
	SoftDeleteMessage(ctx context.Context, id uuid.UUID) error
}

type MCPRepository interface {
	ProcessTask(ctx context.Context, req *pb.ProcessTaskRequest) (*pb.ProcessTaskResponse, error)
	StreamProcessTask(ctx context.Context, req *pb.ProcessTaskRequest) (pb.MCPService_StreamProcessTaskClient, error)
}
