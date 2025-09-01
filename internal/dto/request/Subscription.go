package request

type SubscriptionCreateRequest struct {
	UserId         string `json:"user_id"`
	SubscriptionId string `json:"subscription_id"`
	TypeId         int8   `json:"type_id"`
}

type SubscriptionListRequest struct {
	Page     *int `json:"page"`
	PageSize *int `json:"page_size"`
}

type SubscriptionListByUserRequest struct {
	UserID   string `json:"user_id"`
	Page     *int   `json:"page"`
	PageSize *int   `json:"page_size"`
}

type SubscriptionListBySubObjectRequest struct {
	TypeId         int8   `json:"type_id"`
	SubscriptionId string `json:"subscription_id"`
	Page           *int   `json:"page"`
	PageSize       *int   `json:"page_size"`
}
