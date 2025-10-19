package handlers

import (
	"strconv"

	"github.com/google/uuid"
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
