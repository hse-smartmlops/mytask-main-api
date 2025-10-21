package ports

import (
	"time"

	"github.com/google/uuid"
)

type UserInput struct {
	Email         string
	IsActive      *bool
	TgID          *string
	TgUserID      *int64
	Profession    *string
	EmailVerified *bool
	FirstName     *string
	LastName      *string
}

type UserAvatarInput struct {
	Data        []byte
	ContentType string
}

type UserAvatarFile struct {
	FileName    string
	ContentType string
	Data        []byte
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

type CreateRoleInput struct {
	Name        string
	Description *string
}

type UpdateRoleInput struct {
	Name        *string
	Description *string
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

type CreateBoardInput struct {
	ProjectID   uuid.UUID
	Name        string
	Description *string
}

type UpdateBoardInput struct {
	Name        *string
	Description *string
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

type CreateSubscriptionInput struct {
	UserID         uuid.UUID
	SubscriptionID *uuid.UUID
	TypeID         *int8
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

type CreateProblemInput struct {
	Description []string
	CreatorID   *uuid.UUID
	Name        *string
}

type UpdateProblemInput struct {
	Description *[]string
	Name        *string
}

type CreateForumMessageInput struct {
	ProblemID   uuid.UUID
	Description []string
	CreatorID   *uuid.UUID
}

type UpdateForumMessageInput struct {
	Description *[]string
}

type ImproveReportInput struct {
	Description string
	UserText    string
	TaskID      string
	Meta        map[string]string
	ContentType string
	TimeoutMs   *int64
}

type ImproveReportResult struct {
	ImprovedText string
	ContentType  string
	Meta         map[string]string
	Duration     time.Duration
}

type MCPStreamEventType string

const (
	MCPStreamEventStatus MCPStreamEventType = "status"
	MCPStreamEventChunk  MCPStreamEventType = "chunk"
	MCPStreamEventFinal  MCPStreamEventType = "final"
	MCPStreamEventError  MCPStreamEventType = "error"
)

type MCPStatusEvent struct {
	State    string
	Message  string
	Progress int32
}

type MCPChunkEvent struct {
	Data  string
	Index int32
}

type MCPFinalEvent struct {
	Result      string
	ContentType string
}

type MCPErrorEvent struct {
	Code    int32
	Message string
}

type MCPStreamEvent struct {
	Type   MCPStreamEventType
	Status *MCPStatusEvent
	Chunk  *MCPChunkEvent
	Final  *MCPFinalEvent
	Error  *MCPErrorEvent
}

type MCPStream struct {
	Events <-chan MCPStreamEvent
	Errors <-chan error
	Close  func()
}
