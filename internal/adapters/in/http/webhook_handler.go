package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gedanmx/payments-api/internal/core/services"
)

type WebhookHandler struct {
	service       *services.PaymentService
	webhookSecret string
}

func NewWebhookHandler(service *services.PaymentService, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{service: service, webhookSecret: webhookSecret}
}

type mpWebhookPayload struct {
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
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

	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s", notificationID, xRequestID, ts)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	received, err := hex.DecodeString(v1)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(hex.EncodeToString(received)))
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	xSignature := r.Header.Get("x-signature")
	xRequestID := r.Header.Get("x-request-id")

	if xSignature == "" {
		log.Printf("webhook rejeitado: x-signature ausente, ip=%s", r.RemoteAddr)
		http.Error(w, "assinatura ausente", http.StatusBadRequest)
		return
	}

	var payload mpWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	if !validateSignature(h.webhookSecret, xSignature, xRequestID, payload.Data.ID) {
		log.Printf("webhook rejeitado: assinatura inválida, ip=%s", r.RemoteAddr)
		http.Error(w, "assinatura inválida", http.StatusBadRequest)
		return
	}

	if payload.Type != "payment" {
		w.WriteHeader(http.StatusOK)
		return
	}

	paymentID := payload.Data.ID
	if paymentID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Responde 200 imediatamente para o MP e processa em background
	w.WriteHeader(http.StatusOK)

	go func() {
		if err := h.service.ProcessWebhook(r.Context(), paymentID); err != nil {
			log.Printf("erro ao processar webhook payment_id=%s: %v", paymentID, err)
		}
	}()
}
