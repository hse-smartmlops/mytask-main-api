package handlers

import (
	"emplacc-api/api/v1/dto"
	v1helpers "emplacc-api/api/v1/helpers"
	"strings"

	"github.com/labstack/echo/v4"
)

var (
	parseUUID          = v1helpers.ParseUUID
	parseOptionalInt8  = v1helpers.ParseOptionalInt8
	parseUUIDPointer   = v1helpers.ParseUUIDPointer
	parseDateValue     = v1helpers.ParseDateValue
	extractBearerToken = v1helpers.ExtractBearerToken
	getOptionalString  = v1helpers.GetOptionalString
	toOptionalString   = v1helpers.ToOptionalString
)

func resolvePagination(queryPage, querySize string, defaultPage, defaultSize int) (int, int) {
	return v1helpers.ResolvePagination(queryPage, querySize, defaultPage, defaultSize)
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
