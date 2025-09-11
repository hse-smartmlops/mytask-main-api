package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ===== Users & Roles =====

type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email          *string    `gorm:"size:100"`
	IsActive       *bool
	CreatedAt      *time.Time `gorm:"type:timestamp"`
	TgID           *string    `gorm:"size:50"`
	TgUserID       *int64
	Profession     *string    `gorm:"size:50"`
	EmailVerified  *bool
	FirstName      *string    `gorm:"size:50"`
	LastName       *string    `gorm:"size:50"`
	LastLogin      *time.Time
	Roles          []Role        `gorm:"many2many:user_role;joinForeignKey:UserID;JoinReferences:RoleID"`
	Deleted        *bool         `gorm:"type:boolean"`
	UpdatedAt      *time.Time `gorm:"type:timestamp"`
}

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        *string   `gorm:"size:50"`
	Description *string   `gorm:"size:200"`
	Users       []User    `gorm:"many2many:user_role;joinForeignKey:RoleID;JoinReferences:UserID"`
	Deleted     *bool     `gorm:"type:boolean"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
}

type UserRole struct {
	RoleID     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AssignedAt *time.Time
	AssignedBy *uuid.UUID `gorm:"type:uuid"`
	Deleted    *bool      `gorm:"type:boolean"`
	CreatedAt  *time.Time `gorm:"type:timestamp"`
	UpdatedAt  *time.Time `gorm:"type:timestamp"`
}

// ===== Project / Board / Task =====

type Project struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name            *string    `gorm:"size:100"`
	Description     *string    `gorm:"type:text"`
	CreatedAt       *time.Time `gorm:"type:timestamp"`
	GitlabProjectID *int
	GitlabURL       *string    `gorm:"size:255"`
	Boards          []Board       `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	Tasks           []Task        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	ProjectTeams    []ProjectTeam `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	Deleted         *bool         `gorm:"type:boolean"`
	UpdatedAt       *time.Time    `gorm:"type:timestamp"`
}

type Board struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ProjectID   uuid.UUID      `gorm:"type:uuid;index;not null"`
	Project     *Project       `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	Name        *string        `gorm:"size:100"`
	Description *string        `gorm:"type:text"`
	Deleted     *bool          `gorm:"type:boolean"`
	CreatedAt   *time.Time     `gorm:"type:timestamp"`
	UpdatedAt   *time.Time     `gorm:"type:timestamp"`

	StatusBoards []StatusBoard `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE;"`
}
type Task struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Priority      *int16
	Name          *string    `gorm:"size:100"`
	Description   *string    `gorm:"type:text"`
	CreatedBy     *uuid.UUID `gorm:"type:uuid"`
	AssignedTo    *uuid.UUID `gorm:"type:uuid"`
	Deadline      *time.Time
	TimeSpent     *string    `gorm:"type:interval"`
	StartDate     *time.Time
	GitlabIssueID *int
	ProjectID     uuid.UUID  `gorm:"type:uuid;index"`
	Project       *Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	Category      *int8
	Deleted       *bool `gorm:"type:boolean"`
	CreatedAt     *time.Time `gorm:"type:timestamp"`
	UpdatedAt     *time.Time `gorm:"type:timestamp"`

	StatusTasks []StatusTask `gorm:"foreignKey:TaskID"`
	CreatedByUser  *User `gorm:"foreignKey:CreatedBy"`
    AssignedToUser *User `gorm:"foreignKey:AssignedTo"`
}

// ===== Teams =====

type Team struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey"`
	Name         *string       `gorm:"size:100"`
	Description  *string       `gorm:"type:text"`
	TeamMembers  []TeamMember  `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE;"`
	ProjectTeams []ProjectTeam `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE;"`
	Deleted      *bool         `gorm:"type:boolean"`
	CreatedAt    *time.Time     `gorm:"type:timestamp"`
	UpdatedAt    *time.Time	 	`gorm:"type:timestamp"`
}

type TeamMember struct {
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	TeamID         uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	Specialization *string   `gorm:"size:50"`
	User           *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	Team           *Team     `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE;"`
	Deleted        *bool     `gorm:"type:boolean"`
	CreatedAt      *time.Time `gorm:"type:timestamp"`
	UpdatedAt      *time.Time `gorm:"type:timestamp"`
}

type ProjectTeam struct {
	ProjectID uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	TeamID    uuid.UUID `gorm:"type:uuid;primaryKey;index"`
	Project   *Project  `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE;"`
	Team      *Team     `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE;"`
	Deleted   *bool     `gorm:"type:boolean"`
	CreatedAt *time.Time `gorm:"type:timestamp"`
	UpdatedAt *time.Time `gorm:"type:timestamp"`
}

// ===== Attendance / Reports =====

type Attendance struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID  `gorm:"type:uuid;index"`
	User          *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Date          *time.Time `gorm:"type:date"`
	WorkdayHours  *int16
	PlannedStart  *time.Time `gorm:"type:time"`
	ActualStart   *time.Time `gorm:"type:time"`
	Commits       *int16
	MergeRequests *int16
	CodeReviews   *int16
	EndWork       *time.Time `gorm:"type:time"`
	Deleted       *bool      `gorm:"type:boolean"`
	CreatedAt       *time.Time `gorm:"type:timestamp"`
	UpdatedAt     *time.Time `gorm:"type:timestamp"`
}

