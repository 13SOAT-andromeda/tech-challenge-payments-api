package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gedanmx/payments-api/internal/core/domain"
	"github.com/gedanmx/payments-api/internal/core/ports"
	"github.com/google/uuid"
)

type PaymentRequestedEvent struct {
	EventType string                `json:"event_type"`
	Payload   PaymentRequestPayload `json:"payload"`
}

type PaymentRequestPayload struct {
	OrderID       string        `json:"order_id"`
	CustomerID    string        `json:"customer_id"`
	CustomerEmail string        `json:"customer_email"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Items         []PayloadItem `json:"items"`
}

type PayloadItem struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type PaymentService struct {
	gateway    ports.PaymentGateway
	broker     ports.MessageBroker
	repository ports.PaymentRepository
}

func NewPaymentService(gateway ports.PaymentGateway, broker ports.MessageBroker, repository ports.PaymentRepository) *PaymentService {
	return &PaymentService{
		gateway:    gateway,
		broker:     broker,
		repository: repository,
	}
}

func (s *PaymentService) ProcessPaymentRequest(ctx context.Context, event PaymentRequestedEvent) error {
	start := time.Now()
	slog.Info("service.ProcessPaymentRequest início",
		"op", "ProcessPaymentRequest",
		"order_id", event.Payload.OrderID,
		"event_type", event.EventType,
	)

	if existing, err := s.repository.FindByOrderID(ctx, event.Payload.OrderID); err == nil {
		slog.Info("service.ProcessPaymentRequest idempotente — ignorado",
			"order_id", event.Payload.OrderID,
			"existing_payment_id", existing.ID,
			"business_status", existing.BusinessStatus,
		)
		return nil
	}

	items := make([]ports.PaymentItem, len(event.Payload.Items))
	for i, item := range event.Payload.Items {
		items[i] = ports.PaymentItem{
			ID:         item.ID,
			Title:      item.Title,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			CurrencyID: event.Payload.Currency,
		}
	}

	resp, err := s.gateway.CreatePreference(ctx, ports.CreatePreferenceRequest{
		OrderID:       event.Payload.OrderID,
		CustomerEmail: event.Payload.CustomerEmail,
		Amount:        event.Payload.Amount,
		Currency:      event.Payload.Currency,
		Items:         items,
	})
	if err != nil {
		slog.Error("service.ProcessPaymentRequest erro",
			"op", "ProcessPaymentRequest",
			"order_id", event.Payload.OrderID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return fmt.Errorf("criar preferência mercado pago: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	payment := domain.Payment{
		ID:                uuid.NewString(),
		OrderID:           event.Payload.OrderID,
		Provider:          "MERCADO_PAGO",
		PreferenceID:      resp.PreferenceID,
		CheckoutURL:       resp.CheckoutURL,
		ExpiresAt:         &expiresAt,
		BusinessStatus: domain.BusinessStatusPending,
		Status:         domain.SagaStatusAwaitingPayment,
		PaymentStatus:  domain.StatusPendingCustomerAction,
		TransactionAmount: event.Payload.Amount,
		Currency:          event.Payload.Currency,
		CustomerEmail:     event.Payload.CustomerEmail,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repository.Save(ctx, payment); err != nil {
		slog.Error("service.ProcessPaymentRequest erro ao persistir",
			"order_id", event.Payload.OrderID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return fmt.Errorf("persistir pagamento: %w", err)
	}

	if err := s.broker.PublishPaymentCheckoutCreated(ctx, ports.PaymentCheckoutCreatedEvent{
		OrderID:       payment.OrderID,
		PaymentID:     payment.ID,
		PreferenceID:  payment.PreferenceID,
		CheckoutURL:   payment.CheckoutURL,
		ExpiresAt:     expiresAt,
	}); err != nil {
		slog.Error("service.ProcessPaymentRequest erro ao publicar checkout",
			"order_id", event.Payload.OrderID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return fmt.Errorf("publicar PaymentCheckoutCreated: %w", err)
	}

	slog.Info("service.ProcessPaymentRequest concluído",
		"order_id", event.Payload.OrderID,
		"preference_id", resp.PreferenceID,
		"checkout_url", resp.CheckoutURL,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}

func (s *PaymentService) ProcessWebhook(ctx context.Context, paymentID string) error {
	start := time.Now()
	slog.Info("service.ProcessWebhook início", "op", "ProcessWebhook", "payment_id", paymentID)

	existing, err := s.repository.FindByPaymentID(ctx, paymentID)
	if err == nil && existing.BusinessStatus.IsFinal() {
		slog.Info("service.ProcessWebhook idempotente — ignorado",
			"payment_id", paymentID,
			"business_status", existing.BusinessStatus,
		)
		return nil
	}

	mpStatus, netAmount, orderID, err := s.gateway.GetPaymentStatus(ctx, paymentID)
	if err != nil {
		slog.Error("service.ProcessWebhook erro ao consultar status",
			"payment_id", paymentID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return fmt.Errorf("consultar status pagamento %s: %w", paymentID, err)
	}

	if mpStatus == domain.StatusPending {
		slog.Info("service.ProcessWebhook pagamento ainda pendente",
			"payment_id", paymentID,
			"mp_status", mpStatus,
		)
		return nil
	}

	payment, err := s.repository.FindByOrderID(ctx, orderID)
	if err != nil {
		slog.Warn("service.ProcessWebhook pagamento não encontrado no repositório",
			"payment_id", paymentID,
			"order_id", orderID,
			"error", err,
		)
		payment = domain.Payment{OrderID: orderID, PaymentID: paymentID}
	}

	finalMPStatus := mpStatus
	if mpStatus == domain.StatusCancelled {
		finalMPStatus = domain.StatusFailed
	}

	now := time.Now()
	switch mpStatus {
	case domain.StatusApproved:
		if err := s.repository.UpdatePayment(ctx, payment.OrderID, paymentID, netAmount,
			finalMPStatus, domain.BusinessStatusApproved, domain.SagaStatusPaymentConfirmed); err != nil {
			slog.Warn("service.ProcessWebhook falha ao atualizar repositório",
				"payment_id", paymentID,
				"error", err,
			)
		}
		if err := s.broker.PublishPaymentApproved(ctx, ports.PaymentApprovedEvent{
			OrderID:       payment.OrderID,
			PaymentID:     paymentID,
			PreferenceID:  payment.PreferenceID,
			Amount:        payment.TransactionAmount,
			Currency:      payment.Currency,
			ApprovedAt:    now,
		}); err != nil {
			slog.Error("service.ProcessWebhook erro ao publicar PaymentApproved",
				"payment_id", paymentID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			return err
		}
		slog.Info("service.ProcessWebhook concluído",
			"payment_id", paymentID,
			"mp_status", mpStatus,
			"business_status", domain.BusinessStatusApproved,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil

	case domain.StatusFailed, domain.StatusCancelled:
		if err := s.repository.UpdatePayment(ctx, payment.OrderID, paymentID, netAmount,
			finalMPStatus, domain.BusinessStatusFailed, domain.SagaStatusFailed); err != nil {
			slog.Warn("service.ProcessWebhook falha ao atualizar repositório",
				"payment_id", paymentID,
				"error", err,
			)
		}
		if err := s.broker.PublishPaymentFailed(ctx, ports.PaymentFailedEvent{
			OrderID:      payment.OrderID,
			PaymentID:    paymentID,
			PreferenceID: payment.PreferenceID,
			Amount:       payment.TransactionAmount,
			Currency:     payment.Currency,
			Reason:       string(mpStatus),
			FailedAt:     now,
		}); err != nil {
			slog.Error("service.ProcessWebhook erro ao publicar PaymentFailed",
				"payment_id", paymentID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			return err
		}
		slog.Info("service.ProcessWebhook concluído",
			"payment_id", paymentID,
			"mp_status", mpStatus,
			"business_status", domain.BusinessStatusFailed,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil
	}

	return nil
}

func (s *PaymentService) ProcessPaymentRejected(ctx context.Context, orderID string) error {
	start := time.Now()
	slog.Info("service.ProcessPaymentRejected início", "op", "ProcessPaymentRejected", "order_id", orderID)

	payment, err := s.repository.FindByOrderID(ctx, orderID)
	if err != nil {
		slog.Warn("service.ProcessPaymentRejected pagamento não encontrado no repositório",
			"order_id", orderID,
			"error", err,
		)
		payment = domain.Payment{OrderID: orderID}
	}

	if payment.BusinessStatus.IsFinal() {
		slog.Info("service.ProcessPaymentRejected idempotente — ignorado",
			"order_id", orderID,
			"business_status", payment.BusinessStatus,
		)
		return nil
	}

	if err := s.repository.UpdatePayment(ctx, payment.OrderID, payment.PaymentID, payment.NetAmount,
		domain.StatusFailed, domain.BusinessStatusFailed, domain.SagaStatusFailed); err != nil {
		slog.Warn("service.ProcessPaymentRejected falha ao atualizar repositório",
			"order_id", orderID,
			"error", err,
		)
	}

	now := time.Now()
	if err := s.broker.PublishPaymentRejected(ctx, ports.PaymentFailedEvent{
		OrderID:      payment.OrderID,
		PaymentID:    payment.PaymentID,
		PreferenceID: payment.PreferenceID,
		Amount:       payment.TransactionAmount,
		Currency:     payment.Currency,
		Reason:       "rejected",
		FailedAt:     now,
	}); err != nil {
		slog.Error("service.ProcessPaymentRejected erro ao publicar PaymentRejected",
			"order_id", orderID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	slog.Info("service.ProcessPaymentRejected concluído",
		"order_id", orderID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}
