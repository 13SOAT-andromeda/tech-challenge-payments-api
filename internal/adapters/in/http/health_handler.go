package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"gorm.io/gorm"
)

// HealthResponse is the JSON body returned by GET /health.
type HealthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]CheckResult `json:"checks"`
}

// CheckResult holds the status of a single dependency check.
type CheckResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthHandler verifica o status das dependências da aplicação.
type HealthHandler struct {
	db          *gorm.DB
	sqsClient   *awssqs.Client
	snsClient   *awssns.Client
	sqsQueueURL string
	snsTopicARN string
}

func NewHealthHandler(db *gorm.DB, sqsClient *awssqs.Client, snsClient *awssns.Client, sqsQueueURL, snsTopicARN string) *HealthHandler {
	return &HealthHandler{
		db:          db,
		sqsClient:   sqsClient,
		snsClient:   snsClient,
		sqsQueueURL: sqsQueueURL,
		snsTopicARN: snsTopicARN,
	}
}

func (h *HealthHandler) checkDatabase() CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sqlDB, err := h.db.DB()
	if err != nil {
		return CheckResult{Status: "unhealthy", Error: err.Error()}
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return CheckResult{Status: "unhealthy", Error: err.Error()}
	}
	return CheckResult{Status: "healthy"}
}

func (h *HealthHandler) checkSQS() CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := h.sqsClient.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{
		QueueUrl: &h.sqsQueueURL,
	})
	if err != nil {
		return CheckResult{Status: "unhealthy", Error: err.Error()}
	}
	return CheckResult{Status: "healthy"}
}

func (h *HealthHandler) checkSNS() CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := h.snsClient.GetTopicAttributes(ctx, &awssns.GetTopicAttributesInput{
		TopicArn: &h.snsTopicARN,
	})
	if err != nil {
		return CheckResult{Status: "unhealthy", Error: err.Error()}
	}
	return CheckResult{Status: "healthy"}
}

// Handle verifica a saúde da API e suas dependências.
//
// @Summary      Healthcheck
// @Description  Verifica o status da API, banco de dados PostgreSQL, fila SQS e tópico SNS.
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse  "Todos os componentes saudáveis"
// @Failure      503  {object}  HealthResponse  "Uma ou mais dependências indisponíveis"
// @Router       /health [get]
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	type namedResult struct {
		name   string
		result CheckResult
	}

	checks := []struct {
		name string
		fn   func() CheckResult
	}{
		{"database", h.checkDatabase},
		{"sqs", h.checkSQS},
		{"sns", h.checkSNS},
	}

	results := make(chan namedResult, len(checks))
	var wg sync.WaitGroup

	for _, c := range checks {
		wg.Add(1)
		go func(name string, fn func() CheckResult) {
			defer wg.Done()
			results <- namedResult{name: name, result: fn()}
		}(c.name, c.fn)
	}

	wg.Wait()
	close(results)

	resp := HealthResponse{
		Status: "healthy",
		Checks: make(map[string]CheckResult),
	}
	for r := range results {
		resp.Checks[r.name] = r.result
		if r.result.Status != "healthy" {
			resp.Status = "unhealthy"
		}
	}

	statusCode := http.StatusOK
	if resp.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
