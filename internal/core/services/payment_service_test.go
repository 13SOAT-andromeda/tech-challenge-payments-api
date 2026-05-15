package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gedanmx/payments-api/internal/core/domain"
	"github.com/gedanmx/payments-api/internal/core/ports"
	"github.com/gedanmx/payments-api/internal/core/services"
)

// --- Mocks ---

type mockGateway struct {
	createFn    func(ctx context.Context, req ports.CreatePreferenceRequest) (ports.CreatePreferenceResponse, error)
	getStatusFn func(ctx context.Context, paymentID string) (domain.PaymentStatus, float64, error)
}

func (m *mockGateway) CreatePreference(ctx context.Context, req ports.CreatePreferenceRequest) (ports.CreatePreferenceResponse, error) {
	return m.createFn(ctx, req)
}

func (m *mockGateway) GetPaymentStatus(ctx context.Context, paymentID string) (domain.PaymentStatus, float64, error) {
	return m.getStatusFn(ctx, paymentID)
}

type mockBroker struct {
	checkoutCreatedCalled bool
	approvedCalled        bool
	failedCalled          bool
	checkoutErr           error
	approvedErr           error
	failedErr             error
}

func (m *mockBroker) PublishPaymentCheckoutCreated(_ context.Context, _ ports.PaymentCheckoutCreatedEvent) error {
	m.checkoutCreatedCalled = true
	return m.checkoutErr
}

func (m *mockBroker) PublishPaymentApproved(_ context.Context, _ ports.PaymentApprovedEvent) error {
	m.approvedCalled = true
	return m.approvedErr
}

func (m *mockBroker) PublishPaymentFailed(_ context.Context, _ ports.PaymentFailedEvent) error {
	m.failedCalled = true
	return m.failedErr
}

type mockRepo struct {
	saved               *domain.Payment
	findByOrderID       domain.Payment
	findByPaymentID     domain.Payment
	findByPaymentIDErr  error
	updateErr           error
}

func (m *mockRepo) Save(_ context.Context, p domain.Payment) error {
	m.saved = &p
	return nil
}

func (m *mockRepo) FindByOrderID(_ context.Context, _ string) (domain.Payment, error) {
	return m.findByOrderID, nil
}

func (m *mockRepo) FindByPaymentID(_ context.Context, _ string) (domain.Payment, error) {
	return m.findByPaymentID, m.findByPaymentIDErr
}

func (m *mockRepo) UpdatePayment(_ context.Context, _, _ string, _ float64, _ domain.PaymentStatus, _ domain.BusinessStatus, _ domain.SagaStatus) error {
	return m.updateErr
}

// --- Testes ---

func makeEvent(orderID string) services.PaymentRequestedEvent {
	return services.PaymentRequestedEvent{
		EventType: "payment.requested",
		EventID:   "evt-001",
		Timestamp: time.Now(),
		Payload: services.PaymentRequestPayload{
			OrderID:       orderID,
			CorrelationID: "corr-001",
			CustomerEmail: "test@email.com",
			Amount:        850.0,
			Currency:      "BRL",
			Items: []services.PayloadItem{
				{ID: "SVC-001", Title: "Troca de Óleo", Quantity: 1, UnitPrice: 850.0},
			},
		},
	}
}

