## 1. Scaffolding e Configuração do Projeto

- [x] 1.1 Inicializar módulo Go com `go mod init` e criar estrutura de diretórios (`cmd/api/`, `internal/core/domain/`, `internal/core/ports/`, `internal/core/services/`, `internal/adapters/in/http/`, `internal/adapters/in/sqs/`, `internal/adapters/out/mercadopago/`, `internal/adapters/out/sqs/`, `internal/adapters/out/database/`)
- [x] 1.2 Criar arquivo `.env` com entradas `MERCADOPAGO_PUBLIC_KEY=` e `MERCADOPAGO_ACCESS_TOKEN=` e adicionar `.env` ao `.gitignore`
- [x] 1.3 Adicionar dependências: `github.com/joho/godotenv`, `github.com/aws/aws-sdk-go-v2`, `github.com/google/uuid`, `github.com/mattn/go-sqlite3`
- [x] 1.4 Criar `cmd/api/main.go` com carregamento do `.env` via `godotenv.Load()`, validação das variáveis obrigatórias e ponto de entrada da aplicação

## 2. Domain e Ports (Camada Central)

- [x] 2.1 Definir entidade `Payment` em `internal/core/domain/payment.go` com campos: `ID`, `OrderID`, `PreferenceID`, `PaymentID`, `CheckoutURL`, `Status`, `Amount`, `Currency`, `CustomerEmail`, `CreatedAt`, `UpdatedAt`
- [x] 2.2 Definir tipo `PaymentStatus` com constantes: `PendingCustomerAction`, `Approved`, `Failed`, `Cancelled`, `Pending`
- [x] 2.3 Criar interface `PaymentGateway` em `internal/core/ports/payment_gateway.go` com `CreatePreference` e `GetPaymentStatus` conforme RFC
- [x] 2.4 Criar interface `MessageBroker` em `internal/core/ports/message_broker.go` com `PublishEmailRequested`, `PublishPaymentApproved`, `PublishPaymentFailed`
- [x] 2.5 Criar interface `PaymentRepository` em `internal/core/ports/payment_repository.go` com `Save`, `FindByOrderID`, `FindByPaymentID`, `UpdateStatus`

## 3. Casos de Uso (Services)

- [x] 3.1 Implementar `PaymentService.ProcessPaymentRequest` em `internal/core/services/payment_service.go`: chama `CreatePreference`, persiste `Payment` com status `PendingCustomerAction`, publica `notification.email.requested`
- [x] 3.2 Implementar `PaymentService.ProcessWebhook`: verifica idempotência via repositório, consulta `GetPaymentStatus`, atualiza status local, publica `payment.approved` ou `payment.failed` conforme status; ignora `pending`
- [x] 3.3 Adicionar lógica de idempotência em `ProcessWebhook`: retornar sem reprocessar se status já for `Approved` ou `Failed`

## 4. Adaptador: Cliente Mercado Pago

- [x] 4.1 Implementar `MercadoPagoClient` em `internal/adapters/out/mercadopago/client.go` com `http.Client` configurado com timeout de 10s, lendo `MERCADOPAGO_ACCESS_TOKEN` do ambiente
- [x] 4.2 Implementar método `CreatePreference`: montar body JSON conforme RFC (external_reference, items, payer, notification_url, back_urls, auto_return), fazer `POST /checkout/preferences`, parsear resposta e retornar `PreferenceID` e `CheckoutURL` (init_point)
- [x] 4.3 Implementar método `GetPaymentStatus`: fazer `GET /v1/payments/{id}`, parsear campo `status` e retornar `PaymentStatus` correspondente

## 5. Adaptador: Consumer SQS

