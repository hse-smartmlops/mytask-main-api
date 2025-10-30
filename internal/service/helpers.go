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
	excel "emplacc-api/pkg/excel"
	pb "emplacc-api/pkg/pb/v1"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

func stringPtr(
	value string,
) *string {
	return &value
}

func cloneStringPtr(
	value *string,
) *string {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}
	return &v
}

func sanitizeDescription(
	items []string,
) []string {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return clean
}

func checkPagination(
	p *ports.PaginationParams,
	config *config.PaginationConfig,
) {
	if p.Page <= 0 {
		p.Page = config.DefaultPageLimit
	}
	if p.Page > config.MaxPageLimit {
		p.Page = config.MaxPageLimit
	}
	if p.PageSize <= 0 {
		p.PageSize = config.DefaultPageSizeLimit
	}
	if p.PageSize > config.MaxPageSizeLimit {
		p.PageSize = config.MaxPageSizeLimit
	}
}

func boolValue(
	value *bool,
	fallback bool,
) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func stringValue(
	value *string,
) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func int64Value(
	value *int64,
) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func prepareAvatar(
	data []byte,
) ([]byte, string, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", domain.ErrInvalidInput
	}

	avatar := imaging.Fill(img, 256, 256, imaging.Center, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, avatar, imaging.PNG); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), "image/png", nil
}

func safePointer(
	value *string,
) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func isUniqueViolation(
	err error,
) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

func sanitizeReportInputs(
	input *ports.CreateDailyReportInput,
) {
	input.CompletedWork = sanitizeCompletedWorkInputs(input.CompletedWork)
	input.HelpRequests = sanitizeHelpRequestInputs(input.HelpRequests)
	input.TomorrowPlans = sanitizeTomorrowPlanInputs(input.TomorrowPlans)
}

func reportExportObjectKey(
	fileName string,
) string {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = fmt.Sprintf("reports_%d.xlsx", time.Now().Unix())
	}
	return fmt.Sprintf("exports/%d_%s", time.Now().UnixNano(), name)
}

func sanitizeUpdateReportInputs(
	input *ports.UpdateDailyReportInput,
) {
	input.CompletedWork = sanitizeCompletedWorkInputs(input.CompletedWork)
	input.HelpRequests = sanitizeHelpRequestInputs(input.HelpRequests)
	input.TomorrowPlans = sanitizeTomorrowPlanInputs(input.TomorrowPlans)
}

func sanitizeCompletedWorkInputs(
	items []ports.CompletedWorkInput,
) []ports.CompletedWorkInput {
	for i := range items {
		items[i].Description = cloneStringPtr(items[i].Description)
	}
	return items
}

func sanitizeHelpRequestInputs(
	items []ports.HelpRequestInput,
) []ports.HelpRequestInput {
	for i := range items {
		items[i].Description = cloneStringPtr(items[i].Description)
		items[i].Status = cloneStringPtr(items[i].Status)
	}
	return items
}

func sanitizeTomorrowPlanInputs(
	items []ports.TomorrowPlanInput,
) []ports.TomorrowPlanInput {
	for i := range items {
		items[i].Description = cloneStringPtr(items[i].Description)
	}
	return items
}

func buildDailyReportsWorkbook(
	reports []models.DailyReport,
	startDate,
	endDate time.Time,
) ([]byte, error) {
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
		"[Content_Types].xml":        excel.ContentTypesXML,
		"_rels/.rels":                excel.RelationshipsXML,
		"xl/workbook.xml":            excel.WorkbookXML,
		"xl/_rels/workbook.xml.rels": excel.WorkbookRelationshipsXML,
		"xl/styles.xml":              excel.StylesXML,
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

func enumerateDates(
	start,
	end time.Time,
) []time.Time {
	var result []time.Time
	current := truncateDate(start)
	last := truncateDate(end)
	for !current.After(last) {
		result = append(result, current)
		current = current.AddDate(0, 0, 1)
	}
	return result
}

func truncateDate(
	value time.Time,
) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func buildUserName(
	user *models.User,
) string {
	full := strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
	if full != "" {
		return full
	}
	if user.Email != "" {
		return user.Email
	}
	return user.ID.String()
}

func describeCompletedWork(
	item *models.CompletedWork,
) string {
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

func buildSheetXML(
	users []string,
	dates []time.Time,
	data map[string]map[string][]string,
) string {
	builder := &strings.Builder{}
	builder.WriteString(excel.XmlHeader)
	builder.WriteString("<worksheet " + excel.Namespace + "><sheetData>")

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

func writeInlineCell(
	builder *strings.Builder,
	cellRef,
	value string,
) {
	builder.WriteString(`<c r="`)
	builder.WriteString(cellRef)
	builder.WriteString(`" t="inlineStr"><is><t xml:space="preserve">`)
	builder.WriteString(escapeXML(value))
	builder.WriteString(`</t></is></c>`)
}

func escapeXML(
	value string,
) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(value)); err != nil {
		return value
	}
	return buf.String()
}

func columnName(
	index int,
) string {
	result := ""
	for index > 0 {
		index--
		result = string(rune('A'+(index%26))) + result
		index /= 26
	}
	return result
}

func serializeReport(
	report *models.DailyReport,
) ([]byte, error) {
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

func reportObjectKey(
	id uuid.UUID,
) string {
	return fmt.Sprintf("%s.json", id.String())
}

func splitStoragePath(
	path string,
) (string, string) {
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

func buildAuthTokens(
	tokens *ports.AuthRepositoryTokens,
) ports.AuthTokens {
	if tokens == nil {
		return ports.AuthTokens{}
	}

	now := time.Now().UTC()

	return ports.AuthTokens{
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		ExpiresIn:        tokens.ExpiresIn,
		RefreshExpiresIn: tokens.RefreshExpiresIn,
		TokenType:        tokens.TokenType,
		ExpiresAt:        now.Add(time.Duration(tokens.ExpiresIn) * time.Second),
	}
}

func copyMeta(
	meta map[string]string,
) map[string]string {
	if len(meta) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(meta))
	for k, v := range meta {
		out[k] = v
	}
	return out
}

func withOptionalTimeout(
	ctx context.Context,
	timeoutMs *int64,
) (context.Context, context.CancelFunc) {
	if timeoutMs != nil && *timeoutMs > 0 {
		return context.WithTimeout(ctx, time.Duration(*timeoutMs)*time.Millisecond)
	}
	return context.WithCancel(ctx)
}

func convertProcessEvent(
	ev *pb.ProcessTaskEvent,
) ports.MCPStreamEvent {
	if status := ev.GetStatus(); status != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventStatus,
			Status: &ports.MCPStatusEvent{
				State:    status.GetState(),
				Message:  status.GetMessage(),
				Progress: status.GetProgress(),
			},
		}
	}

	if chunk := ev.GetChunk(); chunk != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventChunk,
			Chunk: &ports.MCPChunkEvent{
				Data:  chunk.GetData(),
				Index: chunk.GetIndex(),
			},
		}
	}

	if final := ev.GetFinal(); final != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventFinal,
			Final: &ports.MCPFinalEvent{
				Result:      final.GetResult(),
				ContentType: final.GetContentType(),
			},
		}
	}

	if err := ev.GetError(); err != nil {
		return ports.MCPStreamEvent{
			Type: ports.MCPStreamEventError,
			Error: &ports.MCPErrorEvent{
				Code:    err.GetCode(),
				Message: err.GetMessage(),
			},
		}
	}

	return ports.MCPStreamEvent{Type: ports.MCPStreamEventStatus}
}
