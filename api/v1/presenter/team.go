package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToTeamDTO(team *models.Team) response.Team {
	if team == nil {
		return response.Team{}
	}

	members := make([]response.TeamMember, 0, len(team.TeamMembers))
	for _, member := range team.TeamMembers {
		if member.User == nil {
			continue
		}
		members = append(members, response.TeamMember{
			UserID:         member.UserID.String(),
			FirstName:      &member.User.FirstName,
			LastName:       &member.User.LastName,
			Email:          member.User.Email,
			Specialization: member.Specialization,
		})
	}

	return response.Team{
		ID:          team.ID.String(),
		Name:        valueOrEmpty(team.Name),
		Description: team.Description,
		CreatedAt:   team.CreatedAt,
		UpdatedAt:   team.UpdatedAt,
		Members:     members,
	}
}

func MapTeamsPage(page *ports.Page[models.Team]) (dto.Pagination, response.TeamsPage) {
	if page == nil {
		return dto.Pagination{}, response.TeamsPage{}
	}

	items := make([]response.Team, len(page.Items))
	for i := range page.Items {
		items[i] = ToTeamDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.TeamsPage{Teams: items}
}
