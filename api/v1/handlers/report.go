package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type DailyReportHandler struct {
	service ports.DailyReportService
}

func NewDailyReportHandler(service ports.DailyReportService) *DailyReportHandler {
	return &DailyReportHandler{service: service}
}

// @Summary Register Daily Report Routes
// @Description Register routes for daily report management
// @Tags Reports
func RegisterDailyReportRoutes(group *echo.Group, service ports.DailyReportService) {
	handler := NewDailyReportHandler(service)

	group.POST("/report", handler.CreateReport)
	group.GET("/report/:id", handler.GetReport)
	group.PATCH("/report/:id", handler.UpdateReport)
	group.DELETE("/report/:id", handler.DeleteReport)
	group.GET("/report/all", handler.ListReports)
	group.GET("/report/user/:id", handler.ListReportsByUser)
	group.GET("/report/project/:id", handler.ListReportsByProject)
	group.GET("/report/task/:id", handler.ListReportsByTask)
	group.PATCH("/report/completed-work/:id", handler.UpdateCompletedWork)
	group.PATCH("/report/help-request/:id", handler.UpdateHelpRequest)
	group.DELETE("/report/help-request/:id", handler.DeleteHelpRequest)
	group.GET("/report/help-requests-by-user-id/:id", handler.ListHelpRequestsByHelper)
	group.PATCH("/report/tomorrow-plans/:id", handler.UpdateTomorrowPlan)
	group.POST("/report/export/xlsx", handler.ExportReportsXLSX)
}

// @Summary List Reports
// @Description Retrieve a paginated list of daily reports
// @Tags Reports
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param user_id query string false "Filter by user ID"
// @Param project_id query string false "Filter by project ID"
// @Param task_id query string false "Filter by task ID"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (YYYY-MM-DD)"
// @Success 200 {object} dto.ReportsListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/all [get]
func (h *DailyReportHandler) ListReports(c echo.Context) error {
	ctx := c.Request().Context()
	userParam := strings.TrimSpace(c.QueryParam("user_id"))
	projectParam := strings.TrimSpace(c.QueryParam("project_id"))
	taskParam := strings.TrimSpace(c.QueryParam("task_id"))
	startParam := strings.TrimSpace(c.QueryParam("start_date"))
	endParam := strings.TrimSpace(c.QueryParam("end_date"))

	if userParam != "" {
		userID, err := parseUUID(userParam)
		if err != nil {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_query", "user_id must be a valid UUID"))
		}
		pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
		params := ports.PaginationParams{Page: pageNum, PageSize: size}

		result, err := h.service.ListReportsByUser(ctx, userID, params)
		if err != nil {
			return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports for user"))
		}

		pagination, payload := presenter.MapReportsPage(result)
		return respondPaginated(c, http.StatusOK, payload, pagination)
	}

	if projectParam != "" {
		projectID, err := parseUUID(projectParam)
		if err != nil {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_query", "project_id must be a valid UUID"))
		}
		reports, err := h.service.ListReportsByProject(ctx, projectID)
		if err != nil {
			return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports by project"))
		}
		return respondSuccess(c, http.StatusOK, response.ReportsPage{Reports: mapReportModels(reports)})
	}

	if taskParam != "" {
		taskID, err := parseUUID(taskParam)
		if err != nil {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_query", "task_id must be a valid UUID"))
		}
		reports, err := h.service.ListReportsByTask(ctx, taskID)
		if err != nil {
			return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports by task"))
		}
		return respondSuccess(c, http.StatusOK, response.ReportsPage{Reports: mapReportModels(reports)})
	}

	if startParam != "" || endParam != "" {
		if startParam == "" || endParam == "" {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_query", "start_date and end_date must both be provided"))
		}
		startDate, err := parseDateValue(startParam)
		if err != nil {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_query", "start_date must be a valid date"))
		}
		endDate, err := parseDateValue(endParam)
		if err != nil {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_query", "end_date must be a valid date"))
		}
		reports, err := h.service.ListReportsByDateRange(ctx, startDate, endDate)
		if err != nil {
			return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports by date range"))
		}
		return respondSuccess(c, http.StatusOK, response.ReportsPage{Reports: mapReportModels(reports)})
	}

	pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: pageNum, PageSize: size}

	result, err := h.service.ListReports(ctx, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports"))
	}

	pagination, payload := presenter.MapReportsPage(result)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary List Reports By User
// @Description Retrieve a paginated list of daily reports for a specific user
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ReportsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/user/{id} [get]
func (h *DailyReportHandler) ListReportsByUser(c echo.Context) error {
	userID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	page, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: page, PageSize: size}
	reports, err := h.service.ListReportsByUser(c.Request().Context(), userID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports for user"))
	}

	pagination, payload := presenter.MapReportsPage(reports)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary List Reports By Project
