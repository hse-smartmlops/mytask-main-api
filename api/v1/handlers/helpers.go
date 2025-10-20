package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"emplacc-api/api/v1/dto"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func parseOptionalInt8(value string) (*int8, error) {
	if value == "" {
		return nil, nil
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 8)
	if err != nil {
		return nil, err
	}
	casted := int8(parsed)
	return &casted, nil
}

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

func parseTimePointer(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func sanitizeStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	res := trimmed
	return &res
}

func parseDatePointer(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			result := parsed
			return &result, nil
		}
	}
	return nil, fmt.Errorf("invalid date format")
}

func parseTimeOfDayPointer(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	layouts := []string{
		time.RFC3339,
		"15:04:05",
		"15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			result := parsed
			return &result, nil
		}
	}
	return nil, fmt.Errorf("invalid time format")
}

func parseDateValue(value string) (time.Time, error) {
	stripped := strings.TrimSpace(value)
	if stripped == "" {
		return time.Time{}, fmt.Errorf("date value is required")
	}
	parsed, err := parseDatePointer(&stripped)
	if err != nil {
		return time.Time{}, err
	}
	if parsed == nil {
		return time.Time{}, fmt.Errorf("date value is required")
	}
	return *parsed, nil
}

func toInt16Ptr(value *int) (*int16, error) {
	if value == nil {
		return nil, nil
	}
	if *value < -32768 || *value > 32767 {
		return nil, fmt.Errorf("value %d is out of range for int16", *value)
	}
	casted := int16(*value)
	return &casted, nil
}

func resolvePagination(pathPage, pathSize, queryPage, querySize string, defaultPage, defaultSize int) (int, int) {
	page := parsePositiveInt(queryPage, defaultPage)
	size := parsePositiveInt(querySize, defaultSize)

	if pathPage != "" {
		if parsed, err := strconv.Atoi(pathPage); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pathSize != "" {
		if parsed, err := strconv.Atoi(pathSize); err == nil && parsed > 0 {
			size = parsed
		}
	}

	return page, size
}

func extractBearerToken(header string) (string, error) {
	if strings.TrimSpace(header) == "" {
		return "", errors.New("authorization header is required")
	}
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		token := strings.TrimSpace(header[len(prefix):])
		if token == "" {
			return "", errors.New("authorization header contains empty token")
		}
		return token, nil
	}
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return "", errors.New("authorization header contains empty token")
	}
	return trimmed, nil
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
