package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type dailyReportService struct {
	repo ports.DailyReportRepository
}

func NewDailyReportService(repo ports.DailyReportRepository) ports.DailyReportService {
	return &dailyReportService{repo: repo}
}

func (s *dailyReportService) ListReports(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.DailyReport], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListReports(ctx, params)
}

func (s *dailyReportService) ListReportsByUser(ctx context.Context, userID uuid.UUID, params ports.PaginationParams) (*ports.Page[models.DailyReport], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListReportsByUser(ctx, userID, params)
}

func (s *dailyReportService) ListReportsByProject(ctx context.Context, projectID uuid.UUID) ([]models.DailyReport, error) {
	return s.repo.ListReportsByProject(ctx, projectID)
}

func (s *dailyReportService) ListReportsByTask(ctx context.Context, taskID uuid.UUID) ([]models.DailyReport, error) {
	return s.repo.ListReportsByTask(ctx, taskID)
}

func (s *dailyReportService) ListReportsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.DailyReport, error) {
	if endDate.Before(startDate) {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.ListReportsByDateRange(ctx, startDate, endDate)
}

func (s *dailyReportService) ListHelpRequestsByHelper(ctx context.Context, helperID uuid.UUID) ([]models.HelpRequest, error) {
	return s.repo.ListHelpRequestsByHelper(ctx, helperID)
}

func (s *dailyReportService) GetReport(ctx context.Context, id uuid.UUID) (*models.DailyReport, error) {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, domain.ErrNotFound
	}
	return report, nil
}

func (s *dailyReportService) CreateReport(ctx context.Context, input ports.CreateDailyReportInput) (*models.DailyReport, error) {
	if input.UserID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}

	sanitizeReportInputs(&input)
	return s.repo.CreateReport(ctx, input)
}

func (s *dailyReportService) UpdateReport(ctx context.Context, id uuid.UUID, input ports.UpdateDailyReportInput) (*models.DailyReport, error) {
	sanitizeUpdateReportInputs(&input)
	return s.repo.UpdateReport(ctx, id, input)
}

func (s *dailyReportService) DeleteReport(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteReport(ctx, id)
}

func (s *dailyReportService) UpdateCompletedWork(ctx context.Context, id uuid.UUID, input ports.UpdateCompletedWorkInput) (*models.CompletedWork, error) {
	input.Description = cloneStringPtr(input.Description)
	return s.repo.UpdateCompletedWork(ctx, id, input)
}

func (s *dailyReportService) UpdateHelpRequest(ctx context.Context, id uuid.UUID, input ports.UpdateHelpRequestInput) (*models.HelpRequest, error) {
	input.Description = cloneStringPtr(input.Description)
	input.Status = cloneStringPtr(input.Status)
	return s.repo.UpdateHelpRequest(ctx, id, input)
}

func (s *dailyReportService) DeleteHelpRequest(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteHelpRequest(ctx, id)
}

func (s *dailyReportService) UpdateTomorrowPlan(ctx context.Context, id uuid.UUID, input ports.UpdateTomorrowPlanInput) (*models.TomorrowPlans, error) {
	input.Description = cloneStringPtr(input.Description)
	return s.repo.UpdateTomorrowPlan(ctx, id, input)
}

func (s *dailyReportService) ExportReportsToXLSX(ctx context.Context, input ports.ReportsByDateInput) ([]byte, error) {
	if input.EndDate.Before(input.StartDate) {
		return nil, domain.ErrInvalidInput
	}

	reports, err := s.repo.ListReportsByDateRange(ctx, input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}

	return buildDailyReportsWorkbook(reports, input.StartDate, input.EndDate)
}

func sanitizeReportInputs(input *ports.CreateDailyReportInput) {
	input.CompletedWork = sanitizeCompletedWorkInputs(input.CompletedWork)
	input.HelpRequests = sanitizeHelpRequestInputs(input.HelpRequests)
	input.TomorrowPlans = sanitizeTomorrowPlanInputs(input.TomorrowPlans)
}

func sanitizeUpdateReportInputs(input *ports.UpdateDailyReportInput) {
	input.CompletedWork = sanitizeCompletedWorkInputs(input.CompletedWork)
	input.HelpRequests = sanitizeHelpRequestInputs(input.HelpRequests)
	input.TomorrowPlans = sanitizeTomorrowPlanInputs(input.TomorrowPlans)
}

func sanitizeCompletedWorkInputs(items []ports.CompletedWorkInput) []ports.CompletedWorkInput {
	for i := range items {
		items[i].Description = cloneStringPtr(items[i].Description)
	}
	return items
}

func sanitizeHelpRequestInputs(items []ports.HelpRequestInput) []ports.HelpRequestInput {
	for i := range items {
		items[i].Description = cloneStringPtr(items[i].Description)
		items[i].Status = cloneStringPtr(items[i].Status)
	}
	return items
}

func sanitizeTomorrowPlanInputs(items []ports.TomorrowPlanInput) []ports.TomorrowPlanInput {
	for i := range items {
		items[i].Description = cloneStringPtr(items[i].Description)
	}
	return items
}

