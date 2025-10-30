package presenter

import (
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

func uuidString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func uuidPtrToString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}

func toUserSummary(user *models.User) *response.UserSummary {
	if user == nil {
		return nil
	}
	return &response.UserSummary{
		ID:        user.ID.String(),
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		Email:     user.Email,
	}
}
