package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

//
// ===== Users & Roles =====
//

type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email         string    `gorm:"size:100"`
	IsActive      bool
	CreatedAt     time.Time `gorm:"type:timestamp"`
	TgID          string    `gorm:"size:50"`
	TgUserID      int64
	Profession    string    `gorm:"size:50"`
	EmailVerified bool
	FirstName     string    `gorm:"size:50"`
	LastName      string    `gorm:"size:50"`
	LastLogin     time.Time
	Deleted       bool      `gorm:"type:boolean"`
	UpdatedAt     time.Time `gorm:"type:timestamp"`

	// вместо many2many — явная джойн-модель
	UserRoles []UserRole `gorm:"foreignKey:UserID;references:ID"`
}

type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name        *string    `gorm:"size:50"`
	Description *string    `gorm:"size:200"`
	Deleted     *bool      `gorm:"type:boolean"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`

	// вместо many2many — явная джойн-модель
	UserRoles []UserRole `gorm:"foreignKey:RoleID;references:ID"`
}

type UserRole struct {
	RoleID     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AssignedAt *time.Time
	AssignedBy *uuid.UUID `gorm:"type:uuid"`
	Deleted    *bool      `gorm:"type:boolean"`
	CreatedAt  *time.Time `gorm:"type:timestamp"`
	UpdatedAt  *time.Time `gorm:"type:timestamp"`

	//Role *Role `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE"`
	//User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (UserRole) TableName() string { return "user_roles" }

//
// ===== Project / Board / Task =====
//

type Project struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name            *string    `gorm:"size:100"`
	Description     *string    `gorm:"type:text"`
	CreatedAt       *time.Time `gorm:"type:timestamp"`
	GitlabProjectID *int
	GitlabURL       *string    `gorm:"size:255"`
	CreatedBy       *uuid.UUID `gorm:"type:uuid"`
	Status          *string    `gorm:"type:varchar(50)"`
	Deleted         *bool      `gorm:"type:boolean"`
	UpdatedAt       *time.Time `gorm:"type:timestamp"`

	Boards       []Board       `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	ProjectTeams []ProjectTeam `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedByUser *User `gorm:"foreignKey:CreatedBy;references:ID"`
}

type Board struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProjectID   uuid.UUID  `gorm:"type:uuid;index;not null"`
	Name        *string    `gorm:"size:100"`
	Description *string    `gorm:"type:text"`
	Deleted     *bool      `gorm:"type:boolean"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`

	Project *Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`

	// явная джойн-таблица
	Statuses   []Status `gorm:"foreignKey:BoardID;references:ID;constraint:OnDelete:CASCADE"`
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
	StatusID       uuid.UUID  `gorm:"type:uuid;index"`
	Category      *int8
	Deleted       *bool      `gorm:"type:boolean"`
	CreatedAt     *time.Time `gorm:"type:timestamp"`
	UpdatedAt     *time.Time `gorm:"type:timestamp"`

	Status *Status `gorm:"foreignKey:StatusID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedByUser  *User    `gorm:"foreignKey:CreatedBy;references:ID"`
	AssignedToUser *User    `gorm:"foreignKey:AssignedTo;references:ID"`
}

//
// ===== Teams =====
//

type Team struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name        *string    `gorm:"size:100"`
	Description *string    `gorm:"type:text"`
	Deleted     *bool      `gorm:"type:boolean"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`

	TeamMembers  []TeamMember  `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
	ProjectTeams []ProjectTeam `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
}

type TeamMember struct {
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	TeamID         uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	Specialization *string    `gorm:"size:50"`
	Deleted        *bool      `gorm:"type:boolean"`
	CreatedAt      *time.Time `gorm:"type:timestamp"`
	UpdatedAt      *time.Time `gorm:"type:timestamp"`

	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Team *Team `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
}

type ProjectTeam struct {
	ProjectID uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	TeamID    uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	Deleted   *bool      `gorm:"type:boolean"`
	CreatedAt *time.Time `gorm:"type:timestamp"`
	UpdatedAt *time.Time `gorm:"type:timestamp"`

	Project *Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	Team    *Team    `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
}

//
// ===== Attendance / Reports =====
//

type Attendance struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID  `gorm:"type:uuid;index"`
	Date          *time.Time `gorm:"type:date"`
	WorkdayHours  *int16
	PlannedStart  *time.Time `gorm:"type:time"`
	ActualStart   *time.Time `gorm:"type:time"`
	Commits       *int16
	MergeRequests *int16
	CodeReviews   *int16
	EndWork       *time.Time `gorm:"type:time"`
	Deleted       *bool      `gorm:"type:boolean"`
	CreatedAt     *time.Time `gorm:"type:timestamp"`
	UpdatedAt     *time.Time `gorm:"type:timestamp"`

	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

