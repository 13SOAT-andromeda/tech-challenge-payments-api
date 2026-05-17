package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gedanmx/payments-api/internal/core/ports"
	"github.com/gedanmx/payments-api/internal/core/services"
)

type WebhookHandler struct {
	service       *services.PaymentService
	gateway       ports.PaymentGateway
	webhookSecret string
}

func NewWebhookHandler(service *services.PaymentService, gateway ports.PaymentGateway, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{service: service, gateway: gateway, webhookSecret: webhookSecret}
}

// MPWebhookPayload is the notification body sent by Mercado Pago.
type MPWebhookPayload struct {
	ID   string `json:"id" example:"40928737633"`
	Type string `json:"type" example:"payment"`
	Data struct {
		ID string `json:"id" example:"123456789"`
	} `json:"data"`
}

// validateSignature verifies the x-signature header sent by Mercado Pago.
// The manifest format is: id:<notificationID>;request-id:<xRequestID>;ts:<ts>
func validateSignature(secret, xSignature, xRequestID, notificationID string) bool {
	var ts, v1 string
	for _, part := range strings.Split(xSignature, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "ts":
			ts = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}
	if ts == "" || v1 == "" {
		return false
	}

	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", notificationID, xRequestID, ts)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	received, err := hex.DecodeString(v1)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(hex.EncodeToString(received)))
}

// Handle processa notificações de pagamento recebidas do Mercado Pago.
//
// @Summary      Webhook Mercado Pago
// @Description  Recebe notificações de eventos de pagamento do Mercado Pago. Valida a assinatura HMAC-SHA256 antes de processar.
// @Tags         webhooks
// @Accept       json
// @Produce      plain
// @Param        x-signature    header    string             true   "Assinatura HMAC-SHA256 no formato ts=<ts>,v1=<hex>"
// @Param        x-request-id   header    string             false  "ID único da requisição enviado pelo Mercado Pago"
// @Param        payload        body      MPWebhookPayload   true   "Payload da notificação"
// @Success      200
// @Failure      400  {string}  string  "Assinatura ausente ou inválida, ou payload malformado"
// @Failure      500  {string}  string  "Erro ao processar o pagamento ou publicar evento"
// @Router       /webhooks/mercadopago [post]
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	xSignature := r.Header.Get("x-signature")
	xRequestID := r.Header.Get("x-request-id")
	dataIDFromURL := r.URL.Query().Get("data.id")

	if xSignature == "" {
		slog.Warn("webhook rejeitado: x-signature ausente", "remote_addr", r.RemoteAddr, "error", "x-signature ausente")
		http.Error(w, "assinatura ausente", http.StatusBadRequest)
		return
	}

	if !validateSignature(h.webhookSecret, xSignature, xRequestID, dataIDFromURL) {
		slog.Warn("webhook rejeitado: assinatura inválida", "remote_addr", r.RemoteAddr, "error", "assinatura inválida")
		http.Error(w, "assinatura inválida", http.StatusBadRequest)
		return
	}

	var payload MPWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	slog.Info("webhook.Handle recebido", "payment_id", dataIDFromURL, "mp_type", payload.Type)

	switch payload.Type {
	case "payment":
		paymentID := payload.Data.ID
		if paymentID == "" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if err := h.service.ProcessWebhook(r.Context(), paymentID); err != nil {
			slog.Error("webhook.Handle erro", "payment_id", paymentID, "error", err, "duration_ms", time.Since(start).Milliseconds())
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}
		slog.Info("webhook.Handle concluído", "payment_id", paymentID, "duration_ms", time.Since(start).Milliseconds())

	case "topic_merchant_order_wh":
		merchantOrderID := payload.ID
		result, err := h.gateway.GetMerchantOrderPaymentID(r.Context(), merchantOrderID)
		if err != nil {
			if errors.Is(err, ports.ErrNoApprovedPayment) {
				if result.OrderID == "" {
					slog.Warn("webhook.Handle merchant order sem pagamento aprovado e sem order_id", "merchant_order_id", merchantOrderID, "duration_ms", time.Since(start).Milliseconds())
					w.WriteHeader(http.StatusOK)
					return
				}
				slog.Warn("webhook.Handle merchant order sem pagamento aprovado — publicando rejeição", "merchant_order_id", merchantOrderID, "order_id", result.OrderID, "duration_ms", time.Since(start).Milliseconds())
				if err := h.service.ProcessPaymentRejected(r.Context(), result.OrderID); err != nil {
					slog.Error("webhook.Handle erro ao processar rejeição", "order_id", result.OrderID, "error", err, "duration_ms", time.Since(start).Milliseconds())
					http.Error(w, "erro interno", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				return
			}
			slog.Error("webhook.Handle erro ao consultar merchant order", "merchant_order_id", merchantOrderID, "error", err, "duration_ms", time.Since(start).Milliseconds())
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}
		if err := h.service.ProcessWebhook(r.Context(), result.PaymentID); err != nil {
			slog.Error("webhook.Handle erro", "payment_id", result.PaymentID, "merchant_order_id", merchantOrderID, "error", err, "duration_ms", time.Since(start).Milliseconds())
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}
		slog.Info("webhook.Handle concluído", "payment_id", result.PaymentID, "merchant_order_id", merchantOrderID, "duration_ms", time.Since(start).Milliseconds())

	default:
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}
