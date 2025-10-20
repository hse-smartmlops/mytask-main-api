package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

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
	group.GET("/report/:id/download", handler.DownloadReportFile)
	group.GET("/report/all/:page/:page_size", handler.ListReports)
	group.GET("/report/all", handler.ListReports)
	group.GET("/report/user/:id/:page/:page_size", handler.ListReportsByUser)
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
// @Success 200 {object} dto.ReportsListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/all [get]
func (h *DailyReportHandler) ListReports(c echo.Context) error {
	page, pageSize := parseLegacyPagination(c.Param("page"), c.Param("page_size"), c.QueryParam("page"), c.QueryParam("page_size"))
	params := ports.PaginationParams{Page: page, PageSize: pageSize}

	reports, err := h.service.ListReports(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports"))
	}

	pagination, payload := presenter.MapReportsPage(reports)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
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
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "user id is required"))
	}
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	page, pageSize := parseLegacyPagination(c.Param("page"), c.Param("page_size"), c.QueryParam("page"), c.QueryParam("page_size"))
	params := ports.PaginationParams{Page: page, PageSize: pageSize}
	reports, err := h.service.ListReportsByUser(c.Request().Context(), userID, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports for user"))
	}

	pagination, payload := presenter.MapReportsPage(reports)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	reports, err := h.service.ListReportsByProject(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports by project"))
	}

	payload := make([]response.Report, len(reports))
	for i := range reports {
		payload[i] = presenter.ToReportDTO(&reports[i])
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.ReportsPage{Reports: payload}))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid task identifier"))
	}

	reports, err := h.service.ListReportsByTask(c.Request().Context(), taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to list reports by task"))
	}

	payload := make([]response.Report, len(reports))
	for i := range reports {
		payload[i] = presenter.ToReportDTO(&reports[i])
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.ReportsPage{Reports: payload}))
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
	var req request.ReportCreate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	userID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil || userID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "user_id must be a valid UUID"))
	}

	reportDate, err := parseDatePointer(req.ReportDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "report_date must be a valid date"))
	}

	completed, err := mapCompletedWorkInputs(req.CompletedWork)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
	}
	helpRequests, err := mapHelpRequestInputs(req.HelpRequests)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
	}
	plans, err := mapTomorrowPlanInputs(req.TomorrowPlans)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
	}
	problems, err := parseUUIDList(req.Problems)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "problems must contain valid UUIDs"))
	}

	input := ports.CreateDailyReportInput{
		UserID:        userID,
		ReportDate:    reportDate,
		CompletedWork: completed,
		HelpRequests:  helpRequests,
		TomorrowPlans: plans,
		Problems:      problems,
	}

	report, err := h.service.CreateReport(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("report_create_failed", "failed to create report"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToReportDTO(report)))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid report identifier"))
	}

	report, err := h.service.GetReport(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("report_not_found", "report not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("reports_fetch_failed", "failed to get report"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToReportDTO(report)))
}

// @Summary Download Report File
// @Description Download the file associated with a daily report
// @Tags Reports
// @Accept json
// @Produce octet-stream
// @Param id path string true "Report ID"
// @Success 200 {file} file
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /report/{id}/download [get]
func (h *DailyReportHandler) DownloadReportFile(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid report identifier"))
	}

	file, err := h.service.DownloadReportFile(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("report_not_found", "report not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("report_file_failed", "failed to download report"))
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileName))
	return c.Blob(http.StatusOK, file.ContentType, file.Data)
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid report identifier"))
	}

	var req request.ReportUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	var checked *int8
	if req.Checked != nil {
		value := int8(*req.Checked)
		checked = &value
	}

	reportDate, err := parseDatePointer(req.ReportDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "report_date must be a valid date"))
	}

	completed, err := mapCompletedWorkInputs(req.CompletedWork)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
	}
	helpRequests, err := mapHelpRequestInputs(req.HelpRequests)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
	}
	plans, err := mapTomorrowPlanInputs(req.TomorrowPlans)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
	}
	problems, err := parseUUIDList(req.Problems)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "problems must contain valid UUIDs"))
	}

	input := ports.UpdateDailyReportInput{
		Checked:       checked,
		ReportDate:    reportDate,
		CompletedWork: completed,
		HelpRequests:  helpRequests,
		TomorrowPlans: plans,
		Problems:      problems,
	}

	report, err := h.service.UpdateReport(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("report_not_found", "report not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("report_update_failed", "failed to update report"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToReportDTO(report)))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid report identifier"))
	}

	if err := h.service.DeleteReport(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("report_not_found", "report not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("report_delete_failed", "failed to delete report"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.ReportUniversal{ID: id.String(), Message: "report deleted"}))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid completed work identifier"))
	}

	var req request.CompletedWorkUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	taskID, err := parseUUIDPointer(req.TaskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "task_id must be a valid UUID"))
	}

	input := ports.UpdateCompletedWorkInput{
		Description: sanitizeStringPtr(req.Description),
		TaskID:      taskID,
	}

	item, err := h.service.UpdateCompletedWork(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("completed_work_not_found", "completed work not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("completed_work_update_failed", "failed to update completed work"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.CompletedWorkItem{
		ID:          item.ID.String(),
		Description: item.Description,
		TaskID:      uuidPtrToString(item.TaskID),
	}))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid help request identifier"))
	}

	var req request.HelpRequestUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	helperID, err := parseUUIDPointer(req.HelperID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "helper_id must be a valid UUID"))
	}

	input := ports.UpdateHelpRequestInput{
		Description: sanitizeStringPtr(req.Description),
		HelperID:    helperID,
		Status:      sanitizeStringPtr(req.Status),
	}

	item, err := h.service.UpdateHelpRequest(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("help_request_not_found", "help request not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("help_request_update_failed", "failed to update help request"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.HelpRequestItem{
		ID:          item.ID.String(),
		Description: item.Description,
		HelperID:    uuidPtrToString(item.HelperID),
		Status:      item.Status,
	}))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid help request identifier"))
	}

	if err := h.service.DeleteHelpRequest(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("help_request_not_found", "help request not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("help_request_delete_failed", "failed to delete help request"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.ReportUniversal{ID: id.String(), Message: "help request deleted"}))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid helper identifier"))
	}

	items, err := h.service.ListHelpRequestsByHelper(c.Request().Context(), helperID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("help_requests_fetch_failed", "failed to list help requests"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.MapHelpRequestsForUser(items)))
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
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid tomorrow plan identifier"))
	}

	var req request.TomorrowPlanUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	taskID, err := parseUUIDPointer(req.TaskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "task_id must be a valid UUID"))
	}

	input := ports.UpdateTomorrowPlanInput{
		Description: sanitizeStringPtr(req.Description),
		TaskID:      taskID,
	}

	item, err := h.service.UpdateTomorrowPlan(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("tomorrow_plan_not_found", "tomorrow plan not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("tomorrow_plan_update_failed", "failed to update tomorrow plan"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(response.TomorrowPlanItem{
		ID:          item.ID.String(),
		Description: item.Description,
		TaskID:      uuidPtrToString(item.TaskID),
	}))
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
	var req request.ReportsByDate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	if strings.TrimSpace(req.StartDate) == "" || strings.TrimSpace(req.EndDate) == "" {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "start_date and end_date are required"))
	}

	start, err := parseDateValue(req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "start_date must be a valid date"))
	}
	end, err := parseDateValue(req.EndDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "end_date must be a valid date"))
	}

	input := ports.ReportsByDateInput{
		StartDate: start,
		EndDate:   end,
	}

	data, err := h.service.ExportReportsToXLSX(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("report_export_failed", "failed to export reports"))
	}

	fileName := fmt.Sprintf("reports_%s_%s.xlsx", start.Format("2006-01-02"), end.Format("2006-01-02"))
	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	return c.Blob(http.StatusOK, contentType, data)
}

