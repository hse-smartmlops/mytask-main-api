package response

import (
	"time"
)

// Полная информация о задаче (для GetTaskByID)
type TaskDetail struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	BoardID       *string    `json:"board_id,omitempty"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	Priority      int        `json:"priority"`
	CreatedBy     UserShort  `json:"created_by"`     // Связь с users
	AssignedTo    *UserShort `json:"assigned_to,omitempty"` // Связь с users
	Deadline      *time.Time `json:"deadline,omitempty"`    // DATE в БД
	TimeSpent     string     `json:"time_spent"`     // INTERVAL в БД
	StartDate     time.Time  `json:"start_date"`     // DATE в БД
	GitlabIssueID *int       `json:"gitlab_issue_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Краткая информация о задаче (для списков)
type TaskShort struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Priority    int       `json:"priority"`
	StartDate   time.Time `json:"start_date"`
	Deadline    *time.Time `json:"deadline,omitempty"`
}

// Ответ для списка задач с пагинацией
type TaskList struct {
	Tasks      []TaskShort `json:"tasks"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page,omitempty"`
	PageSize   int         `json:"page_size,omitempty"`
}

// Универсальный ответ для операций
type TaskOperation struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Вспомогательная структура для пользователя
type UserShort struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}