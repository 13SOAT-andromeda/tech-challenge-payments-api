package ports

import (
	"context"
	"time"
)

type PaymentCheckoutCreatedEvent struct {
	OrderID       string
	PaymentID     string
	PreferenceID  string
	CheckoutURL   string
	ExpiresAt     time.Time
}

type PaymentApprovedEvent struct {
	OrderID       string
	PaymentID     string
	PreferenceID  string
	Amount        float64
	Currency      string
	ApprovedAt    time.Time
}

type PaymentFailedEvent struct {
	OrderID       string
	PaymentID     string
	PreferenceID  string
	Amount        float64
	Currency      string
	Reason        string
	FailedAt      time.Time
}

type MessageBroker interface {
	PublishPaymentCheckoutCreated(ctx context.Context, event PaymentCheckoutCreatedEvent) error
	PublishPaymentApproved(ctx context.Context, event PaymentApprovedEvent) error
	PublishPaymentFailed(ctx context.Context, event PaymentFailedEvent) error
	PublishPaymentRejected(ctx context.Context, event PaymentFailedEvent) error
}
