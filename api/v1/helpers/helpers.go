package helpers

import (
	"fmt"
	"strings"
	"time"
)

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
