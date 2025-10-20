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
	middleware "emplacc-api/internal/transport/http/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ValidatorFunc represents custom validation logic for a request DTO.
type ValidatorFunc[T BindableRequest] func(*T) error

type BindableRequest interface {
	request.AuthLogin |
		request.AuthRefresh |
		request.CreateAttendance |
		request.UpdateAttendance |
		request.CreateBoard |
		request.UpdateBoard |
		request.CreateForumMessage |
		request.UpdateForumMessage |
		request.CreateProblem |
		request.UpdateProblem |
		request.CreateProject |
		request.UpdateProject |
		request.CreateRole |
		request.UpdateRole |
		request.CreateStatus |
		request.UpdateStatus |
		request.CreateTask |
		request.UpdateTask |
		request.CreateTeam |
		request.UpdateTeam |
		request.AddTeamMember |
		request.AddTeamProject |
		request.CreateSubscription |
		request.CreateUser |
		request.ReportCreate |
		request.ReportUpdate |
		request.CompletedWorkUpdate |
		request.HelpRequestUpdate |
		request.TomorrowPlanUpdate |
		request.ReportsByDate |
		request.ImproveTaskReport
}

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
func BindAndValidate[T BindableRequest](c echo.Context, validators ...ValidatorFunc[T]) (T, error) {
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
		resp := dto.NewError("invalid_payload", validationErr.Error())
		resp.Meta.TraceID = middleware.TraceIDFromContext(c)
		return c.JSON(http.StatusBadRequest, resp)
	}
	resp := dto.NewError("invalid_payload", err.Error())
	resp.Meta.TraceID = middleware.TraceIDFromContext(c)
	return c.JSON(http.StatusBadRequest, resp)
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
	start, err := v1helpers.ParseDateValue(req.StartDate)
	if err != nil {
		return ValidationError{Message: "start_date must be a valid date"}
	}
	end, err := v1helpers.ParseDateValue(req.EndDate)
	if err != nil {
		return ValidationError{Message: "end_date must be a valid date"}
	}
	if end.Before(start) {
		return ValidationError{Message: "end_date must be on or after start_date"}
	}
	req.StartValue = start
	req.EndValue = end
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

func ValidateCreateTeamPayload(req *request.CreateTeam) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ValidationError{Message: "name is required"}
	}
	req.Description = v1helpers.SanitizeStringPtr(req.Description)
	return nil
}