// @Description Retrieve a list of daily reports for a specific project
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ReportsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/project/{id} [get]
func (h *DailyReportHandler) ListReportsByProject(c echo.Context) error {
	projectID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	reports, err := h.service.ListReportsByProject(c.Request().Context(), projectID)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports by project"))
	}

	payload := make([]response.Report, len(reports))
	for i := range reports {
		payload[i] = presenter.ToReportDTO(&reports[i])
	}

	return respondSuccess(c, http.StatusOK, response.ReportsPage{Reports: payload})
}

// @Summary List Reports By Task
// @Description Retrieve a list of daily reports for a specific task
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} dto.ReportsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/task/{id} [get]
func (h *DailyReportHandler) ListReportsByTask(c echo.Context) error {
	taskID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	reports, err := h.service.ListReportsByTask(c.Request().Context(), taskID)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports by task"))
	}

	payload := make([]response.Report, len(reports))
	for i := range reports {
		payload[i] = presenter.ToReportDTO(&reports[i])
	}

	return respondSuccess(c, http.StatusOK, response.ReportsPage{Reports: payload})
}

// @Summary Create Report
// @Description Create a new daily report
// @Tags Reports
// @Accept json
// @Produce json
// @Param report body request.ReportCreate true "Report creation payload"
// @Success 201 {object} dto.ReportResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report [post]
func (h *DailyReportHandler) CreateReport(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateReportCreatePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.CreateDailyReportInput{
		UserID:        req.UserUUID,
		ReportDate:    req.ReportDateValue,
		CompletedWork: toCompletedWorkInputs(req.CompletedWork),
		HelpRequests:  toHelpRequestInputs(req.HelpRequests),
		TomorrowPlans: toTomorrowPlanInputs(req.TomorrowPlans),
		Problems:      req.ProblemUUIDs,
	}

	report, err := h.service.CreateReport(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("report_create_failed", "failed to create report"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToReportDTO(report))
}

// @Summary Get Report
// @Description Retrieve a daily report by its ID
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} dto.ReportResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/{id} [get]
func (h *DailyReportHandler) GetReport(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid report identifier"))
	}

	report, err := h.service.GetReport(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("report_not_found", "report not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to get report"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToReportDTO(report))
}

// @Summary Update Report
// @Description Update an existing daily report
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param report body request.ReportUpdate true "Report update payload"
// @Success 200 {object} dto.ReportResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/{id} [put]
func (h *DailyReportHandler) UpdateReport(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid report identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateReportUpdatePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateDailyReportInput{
		Checked:       req.CheckedValue,
		ReportDate:    req.ReportDateValue,
		CompletedWork: toCompletedWorkInputs(req.CompletedWork),
		HelpRequests:  toHelpRequestInputs(req.HelpRequests),
		TomorrowPlans: toTomorrowPlanInputs(req.TomorrowPlans),
		Problems:      req.ProblemUUIDs,
	}

	report, err := h.service.UpdateReport(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("report_not_found", "report not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("report_update_failed", "failed to update report"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToReportDTO(report))
}

// @Summary Delete Report
// @Description Delete a daily report by its ID
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} dto.ReportUniversalResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/{id} [delete]
func (h *DailyReportHandler) DeleteReport(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid report identifier"))
	}

	if err := h.service.DeleteReport(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("report_not_found", "report not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("report_delete_failed", "failed to delete report"))
	}

	return respondSuccess(c, http.StatusOK, response.ReportUniversal{ID: id.String(), Message: "report deleted"})
}

// @Summary Update Completed Work
// @Description Update an existing completed work item
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Completed Work ID"
// @Param completed_work body request.CompletedWorkUpdate true "Completed Work update payload"
// @Success 200 {object} dto.CompletedWorkItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/completed-work/{id} [patch]
func (h *DailyReportHandler) UpdateCompletedWork(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid completed work identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateCompletedWorkUpdatePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateCompletedWorkInput{
		Description: req.Description,
		TaskID:      req.TaskUUID,
	}

	item, err := h.service.UpdateCompletedWork(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("completed_work_not_found", "completed work not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("completed_work_update_failed", "failed to update completed work"))
	}

	return respondSuccess(c, http.StatusOK, response.CompletedWorkItem{
		ID:          item.ID.String(),
		Description: item.Description,
		TaskID:      uuidPtrToString(item.TaskID),
	})
}

// @Summary Update Help Request
// @Description Update an existing help request item
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Help Request ID"
// @Param help_request body request.HelpRequestUpdate true "Help Request update payload"
// @Success 200 {object} dto.HelpRequestItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/help-request/{id} [patch]
func (h *DailyReportHandler) UpdateHelpRequest(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid help request identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateHelpRequestUpdatePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateHelpRequestInput{
		Description: req.Description,
		HelperID:    req.HelperUUID,
		Status:      req.Status,
	}

	item, err := h.service.UpdateHelpRequest(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("help_request_not_found", "help request not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("help_request_update_failed", "failed to update help request"))
	}

	return respondSuccess(c, http.StatusOK, response.HelpRequestItem{
		ID:          item.ID.String(),
		Description: item.Description,
		HelperID:    uuidPtrToString(item.HelperID),
		Status:      item.Status,
	})
}

// @Summary Delete Help Request
// @Description Delete a help request by its ID
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Help Request ID"
// @Success 200 {object} dto.ReportUniversalResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/help-request/{id} [delete]
func (h *DailyReportHandler) DeleteHelpRequest(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid help request identifier"))
	}

	if err := h.service.DeleteHelpRequest(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("help_request_not_found", "help request not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("help_request_delete_failed", "failed to delete help request"))
	}

	return respondSuccess(c, http.StatusOK, response.ReportUniversal{ID: id.String(), Message: "help request deleted"})
}

// @Summary List Help Requests By Helper
// @Description Retrieve a list of help requests assigned to a specific helper
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Helper ID"
// @Success 200 {object} dto.HelpRequestsForUserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/help-requests-by-user-id/{id} [get]
func (h *DailyReportHandler) ListHelpRequestsByHelper(c echo.Context) error {
	helperID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid helper identifier"))
	}

	items, err := h.service.ListHelpRequestsByHelper(c.Request().Context(), helperID)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("help_requests_fetch_failed", "failed to list help requests"))
	}

	return respondSuccess(c, http.StatusOK, presenter.MapHelpRequestsForUser(items))
}

