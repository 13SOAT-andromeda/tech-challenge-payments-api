package ports

import (
	"context"

	"github.com/gedanmx/payments-api/internal/core/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, payment domain.Payment) error
	FindByOrderID(ctx context.Context, orderID string) (domain.Payment, error)
	FindByPaymentID(ctx context.Context, paymentID string) (domain.Payment, error)
	UpdatePayment(ctx context.Context, orderID, paymentID string, netAmount float64, status domain.PaymentStatus, businessStatus domain.BusinessStatus, sagaStatus domain.SagaStatus) error
}
