package middleware

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	v1helpers "emplacc-api/api/v1/helpers"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ValidatorFunc represents custom validation logic for a request DTO.
type ValidatorFunc[T any] func(*T) error

// ValidationError is returned when request validation fails.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	if e.Message == "" {
		return "validation failed"
	}
	return e.Message
}

// BindAndValidate binds the request body to T and runs provided validator functions.
func BindAndValidate[T any](c echo.Context, validators ...ValidatorFunc[T]) (T, error) {
	var req T
	if err := c.Bind(&req); err != nil {
		return req, err
	}

	for _, validate := range validators {
		if validate == nil {
			continue
		}
		if err := validate(&req); err != nil {
			return req, wrapValidationError(err)
		}
	}

	return req, nil
}

// RespondValidationError normalises validation errors into consistent JSON response.
func RespondValidationError(c echo.Context, err error) error {
	var validationErr ValidationError
	if errors.As(err, &validationErr) {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", validationErr.Error()))
	}
	return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
}

func wrapValidationError(err error) error {
	if err == nil {
		return nil
	}
	var validationErr ValidationError
	if errors.As(err, &validationErr) {
		return validationErr
	}
	return ValidationError{Message: err.Error()}
}

// RequireString ensures provided string pointer is not nil/empty.
func RequireString(fieldName string, value *string) error {
	if value == nil || len(trim(*value)) == 0 {
		return ValidationError{Message: fmt.Sprintf("%s is required", fieldName)}
	}
	return nil
}

// trim is defined to avoid importing strings in every handler.
func trim(value string) string {
	start, end := 0, len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t' || value[start] == '\n' || value[start] == '\r') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t' || value[end-1] == '\n' || value[end-1] == '\r') {
		end--
	}
	if start == 0 && end == len(value) {
		return value
	}
	return value[start:end]
}

func ValidateAuthLoginPayload(req *request.AuthLogin) error {
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		return ValidationError{Message: "email and password are required"}
	}
	return nil
}

func ValidateAuthRefreshPayload(req *request.AuthRefresh) error {
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		return ValidationError{Message: "refresh_token is required"}
	}
	return nil
}

func ValidateReportsByDatePayload(req *request.ReportsByDate) error {
	req.StartDate = strings.TrimSpace(req.StartDate)
	req.EndDate = strings.TrimSpace(req.EndDate)
	if req.StartDate == "" || req.EndDate == "" {
		return ValidationError{Message: "start_date and end_date are required"}
	}
	return nil
}

func ValidateCreateProjectPayload(req *request.CreateProject) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ValidationError{Message: "name is required"}
	}

	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}
	if req.GitlabURL != nil {
		trimmed := strings.TrimSpace(*req.GitlabURL)
		if trimmed == "" {
			req.GitlabURL = nil
		} else {
			req.GitlabURL = &trimmed
		}
	}
	if req.Status != nil {
		trimmed := strings.TrimSpace(*req.Status)
		if trimmed == "" {
			req.Status = nil
		} else {
			req.Status = &trimmed
		}
	}
	if req.CreatedBy != nil {
		trimmed := strings.TrimSpace(*req.CreatedBy)
		if trimmed == "" {
			req.CreatedBy = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ValidationError{Message: "created_by must be a valid UUID"}
			}
			req.CreatedBy = &trimmed
		}
	}

	return nil
}

func ValidateUpdateProjectPayload(req *request.UpdateProject) error {
	var hasField bool

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return ValidationError{Message: "name cannot be empty"}
		}
		req.Name = &trimmed
		hasField = true
	}
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
		hasField = true
	}
	if req.GitlabURL != nil {
		trimmed := strings.TrimSpace(*req.GitlabURL)
		if trimmed == "" {
			req.GitlabURL = nil
		} else {
			req.GitlabURL = &trimmed
		}
		hasField = true
	}
	if req.Status != nil {
		trimmed := strings.TrimSpace(*req.Status)
		if trimmed == "" {
			req.Status = nil
		} else {
			req.Status = &trimmed
		}
		hasField = true
	}
	if req.GitlabProjectID != nil {
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}

	return nil
}

func ValidateCreateRolePayload(req *request.CreateRole) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ValidationError{Message: "name is required"}
	}

	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}

	return nil
}

func ValidateUpdateRolePayload(req *request.UpdateRole) error {
	var hasField bool

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return ValidationError{Message: "name cannot be empty"}
		}
		req.Name = &trimmed
		hasField = true
	}
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateCreateProblemPayload(req *request.CreateProblem) error {
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			req.Name = nil
		} else {
			req.Name = &trimmed
		}
	}
	req.Description = sanitizeDescriptionValues(req.Description)
	if req.CreatorID != nil {
		trimmed := strings.TrimSpace(*req.CreatorID)
		if trimmed == "" {
			req.CreatorID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ValidationError{Message: "creator_id must be a valid UUID"}
			}
			req.CreatorID = &trimmed
		}
	}
	return nil
}