// @Summary Update Tomorrow Plan
// @Description Update an existing tomorrow plan item
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Tomorrow Plan ID"
// @Param tomorrow_plan body request.TomorrowPlanUpdate true "Tomorrow Plan update payload"
// @Success 200 {object} dto.TomorrowPlanItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/tomorrow-plans/{id} [patch]
func (h *DailyReportHandler) UpdateTomorrowPlan(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid tomorrow plan identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateTomorrowPlanUpdatePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateTomorrowPlanInput{
		Description: req.Description,
		TaskID:      req.TaskUUID,
	}

	item, err := h.service.UpdateTomorrowPlan(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("tomorrow_plan_not_found", "tomorrow plan not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("tomorrow_plan_update_failed", "failed to update tomorrow plan"))
	}

	return respondSuccess(c, http.StatusOK, response.TomorrowPlanItem{
		ID:          item.ID.String(),
		Description: item.Description,
		TaskID:      uuidPtrToString(item.TaskID),
	})
}

// @Summary Export Reports to XLSX
// @Description Export daily reports within a specified date range to an XLSX file
// @Tags Reports
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param export_request body request.ReportsByDate true "Reports export request payload"
// @Success 200 {file} file "XLSX file containing the exported reports"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/export/xlsx [post]

func (h *DailyReportHandler) ExportReportsXLSX(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateReportsByDatePayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.ReportsByDateInput{StartDate: req.StartValue, EndDate: req.EndValue}
	fileName := fmt.Sprintf("reports_%s_%s.xlsx", req.StartValue.Format("2006-01-02"), req.EndValue.Format("2006-01-02"))

	data, err := h.service.ExportReportsToXLSX(c.Request().Context(), input, fileName)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("report_export_failed", "failed to export reports"))
	}

	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	return c.Blob(http.StatusOK, contentType, data)
}

func toCompletedWorkInputs(items []request.ReportCompletedWork) []ports.CompletedWorkInput {
	if len(items) == 0 {
		return nil
	}
	result := make([]ports.CompletedWorkInput, len(items))
	for i := range items {
		item := items[i]
		result[i] = ports.CompletedWorkInput{
			ID:          item.IDUUID,
			Description: item.Description,
			TaskID:      item.TaskUUID,
		}
	}
	return result
}

func toHelpRequestInputs(items []request.ReportHelpRequest) []ports.HelpRequestInput {
	if len(items) == 0 {
		return nil
	}
	result := make([]ports.HelpRequestInput, len(items))
	for i := range items {
		item := items[i]
		result[i] = ports.HelpRequestInput{
			ID:          item.IDUUID,
			Description: item.Description,
			HelperID:    item.HelperUUID,
			Status:      item.Status,
		}
	}
	return result
}

func toTomorrowPlanInputs(items []request.ReportTomorrowPlan) []ports.TomorrowPlanInput {
	if len(items) == 0 {
		return nil
	}
	result := make([]ports.TomorrowPlanInput, len(items))
	for i := range items {
		item := items[i]
		result[i] = ports.TomorrowPlanInput{
			ID:          item.IDUUID,
			Description: item.Description,
			TaskID:      item.TaskUUID,
		}
	}
	return result
}

func mapReportModels(items []models.DailyReport) []response.Report {
	if len(items) == 0 {
		return nil
	}
	result := make([]response.Report, len(items))
	for i := range items {
		result[i] = presenter.ToReportDTO(&items[i])
	}
	return result
}

func uuidPtrToString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}
