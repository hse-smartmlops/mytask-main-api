package response

import "time"

type Board struct {
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	CreatedAt   *time.Time    `json:"created_at,omitempty"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty"`
	Statuses    []BoardStatus `json:"statuses,omitempty"`
}

type BoardStatus struct {
	ID        string     `json:"id"`
	Key       *string    `json:"key,omitempty"`
	Name      string     `json:"name"`
	Color     *string    `json:"color,omitempty"`
	IsDefault *bool      `json:"is_default,omitempty"`
	SortOrder *int       `json:"sort_order,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	IsOpen    *bool      `json:"is_open,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type BoardsPage struct {
	Boards []Board `json:"boards"`
}
