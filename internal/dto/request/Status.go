package request

type CreateStatusRequest struct {
	Name      string `json:"name" binding:"required"`
	Color     string `json:"color" binding:"required"`
	IsDefault bool   `json:"is_default"`
	IsActive  bool   `json:"is_active"`
	IsOpen    bool   `json:"is_open"`
}

type UpdateStatusRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	IsDefault *bool   `json:"is_default"`
	IsActive  *bool   `json:"is_active"`
	IsOpen    *bool   `json:"is_open"`
}

type AddStatusToTaskRequest struct {
	TaskId   string `json:"task_id"`
	StatusId string `json:"status_id"`
}

type AddStatusToBoardRequest struct {
	BoardId  string `json:"board_id"`
	StatusId string `json:"status_id"`
}

type DeleteStatusFromTaskRequest struct {
	TaskId   string `json:"task_id"`
	StatusId string `json:"status_id"`
}

type DeleteStatusFromBoardRequest struct {
	BoardId  string `json:"board_id"`
	StatusId string `json:"status_id"`
}