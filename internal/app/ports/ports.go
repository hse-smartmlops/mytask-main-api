package ports

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

type SaveUserAvatarInput struct {
	Data        []byte
	ContentType string
}

type UserAvatarFile struct {
	FileName    string
	ContentType string
	Data        []byte
}

type UserRepository interface {
	ListUsers(ctx context.Context, params PaginationParams) (*Page[models.User], error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUserAvatar(ctx context.Context, id uuid.UUID, avatarPath string) error
	ClearUserAvatar(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	ListUsers(ctx context.Context, params PaginationParams) (*Page[models.User], error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
	SaveAvatar(ctx context.Context, id uuid.UUID, input SaveUserAvatarInput) (*models.User, error)
	DeleteAvatar(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetAvatar(ctx context.Context, id uuid.UUID) (*UserAvatarFile, error)
}

type AuthTokens struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	TokenType        string
	ExpiresAt        time.Time
}

type AuthRepositoryTokens struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	TokenType        string
}

type AuthLoginResult struct {
	UserID uuid.UUID
	Email  string
	Tokens AuthTokens
}

type AuthRefreshResult struct {
	Tokens AuthTokens
}

type AuthUserInfo struct {
	Subject           string
	Name              *string
	PreferredUsername *string
	GivenName         *string
	FamilyName        *string
	Email             *string
	EmailVerified     bool
}

type AuthRepository interface {
	Login(ctx context.Context, email, password string) (*AuthRepositoryTokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthRepositoryTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	UserInfo(ctx context.Context, token string) (*AuthUserInfo, error)
	Introspect(ctx context.Context, token string) (bool, error)
	ExchangeToken(ctx context.Context, subjectToken string) (*AuthRepositoryTokens, error)
}

type AuthService interface {
	TokenValidator
	Login(ctx context.Context, email, password string) (*AuthLoginResult, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserInfo(ctx context.Context, token string) (*AuthUserInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthRefreshResult, error)
}

type ObjectStorage interface {
	Upload(ctx context.Context, bucket, object string, data []byte, contentType string, metadata map[string]string) error
	Delete(ctx context.Context, bucket, object string) error
	Get(ctx context.Context, bucket, object string) ([]byte, string, error)
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

type CreateAttendanceInput struct {
	UserID        uuid.UUID
	Date          *time.Time
	WorkdayHours  *int16
	PlannedStart  *time.Time
	ActualStart   *time.Time
	Commits       *int16
	MergeRequests *int16
	CodeReviews   *int16
	EndWork       *time.Time
	Status        *string
}

type UpdateAttendanceInput struct {
	Date          *time.Time
	WorkdayHours  *int16
	PlannedStart  *time.Time
	ActualStart   *time.Time
	Commits       *int16
	MergeRequests *int16
	CodeReviews   *int16
	EndWork       *time.Time
	Status        *string
}

type AttendanceRepository interface {
	ListAttendances(ctx context.Context, params PaginationParams) (*Page[models.Attendance], error)
	ListAttendancesByUser(ctx context.Context, userID uuid.UUID) ([]models.Attendance, error)
	GetAttendanceByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error)
	CreateAttendance(ctx context.Context, attendance *models.Attendance) error
	UpdateAttendance(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Attendance, error)
	SoftDeleteAttendance(ctx context.Context, id uuid.UUID) error
}

type AttendanceService interface {
	ListAttendances(ctx context.Context, params PaginationParams) (*Page[models.Attendance], error)
	ListAttendancesByUser(ctx context.Context, userID uuid.UUID) ([]models.Attendance, error)
	GetAttendance(ctx context.Context, id uuid.UUID) (*models.Attendance, error)
	CreateAttendance(ctx context.Context, input CreateAttendanceInput) (*models.Attendance, error)
	UpdateAttendance(ctx context.Context, id uuid.UUID, input UpdateAttendanceInput) (*models.Attendance, error)
	DeleteAttendance(ctx context.Context, id uuid.UUID) error
}

type CompletedWorkInput struct {
	ID          *uuid.UUID
	Description *string
	TaskID      *uuid.UUID
}

type TomorrowPlanInput struct {
	ID          *uuid.UUID
	Description *string
	TaskID      *uuid.UUID
}

type HelpRequestInput struct {
	ID          *uuid.UUID
	Description *string
	HelperID    *uuid.UUID
	Status      *string
}

type CreateDailyReportInput struct {
	UserID        uuid.UUID
	ReportDate    *time.Time
	CompletedWork []CompletedWorkInput
	HelpRequests  []HelpRequestInput
	TomorrowPlans []TomorrowPlanInput
	Problems      []uuid.UUID
}

type UpdateDailyReportInput struct {
	Checked       *int8
	ReportDate    *time.Time
	CompletedWork []CompletedWorkInput
	HelpRequests  []HelpRequestInput
	TomorrowPlans []TomorrowPlanInput
	Problems      []uuid.UUID
}

type UpdateCompletedWorkInput struct {
	Description *string
	TaskID      *uuid.UUID
}

type UpdateHelpRequestInput struct {
	Description *string
	HelperID    *uuid.UUID
	Status      *string
}

type UpdateTomorrowPlanInput struct {
	Description *string
	TaskID      *uuid.UUID
}

type ReportsByDateInput struct {
	StartDate time.Time
	EndDate   time.Time
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

type ReportFile struct {
	FileName    string
	ContentType string
	Data        []byte
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
	ExportReportsToXLSX(ctx context.Context, input ReportsByDateInput) ([]byte, error)
	DownloadReportFile(ctx context.Context, id uuid.UUID) (*ReportFile, error)
}

type CreateProblemInput struct {
	Description []string
	CreatorID   *uuid.UUID
	Name        *string
}

type UpdateProblemInput struct {
	Description *[]string
	Name        *string
}

type ProblemRepository interface {
	ListProblems(ctx context.Context, params PaginationParams) (*Page[models.Problem], error)
	GetProblemByID(ctx context.Context, id uuid.UUID) (*models.Problem, error)
	CreateProblem(ctx context.Context, problem *models.Problem) error
	UpdateProblem(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Problem, error)
	SoftDeleteProblem(ctx context.Context, id uuid.UUID) error
}

type ProblemService interface {
	ListProblems(ctx context.Context, params PaginationParams) (*Page[models.Problem], error)
	GetProblem(ctx context.Context, id uuid.UUID) (*models.Problem, error)
	CreateProblem(ctx context.Context, input CreateProblemInput) (*models.Problem, error)
	UpdateProblem(ctx context.Context, id uuid.UUID, input UpdateProblemInput) (*models.Problem, error)
	DeleteProblem(ctx context.Context, id uuid.UUID) error
}

type CreateForumMessageInput struct {
	ProblemID   uuid.UUID
	Description []string
	CreatorID   *uuid.UUID
}

type UpdateForumMessageInput struct {
	Description *[]string
}

type ForumMessageRepository interface {
	ListMessages(ctx context.Context, params PaginationParams) (*Page[models.ForumMessage], error)
	GetMessageByID(ctx context.Context, id uuid.UUID) (*models.ForumMessage, error)
	ListMessagesByProblem(ctx context.Context, problemID uuid.UUID, params PaginationParams) (*Page[models.ForumMessage], error)
	CreateMessage(ctx context.Context, message *models.ForumMessage) error
	UpdateMessage(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.ForumMessage, error)
	SoftDeleteMessage(ctx context.Context, id uuid.UUID) error
}

type ForumMessageService interface {
	ListMessages(ctx context.Context, params PaginationParams) (*Page[models.ForumMessage], error)
	GetMessage(ctx context.Context, id uuid.UUID) (*models.ForumMessage, error)
	ListMessagesByProblem(ctx context.Context, problemID uuid.UUID, params PaginationParams) (*Page[models.ForumMessage], error)
	CreateMessage(ctx context.Context, input CreateForumMessageInput) (*models.ForumMessage, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, input UpdateForumMessageInput) (*models.ForumMessage, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) error
}
