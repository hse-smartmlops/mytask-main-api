package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthProvider struct {
	ID            *uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClientID      *string    `gorm:"size:100;uniqueIndex"`
	ClientSecret  *string    `gorm:"size:100"`
	WellKnownURL  *string    `gorm:"size:255"`
	Scopes        *string    `gorm:"type:text"`
	PostLogoutURI *string    `gorm:"size:255"`
	Deleted       *bool      `gorm:"type:boolean"`
	CreatedAt     *time.Time
	UpdatedAt     *time.Time

	Events *[]AuthEvent `gorm:"foreignKey:ProviderID"`
}

type AuthEvent struct {
	ID          *uuid.UUID  `gorm:"type:uuid;primaryKey"`
	EventID     *uuid.UUID  `gorm:"type:uuid;not null"` // ID из вебхука
	EventType   *string     `gorm:"size:50;not null"`
	UserID      *uuid.UUID  `gorm:"type:uuid"`
	ProviderID  *uuid.UUID  `gorm:"type:uuid"`
	ClientID    *string     `gorm:"size:100"`
	RealmID     *uuid.UUID  `gorm:"type:uuid"`
	IPAddress   *string     `gorm:"size:45"`
	ResourcePath *string    `gorm:"type:text"`
	OccurredAt  *time.Time  // время события из вебхука
	Error       *string     `gorm:"type:text"`
	Deleted     *bool       `gorm:"type:boolean"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Provider        *AuthProvider              `gorm:"foreignKey:ProviderID"`
	Details         *[]AuthEventDetail          `gorm:"foreignKey:EventID"`
	Representation  *[]AuthEventUserRepresentation `gorm:"foreignKey:EventID"`
}

type AuthEventDetail struct {
	ID                *uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID           *uuid.UUID `gorm:"type:uuid;not null"`
	AuthMethod        *string    `gorm:"size:50"`
	ClientAuthMethod  *string    `gorm:"size:50"`
	GrantType         *string    `gorm:"size:50"`
	SignatureRequired *bool
	Username          *string    `gorm:"size:255"`
	Scope             *string    `gorm:"type:text"`
	TokenID           *uuid.UUID `gorm:"type:uuid"`
	RefreshTokenID    *uuid.UUID `gorm:"type:uuid"`
	RefreshTokenType  *string    `gorm:"size:50"`
	UpdatedRefreshTokenID *uuid.UUID `gorm:"type:uuid"`

	Event *AuthEvent `gorm:"foreignKey:EventID"`
}

type AuthEventUserRepresentation struct {
	ID        *uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID   *uuid.UUID `gorm:"type:uuid;not null"`
	Username  *string    `gorm:"size:255"`
	FirstName *string    `gorm:"size:100"`
	LastName  *string    `gorm:"size:100"`
	Email     *string    `gorm:"size:255"`
	Enabled   *bool

	Event *AuthEvent `gorm:"foreignKey:EventID"`
}

type User struct {
	ID             *uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email          *string    `gorm:"size:100"`
	IsActive       *bool
	CreatedAt      *time.Time
	TgID           *string `gorm:"size:50"`
	TgUserID       *int64
	Profession     *string `gorm:"size:50"`
	EmailVerified  *bool
	FirstName      *string `gorm:"size:50"`
	LastName       *string `gorm:"size:50"`
	LastLogin      *time.Time
	AuthProviderID *uuid.UUID    `gorm:"type:uuid"`
	AuthProvider   *AuthProvider `gorm:"foreignKey:AuthProviderID;constraint:OnDelete:SET NULL;"`
	Roles          *[]Role        `gorm:"many2many:user_role;joinForeignKey:UserID;JoinReferences:RoleID"`
	Deleted        *bool         `gorm:"type:boolean"`
	UpdatedAt      *time.Time
}

type Role struct {
	ID          *uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        *string    `gorm:"size:50"`
	Description *string    `gorm:"size:200"`
	Users       *[]User     `gorm:"many2many:user_role;joinForeignKey:RoleID;JoinReferences:UserID"`
	Deleted     *bool      `gorm:"type:boolean"`
	UpdatedAt   *time.Time
}

type UserRole struct {
	RoleID     *uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     *uuid.UUID `gorm:"type:uuid;primaryKey"`
	AssignedAt *time.Time
	AssignedBy *uuid.UUID `gorm:"type:uuid"`
	Deleted    *bool      `gorm:"type:boolean"`
	UpdatedAt  *time.Time
}

type Session struct {
	ID           *uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       *uuid.UUID `gorm:"type:uuid;index"`
	User         *User      `gorm:"constraint:OnDelete:CASCADE;"`
	IDToken      *string    `gorm:"type:text"`
	SessionState *string    `gorm:"size:50"`
	ExpiresAt    *time.Time
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Deleted      *bool `gorm:"type:boolean"`
}

type Project struct {
	ID              *uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name            *string    `gorm:"size:100"`
	Description     *string    `gorm:"type:text"`
	CreatedAt       *time.Time
	Status          *string `gorm:"size:20"`
	GitlabProjectID *int
	GitlabURL       *string `gorm:"size:255"`
	Priority        *int16
	Boards          *[]Board       `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	Tasks           *[]Task        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	ProjectTeams    *[]ProjectTeam `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	Deleted         *bool         `gorm:"type:boolean"`
	UpdatedAt       *time.Time
}

type Board struct {
	ID          *uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID   *uuid.UUID `gorm:"type:uuid;index;not null"`
	Project     *Project   `gorm:"constraint:OnDelete:CASCADE;"`
	Name        *string    `gorm:"size:100"`
	Description *string    `gorm:"type:text"`
	Filter      *string    `gorm:"size:50"`
	Deleted     *bool      `gorm:"type:boolean"`
	UpdatedAt   *time.Time
}

type Task struct {
	ID            *uuid.UUID `gorm:"type:uuid;primaryKey"`
	Priority      *int16
	Name          *string    `gorm:"size:100"`
	Description   *string    `gorm:"type:text"`
	Status        *string    `gorm:"size:20"`
	CreatedBy     *uuid.UUID `gorm:"type:uuid"`
	AssignedTo    *uuid.UUID `gorm:"type:uuid"`
	Deadline      *time.Time
	TimeSpent     *string `gorm:"type:interval"`
	StartDate     *time.Time
	GitlabIssueID *int
	ProjectID     *uuid.UUID `gorm:"type:uuid;index"`
	Project       *Project   `gorm:"constraint:OnDelete:CASCADE;"`
	Category      *int8
	Deleted       *bool `gorm:"type:boolean"`
	UpdatedAt     *time.Time
}

type Team struct {
	ID           *uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name         *string       `gorm:"size:100" json:"name"`
	Description  *string       `gorm:"type:text" json:"description"`
	TeamMembers  *[]TeamMember  `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE;"`
	ProjectTeams *[]ProjectTeam `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE;"`
	Deleted      *bool         `gorm:"type:boolean"`
	UpdatedAt    *time.Time
}

