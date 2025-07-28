package request

// DTO для создания команды
// POST /team

type TeamCreateRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// DTO для обновления команды
// PATCH /team/:id

type TeamUpdateRequest struct {
	ID          string  `json:"id" binding:"required"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// DTO для удаления команды
// DELETE /team/:id

type DeleteTeamRequest struct {
	ID string `json:"id" binding:"required"`
}
