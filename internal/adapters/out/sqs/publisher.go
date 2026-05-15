package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gedanmx/payments-api/internal/core/ports"
	"github.com/google/uuid"
)

type Publisher struct {
	client                   *sqs.Client
	queueNotificationEmail   string
	queuePaymentApproved     string
	queuePaymentFailed       string
}

func NewPublisher(client *sqs.Client, notificationEmailQueue, paymentApprovedQueue, paymentFailedQueue string) *Publisher {
	return &Publisher{
		client:                 client,
		queueNotificationEmail: notificationEmailQueue,
		queuePaymentApproved:   paymentApprovedQueue,
		queuePaymentFailed:     paymentFailedQueue,
	}
}

type sqsEvent struct {
	EventType string      `json:"event_type"`
	EventID   string      `json:"event_id"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

func (p *Publisher) PublishEmailRequested(ctx context.Context, orderID, email, checkoutURL, preferenceID string) error {
	payload := map[string]string{
		"order_id":          orderID,
		"customer_email":    email,
		"notification_type": "PAYMENT_LINK",
		"checkout_url":      checkoutURL,
		"preference_id":     preferenceID,
	}
	return p.publish(ctx, p.queueNotificationEmail, "notification.email.requested", payload)
}

func (p *Publisher) PublishPaymentApproved(ctx context.Context, event ports.PaymentApprovedEvent) error {
	payload := map[string]interface{}{
		"order_id":      event.OrderID,
		"payment_id":    event.PaymentID,
		"preference_id": event.PreferenceID,
		"amount":        event.Amount,
		"currency":      event.Currency,
		"approved_at":   event.ApprovedAt.Format(time.RFC3339),
	}
	return p.publish(ctx, p.queuePaymentApproved, "payment.approved", payload)
}

func (p *Publisher) PublishPaymentFailed(ctx context.Context, event ports.PaymentFailedEvent) error {
	payload := map[string]interface{}{
		"order_id":      event.OrderID,
		"payment_id":    event.PaymentID,
		"preference_id": event.PreferenceID,
		"amount":        event.Amount,
		"currency":      event.Currency,
		"reason":        event.Reason,
		"failed_at":     event.FailedAt.Format(time.RFC3339),
	}
	return p.publish(ctx, p.queuePaymentFailed, "payment.failed", payload)
}

func (p *Publisher) publish(ctx context.Context, queueURL, eventType string, payload interface{}) error {
	envelope := sqsEvent{
		EventType: eventType,
		EventID:   uuid.NewString(),
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("serializar evento %s: %w", eventType, err)
	}
	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}
