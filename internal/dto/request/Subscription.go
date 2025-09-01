package request

type SubscriptionCreateRequest struct {
	UserId         string `json:"user_id"`
	SubscriptionId string `json:"subscription_id"`
	TypeId         int8   `json:"type_id"`
}

type SubscriptionListByUserRequest struct {
	UserID string `json:"user_id"`
}

type SubscriptionListBySubObjectRequest struct {
	TypeId         int8   `json:"type_id"`
	SubscriptionId string `json:"subscription_id"`
}
