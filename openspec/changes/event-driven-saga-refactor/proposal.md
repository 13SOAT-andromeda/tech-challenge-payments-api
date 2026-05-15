## Why

A Payment API atual publica eventos diretamente em filas SQS e não suporta o padrão Saga, misturando status de negócio com estado de orquestração em um único campo `Status`. Isso impede a integração com serviços externos via SNS e elimina a possibilidade de rastreamento de compensação distribuída em caso de falha.

## What Changes

- **BREAKING** Substituir publisher SQS por publisher SNS (tópico `payment.events`)
- **BREAKING** Separar `PaymentStatus` em `BusinessStatus` (externo) e `SagaStatus` (interno de orquestração)
- **BREAKING** Adicionar campos `CorrelationID` e `Provider` ao domínio `Payment`
- Adicionar unwrap do envelope SNS no SQS consumer (mensagens chegam via SNS → SQS)
- Substituir evento `PaymentRequestedEvent` por `OrderCreated` com schema alinhado ao contrato do serviço de pedidos
- Substituir `PublishEmailRequested` por `PublishPaymentCheckoutCreated` (evento Saga)
- Adicionar `PublishPaymentApproved` e `PublishPaymentFailed` com correlationId e sagaStatus
- Implementar idempotência no webhook handler baseada em `providerPaymentId` + `businessStatus` atual
- Migrar banco de SQLite para PostgreSQL (produção-ready)
- Adicionar DLQ no consumer SQS para mensagens que excedem retries

## Capabilities

### New Capabilities

- `order-event-consumer`: Consumo de `OrderCreated` via SQS inscrita no SNS `order.events`, com unwrap do envelope SNS e criação de pagamento no banco
- `sns-event-publisher`: Publicação de eventos no SNS `payment.events` (`PaymentCheckoutCreated`, `PaymentApproved`, `PaymentFailed`)
- `saga-status-tracking`: Separação de `BusinessStatus` e `SagaStatus` no domínio `Payment` para suporte a orquestração Saga externa
- `webhook-idempotency`: Idempotência no processamento de webhooks do provedor baseada em `providerPaymentId` e transição de status

### Modified Capabilities

- Nenhuma (as capabilities existentes são substituídas pelas novas — não há spec anterior para atualizar)

## Impact

- **`internal/core/domain/payment.go`**: adicionar `BusinessStatus`, `SagaStatus`, `CorrelationID`, `Provider`; remover `PaymentStatus` unificado
- **`internal/core/ports/message_broker.go`**: substituir interface `MessageBroker` com métodos alinhados aos novos eventos SNS
- **`internal/core/services/payment_service.go`**: adaptar para receber `OrderCreated` e produzir novos eventos
- **`internal/adapters/in/sqs/consumer.go`**: adicionar unwrap do envelope SNS (`{"Type":"Notification","Message":"..."}`)
- **`internal/adapters/out/sqs/publisher.go`**: substituir por `internal/adapters/out/sns/publisher.go` usando AWS SDK v2 SNS
- **`internal/adapters/out/database/sqlite_repository.go`**: migrar para PostgreSQL
- **`internal/adapters/in/http/webhook_handler.go`**: adicionar lógica de idempotência
- **`go.mod`**: adicionar `github.com/aws/aws-sdk-go-v2/service/sns` e driver PostgreSQL (`lib/pq` ou `pgx`)
