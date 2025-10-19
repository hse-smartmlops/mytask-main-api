package dto

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

type ResponseMeta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
	TraceID    string      `json:"trace_id,omitempty"`
}

type SuccessResponse[T any] struct {
	Data T            `json:"data"`
	Meta ResponseMeta `json:"meta,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail  `json:"error"`
	Meta  ResponseMeta `json:"meta,omitempty"`
}

func NewSuccessResponse[T any](data T) SuccessResponse[T] {
	return SuccessResponse[T]{Data: data}
}

func NewPaginatedResponse[T any](data T, pagination Pagination) SuccessResponse[T] {
	return SuccessResponse[T]{
		Data: data,
		Meta: ResponseMeta{Pagination: &pagination},
	}
}

func NewError(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}
