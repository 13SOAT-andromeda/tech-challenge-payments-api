package mercadopago

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	mpconfig "github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/merchantorder"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"

	"github.com/gedanmx/payments-api/internal/core/domain"
	"github.com/gedanmx/payments-api/internal/core/ports"
)


type Client struct {
	preferenceClient    preference.Client
	paymentClient       payment.Client
	merchantOrderClient merchantorder.Client
}

// NewClient builds a MercadoPago client. The environment (test vs production)
// is inferred from the token prefix: "TEST-" → sandbox, "APP_USR-" → production.
func NewClient(accessToken string) (*Client, error) {
	cfg, err := mpconfig.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("configurar SDK mercado pago: %w", err)
	}
	return &Client{
		preferenceClient:    preference.NewClient(cfg),
		paymentClient:       payment.NewClient(cfg),
		merchantOrderClient: merchantorder.NewClient(cfg),
	}, nil
}

func (c *Client) CreatePreference(ctx context.Context, req ports.CreatePreferenceRequest) (ports.CreatePreferenceResponse, error) {
	start := time.Now()
	slog.Info("mp.CreatePreference inicio", "order_id", req.OrderID)

	items := make([]preference.ItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = preference.ItemRequest{
			ID:         item.ID,
			Title:      item.Title,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			CurrencyID: item.CurrencyID,
		}
	}

	request := preference.Request{
		ExternalReference: req.OrderID,
		Items:             items,
		Payer:             &preference.PayerRequest{Email: req.CustomerEmail},
	}

	if req.BackURLs.Success != "" {
		request.BackURLs = &preference.BackURLsRequest{
			Success: req.BackURLs.Success,
			Failure: req.BackURLs.Failure,
			Pending: req.BackURLs.Pending,
		}
		request.AutoReturn = "approved"
	}

	resp, err := c.preferenceClient.Create(ctx, request)
	if err != nil {
		slog.Error("mp.CreatePreference erro", "order_id", req.OrderID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return ports.CreatePreferenceResponse{}, fmt.Errorf("criar preferência: %w", err)
	}

	checkoutURL := resp.InitPoint
	if os.Getenv("MERCADOPAGO_SANDBOX") == "true" {
		checkoutURL = resp.SandboxInitPoint
	}

	slog.Info("mp.CreatePreference concluído", "order_id", req.OrderID, "preference_id", resp.ID, "duration_ms", time.Since(start).Milliseconds())
	return ports.CreatePreferenceResponse{
		PreferenceID: resp.ID,
		CheckoutURL:  checkoutURL,
	}, nil
}

func (c *Client) GetPaymentStatus(ctx context.Context, paymentID string) (domain.PaymentStatus, float64, string, error) {
	start := time.Now()
	slog.Info("mp.GetPaymentStatus inicio", "payment_id", paymentID)

	id, err := strconv.Atoi(paymentID)
	if err != nil {
		slog.Error("mp.GetPaymentStatus payment_id inválido", "payment_id", paymentID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return "", 0, "", fmt.Errorf("payment_id inválido %q: %w", paymentID, err)
	}

	resp, err := c.paymentClient.Get(ctx, id)
	if err != nil {
		slog.Error("mp.GetPaymentStatus erro", "payment_id", paymentID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return "", 0, "", fmt.Errorf("consultar pagamento %s: %w", paymentID, err)
	}

	status := mapStatus(resp.Status)
	slog.Info("mp.GetPaymentStatus concluído", "payment_id", paymentID, "mp_status", status, "order_id", resp.ExternalReference, "duration_ms", time.Since(start).Milliseconds())
	return status, resp.NetAmount, resp.ExternalReference, nil
}

func (c *Client) GetMerchantOrderPaymentID(ctx context.Context, merchantOrderID string) (ports.MerchantOrderResult, error) {
	start := time.Now()
	slog.Info("mp.GetMerchantOrderPaymentID inicio", "merchant_order_id", merchantOrderID)

	id, err := strconv.Atoi(merchantOrderID)
	if err != nil {
		slog.Error("mp.GetMerchantOrderPaymentID id inválido", "merchant_order_id", merchantOrderID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return ports.MerchantOrderResult{}, fmt.Errorf("merchant_order_id inválido %q: %w", merchantOrderID, err)
	}

	resp, err := c.merchantOrderClient.Get(ctx, id)
	if err != nil {
		slog.Error("mp.GetMerchantOrderPaymentID erro", "merchant_order_id", merchantOrderID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return ports.MerchantOrderResult{}, fmt.Errorf("consultar merchant order %s: %w", merchantOrderID, err)
	}

	for _, p := range resp.Payments {
		if p.Status == "approved" {
			paymentID := strconv.Itoa(p.ID)
			slog.Info("mp.GetMerchantOrderPaymentID concluído", "merchant_order_id", merchantOrderID, "payment_id", paymentID, "duration_ms", time.Since(start).Milliseconds())
			return ports.MerchantOrderResult{PaymentID: paymentID, OrderID: resp.ExternalReference}, nil
		}
	}

	slog.Warn("mp.GetMerchantOrderPaymentID sem pagamento aprovado", "merchant_order_id", merchantOrderID, "payments_count", len(resp.Payments), "duration_ms", time.Since(start).Milliseconds())
	return ports.MerchantOrderResult{OrderID: resp.ExternalReference}, ports.ErrNoApprovedPayment
}

func mapStatus(s string) domain.PaymentStatus {
	switch s {
	case "approved":
		return domain.StatusApproved
	case "rejected":
		return domain.StatusFailed
	case "cancelled":
		return domain.StatusCancelled
	default:
		return domain.StatusPending
	}
}
