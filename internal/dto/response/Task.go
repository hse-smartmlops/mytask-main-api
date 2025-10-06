package response

import (
	"time"
)

// Полная информация о задаче (для GetTaskByID)
type GetTaskByIDResponse struct {
	ID            string    `json:"id"`
	StatusID     string    `json:"status_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Priority      int16     `json:"priority"`
	CreatedBy     UserShort `json:"created_by"`            // Связь с users
	AssignedTo    UserShort `json:"assigned_to,omitempty"` // Связь с users
	Deadline      time.Time `json:"deadline,omitempty"`    // DATE в БД
	TimeSpent     string    `json:"time_spent"`            // INTERVAL в БД
	StartDate     time.Time `json:"start_date"`            // DATE в БД
	GitlabIssueID int       `json:"gitlab_issue_id,omitempty"`
	Category      int8      `json:"community"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// Краткая информация о задаче (для списков)
type TaskShort struct {
	ID        string    `json:"id"`
	StatusID     string    `json:"status_id"`
	Name      string    `json:"name"`
	Priority  int16     `json:"priority"`
	StartDate time.Time `json:"start_date"`
	Deadline  time.Time `json:"deadline,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// Ответ для списка задач с пагинацией
type TaskListResponse struct {
	Tasks      []TaskShort `json:"tasks"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page,omitempty"`
	PageSize   int         `json:"page_size,omitempty"`
}

// Универсальный ответ для операций
type TaskUniversaResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Вспомогательная структура для пользователя
type UserShort struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
