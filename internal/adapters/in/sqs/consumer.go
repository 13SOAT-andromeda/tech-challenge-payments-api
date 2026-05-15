package sqs

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gedanmx/payments-api/internal/core/services"
)

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
	log.Println("SQS consumer iniciado")
	for {
		select {
		case <-ctx.Done():
			log.Println("SQS consumer encerrado")
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
		log.Printf("erro ao receber mensagens SQS: %v", err)
		return
	}

	for _, msg := range out.Messages {
		if err := c.process(ctx, *msg.Body); err != nil {
			log.Printf("erro ao processar mensagem %s: %v — mantendo na fila para retry", *msg.MessageId, err)
			continue
		}

		_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(c.queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			log.Printf("erro ao deletar mensagem %s do SQS: %v", *msg.MessageId, err)
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

func (c *Consumer) process(ctx context.Context, body string) error {
	payload := unwrapSNS(body)

	var event services.PaymentRequestedEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return err
	}
	return c.service.ProcessPaymentRequest(ctx, event)
}