- [x] 5.1 Implementar consumer em `internal/adapters/in/sqs/consumer.go` usando `aws-sdk-go-v2/service/sqs`: loop de poll com `ReceiveMessage`, desserializar payload para `PaymentRequestedEvent`, invocar `PaymentService.ProcessPaymentRequest`
- [x] 5.2 Implementar NACK (não deletar mensagem do SQS) em caso de erro no processamento, e `DeleteMessage` apenas em caso de sucesso
- [x] 5.3 Integrar `context.Context` no consumer para suporte a graceful shutdown via cancelamento de contexto

## 6. Adaptador: Publisher SQS

- [x] 6.1 Implementar `SQSPublisher` em `internal/adapters/out/sqs/publisher.go` com métodos `PublishEmailRequested`, `PublishPaymentApproved`, `PublishPaymentFailed`
- [x] 6.2 Cada método SHALL serializar o payload conforme contrato de eventos da RFC (com `event_type`, `event_id` UUID v4, `timestamp` RFC3339, `payload`)
- [x] 6.3 Configurar URLs das filas SQS via variáveis de ambiente: `SQS_QUEUE_PAYMENT_REQUESTED`, `SQS_QUEUE_NOTIFICATION_EMAIL`, `SQS_QUEUE_PAYMENT_APPROVED`, `SQS_QUEUE_PAYMENT_FAILED`

## 7. Adaptador: Banco de Dados (Repositório)

- [x] 7.1 Implementar `SQLiteRepository` em `internal/adapters/out/database/sqlite_repository.go` com `database/sql` + `go-sqlite3`
- [x] 7.2 Implementar schema SQL: tabela `payments` com colunas `id`, `order_id`, `preference_id`, `payment_id` (nullable), `transaction_amount`, `net_amount` (nullable), `status`, `created_at`, `updated_at` — índices em `order_id` e `payment_id`
- [x] 7.3 Implementar método `Save`: persiste a transação inicial com `order_id`, `preference_id`, `transaction_amount` e status `PENDING_CUSTOMER_ACTION`
- [x] 7.4 Implementar método `UpdatePayment`: atualiza `payment_id`, `net_amount`, `status` e `updated_at` ao processar o webhook do MP
- [x] 7.5 Implementar método `FindByPaymentID`: retorna o registro pelo `payment_id`; o serviço usa para verificar idempotência antes de chamar a Payment API do MP

## 8. Adaptador: Handler HTTP (Webhook)

- [x] 8.1 Implementar handler `WebhookHandler` em `internal/adapters/in/http/webhook_handler.go`: parsear payload do Mercado Pago, ignorar eventos com `type != "payment"`, extrair `data.id` e invocar `PaymentService.ProcessWebhook`
- [x] 8.2 Garantir resposta `HTTP 200 OK` para todos os casos válidos (incluindo pagamentos já processados e status `pending`)
- [x] 8.3 Retornar `HTTP 400 Bad Request` apenas para payload JSON inválido

## 9. Entrypoint e Inicialização

- [x] 9.1 Montar todas as dependências em `cmd/api/main.go`: instanciar repositório, cliente MP, publisher SQS, broker, serviço e handlers com injeção de dependência
- [x] 9.2 Registrar rota `POST /webhooks/mercadopago` no servidor HTTP (usar `net/http` padrão ou `chi`)
- [x] 9.3 Implementar graceful shutdown: capturar `SIGINT`/`SIGTERM`, cancelar contexto raiz, aguardar finalização do consumer SQS e do servidor HTTP com `http.Server.Shutdown`

## 10. Testes e Validação

- [x] 10.1 Escrever testes unitários para `PaymentService.ProcessPaymentRequest` e `ProcessWebhook` com mocks das interfaces `PaymentGateway`, `MessageBroker` e `PaymentRepository`
- [x] 10.2 Escrever testes para idempotência: verificar que `ProcessWebhook` retorna sem publicar evento quando status já é final
- [x] 10.3 Validar fluxo completo em sandbox do Mercado Pago: criar preferência, acessar `sandbox_init_point`, verificar recebimento do webhook e publicação dos eventos SQS
