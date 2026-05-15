package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gedanmx/payments-api/internal/core/domain"
	"github.com/gedanmx/payments-api/internal/core/ports"
	"github.com/google/uuid"
)

type PaymentRequestedEvent struct {
	EventType string                `json:"event_type"`
	EventID   string                `json:"event_id"`
	Timestamp time.Time             `json:"timestamp"`
	Payload   PaymentRequestPayload `json:"payload"`
}

type PaymentRequestPayload struct {
	OrderID       string        `json:"order_id"`
	CorrelationID string        `json:"correlation_id"`
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

	webhookURL := os.Getenv("WEBHOOK_BASE_URL") + "/webhooks/mercadopago"

	resp, err := s.gateway.CreatePreference(ctx, ports.CreatePreferenceRequest{
		OrderID:       event.Payload.OrderID,
		CustomerEmail: event.Payload.CustomerEmail,
		Amount:        event.Payload.Amount,
		Currency:      event.Payload.Currency,
		Items:         items,
		WebhookURL:    webhookURL,
		BackURLs: ports.BackURLs{
			Success: os.Getenv("BACK_URL_SUCCESS"),
			Failure: os.Getenv("BACK_URL_FAILURE"),
			Pending: os.Getenv("BACK_URL_PENDING"),
		},
	})
	if err != nil {
		return fmt.Errorf("criar preferência mercado pago: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	payment := domain.Payment{
		ID:                uuid.NewString(),
		OrderID:           event.Payload.OrderID,
		CorrelationID:     event.Payload.CorrelationID,
		Provider:          "MERCADO_PAGO",
		PreferenceID:      resp.PreferenceID,
		CheckoutURL:       resp.CheckoutURL,
		ExpiresAt:         &expiresAt,
		BusinessStatus:    domain.BusinessStatusPending,
		SagaStatus:        domain.SagaStatusAwaitingPayment,
		Status:            domain.StatusPendingCustomerAction,
		TransactionAmount: event.Payload.Amount,
		Currency:          event.Payload.Currency,
		CustomerEmail:     event.Payload.CustomerEmail,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repository.Save(ctx, payment); err != nil {
		return fmt.Errorf("persistir pagamento: %w", err)
	}

	if err := s.broker.PublishPaymentCheckoutCreated(ctx, ports.PaymentCheckoutCreatedEvent{
		CorrelationID: payment.CorrelationID,
		OrderID:       payment.OrderID,
		PaymentID:     payment.ID,
		PreferenceID:  payment.PreferenceID,
		CheckoutURL:   payment.CheckoutURL,
		ExpiresAt:     expiresAt,
	}); err != nil {
		return fmt.Errorf("publicar PaymentCheckoutCreated: %w", err)
	}

	log.Printf("pagamento iniciado: order_id=%s preference_id=%s correlation_id=%s",
		event.Payload.OrderID, resp.PreferenceID, event.Payload.CorrelationID)
	return nil
}

func (s *PaymentService) ProcessWebhook(ctx context.Context, paymentID string) error {
	existing, err := s.repository.FindByPaymentID(ctx, paymentID)
	if err == nil && existing.BusinessStatus.IsFinal() {
		log.Printf("webhook idempotente ignorado: payment_id=%s business_status=%s", paymentID, existing.BusinessStatus)
		return nil
	}

	mpStatus, netAmount, err := s.gateway.GetPaymentStatus(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("consultar status pagamento %s: %w", paymentID, err)
	}

	if mpStatus == domain.StatusPending {
		log.Printf("pagamento ainda pendente, aguardando novo webhook: payment_id=%s", paymentID)
		return nil
	}

	payment, err := s.repository.FindByPaymentID(ctx, paymentID)
	if err != nil {
		log.Printf("pagamento não encontrado no repositório para payment_id=%s, publicando evento sem correlação local", paymentID)
		payment = domain.Payment{PaymentID: paymentID}
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
			log.Printf("aviso: falha ao atualizar repositório para payment_id=%s: %v", paymentID, err)
		}
		return s.broker.PublishPaymentApproved(ctx, ports.PaymentApprovedEvent{
			CorrelationID: payment.CorrelationID,
			OrderID:       payment.OrderID,
			PaymentID:     paymentID,
			PreferenceID:  payment.PreferenceID,
			Amount:        payment.TransactionAmount,
			Currency:      payment.Currency,
			ApprovedAt:    now,
		})
	case domain.StatusFailed, domain.StatusCancelled:
		if err := s.repository.UpdatePayment(ctx, payment.OrderID, paymentID, netAmount,
			finalMPStatus, domain.BusinessStatusFailed, domain.SagaStatusFailed); err != nil {
			log.Printf("aviso: falha ao atualizar repositório para payment_id=%s: %v", paymentID, err)
		}
		return s.broker.PublishPaymentFailed(ctx, ports.PaymentFailedEvent{
			CorrelationID: payment.CorrelationID,
			OrderID:       payment.OrderID,
			PaymentID:     paymentID,
			PreferenceID:  payment.PreferenceID,
			Amount:        payment.TransactionAmount,
			Currency:      payment.Currency,
			Reason:        string(mpStatus),
			FailedAt:      now,
		})
	}

	return nil
}