func TestProcessPaymentRequest_Success(t *testing.T) {
	gateway := &mockGateway{
		createFn: func(_ context.Context, _ ports.CreatePreferenceRequest) (ports.CreatePreferenceResponse, error) {
			return ports.CreatePreferenceResponse{
				PreferenceID: "pref-123",
				CheckoutURL:  "https://mp.com/checkout",
			}, nil
		},
	}
	broker := &mockBroker{}
	repo := &mockRepo{}
	svc := services.NewPaymentService(gateway, broker, repo)

	err := svc.ProcessPaymentRequest(context.Background(), makeEvent("order-001"))
	if err != nil {
		t.Fatalf("esperava nil, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("esperava que o pagamento fosse persistido")
	}
	if repo.saved.PreferenceID != "pref-123" {
		t.Errorf("preference_id incorreto: got %s", repo.saved.PreferenceID)
	}
	if repo.saved.BusinessStatus != domain.BusinessStatusPending {
		t.Errorf("business_status esperado PENDING, got %s", repo.saved.BusinessStatus)
	}
	if repo.saved.SagaStatus != domain.SagaStatusAwaitingPayment {
		t.Errorf("saga_status esperado AWAITING_PAYMENT, got %s", repo.saved.SagaStatus)
	}
	if repo.saved.CorrelationID != "corr-001" {
		t.Errorf("correlation_id incorreto: got %s", repo.saved.CorrelationID)
	}
	if !broker.checkoutCreatedCalled {
		t.Error("esperava PublishPaymentCheckoutCreated ser chamado")
	}
}

func TestProcessPaymentRequest_GatewayError(t *testing.T) {
	gateway := &mockGateway{
		createFn: func(_ context.Context, _ ports.CreatePreferenceRequest) (ports.CreatePreferenceResponse, error) {
			return ports.CreatePreferenceResponse{}, errors.New("mp timeout")
		},
	}
	broker := &mockBroker{}
	repo := &mockRepo{}
	svc := services.NewPaymentService(gateway, broker, repo)

	err := svc.ProcessPaymentRequest(context.Background(), makeEvent("order-002"))
	if err == nil {
		t.Fatal("esperava erro quando gateway falha")
	}
	if broker.checkoutCreatedCalled {
		t.Error("não deveria publicar evento quando gateway falha")
	}
}

func TestProcessWebhook_Idempotency_AlreadyApproved(t *testing.T) {
	gateway := &mockGateway{}
	broker := &mockBroker{}
	repo := &mockRepo{
		findByPaymentID: domain.Payment{
			PaymentID:      "pay-999",
			BusinessStatus: domain.BusinessStatusApproved,
		},
		findByPaymentIDErr: nil,
	}
	svc := services.NewPaymentService(gateway, broker, repo)

	err := svc.ProcessWebhook(context.Background(), "pay-999")
	if err != nil {
		t.Fatalf("esperava nil para idempotência, got: %v", err)
	}
	if broker.approvedCalled || broker.failedCalled {
		t.Error("não deveria publicar eventos para pagamento já em estado final")
	}
}

func TestProcessWebhook_Idempotency_AlreadyFailed(t *testing.T) {
	gateway := &mockGateway{}
	broker := &mockBroker{}
	repo := &mockRepo{
		findByPaymentID: domain.Payment{
			PaymentID:      "pay-888",
			BusinessStatus: domain.BusinessStatusFailed,
		},
	}
	svc := services.NewPaymentService(gateway, broker, repo)

	err := svc.ProcessWebhook(context.Background(), "pay-888")
	if err != nil {
		t.Fatalf("esperava nil, got: %v", err)
	}
	if broker.approvedCalled || broker.failedCalled {
		t.Error("não deveria publicar eventos para pagamento já em estado final")
	}
}

func TestProcessWebhook_Approved(t *testing.T) {
	gateway := &mockGateway{
		getStatusFn: func(_ context.Context, _ string) (domain.PaymentStatus, float64, error) {
			return domain.StatusApproved, 820.0, nil
		},
	}
	broker := &mockBroker{}
	repo := &mockRepo{
		findByPaymentIDErr: errors.New("não encontrado"),
		findByOrderID: domain.Payment{
			OrderID:           "order-003",
			PreferenceID:      "pref-456",
			TransactionAmount: 850.0,
			Currency:          "BRL",
		},
	}
	svc := services.NewPaymentService(gateway, broker, repo)

	err := svc.ProcessWebhook(context.Background(), "pay-777")
	if err != nil {
		t.Fatalf("esperava nil, got: %v", err)
	}
	if !broker.approvedCalled {
		t.Error("esperava PublishPaymentApproved ser chamado")
	}
	if broker.failedCalled {
		t.Error("não deveria chamar PublishPaymentFailed para pagamento aprovado")
	}
}

func TestProcessWebhook_Failed(t *testing.T) {
	gateway := &mockGateway{
		getStatusFn: func(_ context.Context, _ string) (domain.PaymentStatus, float64, error) {
			return domain.StatusFailed, 0, nil
		},
	}
	broker := &mockBroker{}
	repo := &mockRepo{
		findByPaymentIDErr: errors.New("não encontrado"),
	}
	svc := services.NewPaymentService(gateway, broker, repo)

	err := svc.ProcessWebhook(context.Background(), "pay-666")
	if err != nil {
		t.Fatalf("esperava nil, got: %v", err)
	}
	if !broker.failedCalled {
		t.Error("esperava PublishPaymentFailed ser chamado")
	}
}

func TestProcessWebhook_Pending_NoEvent(t *testing.T) {
	gateway := &mockGateway{
		getStatusFn: func(_ context.Context, _ string) (domain.PaymentStatus, float64, error) {
			return domain.StatusPending, 0, nil
		},
	}
	broker := &mockBroker{}
	repo := &mockRepo{
		findByPaymentIDErr: errors.New("não encontrado"),
	}
	svc := services.NewPaymentService(gateway, broker, repo)

	err := svc.ProcessWebhook(context.Background(), "pay-555")
	if err != nil {
		t.Fatalf("esperava nil para status pending, got: %v", err)
	}
	if broker.approvedCalled || broker.failedCalled {
		t.Error("não deveria publicar eventos para status pending")
	}
}
