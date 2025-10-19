package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToSubscriptionDTO(sub *models.Subscription) response.Subscription {
	if sub == nil {
		return response.Subscription{}
	}

	return response.Subscription{
		ID:             sub.ID.String(),
		UserID:         sub.UserID.String(),
		SubscriptionID: uuidPtrToString(sub.SubscriptionID),
		TypeID:         sub.TypeID,
		CreatedAt:      sub.CreatedAt,
		UpdatedAt:      sub.UpdatedAt,
		User:           toUserSummary(sub.User),
	}
}

func MapSubscriptionsPage(page *ports.Page[models.Subscription]) (dto.Pagination, response.SubscriptionsPage) {
	if page == nil {
		return dto.Pagination{}, response.SubscriptionsPage{}
	}

	items := make([]response.Subscription, len(page.Items))
	for i := range page.Items {
		items[i] = ToSubscriptionDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.SubscriptionsPage{Subscriptions: items}
}
