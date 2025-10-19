package service

import "strings"

func stringPtr(value string) *string {
	return &value
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}
	return &v
}
