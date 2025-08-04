package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	Email          string
	IsActive       bool
	CreatedAt      time.Time
	TgID           string
	TgUserID       int64
	Profession     *string
	EmailVerified  bool
	FirstName      string
	LastName       string
	LastLogin      *time.Time
	AuthProviderID *uuid.UUID
}

type Roles struct {
	ID          uuid.UUID
	Name        string
	Description *string
}

type UserRole struct {
	RoleID     uuid.UUID
	UserID     uuid.UUID
	AssignedAt *time.Time
	AssignedBy uuid.UUID
}

type Sessions struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	IDToken      string
	SessionState string
	ExpressedAt  *time.Time
	CreatedAt    *time.Time
}

type AuthEvents struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	EventType string
	Provider  string
	IPAddress string
}

type AuthProvider struct {
	ID            uuid.UUID
	ClientID      string
	ClientSecret  string
	WellKnownURL  string
	Scopes        string
	PostLogoutURI string
	CreatedAt     *time.Time
	UpdateddAt    *time.Time
}

type Task struct {
	ID            uuid.UUID
	Priority      int16
	Name          string
	Description   string
	Status        string
	CreatedBy     uuid.UUID
	AssignedTo    uuid.UUID
	Deadline      *time.Time
	TimeSpent     *time.Time
	StartDate     *time.Time
	GitlabIssueID int16
	ProjectID     uuid.UUID
}

type Board struct {
	ID          uuid.UUID
	Name        string
	Description string
	ProjectID   uuid.UUID
	Filter      string
}

type Project struct {
	ID              uuid.UUID
	Name            string
	Description     string
	CreatedAt       *time.Time
	Status          string
	GitlabProjectID int64
	GitlabURL       string
	Priority        int16
}

type Teams struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type TeamMembers struct {
	UserID         uuid.UUID
	TeamID         uuid.UUID
	Specialization string
}

type ProjectTeams struct {
	ProjectID uuid.UUID
	TeamID    uuid.UUID
}

type Attendances struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Date          time.Time
	WorkdayHours  int16
	PlannedStart  time.Time
	ActualStart   time.Time
	EndWork       time.Time
	Status        string
	Commits       int16
	MergeRequests int16
	CodeReviews   int16
}

type DailyReport struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ReportDate    time.Time
	CompletedWork JSONB
	PlanTomorrow  JSONB
	Problems      JSONB
	HelpRequest   JSONB
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// JSONB тип для работы с jsonb в Postgres
type JSONB map[string]interface{}

func (j JSONB) GormDataType() string {
	return "jsonb"
}