func buildDailyReportsWorkbook(reports []models.DailyReport, startDate, endDate time.Time) ([]byte, error) {
	dates := enumerateDates(startDate, endDate)
	userData := make(map[string]map[string][]string)
	userOrder := make([]string, 0)

	for i := range reports {
		report := reports[i]
		if report.User == nil {
			continue
		}
		userName := buildUserName(report.User)
		if _, ok := userData[userName]; !ok {
			userData[userName] = make(map[string][]string)
			userOrder = append(userOrder, userName)
		}
		if report.ReportDate == nil {
			continue
		}
		dateKey := report.ReportDate.Format("2006-01-02")
		if _, ok := userData[userName][dateKey]; !ok {
			userData[userName][dateKey] = make([]string, 0)
		}
		for j := range report.CompletedWork {
			item := report.CompletedWork[j]
			if item.Deleted != nil && *item.Deleted {
				continue
			}
			if desc := strings.TrimSpace(describeCompletedWork(&item)); desc != "" {
				userData[userName][dateKey] = append(userData[userName][dateKey], desc)
			}
		}
	}

	sort.Strings(userOrder)

	sheetXML := buildSheetXML(userOrder, dates, userData)

	buffer := &bytes.Buffer{}
	writer := zip.NewWriter(buffer)
	files := map[string]string{
		"[Content_Types].xml":        contentTypesXML,
		"_rels/.rels":                relationshipsXML,
		"xl/workbook.xml":            workbookXML,
		"xl/_rels/workbook.xml.rels": workbookRelationshipsXML,
		"xl/styles.xml":              stylesXML,
		"xl/worksheets/sheet1.xml":   sheetXML,
	}

	for name, content := range files {
		f, err := writer.Create(name)
		if err != nil {
			writer.Close()
			return nil, err
		}
		if _, err := f.Write([]byte(content)); err != nil {
			writer.Close()
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func enumerateDates(start, end time.Time) []time.Time {
	var result []time.Time
	current := truncateDate(start)
	last := truncateDate(end)
	for !current.After(last) {
		result = append(result, current)
		current = current.AddDate(0, 0, 1)
	}
	return result
}

func truncateDate(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func buildUserName(user *models.User) string {
	full := strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
	if full != "" {
		return full
	}
	if user.Email != "" {
		return user.Email
	}
	return user.ID.String()
}

func describeCompletedWork(item *models.CompletedWork) string {
	if item.Description != nil {
		if trimmed := strings.TrimSpace(*item.Description); trimmed != "" {
			return trimmed
		}
	}
	if item.Task != nil && item.Task.Name != nil {
		if trimmed := strings.TrimSpace(*item.Task.Name); trimmed != "" {
			return trimmed
		}
	}
	return item.ID.String()
}

func buildSheetXML(users []string, dates []time.Time, data map[string]map[string][]string) string {
	const (
		xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
		namespace = `xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"`
	)

	builder := &strings.Builder{}
	builder.WriteString(xmlHeader)
	builder.WriteString("<worksheet " + namespace + "><sheetData>")

	builder.WriteString(`<row r="1">`)
	writeInlineCell(builder, "A1", "User")
	for idx, date := range dates {
		col := columnName(idx + 2)
		writeInlineCell(builder, fmt.Sprintf("%s1", col), date.Format("2006-01-02"))
	}
	builder.WriteString(`</row>`)

	for rowIdx, user := range users {
		rowNumber := rowIdx + 2
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowNumber))
		writeInlineCell(builder, fmt.Sprintf("A%d", rowNumber), user)
		for colIdx, date := range dates {
			col := columnName(colIdx + 2)
			values := data[user][date.Format("2006-01-02")]
			writeInlineCell(builder, fmt.Sprintf("%s%d", col, rowNumber), strings.Join(values, "\n"))
		}
		builder.WriteString(`</row>`)
	}

	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func writeInlineCell(builder *strings.Builder, cellRef, value string) {
	builder.WriteString(`<c r="`)
	builder.WriteString(cellRef)
	builder.WriteString(`" t="inlineStr"><is><t xml:space="preserve">`)
	builder.WriteString(escapeXML(value))
	builder.WriteString(`</t></is></c>`)
}

func escapeXML(value string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(value)); err != nil {
		return value
	}
	return buf.String()
}

func columnName(index int) string {
	result := ""
	for index > 0 {
		index--
		result = string('A'+(index%26)) + result
		index /= 26
	}
	return result
}

const (
	contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
	<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
	<Default Extension="xml" ContentType="application/xml"/>
	<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
	<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
	<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`
	relationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
	<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`
	workbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
	<sheets>
		<sheet name="Reports" sheetId="1" r:id="rId1"/>
	</sheets>
</workbook>`
	workbookRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
	<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
	<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
	stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"></styleSheet>`
)

var _ ports.DailyReportService = (*dailyReportService)(nil)
