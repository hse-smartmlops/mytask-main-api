package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"emplacc-api/api/v1/dto"

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
