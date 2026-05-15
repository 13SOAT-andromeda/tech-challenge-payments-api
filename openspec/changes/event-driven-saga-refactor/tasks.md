## 1. Domínio e Modelo de Dados

- [x] 1.1 Refatorar `internal/core/domain/payment.go`: adicionar `BusinessStatus`, `SagaStatus`, `CorrelationID`, `Provider`, `ExpiresAt`; remover `PaymentStatus` unificado
- [x] 1.2 Definir constantes de `BusinessStatus` (`PENDING`, `APPROVED`, `FAILED`) e `SagaStatus` (`STARTED`, `AWAITING_PAYMENT`, `PAYMENT_CONFIRMED`, `FAILED`) com método `IsFinal()` em `BusinessStatus`
- [ ] 1.3 Criar migration SQL para PostgreSQL: adicionar colunas `business_status`, `saga_status`, `correlation_id`, `provider`, `expires_at` na tabela `payments`

## 2. Eventos (Structs Go)

- [x] 2.1 Definir structs de eventos: `PaymentCheckoutCreatedEvent`, `PaymentApprovedEvent`, `PaymentFailedEvent` em `internal/core/ports/message_broker.go`; `PaymentRequestedEvent` em `internal/core/services/payment_service.go`
- [x] 2.2 Definir envelope SNS genérico `snsEnvelope{Type, Message string}` em `internal/adapters/in/sqs/consumer.go` e envelope de publicação `snsEvent` em `internal/adapters/out/sns/publisher.go`
- [x] 2.3 Adicionar `CorrelationID` em `PaymentApprovedEvent` e `PaymentFailedEvent`; adicionar `SagaStatus` no payload publicado

## 3. Portas (Interfaces)

- [x] 3.1 Atualizar `internal/core/ports/message_broker.go`: substituir `PublishEmailRequested` por `PublishPaymentCheckoutCreated`; manter `PublishPaymentApproved` e `PublishPaymentFailed`
- [x] 3.2 Atualizar `internal/core/ports/payment_repository.go`: adicionar `BusinessStatus` e `SagaStatus` na assinatura de `UpdatePayment`

## 4. Adaptador SNS Publisher

- [x] 4.1 Adicionar dependência `github.com/aws/aws-sdk-go-v2/service/sns v1.39.17` no `go.mod`
- [x] 4.2 Criar `internal/adapters/out/sns/publisher.go` implementando `MessageBroker` com client SNS AWS SDK v2
- [x] 4.3 Implementar método `publish` interno com envelope `snsEvent`, `MessageAttribute` `event_type` e serialização JSON
- [x] 4.4 Implementar `PublishPaymentCheckoutCreated`, `PublishPaymentApproved`, `PublishPaymentFailed` publicando no tópico `payment.topic`

## 5. Consumer SQS com Unwrap SNS

- [x] 5.1 Adicionar função `unwrapSNS(body string) string` em `internal/adapters/in/sqs/consumer.go`: detectar `Type: "Notification"` e extrair campo `Message`
- [x] 5.2 Refatorar método `process` do consumer para chamar `unwrapSNS` antes de fazer unmarshal do evento
- [ ] 5.3 Implementar roteamento de eventos por `event_type`: rotear diferentes tipos de evento para handlers correspondentes
- [ ] 5.4 Validar campos obrigatórios (`orderId`, `correlationId`) e retornar erro sem delete se inválido

## 6. Handler de Evento de Pagamento (PaymentService)

- [x] 6.1 Atualizar `ProcessPaymentRequest` no `PaymentService` para receber `PaymentRequestedEvent` com `CorrelationID`
- [x] 6.2 Implementar criação do `Payment` com `BusinessStatus = PENDING`, `SagaStatus = AWAITING_PAYMENT`, `CorrelationID` e `Provider = MERCADO_PAGO`
- [x] 6.3 Calcular `expiresAt = now + 30min` e persistir no `Payment` junto com `CheckoutURL`
- [x] 6.4 Publicar `PaymentCheckoutCreated` via `MessageBroker` após checkout gerado com sucesso
- [x] 6.5 Tratar falha do provedor: retornar erro sem deletar mensagem SQS (retry via VisibilityTimeout)

## 7. Webhook Handler com Idempotência

- [x] 7.1 Atualizar `ProcessWebhook` no `PaymentService`: verificar `BusinessStatus.IsFinal()` para idempotência antes de processar
- [x] 7.2 Se status já final: retornar nil com log de auditoria contendo `payment_id` e `business_status` atual
- [x] 7.3 Ao aprovar: atualizar `BusinessStatus = APPROVED` e `SagaStatus = PAYMENT_CONFIRMED`; publicar `PaymentApproved` no `payment.topic`
- [x] 7.4 Ao reprovar: atualizar `BusinessStatus = FAILED` e `SagaStatus = FAILED`; publicar `PaymentFailed` no `payment.topic`
- [ ] 7.5 Retornar HTTP 500 se publicação SNS falhar (atualmente o webhook responde 200 imediatamente e processa em goroutine)

## 8. Repositório PostgreSQL

- [ ] 8.1 Adicionar dependência `github.com/jackc/pgx/v5` no `go.mod`
- [ ] 8.2 Criar `internal/adapters/out/database/postgres_repository.go` implementando `PaymentRepository` para PostgreSQL
- [ ] 8.3 Implementar `FindByPaymentID` no repositório PostgreSQL para suporte à idempotência
- [ ] 8.4 Atualizar `cmd/` para inicializar PostgreSQL via env var `DATABASE_URL` com fallback para SQLite em modo de desenvolvimento

## 9. Configuração e Wiring

- [x] 9.1 Atualizar variáveis de ambiente: `SNS_TOPIC_ARN_PAYMENT` (tópico `payment.topic`), `SQS_QUEUE_URL_ORDER_EVENTS` (fila `payment-order-events-queue`); remover variáveis SQS antigas de publicação
- [x] 9.2 Atualizar inicialização em `cmd/api/main.go` para instanciar `sns.Publisher` no lugar de `sqs.Publisher`
- [x] 9.3 Atualizar `cmd/api/main.go` para instanciar `sqs.Client` separado do `sns.Client` e passar `MessageBroker` ao `PaymentService`

## 10. Testes

- [x] 10.1 Atualizar `internal/core/services/payment_service_test.go`: substituir `mockBroker.PublishEmailRequested` por `PublishPaymentCheckoutCreated`; atualizar `mockRepo.UpdatePayment` com novos parâmetros
- [ ] 10.2 Adicionar testes unitários para `unwrapSNS`: casos com envelope SNS, sem envelope, JSON inválido
- [x] 10.3 Atualizar testes de idempotência: usar `BusinessStatus` no lugar de `Status` para verificar estado final
- [x] 10.4 Atualizar testes de transição de status: verificar `BusinessStatus` e `SagaStatus` salvos no repositório
