package request

import "github.com/google/uuid"

type CreateStatus struct {
	BoardID   string  `json:"board_id"`
	Name      string  `json:"name"`
	Key       *string `json:"key,omitempty"`
	Color     *string `json:"color,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	IsOpen    *bool   `json:"is_open,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`

	BoardUUID uuid.UUID `json:"-"`
}

type UpdateStatus struct {
	Name      *string `json:"name,omitempty"`
	Key       *string `json:"key,omitempty"`
	Color     *string `json:"color,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	IsOpen    *bool   `json:"is_open,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}
