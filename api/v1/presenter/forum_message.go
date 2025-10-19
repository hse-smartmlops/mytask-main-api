package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToForumMessageDTO(msg *models.ForumMessage) response.ForumMessage {
	if msg == nil {
		return response.ForumMessage{}
	}

	var creatorID *string
	if msg.CreatorID != nil {
		id := msg.CreatorID.String()
		creatorID = &id
	}

	return response.ForumMessage{
		ID:          msg.ID.String(),
		ProblemID:   msg.ProblemID.String(),
		Description: msg.Description,
		CreatorID:   creatorID,
		CreatedAt:   msg.CreatedAt,
		UpdatedAt:   msg.UpdatedAt,
	}
}

func MapForumMessagesPage(page *ports.Page[models.ForumMessage]) (dto.Pagination, response.ForumMessagesPage) {
	if page == nil {
		return dto.Pagination{}, response.ForumMessagesPage{}
	}

	items := make([]response.ForumMessage, len(page.Items))
	for i := range page.Items {
		items[i] = ToForumMessageDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.ForumMessagesPage{Messages: items}
}
