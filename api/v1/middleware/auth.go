package middleware

import (
	"errors"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/internal/app/ports"

	"github.com/labstack/echo/v4"
)

func KeycloakAuth(validator ports.TokenValidator) echo.MiddlewareFunc {
	if validator == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if shouldBypassAuth(c) {
				return next(c)
			}

			token, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if err != nil {
				status := http.StatusUnauthorized
				if isBadRequestError(err) {
					status = http.StatusBadRequest
				}
				return c.JSON(status, dto.NewError("invalid_authorization_header", err.Error()))
			}

			if err := validator.ValidateToken(c.Request().Context(), token); err != nil {
				return c.JSON(http.StatusUnauthorized, dto.NewError("token_invalid", "token validation failed"))
			}

			c.Set("auth_token", token)
			return next(c)
		}
	}
}

func shouldBypassAuth(c echo.Context) bool {
	if c.Request().Method == http.MethodOptions {
		return true
	}

	path := normalizePath(c)
	for _, prefix := range []string{"/auth", "/swagger", "/healthz", "/metrics"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func normalizePath(c echo.Context) string {
	path := c.Path()
	if path == "" {
		path = c.Request().URL.Path
	}
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	lower := strings.ToLower(path)
	if strings.HasPrefix(lower, "/v1") {
		lower = strings.TrimPrefix(lower, "/v1")
		if !strings.HasPrefix(lower, "/") {
			lower = "/" + lower
		}
	}
	return lower
}

func extractBearerToken(header string) (string, error) {
	if strings.TrimSpace(header) == "" {
		return "", errMissingAuthorization
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", errInvalidScheme
	}

	token := strings.TrimSpace(header[len(bearerPrefix):])
	if token == "" {
		return "", errEmptyToken
	}

	return token, nil
}

var (
	errMissingAuthorization = &authHeaderError{message: "authorization header is required", badRequest: false}
	errInvalidScheme        = &authHeaderError{message: "authorization header must use Bearer scheme", badRequest: true}
	errEmptyToken           = &authHeaderError{message: "authorization header contains empty token", badRequest: true}
)

type authHeaderError struct {
	message    string
	badRequest bool
}

func (e *authHeaderError) Error() string {
	return e.message
}

func isBadRequestError(err error) bool {
	if err == nil {
		return false
	}
	var target *authHeaderError
	if errors.As(err, &target) {
		return target.badRequest
	}
	return false
}