type TeamMember struct {
	UserID         *uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	TeamID         *uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	Specialization *string    `gorm:"size:50"`
	User           *User       `gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	Team           *Team       `gorm:"constraint:OnDelete:CASCADE;foreignKey:TeamID;references:ID"`
	Deleted        *bool      `gorm:"type:boolean"`
	UpdatedAt      *time.Time
}

type ProjectTeam struct {
	ProjectID *uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	TeamID    *uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	Project   *Project    `gorm:"constraint:OnDelete:CASCADE;foreignKey:ProjectID;references:ID"`
	Team      *Team       `gorm:"constraint:OnDelete:CASCADE;foreignKey:TeamID;references:ID"`
	Deleted   *bool      `gorm:"type:boolean"`
	UpdatedAt *time.Time
}

type Attendance struct {
	ID            *uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID        *uuid.UUID `gorm:"type:uuid;index"`
	User          *User       `gorm:"constraint:OnDelete:CASCADE;"`
	Date          *time.Time `gorm:"type:date"`
	WorkdayHours  *int16
	PlannedStart  *time.Time `gorm:"type:time"`
	ActualStart   *time.Time `gorm:"type:time"`
	Status        *string    `gorm:"size:20"`
	Commits       *int16
	MergeRequests *int16
	CodeReviews   *int16
	EndWork       *time.Time `gorm:"type:time"`
	Deleted       *bool      `gorm:"type:boolean"`
	UpdatedAt     *time.Time
}

type DailyReport struct {
	ID         *uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     *uuid.UUID `gorm:"type:uuid;index"`
	User       *User       `gorm:"constraint:OnDelete:CASCADE;"`
	TaskID     *uuid.UUID `gorm:"type:uuid;index"`
	Task       *Task       `gorm:"foreignKey:TaskID;references:ID;"`
	ReportDate *time.Time `gorm:"type:date"`
	Status     *string    `gorm:"size:20"`
	CreatedAt  *time.Time `gorm:"type:date"`
	UpdatedAt  *time.Time `gorm:"type:date"`
	Deleted    *bool      `gorm:"type:boolean"`
}

type Problem struct {
	ID          *uuid.UUID `gorm:"type:uuid;primaryKey"`
	Description *[]string  `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID `gorm:"type:uuid;index"`
	Name        *string     `gorm:"type:varchar(255)"`
	User        *User       `gorm:"constraint:OnDelete:CASCADE;"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	Deleted     *bool      `gorm:"type:boolean"`
	Forum       *[]ForumMessage    `gorm:"foreignKey:ProblemID"`
}

type ForumMessage struct {
	ID          *uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProblemID   *uuid.UUID `gorm:"type:uuid;index"`
	Problem     *Problem    `gorm:"constraint:OnDelete:CASCADE;"`
	Description *[]string  `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID `gorm:"type:uuid;index"`
	User        *User       `gorm:"constraint:OnDelete:CASCADE;"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	Deleted     *bool      `gorm:"type:boolean"`
}

type ReportProblem struct {
	ReportID  *uuid.UUID  `gorm:"type:uuid;primaryKey;column:report_id"`
	ProblemID *uuid.UUID  `gorm:"type:uuid;primaryKey;column:problems_id"`
	UpdatedAt *time.Time  `gorm:"type:timestamp"`
	Deleted   *bool       `gorm:"type:boolean;default:false"`
	Report    *DailyReport `gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE;"`
	Problem   *Problem     `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE;"`
}

type HelpRequest struct {
	ID          *uuid.UUID  `gorm:"type:uuid;primaryKey"`
	HelperID    *uuid.UUID  `gorm:"type:uuid"`
	Description *string     `gorm:"type:varchar(255)"`
	Deleted     *bool       `gorm:"default:false"`
	Helper      *User        `gorm:"foreignKey:HelperID;references:ID"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	Report      *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
}

type CompletedWork struct {
	ID          *uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Description *string     `gorm:"type:text"`
	Deleted     *bool       `gorm:"default:false"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	Report      *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
}

type TomorrowPlans struct {
	ID          *uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Description *string     `gorm:"type:text"`
	Deleted     *bool       `gorm:"default:false"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	Report      *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
}
