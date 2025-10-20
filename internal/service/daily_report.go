package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"
	"emplacc-api/pkg/exel"

	"github.com/google/uuid"
)

type dailyReportService struct {
	repo    ports.DailyReportRepository
	storage ports.ObjectStorage
	cfg     config.MinioConfig
}

func NewDailyReportService(repo ports.DailyReportRepository, storage ports.ObjectStorage, cfg config.MinioConfig) ports.DailyReportService {
	return &dailyReportService{repo: repo, storage: storage, cfg: cfg}
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
	report, err := s.repo.CreateReport(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.syncReportFile(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *dailyReportService) UpdateReport(ctx context.Context, id uuid.UUID, input ports.UpdateDailyReportInput) (*models.DailyReport, error) {
	sanitizeUpdateReportInputs(&input)
	report, err := s.repo.UpdateReport(ctx, id, input)
	if err != nil {
		return nil, err
	}

	if err := s.syncReportFile(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *dailyReportService) DeleteReport(ctx context.Context, id uuid.UUID) error {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDeleteReport(ctx, id); err != nil {
		return err
	}

	return s.removeReportFile(ctx, report)
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

func (s *dailyReportService) DownloadReportFile(ctx context.Context, id uuid.UUID) (*ports.ReportFile, error) {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s.json", id.String())

	if s.storage == nil {
		data, err := serializeReport(report)
		if err != nil {
			return nil, err
		}
		return &ports.ReportFile{FileName: filename, ContentType: "application/json", Data: data}, nil
	}

	bucket, object := splitStoragePath(report.StorageObject)
	if bucket == "" || object == "" {
		if err := s.syncReportFile(ctx, report); err != nil {
			return nil, err
		}
		bucket, object = splitStoragePath(report.StorageObject)
	}

	data, contentType, err := s.storage.Get(ctx, bucket, object)
	if err != nil {
		payload, serr := serializeReport(report)
		if serr != nil {
			return nil, err
		}
		if uploadErr := s.storage.Upload(ctx, s.reportBucket(), reportObjectKey(report.ID), payload, "application/json", map[string]string{"report_id": report.ID.String()}); uploadErr == nil {
			path := fmt.Sprintf("%s/%s", s.reportBucket(), reportObjectKey(report.ID))
			_ = s.repo.UpdateReportStorage(ctx, report.ID, path)
			report.StorageObject = path
		}
		return &ports.ReportFile{FileName: filename, ContentType: "application/json", Data: payload}, nil
	}

	if strings.TrimSpace(contentType) == "" {
		contentType = "application/json"
	}

	return &ports.ReportFile{FileName: filename, ContentType: contentType, Data: data}, nil
}

func sanitizeReportInputs(input *ports.CreateDailyReportInput) {
	input.CompletedWork = sanitizeCompletedWorkInputs(input.CompletedWork)
	input.HelpRequests = sanitizeHelpRequestInputs(input.HelpRequests)
	input.TomorrowPlans = sanitizeTomorrowPlanInputs(input.TomorrowPlans)
}

func (s *dailyReportService) syncReportFile(ctx context.Context, report *models.DailyReport) error {
	if s.storage == nil || report == nil {
		return nil
	}

	data, err := serializeReport(report)
	if err != nil {
		return err
	}

	bucket := s.reportBucket()
	object := reportObjectKey(report.ID)
	path := fmt.Sprintf("%s/%s", bucket, object)

	if err := s.storage.Upload(ctx, bucket, object, data, "application/json", map[string]string{"report_id": report.ID.String()}); err != nil {
		return err
	}

	if err := s.repo.UpdateReportStorage(ctx, report.ID, path); err != nil {
		return err
	}

	report.StorageObject = path
	return nil
}

func (s *dailyReportService) removeReportFile(ctx context.Context, report *models.DailyReport) error {
	if s.storage == nil || report == nil {
		return nil
	}

	bucket, object := splitStoragePath(report.StorageObject)
	if bucket == "" || object == "" {
		return nil
	}

	return s.storage.Delete(ctx, bucket, object)
}

func (s *dailyReportService) reportBucket() string {
	bucket := strings.TrimSpace(s.cfg.ReportBucket)
	if bucket == "" {
		return "reports"
	}
	return bucket
}

func serializeReport(report *models.DailyReport) ([]byte, error) {
	if report == nil {
		return nil, domain.ErrInvalidInput
	}

	payload := reportFilePayload{
		ID:            report.ID.String(),
		UserID:        report.UserID.String(),
		Checked:       report.Checked,
		ReportDate:    report.ReportDate,
		CreatedAt:     report.CreatedAt,
		UpdatedAt:     report.UpdatedAt,
		CompletedWork: report.CompletedWork,
		HelpRequests:  report.HelpRequests,
		TomorrowPlans: report.TomorrowPlans,
		Problems:      report.ReportProblems,
	}

	return json.Marshal(payload)
}

func reportObjectKey(id uuid.UUID) string {
	return fmt.Sprintf("%s.json", id.String())
}

func splitStoragePath(path string) (string, string) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", ""
	}
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

type reportFilePayload struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	Checked       *int8                  `json:"checked,omitempty"`
	ReportDate    *time.Time             `json:"report_date,omitempty"`
	CreatedAt     *time.Time             `json:"created_at,omitempty"`
	UpdatedAt     *time.Time             `json:"updated_at,omitempty"`
	CompletedWork []models.CompletedWork `json:"completed_work,omitempty"`
	HelpRequests  []models.HelpRequest   `json:"help_requests,omitempty"`
	TomorrowPlans []models.TomorrowPlans `json:"tomorrow_plans,omitempty"`
	Problems      []models.ReportProblem `json:"problems,omitempty"`
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
		"[Content_Types].xml":        exel.ContentTypesXML,
		"_rels/.rels":                exel.RelationshipsXML,
		"xl/workbook.xml":            exel.WorkbookXML,
		"xl/_rels/workbook.xml.rels": exel.WorkbookRelationshipsXML,
		"xl/styles.xml":              exel.StylesXML,
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
	builder := &strings.Builder{}
	builder.WriteString(exel.XmlHeader)
	builder.WriteString("<worksheet " + exel.Namespace + "><sheetData>")

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
		result = string(rune('A'+(index%26))) + result
		index /= 26
	}
	return result
}

var _ ports.DailyReportService = (*dailyReportService)(nil)
