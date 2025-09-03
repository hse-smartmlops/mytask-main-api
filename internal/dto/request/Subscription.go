package request

type SubscriptionCreateRequest struct {
	UserId         string `json:"user_id"`
	SubscriptionId string `json:"subscription_id"`
	TypeId         *int8  `json:"type_id"`
}
