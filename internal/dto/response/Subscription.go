package response

import (
	"time"
)

type SubscriptionUniversalResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type SubscriptionResponse struct {
	ID             string `json:"id"`
	UserId         string `json:"user_id"`
	SubscriptionId string `json:"subscription_id"`
	TypeId         int8 `json:"type_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

type SubscriptionListResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

type SubscriptionListByUserIdResponse struct {
	UserId        string                 `json:"user_id"`
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

type SubscriptionListBySubObjectResponse struct {
	TypeId         int8                  `json:"type_id"`
	SubscriptionId string                `json:"subscription_id"`
	Subscriptions  []SubscriptionResponse `json:"subscriptions"`
}