func ValidateUpdateProblemPayload(req *request.UpdateProblem) error {
	var hasField bool
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			req.Name = nil
		} else {
			req.Name = &trimmed
		}
		hasField = true
	}
	if req.Description != nil {
		clean := sanitizeDescriptionValues(*req.Description)
		req.Description = &clean
		hasField = true
	}
	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateCreateForumMessagePayload(req *request.CreateForumMessage) error {
	trimmedProblemID := strings.TrimSpace(req.ProblemID)
	if trimmedProblemID == "" {
		return ValidationError{Message: "problem_id is required"}
	}
	problemID, err := uuid.Parse(trimmedProblemID)
	if err != nil {
		return ValidationError{Message: "problem_id must be a valid UUID"}
	}
	req.ProblemUUID = problemID
	req.CleanDesc = sanitizeDescriptionValues(req.Description)
	if len(req.CleanDesc) == 0 {
		return ValidationError{Message: "description is required"}
	}
	if req.CreatorID != nil {
		trimmed := strings.TrimSpace(*req.CreatorID)
		if trimmed == "" {
			req.CreatorID = nil
		} else {
			creator, err := uuid.Parse(trimmed)
			if err != nil {
				return ValidationError{Message: "creator_id must be a valid UUID"}
			}
			req.CreatorUUID = &creator
		}
	}
	return nil
}

func ValidateUpdateForumMessagePayload(req *request.UpdateForumMessage) error {
	var hasField bool
	if req.Description != nil {
		clean := sanitizeDescriptionValues(*req.Description)
		if len(clean) == 0 {
			return ValidationError{Message: "description cannot be empty"}
		}
		req.CleanDesc = &clean
		hasField = true
	}
	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateCreateUserPayload(req *request.CreateUser) error {
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return ValidationError{Message: "email is required"}
	}
	req.TgID = v1helpers.SanitizeStringPtr(req.TgID)
	req.Profession = v1helpers.SanitizeStringPtr(req.Profession)
	req.FirstName = v1helpers.SanitizeStringPtr(req.FirstName)
	req.LastName = v1helpers.SanitizeStringPtr(req.LastName)
	return nil
}

func ValidateCreateBoardPayload(req *request.CreateBoard) error {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	if req.ProjectID == "" {
		return ValidationError{Message: "project_id is required"}
	}
	if _, err := uuid.Parse(req.ProjectID); err != nil {
		return ValidationError{Message: "project_id must be a valid UUID"}
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ValidationError{Message: "name is required"}
	}

	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}

	return nil
}

func ValidateUpdateBoardPayload(req *request.UpdateBoard) error {
	var hasField bool

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return ValidationError{Message: "name cannot be empty"}
		}
		req.Name = &trimmed
		hasField = true
	}

	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}

	return nil
}

func ValidateCreateStatusPayload(req *request.CreateStatus) error {
	req.BoardID = strings.TrimSpace(req.BoardID)
	if req.BoardID == "" {
		return ValidationError{Message: "board_id is required"}
	}
	boardID, err := uuid.Parse(req.BoardID)
	if err != nil {
		return ValidationError{Message: "board_id must be a valid UUID"}
	}
	req.BoardUUID = boardID

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ValidationError{Message: "name is required"}
	}
	req.Key = v1helpers.SanitizeStringPtr(req.Key)
	req.Color = v1helpers.SanitizeStringPtr(req.Color)
	return nil
}

func ValidateUpdateStatusPayload(req *request.UpdateStatus) error {
	var hasField bool

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return ValidationError{Message: "name cannot be empty"}
		}
		req.Name = &trimmed
		hasField = true
	}
	if req.Key != nil {
		req.Key = v1helpers.SanitizeStringPtr(req.Key)
		if req.Key != nil {
			hasField = true
		}
	}
	if req.Color != nil {
		req.Color = v1helpers.SanitizeStringPtr(req.Color)
		if req.Color != nil {
			hasField = true
		}
	}
	if req.IsDefault != nil || req.IsActive != nil || req.IsOpen != nil || req.SortOrder != nil {
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateLegacyCreateBoardPayload(req *request.CreateBoard) error {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	if req.ProjectID == "" {
		return ValidationError{Message: "project_id is required"}
	}
	if _, err := uuid.Parse(req.ProjectID); err != nil {
		return ValidationError{Message: "project_id must be a valid UUID"}
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ValidationError{Message: "name is required"}
	}

	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}

	return nil
}

func sanitizeDescriptionValues(values []string) []string {
	clean := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return clean
}

