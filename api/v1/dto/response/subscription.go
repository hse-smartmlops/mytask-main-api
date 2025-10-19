package response

import "time"

type Subscription struct {
	ID             string       `json:"id"`
	UserID         string       `json:"user_id"`
	SubscriptionID *string      `json:"subscription_id,omitempty"`
	TypeID         *int8        `json:"type_id,omitempty"`
	CreatedAt      *time.Time   `json:"created_at,omitempty"`
	UpdatedAt      *time.Time   `json:"updated_at,omitempty"`
	User           *UserSummary `json:"user,omitempty"`
}

type SubscriptionsPage struct {
	Subscriptions []Subscription `json:"subscriptions"`
}
