package ports

import (
	"context"
	"errors"

	"github.com/gedanmx/payments-api/internal/core/domain"
)

var ErrNoApprovedPayment = errors.New("merchant order has no approved payment")

type CreatePreferenceRequest struct {
	OrderID       string
	CustomerEmail string
	Amount        float64
	Currency      string
	Items         []PaymentItem
	BackURLs      BackURLs
}

type PaymentItem struct {
	ID         string
	Title      string
	Quantity   int
	UnitPrice  float64
	CurrencyID string
}

type BackURLs struct {
	Success string
	Failure string
	Pending string
}

type CreatePreferenceResponse struct {
	PreferenceID string
	CheckoutURL  string
}

type PaymentGateway interface {
	CreatePreference(ctx context.Context, req CreatePreferenceRequest) (CreatePreferenceResponse, error)
	GetPaymentStatus(ctx context.Context, paymentID string) (domain.PaymentStatus, float64, string, error)
	GetMerchantOrderPaymentID(ctx context.Context, merchantOrderID string) (string, error)
}
