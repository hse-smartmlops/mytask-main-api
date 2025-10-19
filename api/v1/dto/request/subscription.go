package request

type CreateSubscription struct {
	UserID         string  `json:"user_id"`
	SubscriptionID *string `json:"subscription_id,omitempty"`
	TypeID         *int8   `json:"type_id,omitempty"`
}
