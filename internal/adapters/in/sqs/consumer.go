package sqs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gedanmx/payments-api/internal/core/services"
)

// errPermanent marks errors where retry will never succeed (malformed message, missing required fields).
var errPermanent = errors.New("permanent")

type Consumer struct {
	client   *sqs.Client
	queueURL string
	service  *services.PaymentService
}

func NewConsumer(client *sqs.Client, queueURL string, service *services.PaymentService) *Consumer {
	return &Consumer{
		client:   client,
		queueURL: queueURL,
		service:  service,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	slog.Info("sqs.consumer iniciado")
	for {
		select {
		case <-ctx.Done():
			slog.Info("sqs.consumer encerrado")
			return
		default:
			c.poll(ctx)
		}
	}
}

func (c *Consumer) poll(ctx context.Context) {
	out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		slog.Error("sqs.receive erro", "error", err)
		return
	}

	for _, msg := range out.Messages {
		msgID := *msg.MessageId
		start := time.Now()

		err := c.process(ctx, msgID, *msg.Body)
		if err != nil && !errors.Is(err, errPermanent) {
			slog.Error("sqs.message erro — mantendo na fila para retry",
				"msg_id", msgID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			continue
		}
		if err != nil {
			slog.Warn("sqs.message descartada — erro permanente",
				"msg_id", msgID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		} else {
			slog.Info("sqs.message processada com sucesso",
				"msg_id", msgID,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}

		_, err = c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(c.queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			slog.Error("sqs.delete erro", "msg_id", msgID, "error", err)
		}
	}
}

// snsEnvelope is the wrapper SNS adds when delivering to an SQS subscription
type snsEnvelope struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}

// unwrapSNS extracts the inner message when SNS delivers via SQS subscription.
// If the body is not an SNS notification envelope, it is returned unchanged.
func unwrapSNS(body string) string {
	var env snsEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		return body
	}
	if env.Type == "Notification" && env.Message != "" {
		return env.Message
	}
	return body
}

func (c *Consumer) process(ctx context.Context, msgID, body string) error {
	payload := unwrapSNS(body)

	var header struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal([]byte(payload), &header); err != nil {
		return fmt.Errorf("%w: parse event envelope: %v", errPermanent, err)
	}

	slog.Info("sqs.message recebida", "msg_id", msgID, "event_type", header.EventType)

	switch header.EventType {
	case "order.finished":
		var event services.PaymentRequestedEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return fmt.Errorf("%w: parse order.finished: %v", errPermanent, err)
		}
		if event.Payload.OrderID == "" {
			return fmt.Errorf("%w: order_id ausente", errPermanent)
		}
		return c.service.ProcessPaymentRequest(ctx, event)
	default:
		slog.Warn("sqs.event_type desconhecido — mensagem ignorada",
			"msg_id", msgID,
			"event_type", header.EventType,
		)
		return nil
	}
}
