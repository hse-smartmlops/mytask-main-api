package response

// Для GET /team/all

type TeamsListResponse struct {
	Teams []TeamResponse `json:"teams"`
}

// Для GET /team/:id, POST /team, PATCH /team/:id

type TeamResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Members     []TeamMemberResponse `json:"members"`
}

// Для участника команды

type TeamMemberResponse struct {
	UserID         string `json:"user_id"`
	Specialization string `json:"specialization"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
}

// Для DELETE /team/:id (если нужен ответ)

type TeamDeleteResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
