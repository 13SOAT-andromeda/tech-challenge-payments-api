package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"

	inhttp "github.com/gedanmx/payments-api/internal/adapters/in/http"
	insqs "github.com/gedanmx/payments-api/internal/adapters/in/sqs"
	"github.com/gedanmx/payments-api/internal/adapters/out/database"
	mercadopago "github.com/gedanmx/payments-api/internal/adapters/out/mercadopago"
	outsns "github.com/gedanmx/payments-api/internal/adapters/out/sns"
	"github.com/gedanmx/payments-api/internal/core/services"
)

func main() {
	_ = godotenv.Load()

	requireEnv("MERCADOPAGO_ACCESS_TOKEN")
	requireEnv("MERCADOPAGO_PUBLIC_KEY")
	requireEnv("MERCADOPAGO_WEBHOOK_SECRET")
	requireEnv("SQS_QUEUE_URL_ORDER_EVENTS")
	requireEnv("SNS_TOPIC_ARN_PAYMENT")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Banco de dados
	dsn := getEnv("DATABASE_URL", "payments.db")
	repo, err := database.NewSQLiteRepository(dsn)
	if err != nil {
		log.Fatalf("inicializar banco de dados: %v", err)
	}

	// Cliente Mercado Pago
	mpClient := mercadopago.NewClient(os.Getenv("MERCADOPAGO_ACCESS_TOKEN"))

	// AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(getEnv("AWS_REGION", "us-east-1")))
	if err != nil {
		log.Fatalf("carregar config AWS: %v", err)
	}

	// Publisher SNS → payment.topic
	snsClient := sns.NewFromConfig(awsCfg)
	publisher := outsns.NewPublisher(snsClient, os.Getenv("SNS_TOPIC_ARN_PAYMENT"))

	// Serviço de domínio
	paymentService := services.NewPaymentService(mpClient, publisher, repo)

	// Consumer SQS ← payment-order-events-queue (inscrita no order.events via SNS)
	sqsClient := sqs.NewFromConfig(awsCfg)
	consumer := insqs.NewConsumer(sqsClient, os.Getenv("SQS_QUEUE_URL_ORDER_EVENTS"), paymentService)
	go consumer.Start(ctx)

	// Servidor HTTP
	webhookHandler := inhttp.NewWebhookHandler(paymentService, os.Getenv("MERCADOPAGO_WEBHOOK_SECRET"))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/mercadopago", webhookHandler.Handle)

	server := &http.Server{
		Addr:    ":" + getEnv("PORT", "8080"),
		Handler: mux,
	}

	go func() {
		log.Printf("servidor HTTP iniciado na porta %s", getEnv("PORT", "8080"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("encerrando serviço...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("erro no shutdown do servidor HTTP: %v", err)
	}

	log.Println("serviço encerrado")
}

func requireEnv(key string) {
	if os.Getenv(key) == "" {
		log.Fatalf("variável de ambiente obrigatória não definida: %s", key)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
