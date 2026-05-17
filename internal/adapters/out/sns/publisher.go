package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/gedanmx/payments-api/internal/core/ports"
	"github.com/google/uuid"
)

type Publisher struct {
	client   *sns.Client
	topicARN string
}

func NewPublisher(client *sns.Client, topicARN string) *Publisher {
	return &Publisher{
		client:   client,
		topicARN: topicARN,
	}
}

type snsEvent struct {
	EventType string      `json:"event_type"`
	EventID   string      `json:"event_id"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

func (p *Publisher) publish(ctx context.Context, eventType string, payload interface{}) error {
	start := time.Now()
	slog.Info("sns.publish inicio", "event_type", eventType)

	envelope := snsEvent{
		EventType: eventType,
		EventID:   uuid.NewString(),
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		slog.Error("sns.publish erro ao serializar", "event_type", eventType, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("serializar evento %s: %w", eventType, err)
	}
	_, err = p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(p.topicARN),
		Message:  aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"event_type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(eventType),
			},
		},
	})
	if err != nil {
		slog.Error("sns.publish erro", "event_type", eventType, "error", err, "duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("publicar %s no SNS: %w", eventType, err)
	}

	slog.Info("sns.publish concluído", "event_type", eventType, "duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (p *Publisher) PublishPaymentCheckoutCreated(ctx context.Context, event ports.PaymentCheckoutCreatedEvent) error {
	payload := map[string]interface{}{
		"order_id":      event.OrderID,
		"payment_id":    event.PaymentID,
		"preference_id": event.PreferenceID,
		"checkout_url":  event.CheckoutURL,
		"expires_at":    event.ExpiresAt.UTC().Format(time.RFC3339),
		"status":        "PENDING",
	}
	return p.publish(ctx, "payment.checkout_created", payload)
}

func (p *Publisher) PublishPaymentApproved(ctx context.Context, event ports.PaymentApprovedEvent) error {
	payload := map[string]interface{}{
		"order_id":     event.OrderID,
		"payment_id":   event.PaymentID,
		"preference_id": event.PreferenceID,
		"amount":       event.Amount,
		"currency":     event.Currency,
		"approved_at":  event.ApprovedAt.UTC().Format(time.RFC3339),
		"saga_status":  "PAYMENT_CONFIRMED",
	}
	return p.publish(ctx, "payment.approved", payload)
}

func (p *Publisher) PublishPaymentFailed(ctx context.Context, event ports.PaymentFailedEvent) error {
	payload := map[string]interface{}{
		"order_id":     event.OrderID,
		"payment_id":   event.PaymentID,
		"preference_id": event.PreferenceID,
		"amount":       event.Amount,
		"currency":     event.Currency,
		"reason":       event.Reason,
		"failed_at":    event.FailedAt.UTC().Format(time.RFC3339),
		"saga_status":  "FAILED",
	}
	return p.publish(ctx, "payment.failed", payload)
}