func ValidateCreateAttendancePayload(req *request.CreateAttendance) error {
	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID == "" {
		return ValidationError{Message: "user_id is required"}
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return ValidationError{Message: "user_id must be a valid UUID"}
	}
	req.UserUUID = userID

	req.Status = v1helpers.SanitizeStringPtr(req.Status)

	date, err := v1helpers.ParseDatePointer(req.Date)
	if err != nil {
		return ValidationError{Message: "date must be a valid ISO date"}
	}
	req.ParsedDate = date

	planned, err := v1helpers.ParseTimeOfDayPointer(req.PlannedStart)
	if err != nil {
		return ValidationError{Message: "planned_start must be a valid time"}
	}
	req.ParsedPlannedStart = planned

	actual, err := v1helpers.ParseTimeOfDayPointer(req.ActualStart)
	if err != nil {
		return ValidationError{Message: "actual_start must be a valid time"}
	}
	req.ParsedActualStart = actual

	end, err := v1helpers.ParseTimeOfDayPointer(req.EndWork)
	if err != nil {
		return ValidationError{Message: "end_work must be a valid time"}
	}
	req.ParsedEndWork = end

	if value, err := v1helpers.ToInt16Ptr(req.WorkdayHours); err != nil {
		return ValidationError{Message: err.Error()}
	} else {
		req.WorkdayHoursValue = value
	}
	if value, err := v1helpers.ToInt16Ptr(req.Commits); err != nil {
		return ValidationError{Message: err.Error()}
	} else {
		req.CommitsValue = value
	}
	if value, err := v1helpers.ToInt16Ptr(req.MergeRequests); err != nil {
		return ValidationError{Message: err.Error()}
	} else {
		req.MergeRequestsValue = value
	}
	if value, err := v1helpers.ToInt16Ptr(req.CodeReviews); err != nil {
		return ValidationError{Message: err.Error()}
	} else {
		req.CodeReviewsValue = value
	}

	return nil
}

func ValidateUpdateAttendancePayload(req *request.UpdateAttendance) error {
	var hasField bool

	if req.Date != nil {
		date, err := v1helpers.ParseDatePointer(req.Date)
		if err != nil {
			return ValidationError{Message: "date must be a valid ISO date"}
		}
		req.ParsedDate = date
		hasField = true
	}

	if req.PlannedStart != nil {
		planned, err := v1helpers.ParseTimeOfDayPointer(req.PlannedStart)
		if err != nil {
			return ValidationError{Message: "planned_start must be a valid time"}
		}
		req.ParsedPlannedStart = planned
		hasField = true
	}

	if req.ActualStart != nil {
		actual, err := v1helpers.ParseTimeOfDayPointer(req.ActualStart)
		if err != nil {
			return ValidationError{Message: "actual_start must be a valid time"}
		}
		req.ParsedActualStart = actual
		hasField = true
	}

	if req.EndWork != nil {
		end, err := v1helpers.ParseTimeOfDayPointer(req.EndWork)
		if err != nil {
			return ValidationError{Message: "end_work must be a valid time"}
		}
		req.ParsedEndWork = end
		hasField = true
	}

	if req.WorkdayHours != nil {
		value, err := v1helpers.ToInt16Ptr(req.WorkdayHours)
		if err != nil {
			return ValidationError{Message: err.Error()}
		}
		req.WorkdayHoursValue = value
		hasField = true
	}

	if req.Commits != nil {
		value, err := v1helpers.ToInt16Ptr(req.Commits)
		if err != nil {
			return ValidationError{Message: err.Error()}
		}
		req.CommitsValue = value
		hasField = true
	}

	if req.MergeRequests != nil {
		value, err := v1helpers.ToInt16Ptr(req.MergeRequests)
		if err != nil {
			return ValidationError{Message: err.Error()}
		}
		req.MergeRequestsValue = value
		hasField = true
	}

	if req.CodeReviews != nil {
		value, err := v1helpers.ToInt16Ptr(req.CodeReviews)
		if err != nil {
			return ValidationError{Message: err.Error()}
		}
		req.CodeReviewsValue = value
		hasField = true
	}

	if req.Status != nil {
		req.Status = v1helpers.SanitizeStringPtr(req.Status)
		if req.Status != nil {
			hasField = true
		}
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func BindAvatarUpload(c echo.Context, fieldName string, maxSize int64) (*request.AvatarUpload, error) {
	if maxSize <= 0 {
		maxSize = 5 << 20
	}

	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, ValidationError{Message: fmt.Sprintf("%s file is required", fieldName)}
	}

	src, err := file.Open()
	if err != nil {
		return nil, ValidationError{Message: "unable to read uploaded file"}
	}
	defer src.Close()

	limited := io.LimitReader(src, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, ValidationError{Message: "unable to process uploaded file"}
	}
	if len(data) == 0 {
		return nil, ValidationError{Message: fmt.Sprintf("%s file is empty", fieldName)}
	}
	if int64(len(data)) > maxSize {
		return nil, ValidationError{Message: fmt.Sprintf("%s file exceeds size limit", fieldName)}
	}

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &request.AvatarUpload{
		Data:        data,
		ContentType: contentType,
	}, nil
}
