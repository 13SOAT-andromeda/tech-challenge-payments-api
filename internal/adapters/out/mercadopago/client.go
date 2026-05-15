package mercadopago

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gedanmx/payments-api/internal/core/domain"
	"github.com/gedanmx/payments-api/internal/core/ports"
)

const baseURL = "https://api.mercadopago.com"

type Client struct {
	httpClient  *http.Client
	accessToken string
}

func NewClient(accessToken string) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		accessToken: accessToken,
	}
}

func (c *Client) CreatePreference(ctx context.Context, req ports.CreatePreferenceRequest) (ports.CreatePreferenceResponse, error) {
	items := make([]preferenceItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = preferenceItem{
			ID:         item.ID,
			Title:      item.Title,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			CurrencyID: item.CurrencyID,
		}
	}

	body := preferenceRequest{
		ExternalReference: req.OrderID,
		Items:             items,
		Payer:             payer{Email: req.CustomerEmail},
		NotificationURL:   req.WebhookURL,
		BackURLs: backURLs{
			Success: req.BackURLs.Success,
			Failure: req.BackURLs.Failure,
			Pending: req.BackURLs.Pending,
		},
		AutoReturn: "approved",
	}

	data, err := json.Marshal(body)
	if err != nil {
		return ports.CreatePreferenceResponse{}, fmt.Errorf("serializar body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/checkout/preferences", bytes.NewReader(data))
	if err != nil {
		return ports.CreatePreferenceResponse{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ports.CreatePreferenceResponse{}, fmt.Errorf("chamada mercado pago: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return ports.CreatePreferenceResponse{}, fmt.Errorf("mercado pago retornou %d: %s", resp.StatusCode, respBody)
	}

	var result preferenceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return ports.CreatePreferenceResponse{}, fmt.Errorf("parsear resposta: %w", err)
	}

	return ports.CreatePreferenceResponse{
		PreferenceID: result.ID,
		CheckoutURL:  result.InitPoint,
	}, nil
}

func (c *Client) GetPaymentStatus(ctx context.Context, paymentID string) (domain.PaymentStatus, float64, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/payments/%s", baseURL, paymentID), nil)
	if err != nil {
		return "", 0, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("consultar pagamento %s: %w", paymentID, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return "", 0, fmt.Errorf("pagamento %s não encontrado no mercado pago", paymentID)
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("mercado pago retornou %d: %s", resp.StatusCode, respBody)
	}

	var result paymentResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", 0, fmt.Errorf("parsear resposta de pagamento: %w", err)
	}

	status := mapStatus(result.Status)
	return status, result.NetAmount, nil
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

type preferenceRequest struct {
	ExternalReference string          `json:"external_reference"`
	Items             []preferenceItem `json:"items"`
	Payer             payer           `json:"payer"`
	NotificationURL   string          `json:"notification_url"`
	BackURLs          backURLs        `json:"back_urls"`
	AutoReturn        string          `json:"auto_return"`
}

type preferenceItem struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	CurrencyID string  `json:"currency_id"`
}

type payer struct {
	Email string `json:"email"`
}

type backURLs struct {
	Success string `json:"success"`
	Failure string `json:"failure"`
	Pending string `json:"pending"`
}

type preferenceResponse struct {
	ID        string `json:"id"`
	InitPoint string `json:"init_point"`
}

type paymentResponse struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	NetAmount float64 `json:"net_amount"`
}
