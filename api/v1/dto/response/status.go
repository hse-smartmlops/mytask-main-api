package response

import "time"

type Status struct {
	ID        string     `json:"id"`
	BoardID   string     `json:"board_id"`
	Name      string     `json:"name"`
	Key       *string    `json:"key,omitempty"`
	Color     *string    `json:"color,omitempty"`
	IsDefault *bool      `json:"is_default,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	IsOpen    *bool      `json:"is_open,omitempty"`
	SortOrder *int       `json:"sort_order,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type StatusesPage struct {
	Statuses []Status `json:"statuses"`
}
