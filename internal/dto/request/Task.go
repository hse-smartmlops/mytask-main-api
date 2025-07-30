package request

type CreateTask struct {
	ProjectID   uint   `json:"project_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
}
