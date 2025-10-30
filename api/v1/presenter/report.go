package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToReportDTO(report *models.DailyReport) response.Report {
	if report == nil {
		return response.Report{}
	}

	items := make([]response.CompletedWorkItem, 0, len(report.CompletedWork))
	for i := range report.CompletedWork {
		item := report.CompletedWork[i]
		if item.Deleted != nil && *item.Deleted {
			continue
		}
		items = append(items, response.CompletedWorkItem{
			ID:          item.ID.String(),
			Description: item.Description,
			TaskID:      uuidPtrToString(item.TaskID),
		})
	}

	plans := make([]response.TomorrowPlanItem, 0, len(report.TomorrowPlans))
	for i := range report.TomorrowPlans {
		item := report.TomorrowPlans[i]
		if item.Deleted != nil && *item.Deleted {
			continue
		}
		plans = append(plans, response.TomorrowPlanItem{
			ID:          item.ID.String(),
			Description: item.Description,
			TaskID:      uuidPtrToString(item.TaskID),
		})
	}

	helpRequests := make([]response.HelpRequestItem, 0, len(report.HelpRequests))
	for i := range report.HelpRequests {
		item := report.HelpRequests[i]
		if item.Deleted != nil && *item.Deleted {
			continue
		}
		helpRequests = append(helpRequests, response.HelpRequestItem{
			ID:          item.ID.String(),
			Description: item.Description,
			HelperID:    uuidPtrToString(item.HelperID),
			Status:      item.Status,
		})
	}

	problems := make([]response.ProblemItem, 0, len(report.ReportProblems))
	for i := range report.ReportProblems {
		item := report.ReportProblems[i]
		if item.Deleted != nil && *item.Deleted {
			continue
		}
		if item.Problem == nil {
			continue
		}
		problems = append(problems, response.ProblemItem{
			ID:          item.Problem.ID.String(),
			Description: item.Problem.Description,
		})
	}

	return response.Report{
		ID:            report.ID.String(),
		UserID:        report.UserID.String(),
		ReportDate:    report.ReportDate,
		Checked:       report.Checked,
		CreatedAt:     report.CreatedAt,
		UpdatedAt:     report.UpdatedAt,
		CompletedWork: items,
		TomorrowPlans: plans,
		HelpRequests:  helpRequests,
		Problems:      problems,
		UserInfo:      toUserSummary(report.User),
	}
}

func MapReportsPage(page *ports.Page[models.DailyReport]) (dto.Pagination, response.ReportsPage) {
	if page == nil {
		return dto.Pagination{}, response.ReportsPage{}
	}

	items := make([]response.Report, len(page.Items))
	for i := range page.Items {
		items[i] = ToReportDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.ReportsPage{Reports: items}
}

func MapHelpRequestsForUser(items []models.HelpRequest) response.HelpRequestsForUser {
	result := make([]response.HelpRequestWithAssigner, 0, len(items))
	for i := range items {
		item := items[i]
		if item.Deleted != nil && *item.Deleted {
			continue
		}
		entry := response.HelpRequestWithAssigner{
			HelpRequest: response.HelpRequestItem{
				ID:          item.ID.String(),
				Description: item.Description,
				HelperID:    uuidPtrToString(item.HelperID),
				Status:      item.Status,
			},
		}
		if item.Report != nil && item.Report.User != nil {
			entry.UserFirstName = &item.Report.User.FirstName
			entry.UserLastName = &item.Report.User.LastName
		}
		result = append(result, entry)
	}
	return response.HelpRequestsForUser{HelpRequests: result}
}

func ToHelpRequestDTO(helpRequest *models.HelpRequest) response.HelpRequestWithAssigner {
	if helpRequest == nil {
		return response.HelpRequestWithAssigner{}
	}

	if helpRequest.Deleted != nil && *helpRequest.Deleted {
		return response.HelpRequestWithAssigner{}
	}

	dto := response.HelpRequestWithAssigner{
		HelpRequest: response.HelpRequestItem{
			ID:          helpRequest.ID.String(),
			Description: helpRequest.Description,
			HelperID:    uuidPtrToString(helpRequest.HelperID),
			Status:      helpRequest.Status,
		},
	}

	if helpRequest.Report != nil && helpRequest.Report.User != nil {
		dto.UserFirstName = &helpRequest.Report.User.FirstName
		dto.UserLastName = &helpRequest.Report.User.LastName
	}

	return dto
}

func MapHelpRequestsPage(page *ports.Page[models.HelpRequest]) (dto.Pagination, response.HelpRequestsForUser) {
	if page == nil {
		return dto.Pagination{}, response.HelpRequestsForUser{}
	}

	helpRequests := make([]response.HelpRequestWithAssigner, 0, len(page.Items))
	for i := range page.Items {
		item := page.Items[i]
		if item.Deleted != nil && *item.Deleted {
			continue
		}
		helpRequests = append(helpRequests, ToHelpRequestDTO(&item))
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.HelpRequestsForUser{HelpRequests: helpRequests}
}