type DailyReport struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID  `gorm:"type:uuid;index"`
	User          *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	TaskID        *uuid.UUID `gorm:"type:uuid;index"`
	Task          *Task      `gorm:"foreignKey:TaskID;references:ID"`
	ReportDate    *time.Time `gorm:"type:date"`
	CreatedAt     *time.Time `gorm:"type:date"`
	UpdatedAt     *time.Time `gorm:"type:date"`
	Deleted       *bool      `gorm:"type:boolean"`

	HelpRequest   *HelpRequest     `gorm:"foreignKey:ReportID"`
	CompletedWork []CompletedWork  `gorm:"foreignKey:ReportID"`
	TomorrowPlans []TomorrowPlans  `gorm:"foreignKey:ReportID"`
	ReportProblems []ReportProblem `gorm:"foreignKey:ReportID"`
}

// ===== Problems / Forum =====

type Problem struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Description pq.StringArray  `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID `gorm:"type:uuid;index"`
	Name        *string    `gorm:"type:varchar(255)"`

	User        *User      `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:CASCADE;"`

	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	Deleted     *bool      `gorm:"type:boolean"`
	Forum       []ForumMessage `gorm:"foreignKey:ProblemID"`
}

type ForumMessage struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProblemID   uuid.UUID  `gorm:"type:uuid;index"`
	Problem     *Problem   `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE;"`
	Description pq.StringArray  `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID `gorm:"type:uuid;index"`
	User        *User      `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:CASCADE;"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	Deleted     *bool      `gorm:"type:boolean"`
}

type ReportProblem struct {
	ReportID  uuid.UUID   `gorm:"type:uuid;primaryKey;column:report_id"`
	ProblemID uuid.UUID   `gorm:"type:uuid;primaryKey;column:problem_id"`
	CreatedAt       *time.Time `gorm:"type:timestamp"`
	UpdatedAt *time.Time  `gorm:"type:timestamp"`
	Deleted   *bool       `gorm:"type:boolean;default:false"`
	Report    *DailyReport `gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE;"`
	Problem   *Problem     `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE;"`
}

// ===== Help / Completed / Plans =====

type HelpRequest struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	HelperID    *uuid.UUID  `gorm:"type:uuid"`
	Description *string     `gorm:"type:varchar(255)"`
	Deleted     *bool       `gorm:"default:false"`
	Helper      *User       `gorm:"foreignKey:HelperID;references:ID"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	Report      *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	CreatedAt       *time.Time `gorm:"type:timestamp"`
}

type CompletedWork struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	Description *string     `gorm:"type:text"`
	Deleted     *bool       `gorm:"default:false"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	Report      *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
}

type TomorrowPlans struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	Description *string     `gorm:"type:text"`
	Deleted     *bool       `gorm:"default:false"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	Report      *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	CreatedAt       *time.Time `gorm:"type:timestamp"`
}

type Subscription struct {
	ID        *uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	User      *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	SubscriptionId *uuid.UUID  `gorm:"type:uuid"`
	TypeID    *int8     `gorm:"type:int8"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	Deleted     *bool       `gorm:"default:false"`
}

type Status struct {
	ID        *uuid.UUID `gorm:"type:uuid;primaryKey"`
	Key       *string    `gorm:"type:varchar(8);uniqueIndex"`
	Name      *string    `gorm:"type:varchar(50)"`
	Color     *string    `gorm:"type:varchar(16)"`
	IsDefault *bool      `gorm:"default:false"`
	IsActive  *bool      `gorm:"default:true"`
	IsOpen    *bool      `gorm:"default:true"`
	CreatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Deleted   *bool      `gorm:"default:false"`

	// Relations
	Boards []*Board `gorm:"many2many:status_boards;joinForeignKey:StatusID;joinReferences:BoardID"`
	Tasks []*Task `gorm:"many2many:status_tasks;joinForeignKey:StatusID;joinReferences:TaskID"`
}

type StatusBoard struct {
	StatusID *uuid.UUID `gorm:"type:uuid;primaryKey"`
	BoardID  *uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Deleted   *bool      `gorm:"default:false"`

	// Relations
	Status *Status `gorm:"foreignKey:StatusID;references:ID;constraint:OnDelete:CASCADE;"`
	Board  *Board  `gorm:"foreignKey:BoardID;references:ID;constraint:OnDelete:CASCADE;"`
}

type StatusTask struct {
	StatusID *uuid.UUID `gorm:"type:uuid;primaryKey"`
	TaskID   *uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Deleted   *bool      `gorm:"default:false"`

	// Relations
	Status *Status `gorm:"foreignKey:StatusID;references:ID;constraint:OnDelete:CASCADE;"`
	Task   *Task   `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:CASCADE;"`
}