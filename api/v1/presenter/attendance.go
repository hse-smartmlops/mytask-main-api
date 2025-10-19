package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToAttendanceDTO(att *models.Attendance) response.Attendance {
	if att == nil {
		return response.Attendance{}
	}

	return response.Attendance{
		ID:            att.ID.String(),
		UserID:        att.UserID.String(),
		Date:          att.Date,
		WorkdayHours:  att.WorkdayHours,
		PlannedStart:  att.PlannedStart,
		ActualStart:   att.ActualStart,
		Commits:       att.Commits,
		MergeRequests: att.MergeRequests,
		CodeReviews:   att.CodeReviews,
		EndWork:       att.EndWork,
		Status:        att.Status,
		CreatedAt:     att.CreatedAt,
		UpdatedAt:     att.UpdatedAt,
	}
}

func MapAttendancesPage(page *ports.Page[models.Attendance]) (dto.Pagination, response.AttendancesPage) {
	if page == nil {
		return dto.Pagination{}, response.AttendancesPage{}
	}

	items := make([]response.Attendance, len(page.Items))
	for i := range page.Items {
		items[i] = ToAttendanceDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.AttendancesPage{Attendances: items}
}

func MapAttendancesByUser(userID string, attendances []models.Attendance) response.AttendancesByUser {
	items := make([]response.Attendance, len(attendances))
	for i := range attendances {
		items[i] = ToAttendanceDTO(&attendances[i])
	}

	return response.AttendancesByUser{
		UserID:      userID,
		Attendances: items,
	}
}
