package response

type ImprovedTaskReport struct {
	TaskID          string            `json:"task_id"`
	OriginalText    string            `json:"original_text"`
	ImprovedText    string            `json:"improved_text"`
	ContentType     string            `json:"content_type,omitempty"`
	TaskTitle       *string           `json:"task_title,omitempty"`
	TaskDescription *string           `json:"task_description,omitempty"`
	Meta            map[string]string `json:"meta,omitempty"`
}

type TaskBoardProject struct {
	TaskID    string  `json:"task_id"`
	BoardID   *string `json:"board_id,omitempty"`
	ProjectID *string `json:"project_id,omitempty"`
}