type DailyReport struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID  `gorm:"type:uuid;index"`
	Checked    *int8      `gorm:"type:int8"`
	ReportDate *time.Time `gorm:"type:date"`
	CreatedAt  *time.Time `gorm:"type:date"`
	UpdatedAt  *time.Time `gorm:"type:date"`
	Deleted    *bool      `gorm:"type:boolean"`

	User           *User           `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	HelpRequests    []HelpRequest    `gorm:"foreignKey:ReportID;references:ID"`
	CompletedWork  []CompletedWork `gorm:"foreignKey:ReportID;references:ID"`
	TomorrowPlans  []TomorrowPlans `gorm:"foreignKey:ReportID;references:ID"`
	ReportProblems []ReportProblem `gorm:"foreignKey:ReportID;references:ID"`
}

//
// ===== Problems / Forum =====
//

type Problem struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Description pq.StringArray `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID     `gorm:"type:uuid;index"`
	Name        *string        `gorm:"type:varchar(255)"`
	CreatedAt   *time.Time     `gorm:"type:timestamp"`
	UpdatedAt   *time.Time     `gorm:"type:timestamp"`
	Deleted     *bool          `gorm:"type:boolean"`

	User  *User          `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:CASCADE"`
	Forum []ForumMessage `gorm:"foreignKey:ProblemID;references:ID"`
}

type ForumMessage struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ProblemID   uuid.UUID      `gorm:"type:uuid;index"`
	Description pq.StringArray `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID     `gorm:"type:uuid;index"`
	CreatedAt   *time.Time     `gorm:"type:timestamp"`
	UpdatedAt   *time.Time     `gorm:"type:timestamp"`
	Deleted     *bool          `gorm:"type:boolean"`

	Problem *Problem `gorm:"foreignKey:ProblemID;references:ID;constraint:OnDelete:CASCADE"`
	User    *User    `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:CASCADE"`
}

type ReportProblem struct {
	ReportID  uuid.UUID  `gorm:"type:uuid;primaryKey;column:report_id"`
	ProblemID uuid.UUID  `gorm:"type:uuid;primaryKey;column:problem_id"`
	CreatedAt *time.Time `gorm:"type:timestamp"`
	UpdatedAt *time.Time `gorm:"type:timestamp"`
	Deleted   *bool      `gorm:"type:boolean;default:false"`

	Report  *DailyReport `gorm:"foreignKey:ReportID;references:ID;constraint:OnDelete:CASCADE"`
	Problem *Problem     `gorm:"foreignKey:ProblemID;references:ID;constraint:OnDelete:CASCADE"`
}

//
// ===== Help / Completed / Plans =====
//

type HelpRequest struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	HelperID    *uuid.UUID  `gorm:"type:uuid"`
	Description *string     `gorm:"type:varchar(255)"`
	Deleted     *bool       `gorm:"type:boolean;default:false"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	Status      *string     `gorm:"type:varchar(30)"`
	UpdatedAt   *time.Time  `gorm:"type:timestamp"`
	CreatedAt   *time.Time  `gorm:"type:timestamp"`

	Helper *User        `gorm:"foreignKey:HelperID;references:ID"`
	Report *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
}

type CompletedWork struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	Description *string     `gorm:"type:text"`
	Deleted     *bool       `gorm:"type:boolean;default:false"`
	ReportID    *uuid.UUID  `gorm:"type:uuid"`
	TaskID      *uuid.UUID  `gorm:"type:uuid"`
	UpdatedAt   *time.Time  `gorm:"type:timestamp"`
	CreatedAt   *time.Time  `gorm:"type:timestamp"`

	Report *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
	Task   *Task        `gorm:"foreignKey:TaskID;references:ID"`
}

type TomorrowPlans struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Description *string    `gorm:"type:text"`
	Deleted     *bool      `gorm:"type:boolean;default:false"`
	ReportID    *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`

	Report *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
}

type Subscription struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;index"`
	SubscriptionId *uuid.UUID `gorm:"type:uuid"`
	TypeID         *int8      `gorm:"type:int8"`
	CreatedAt      *time.Time `gorm:"type:timestamp"`
	Deleted        *bool      `gorm:"type:boolean;default:false"`

	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

//
// ===== Statuses & relations =====
//

type Status struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	BoardID   uuid.UUID  `gorm:"type:uuid;index;not null"` // <-- ДОБАВЛЕНО
	SortOrder     *int        `gorm:"type:int"`
	Key       *string    `gorm:"type:varchar(8);uniqueIndex"`
	Name      *string    `gorm:"type:varchar(50)"`
	Color     *string    `gorm:"type:varchar(16)"`
	IsDefault *bool      `gorm:"type:boolean;default:false"`
	IsActive  *bool      `gorm:"type:boolean;default:true"`
	IsOpen    *bool      `gorm:"type:boolean;default:true"`
	CreatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Deleted   *bool      `gorm:"type:boolean;default:false"`

	// Связь: Status принадлежит Board
	Board *Board `gorm:"foreignKey:BoardID;references:ID;constraint:OnDelete:CASCADE"`

	// Связь: много задач могут иметь этот статус
	Tasks []Task `gorm:"foreignKey:StatusID;references:ID"`
}

