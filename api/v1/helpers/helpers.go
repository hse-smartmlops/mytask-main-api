package helpers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func ParsePositiveInt(value string, fallback int) int {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func ParseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(value))
}

func ParseOptionalInt8(value string) (*int8, error) {
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

func ParseUUIDPointer(value *string) (*uuid.UUID, error) {
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

func ParseTimePointer(value *string) (*time.Time, error) {
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

func SanitizeStringPtr(value *string) *string {
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

func ParseDatePointer(value *string) (*time.Time, error) {
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
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("invalid date format")
}

func ParseTimeOfDayPointer(value *string) (*time.Time, error) {
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
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("invalid time format")
}

func ParseDateValue(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("date value is required")
	}
	parsed, err := ParseDatePointer(&trimmed)
	if err != nil {
		return time.Time{}, err
	}
	if parsed == nil {
		return time.Time{}, fmt.Errorf("date value is required")
	}
	return *parsed, nil
}

func ToInt16Ptr(value *int) (*int16, error) {
	if value == nil {
		return nil, nil
	}
	if *value < -32768 || *value > 32767 {
		return nil, fmt.Errorf("value %d is out of range for int16", *value)
	}
	casted := int16(*value)
	return &casted, nil
}

func ResolvePagination(queryPage, querySize string, defaultPage, defaultSize int) (int, int) {
	page := ParsePositiveInt(queryPage, defaultPage)
	size := ParsePositiveInt(querySize, defaultSize)
	return page, size
}

func ExtractBearerToken(header string) (string, error) {
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

func GetOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func ToOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
