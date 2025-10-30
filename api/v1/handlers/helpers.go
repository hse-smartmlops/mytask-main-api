package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	v1helpers "emplacc-api/api/v1/helpers"
	"emplacc-api/internal/app/ports"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	parseDateValue = v1helpers.ParseDateValue

	errMissingPagination = errors.New("page and page_size query parameters are required")
)

func parseUUIDPointer(value *string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseOptionalInt8(value string) (*int8, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 8)
	if err != nil {
		return nil, err
	}
	casted := int8(parsed)
	return &casted, nil
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(value))
}

func paginationParams(c echo.Context) (ports.PaginationParams, error) {
	pageStr := strings.TrimSpace(c.QueryParam("page"))
	sizeStr := strings.TrimSpace(c.QueryParam("page_size"))

	if pageStr == "" || sizeStr == "" {
		return ports.PaginationParams{}, errMissingPagination
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return ports.PaginationParams{}, fmt.Errorf("page must be a integer")
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return ports.PaginationParams{}, fmt.Errorf("page_size must be a integer")
	}

	return ports.PaginationParams{Page: page, PageSize: size}, nil
}

func resolveAuthToken(c echo.Context) string {
	if raw := c.Get("auth_token"); raw != nil {
		if token, ok := raw.(string); ok {
			trimmed := strings.TrimSpace(token)
			if trimmed != "" {
				return trimmed
			}
		}
	}

	token, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
	if err != nil {
		return ""
	}
	return token
}

func respondSuccess[T any](c echo.Context, status int, data T) error {
	resp := dto.NewSuccessResponse(data)
	resp.Meta.TraceID = requestTraceID(c)
	return c.JSON(status, resp)
}

func respondPaginated[T any](c echo.Context, status int, data T, pagination dto.Pagination) error {
	resp := dto.NewPaginatedResponse(data, pagination)
	resp.Meta.TraceID = requestTraceID(c)
	return c.JSON(status, resp)
}

func respondError(c echo.Context, status int, err dto.ErrorResponse) error {
	err.Meta.TraceID = requestTraceID(c)
	return c.JSON(status, err)
}

func requestTraceID(c echo.Context) string {
	if traceID := c.Response().Header().Get(echo.HeaderXRequestID); traceID != "" {
		return traceID
	}
	return c.Request().Header.Get(echo.HeaderXRequestID)
}

func toCompletedWorkInputs(items []request.ReportCompletedWork) []ports.CompletedWorkInput {
	if len(items) == 0 {
		return nil
	}
	result := make([]ports.CompletedWorkInput, len(items))
	for i := range items {
		item := items[i]
		result[i] = ports.CompletedWorkInput{
			ID:          item.IDUUID,
			Description: item.Description,
			TaskID:      item.TaskUUID,
		}
	}
	return result
}

func toHelpRequestInputs(items []request.ReportHelpRequest) []ports.HelpRequestInput {
	if len(items) == 0 {
		return nil
	}
	result := make([]ports.HelpRequestInput, len(items))
	for i := range items {
		item := items[i]
		result[i] = ports.HelpRequestInput{
			ID:          item.IDUUID,
			Description: item.Description,
			HelperID:    item.HelperUUID,
			Status:      item.Status,
		}
	}
	return result
}

func toTomorrowPlanInputs(items []request.ReportTomorrowPlan) []ports.TomorrowPlanInput {
	if len(items) == 0 {
		return nil
	}
	result := make([]ports.TomorrowPlanInput, len(items))
	for i := range items {
		item := items[i]
		result[i] = ports.TomorrowPlanInput{
			ID:          item.IDUUID,
			Description: item.Description,
			TaskID:      item.TaskUUID,
		}
	}
	return result
}

func uuidPtrToString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}

func toOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func getOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func extractBearerToken(header string) (string, error) {
	if strings.TrimSpace(header) == "" {
		return "", errors.New("authorization header is required")
	}
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		trimmed := strings.TrimSpace(header[len(prefix):])
		if trimmed == "" {
			return "", errors.New("authorization header contains empty token")
		}
		return trimmed, nil
	}
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return "", errors.New("authorization header contains empty token")
	}
	return trimmed, nil
}