func ValidateUpdateTeamPayload(req *request.UpdateTeam) error {
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
		req.Description = v1helpers.SanitizeStringPtr(req.Description)
		if req.Description != nil {
			hasField = true
		}
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateAddTeamMemberPayload(req *request.AddTeamMember) error {
	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID == "" {
		return ValidationError{Message: "user_id is required"}
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return ValidationError{Message: "user_id must be a valid UUID"}
	}
	req.UserUUID = userID
	req.Specialization = v1helpers.SanitizeStringPtr(req.Specialization)
	return nil
}

func ValidateAddTeamProjectPayload(req *request.AddTeamProject) error {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	if req.ProjectID == "" {
		return ValidationError{Message: "project_id is required"}
	}
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return ValidationError{Message: "project_id must be a valid UUID"}
	}
	req.ProjectUUID = projectID
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

func ValidateCreateSubscriptionPayload(req *request.CreateSubscription) error {
	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID == "" {
		return ValidationError{Message: "user_id is required"}
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return ValidationError{Message: "user_id must be a valid UUID"}
	}
	req.UserUUID = userID

	parsed, err := sanitizeOptionalUUID(&req.SubscriptionID, "subscription_id")
	if err != nil {
		return err
	}
	req.SubscriptionUUID = parsed

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

func ValidateCreateTaskPayload(req *request.CreateTask) error {
	req.StatusID = strings.TrimSpace(req.StatusID)
	if req.StatusID == "" {
		return ValidationError{Message: "status_id is required"}
	}
	statusID, err := uuid.Parse(req.StatusID)
	if err != nil {
		return ValidationError{Message: "status_id must be a valid UUID"}
	}
	req.StatusUUID = statusID

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ValidationError{Message: "name is required"}
	}
	req.Description = v1helpers.SanitizeStringPtr(req.Description)

	if req.CreatedBy != nil {
		trimmed := strings.TrimSpace(*req.CreatedBy)
		if trimmed == "" {
			req.CreatedBy = nil
		} else {
			creator, err := uuid.Parse(trimmed)
			if err != nil {
				return ValidationError{Message: "created_by must be a valid UUID"}
			}
			req.CreatedByUUID = &creator
		}
	}

	if req.AssignedTo != nil {
		trimmed := strings.TrimSpace(*req.AssignedTo)
		if trimmed == "" {
			req.AssignedTo = nil
		} else {
			assigned, err := uuid.Parse(trimmed)
			if err != nil {
				return ValidationError{Message: "assigned_to must be a valid UUID"}
			}
			req.AssignedToUUID = &assigned
		}
	}

	deadline, err := v1helpers.ParseTimePointer(req.Deadline)
	if err != nil {
		return ValidationError{Message: "deadline must be RFC3339 timestamp"}
	}
	req.DeadlineTime = deadline

	start, err := v1helpers.ParseTimePointer(req.StartDate)
	if err != nil {
		return ValidationError{Message: "start_date must be RFC3339 timestamp"}
	}
	req.StartDateTime = start

	return nil
}

func ValidateUpdateTaskPayload(req *request.UpdateTask) error {
	var hasField bool

	if req.StatusID != nil {
		trimmed := strings.TrimSpace(*req.StatusID)
		if trimmed == "" {
			return ValidationError{Message: "status_id cannot be empty"}
		}
		statusID, err := uuid.Parse(trimmed)
		if err != nil {
			return ValidationError{Message: "status_id must be a valid UUID"}
		}
		req.StatusUUID = &statusID
		hasField = true
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return ValidationError{Message: "name cannot be empty"}
		}
		req.Name = &trimmed
		hasField = true
	}

	if req.Description != nil {
		req.Description = v1helpers.SanitizeStringPtr(req.Description)
		if req.Description != nil {
			hasField = true
		}
	}

	if req.CreatedBy != nil {
		trimmed := strings.TrimSpace(*req.CreatedBy)
		if trimmed == "" {
			req.CreatedBy = nil
		} else {
			creator, err := uuid.Parse(trimmed)
			if err != nil {
				return ValidationError{Message: "created_by must be a valid UUID"}
			}
			req.CreatedByUUID = &creator
		}
		hasField = true
	}

	if req.AssignedTo != nil {
		trimmed := strings.TrimSpace(*req.AssignedTo)
		if trimmed == "" {
			req.AssignedTo = nil
		} else {
			assigned, err := uuid.Parse(trimmed)
			if err != nil {
				return ValidationError{Message: "assigned_to must be a valid UUID"}
			}
			req.AssignedToUUID = &assigned
		}
		hasField = true
	}

	if req.Deadline != nil {
		deadline, err := v1helpers.ParseTimePointer(req.Deadline)
		if err != nil {
			return ValidationError{Message: "deadline must be RFC3339 timestamp"}
		}
		req.DeadlineTime = deadline
		hasField = true
	}
	if req.StartDate != nil {
		start, err := v1helpers.ParseTimePointer(req.StartDate)
		if err != nil {
			return ValidationError{Message: "start_date must be RFC3339 timestamp"}
		}
		req.StartDateTime = start
		hasField = true
	}

	if req.Priority != nil || req.GitlabIssueID != nil || req.Category != nil {
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateReportCreatePayload(req *request.ReportCreate) error {
	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID == "" {
		return ValidationError{Message: "user_id is required"}
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return ValidationError{Message: "user_id must be a valid UUID"}
	}
	req.UserUUID = userID

	if req.ReportDate != nil {
		trimmed := strings.TrimSpace(*req.ReportDate)
		if trimmed == "" {
			req.ReportDate = nil
			req.ReportDateValue = nil
		} else {
			req.ReportDate = &trimmed
			date, err := v1helpers.ParseDatePointer(req.ReportDate)
			if err != nil {
				return ValidationError{Message: "report_date must be a valid date"}
			}
			req.ReportDateValue = date
		}
	}

	if err := sanitizeReportCompletedWorkItems(req.CompletedWork, "complete_work"); err != nil {
		return err
	}
	if err := sanitizeReportHelpRequests(req.HelpRequests, "help"); err != nil {
		return err
	}
	if err := sanitizeReportTomorrowPlans(req.TomorrowPlans, "plan_tomorrow"); err != nil {
		return err
	}

	problems, err := sanitizeUUIDList(req.Problems, "problems must contain valid UUIDs")
	if err != nil {
		return err
	}
	req.ProblemUUIDs = problems

	return nil
}

func ValidateReportUpdatePayload(req *request.ReportUpdate) error {
	if req.Checked != nil {
		if *req.Checked < -128 || *req.Checked > 127 {
			return ValidationError{Message: "checked must be between -128 and 127"}
		}
		value := int8(*req.Checked)
		req.CheckedValue = &value
	}

	if req.ReportDate != nil {
		trimmed := strings.TrimSpace(*req.ReportDate)
		if trimmed == "" {
			req.ReportDate = nil
			req.ReportDateValue = nil
		} else {
			req.ReportDate = &trimmed
			date, err := v1helpers.ParseDatePointer(req.ReportDate)
			if err != nil {
				return ValidationError{Message: "report_date must be a valid date"}
			}
			req.ReportDateValue = date
		}
	}

	if err := sanitizeReportCompletedWorkItems(req.CompletedWork, "complete_work"); err != nil {
		return err
	}
	if err := sanitizeReportHelpRequests(req.HelpRequests, "help"); err != nil {
		return err
	}
	if err := sanitizeReportTomorrowPlans(req.TomorrowPlans, "plan_tomorrow"); err != nil {
		return err
	}

	problems, err := sanitizeUUIDList(req.Problems, "problems must contain valid UUIDs")
	if err != nil {
		return err
	}
	req.ProblemUUIDs = problems

	if req.CheckedValue == nil && req.ReportDateValue == nil && len(req.CompletedWork) == 0 && len(req.HelpRequests) == 0 && len(req.TomorrowPlans) == 0 && len(req.ProblemUUIDs) == 0 {
		return ValidationError{Message: "no fields to update"}
	}

	return nil
}

func ValidateCompletedWorkUpdatePayload(req *request.CompletedWorkUpdate) error {
	var hasField bool

	req.Description = v1helpers.SanitizeStringPtr(req.Description)
	if req.Description != nil {
		hasField = true
	}

	taskID, err := sanitizeOptionalUUID(&req.TaskID, "task_id")
	if err != nil {
		return err
	}
	req.TaskUUID = taskID
	if req.TaskUUID != nil {
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateHelpRequestUpdatePayload(req *request.HelpRequestUpdate) error {
	var hasField bool

	req.Description = v1helpers.SanitizeStringPtr(req.Description)
	if req.Description != nil {
		hasField = true
	}

	helperID, err := sanitizeOptionalUUID(&req.HelperID, "helper_id")
	if err != nil {
		return err
	}
	req.HelperUUID = helperID
	if req.HelperUUID != nil {
		hasField = true
	}

	req.Status = v1helpers.SanitizeStringPtr(req.Status)
	if req.Status != nil {
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateTomorrowPlanUpdatePayload(req *request.TomorrowPlanUpdate) error {
	var hasField bool

	req.Description = v1helpers.SanitizeStringPtr(req.Description)
	if req.Description != nil {
		hasField = true
	}

	taskID, err := sanitizeOptionalUUID(&req.TaskID, "task_id")
	if err != nil {
		return err
	}
	req.TaskUUID = taskID
	if req.TaskUUID != nil {
		hasField = true
	}

	if !hasField {
		return ValidationError{Message: "no fields to update"}
	}
	return nil
}

func ValidateImproveTaskReportPayload(req *request.ImproveTaskReport) error {
	req.UserText = strings.TrimSpace(req.UserText)
	if req.UserText == "" {
		return ValidationError{Message: "user_text is required"}
	}
	if len(req.UserText) > 2000 {
		return ValidationError{Message: "user_text cannot exceed 2000 characters"}
	}

	if req.Meta != nil {
		clean := make(map[string]string, len(req.Meta))
		for k, v := range req.Meta {
			key := strings.TrimSpace(k)
			value := strings.TrimSpace(v)
			if key != "" {
				clean[key] = value
			}
		}
		req.Meta = clean
	}

	if req.ContentType != nil {
		trimmed := strings.TrimSpace(*req.ContentType)
		if trimmed == "" {
			req.ContentType = nil
		} else {
			req.ContentType = &trimmed
		}
	}

	return nil
}

func sanitizeReportCompletedWorkItems(items []request.ReportCompletedWork, field string) error {
	for i := range items {
		entry := &items[i]

		id, err := sanitizeOptionalUUID(&entry.ID, field+".id")
		if err != nil {
			return err
		}
		entry.IDUUID = id

		taskID, err := sanitizeOptionalUUID(&entry.TaskID, field+".task_id")
		if err != nil {
			return err
		}
		entry.TaskUUID = taskID

		entry.Description = v1helpers.SanitizeStringPtr(entry.Description)
	}
	return nil
}

func sanitizeReportHelpRequests(items []request.ReportHelpRequest, field string) error {
	for i := range items {
		entry := &items[i]

		id, err := sanitizeOptionalUUID(&entry.ID, field+".id")
		if err != nil {
			return err
		}
		entry.IDUUID = id

		helperID, err := sanitizeOptionalUUID(&entry.HelperID, field+".helper_id")
		if err != nil {
			return err
		}
		entry.HelperUUID = helperID

		entry.Description = v1helpers.SanitizeStringPtr(entry.Description)
		entry.Status = v1helpers.SanitizeStringPtr(entry.Status)
	}
	return nil
}

func sanitizeReportTomorrowPlans(items []request.ReportTomorrowPlan, field string) error {
	for i := range items {
		entry := &items[i]

		id, err := sanitizeOptionalUUID(&entry.ID, field+".id")
		if err != nil {
			return err
		}
		entry.IDUUID = id

		taskID, err := sanitizeOptionalUUID(&entry.TaskID, field+".task_id")
		if err != nil {
			return err
		}
		entry.TaskUUID = taskID

		entry.Description = v1helpers.SanitizeStringPtr(entry.Description)
	}
	return nil
}

func sanitizeOptionalUUID(value **string, field string) (*uuid.UUID, error) {
	if value == nil || *value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(**value)
	if trimmed == "" {
		*value = nil
		return nil, nil
	}

	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, ValidationError{Message: fmt.Sprintf("%s must be a valid UUID", field)}
	}

	copyValue := trimmed
	*value = &copyValue
	return &parsed, nil
}

func sanitizeUUIDList(values []string, errMsg string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]uuid.UUID, len(values))
	for i := range values {
		trimmed := strings.TrimSpace(values[i])
		if trimmed == "" {
			return nil, ValidationError{Message: errMsg}
		}
		id, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, ValidationError{Message: errMsg}
		}
		result[i] = id
	}
	return result, nil
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
