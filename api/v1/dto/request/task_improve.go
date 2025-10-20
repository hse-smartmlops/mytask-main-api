package request

type ImproveTaskReport struct {
	UserText    string            `json:"user_text"`
	Meta        map[string]string `json:"meta,omitempty"`
	TimeoutMs   *int64            `json:"timeout_ms,omitempty"`
	ContentType *string           `json:"content_type,omitempty"`
}
