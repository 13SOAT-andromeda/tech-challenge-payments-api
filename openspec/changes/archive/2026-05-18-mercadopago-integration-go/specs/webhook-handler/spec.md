## ADDED Requirements

### Requirement: Endpoint POST /webhooks/mercadopago
O serviço SHALL expor o endpoint `POST /webhooks/mercadopago` para receber notificações de status de pagamento enviadas pelo Mercado Pago.

#### Scenario: Webhook válido de pagamento recebido
- **WHEN** o Mercado Pago envia `POST /webhooks/mercadopago` com `type: "payment"` e `data.id` preenchido
- **THEN** o handler SHALL extrair o `payment_id` de `data.id` e invocar `PaymentService.ProcessWebhook`
- **THEN** o handler SHALL retornar `HTTP 200 OK` ao Mercado Pago em todos os casos de sucesso

#### Scenario: Payload com tipo diferente de "payment"
- **WHEN** o webhook recebido tem `type` diferente de `"payment"` (ex: `"merchant_order"`)
- **THEN** o handler SHALL retornar `HTTP 200 OK` sem processar

#### Scenario: Payload JSON inválido
- **WHEN** o corpo da requisição não é JSON válido
- **THEN** o handler SHALL retornar `HTTP 400 Bad Request` com mensagem de erro

### Requirement: Resposta rápida ao Mercado Pago
O handler SHALL responder ao Mercado Pago em menos de 500ms para evitar reenvios desnecessários.

#### Scenario: Processamento dentro do tempo limite
- **WHEN** `PaymentService.ProcessWebhook` completa dentro de 500ms
- **THEN** o handler SHALL retornar `HTTP 200` diretamente ao final do processamento

#### Scenario: Processamento excede 500ms
- **WHEN** o processamento pode exceder o limite de tempo
- **THEN** o handler SHALL responder `HTTP 200` imediatamente e processar de forma assíncrona via goroutine

### Requirement: Idempotência no processamento do webhook
O handler SHALL garantir que o mesmo evento de pagamento não seja processado mais de uma vez com resultado final.

#### Scenario: Webhook duplicado para pagamento já em estado final
- **WHEN** `ProcessWebhook` é invocado com `payment_id` de um pagamento já com status `approved` ou `failed`
- **THEN** o serviço SHALL retornar imediatamente sem consultar o Mercado Pago nem publicar novos eventos SAGA
- **THEN** o handler SHALL retornar `HTTP 200 OK`

#### Scenario: Primeiro processamento do webhook
- **WHEN** `ProcessWebhook` é invocado com `payment_id` ainda não em estado final
- **THEN** o serviço SHALL consultar `GET /v1/payments/{id}` no Mercado Pago, atualizar o repositório e publicar o evento SAGA correspondente
