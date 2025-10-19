package handlers

import (
	"strings"

	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/middleware"

	"github.com/google/uuid"
)

func validateCreateBoardPayload(req *request.CreateBoard) error {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	if req.ProjectID == "" {
		return middleware.ValidationError{Message: "project_id is required"}
	}
	if _, err := uuid.Parse(req.ProjectID); err != nil {
		return middleware.ValidationError{Message: "project_id must be a valid UUID"}
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return middleware.ValidationError{Message: "name is required"}
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

func validateUpdateBoardPayload(req *request.UpdateBoard) error {
	var hasField bool

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return middleware.ValidationError{Message: "name cannot be empty"}
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
		return middleware.ValidationError{Message: "no fields to update"}
	}

	return nil
}

func validateLegacyCreateBoardPayload(req *request.CreateBoard) error {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	if req.ProjectID == "" {
		return middleware.ValidationError{Message: "project_id is required"}
	}
	if _, err := uuid.Parse(req.ProjectID); err != nil {
		return middleware.ValidationError{Message: "project_id must be a valid UUID"}
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return middleware.ValidationError{Message: "name is required"}
	}

	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}

	// for i, raw := range req.UserIDs {
	// 	trimmed := strings.TrimSpace(raw)
	// 	if trimmed == "" {
	// 		return middleware.ValidationError{Message: "user_ids must contain valid UUIDs"}
	// 	}
	// 	if _, err := uuid.Parse(trimmed); err != nil {
	// 		return middleware.ValidationError{Message: "user_ids must contain valid UUIDs"}
	// 	}
	// 	req.UserIDs[i] = trimmed
	// }

	return nil
}
