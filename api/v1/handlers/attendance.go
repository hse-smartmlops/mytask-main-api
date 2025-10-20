package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	_ request.CreateAttendance
	_ request.UpdateAttendance
)

type AttendanceHandler struct {
	service ports.AttendanceService
}

func NewAttendanceHandler(service ports.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{service: service}
}

// @Summary Register Attendance Routes
// @Description Register routes for attendance management
// @Tags Attendance
func RegisterAttendanceRoutes(group *echo.Group, service ports.AttendanceService) {
	handler := NewAttendanceHandler(service)

	group.GET("/attendance", handler.ListAttendances)
	group.GET("/attendance/all/:page/:page_size", handler.ListAttendances)
	group.GET("/attendance/all", handler.ListAttendances)
	group.GET("/attendance/:id", handler.GetAttendance)
	group.GET("/users/:user_id/attendance", handler.ListAttendancesByUser)
	group.GET("/attendance/user/:user_id", handler.ListAttendancesByUser)
	group.POST("/attendance", handler.CreateAttendance)
	group.PATCH("/attendance/:id", handler.UpdateAttendance)
	group.DELETE("/attendance/:id", handler.DeleteAttendance)
}

// @Summary List Attendances
// @Description Retrieve a paginated list of attendances
// @Tags Attendance
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Number of items per page"
// @Success 200 {object} dto.AttendancesListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /attendance [get]
func (h *AttendanceHandler) ListAttendances(c echo.Context) error {
	page, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: page, PageSize: size}

	attendancesPage, err := h.service.ListAttendances(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("attendances_fetch_failed", "failed to list attendances"))
	}

	pagination, payload := presenter.MapAttendancesPage(attendancesPage)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Attendance
// @Description Retrieve a specific attendance by its ID
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "Attendance ID"
// @Success 200 {object} dto.AttendanceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /attendance/{id} [get]
func (h *AttendanceHandler) GetAttendance(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid attendance identifier"))
	}

	attendance, err := h.service.GetAttendance(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("attendance_not_found", "attendance not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("attendances_fetch_failed", "failed to get attendance"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToAttendanceDTO(attendance))
}

// @Summary List Attendances by User
// @Description Retrieve a list of attendances for a specific user
// @Tags Attendance
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} dto.AttendancesByUserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{user_id}/attendance [get]
func (h *AttendanceHandler) ListAttendancesByUser(c echo.Context) error {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		userIDStr = c.Param("id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	items, err := h.service.ListAttendancesByUser(c.Request().Context(), userID)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("attendances_fetch_failed", "failed to list attendances for user"))
	}

	payload := presenter.MapAttendancesByUser(userID.String(), items)
	return respondSuccess(c, http.StatusOK, payload)
}

// @Summary Create Attendance
// @Description Create a new attendance record
// @Tags Attendance
// @Accept json
// @Produce json
// @Param attendance body request.CreateAttendance true "Attendance data"
// @Success 201 {object} dto.AttendanceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /attendance [post]
func (h *AttendanceHandler) CreateAttendance(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateAttendancePayload) // don't show this error
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.CreateAttendanceInput{
		UserID:        req.UserUUID,
		Date:          req.ParsedDate,
		WorkdayHours:  req.WorkdayHoursValue,
		PlannedStart:  req.ParsedPlannedStart,
		ActualStart:   req.ParsedActualStart,
		EndWork:       req.ParsedEndWork,
		Commits:       req.CommitsValue,
		MergeRequests: req.MergeRequestsValue,
		CodeReviews:   req.CodeReviewsValue,
		Status:        req.Status,
	}

	attendance, err := h.service.CreateAttendance(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("attendance_create_failed", "failed to create attendance"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToAttendanceDTO(attendance))
}

// @Summary Update Attendance
// @Description Update an existing attendance record
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "Attendance ID"
// @Param attendance body request.UpdateAttendance true "Updated attendance data"
// @Success 200 {object} dto.AttendanceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /attendance/{id} [patch]
func (h *AttendanceHandler) UpdateAttendance(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid attendance identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateUpdateAttendancePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateAttendanceInput{
		Date:          req.ParsedDate,
		WorkdayHours:  req.WorkdayHoursValue,
		PlannedStart:  req.ParsedPlannedStart,
		ActualStart:   req.ParsedActualStart,
		EndWork:       req.ParsedEndWork,
		Commits:       req.CommitsValue,
		MergeRequests: req.MergeRequestsValue,
		CodeReviews:   req.CodeReviewsValue,
		Status:        req.Status,
	}

	attendance, err := h.service.UpdateAttendance(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("attendance_not_found", "attendance not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("attendance_update_failed", "failed to update attendance"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToAttendanceDTO(attendance))
}

// @Summary Delete Attendance
// @Description Delete an attendance record by its ID
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "Attendance ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /attendance/{id} [delete]
func (h *AttendanceHandler) DeleteAttendance(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid attendance identifier"))
	}

	if err := h.service.DeleteAttendance(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("attendance_not_found", "attendance not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("attendance_delete_failed", "failed to delete attendance"))
	}

	return c.NoContent(http.StatusNoContent)
}