func mapCompletedWorkInputs(items []request.ReportCompletedWork) ([]ports.CompletedWorkInput, error) {
	if len(items) == 0 {
		return nil, nil
	}
	result := make([]ports.CompletedWorkInput, len(items))
	for i := range items {
		item := items[i]
		var idPtr *uuid.UUID
		if item.ID != nil {
			id, err := parseUUID(*item.ID)
			if err != nil {
				return nil, errors.New("complete_work.id must be a valid UUID")
			}
			idPtr = &id
		}
		taskID, err := parseUUIDPointer(item.TaskID)
		if err != nil {
			return nil, errors.New("complete_work.task_id must be a valid UUID")
		}
		result[i] = ports.CompletedWorkInput{
			ID:          idPtr,
			Description: sanitizeStringPtr(item.Description),
			TaskID:      taskID,
		}
	}
	return result, nil
}

func mapHelpRequestInputs(items []request.ReportHelpRequest) ([]ports.HelpRequestInput, error) {
	if len(items) == 0 {
		return nil, nil
	}
	result := make([]ports.HelpRequestInput, len(items))
	for i := range items {
		item := items[i]
		var idPtr *uuid.UUID
		if item.ID != nil {
			id, err := parseUUID(*item.ID)
			if err != nil {
				return nil, errors.New("help.id must be a valid UUID")
			}
			idPtr = &id
		}
		helperID, err := parseUUIDPointer(item.HelperID)
		if err != nil {
			return nil, errors.New("help.helper_id must be a valid UUID")
		}
		result[i] = ports.HelpRequestInput{
			ID:          idPtr,
			Description: sanitizeStringPtr(item.Description),
			HelperID:    helperID,
			Status:      sanitizeStringPtr(item.Status),
		}
	}
	return result, nil
}

func mapTomorrowPlanInputs(items []request.ReportTomorrowPlan) ([]ports.TomorrowPlanInput, error) {
	if len(items) == 0 {
		return nil, nil
	}
	result := make([]ports.TomorrowPlanInput, len(items))
	for i := range items {
		item := items[i]
		var idPtr *uuid.UUID
		if item.ID != nil {
			id, err := parseUUID(*item.ID)
			if err != nil {
				return nil, errors.New("plan_tomorrow.id must be a valid UUID")
			}
			idPtr = &id
		}
		taskID, err := parseUUIDPointer(item.TaskID)
		if err != nil {
			return nil, errors.New("plan_tomorrow.task_id must be a valid UUID")
		}
		result[i] = ports.TomorrowPlanInput{
			ID:          idPtr,
			Description: sanitizeStringPtr(item.Description),
			TaskID:      taskID,
		}
	}
	return result, nil
}

func parseUUIDList(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]uuid.UUID, len(values))
	for i := range values {
		id, err := parseUUID(strings.TrimSpace(values[i]))
		if err != nil {
			return nil, err
		}
		result[i] = id
	}
	return result, nil
}

func parseLegacyPagination(pageParam, pageSizeParam, qPage, qPageSize string) (int, int) {
	page := parsePositiveInt(qPage, 1)
	pageSize := parsePositiveInt(qPageSize, 20)
	if pageParam != "" {
		if parsed, err := strconv.Atoi(pageParam); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pageSizeParam != "" {
		if parsed, err := strconv.Atoi(pageSizeParam); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	return page, pageSize
}

func uuidPtrToString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}
