## Why

Os logs atuais usam `log.Printf` com mensagens em texto livre, sem campos estruturados, tornando difícil correlacionar eventos entre etapas (consumo da fila → criação de preferência MP → persistência → publicação SNS) e filtrar por `order_id`, `correlation_id` ou `payment_id` em sistemas de observabilidade. A visibilidade insuficiente dificulta o diagnóstico de falhas em produção.

## What Changes

- Substituir todas as chamadas a `log.Printf` por logs estruturados em JSON usando `log/slog` (stdlib Go 1.21+)
- Adicionar log de entrada e saída em cada etapa do fluxo principal:
  - **Consumer SQS**: recebimento de mensagem, parse do evento, resultado do processamento
  - **Integração MercadoPago**: request para `CreatePreference` e `GetPaymentStatus` com campos relevantes e duração
  - **PaymentService**: início e fim de `ProcessPaymentRequest` e `ProcessWebhook` com `order_id`, `correlation_id`, `payment_id`
  - **SNS Publisher**: publicação de cada evento com `event_type` e resultado
  - **Webhook handler**: recebimento, validação de assinatura e resultado do processamento
- Inicializar o logger global com `slog.SetDefault` em `main.go` com handler JSON
- Propagar `logger` via contexto ou campos fixos de `correlation_id` onde disponível

## Capabilities

### New Capabilities

- `structured-logging`: Logger JSON estruturado (`log/slog`) com campos padronizados em todos os pontos de entrada e saída do fluxo de pagamento

### Modified Capabilities

- `order-event-consumer`: Adicionar logs estruturados de recebimento, parse, roteamento e resultado de cada mensagem SQS
- `webhook-hmac-validation`: Adicionar logs estruturados de recebimento e resultado do processamento do webhook

## Impact

- **Nenhuma dependência nova**: `log/slog` é stdlib desde Go 1.21
- **Arquivos alterados**: `cmd/api/main.go`, `internal/adapters/in/sqs/consumer.go`, `internal/adapters/in/http/webhook_handler.go`, `internal/adapters/out/mercadopago/client.go`, `internal/adapters/out/sns/publisher.go`, `internal/core/services/payment_service.go`
- **Formato de saída**: JSON lines — compatível com CloudWatch, Datadog, Loki e similares
- **Campos padronizados**: `time`, `level`, `msg`, `order_id`, `correlation_id`, `payment_id`, `event_type`, `duration_ms`, `error`
