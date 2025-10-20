package request

import "github.com/google/uuid"

type CreateSubscription struct {
	UserID         string  `json:"user_id"`
	SubscriptionID *string `json:"subscription_id,omitempty"`
	TypeID         *int8   `json:"type_id,omitempty"`

	UserUUID         uuid.UUID  `json:"-"`
	SubscriptionUUID *uuid.UUID `json:"-"`
}
