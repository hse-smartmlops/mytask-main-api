package request

type GetAllRolesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type RoleCreateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type RoleUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type RoleUniversalRequest struct {
	ID          string  `json:"id"`
	Description *string `json:"description"`
}
