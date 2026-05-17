package domain

import "time"

// PaymentStatus tracks the raw Mercado Pago payment status
type PaymentStatus string

const (
	StatusPendingCustomerAction PaymentStatus = "PENDING_CUSTOMER_ACTION"
	StatusApproved              PaymentStatus = "APPROVED"
	StatusFailed                PaymentStatus = "FAILED"
	StatusCancelled             PaymentStatus = "CANCELLED"
	StatusPending               PaymentStatus = "PENDING"
)

func (s PaymentStatus) IsFinal() bool {
	return s == StatusApproved || s == StatusFailed || s == StatusCancelled
}

// BusinessStatus is the external-facing payment status published in events
type BusinessStatus string

const (
	BusinessStatusPending  BusinessStatus = "PENDING"
	BusinessStatusApproved BusinessStatus = "APPROVED"
	BusinessStatusFailed   BusinessStatus = "FAILED"
)

func (s BusinessStatus) IsFinal() bool {
	return s == BusinessStatusApproved || s == BusinessStatusFailed
}

// SagaStatus tracks internal Saga orchestration state
type SagaStatus string

const (
	SagaStatusStarted          SagaStatus = "STARTED"
	SagaStatusAwaitingPayment  SagaStatus = "AWAITING_PAYMENT"
	SagaStatusPaymentConfirmed SagaStatus = "PAYMENT_CONFIRMED"
	SagaStatusFailed           SagaStatus = "FAILED"
)

type Payment struct {
	ID                string
	OrderID           string
	Provider          string
	PreferenceID      string
	PaymentID         string
	CheckoutURL       string
	ExpiresAt         *time.Time
	BusinessStatus BusinessStatus
	Status         SagaStatus
	PaymentStatus  PaymentStatus
	TransactionAmount float64
	NetAmount         float64
	Currency          string
	CustomerEmail     string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
