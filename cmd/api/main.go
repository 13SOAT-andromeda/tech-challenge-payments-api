// @title           Payments API
// @version         1.0
// @description     Serviço de pagamentos integrado com Mercado Pago via Checkout Pro.
// @host            localhost:8080
// @BasePath        /

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"

	_ "github.com/gedanmx/payments-api/docs"
	inhttp "github.com/gedanmx/payments-api/internal/adapters/in/http"
	insqs "github.com/gedanmx/payments-api/internal/adapters/in/sqs"
	"github.com/gedanmx/payments-api/internal/adapters/out/database"
	mercadopago "github.com/gedanmx/payments-api/internal/adapters/out/mercadopago"
	outsns "github.com/gedanmx/payments-api/internal/adapters/out/sns"
	"github.com/gedanmx/payments-api/internal/core/services"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	_ = godotenv.Load()

	requireEnv("MERCADOPAGO_ACCESS_TOKEN")
	requireEnv("MERCADOPAGO_WEBHOOK_SECRET")
	requireEnv("SQS_QUEUE_URL_ORDER_EVENTS")
	requireEnv("SNS_TOPIC_ARN_PAYMENT")
	requireEnv("DATABASE_URL")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	mpClient, err := mercadopago.NewClient(os.Getenv("MERCADOPAGO_ACCESS_TOKEN"))
	if err != nil {
		slog.Error("inicializar cliente mercado pago", "error", err)
		os.Exit(1)
	}

	repo, err := database.NewGORMRepository(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("inicializar banco de dados", "error", err)
		os.Exit(1)
	}

	awsOpts := []func(*config.LoadOptions) error{
		config.WithRegion(getEnv("AWS_REGION", "us-east-1")),
	}
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		awsOpts = append(awsOpts, config.WithBaseEndpoint(endpoint))
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, awsOpts...)
	if err != nil {
		slog.Error("carregar config AWS", "error", err)
		os.Exit(1)
	}

	snsClient := sns.NewFromConfig(awsCfg)
	publisher := outsns.NewPublisher(snsClient, os.Getenv("SNS_TOPIC_ARN_PAYMENT"))

	paymentService := services.NewPaymentService(mpClient, publisher, repo)

	sqsClient := sqs.NewFromConfig(awsCfg)
	consumer := insqs.NewConsumer(sqsClient, os.Getenv("SQS_QUEUE_URL_ORDER_EVENTS"), paymentService)
	go consumer.Start(ctx)

	webhookHandler := inhttp.NewWebhookHandler(paymentService, mpClient, os.Getenv("MERCADOPAGO_WEBHOOK_SECRET"))
	healthHandler := inhttp.NewHealthHandler(repo.DB(), sqsClient, snsClient, os.Getenv("SQS_QUEUE_URL_ORDER_EVENTS"), os.Getenv("SNS_TOPIC_ARN_PAYMENT"))
	router := inhttp.SetupRouter(webhookHandler, healthHandler)

	server := &http.Server{
		Addr:    ":" + getEnv("PORT", "8080"),
		Handler: router,
	}

	go func() {
		slog.Info("servidor HTTP iniciado", "port", getEnv("PORT", "8080"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("servidor HTTP encerrado inesperadamente", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("encerrando serviço")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("erro no shutdown do servidor HTTP", "error", err)
	}

	slog.Info("serviço encerrado")
}

func requireEnv(key string) {
	if os.Getenv(key) == "" {
		slog.Error("variável de ambiente obrigatória não definida", "key", key)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
