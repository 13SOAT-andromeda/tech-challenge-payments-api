## ADDED Requirements

### Requirement: Publicar notification.email.requested após criar preferência
Após criar a preferência no Mercado Pago, o serviço SHALL publicar o evento `notification.email.requested` no SQS com o link de checkout.

#### Scenario: Publicação bem-sucedida do link de checkout
- **WHEN** `MessageBroker.PublishEmailRequested` é invocado com `order_id`, `customer_email`, `checkout_url` e `preference_id` válidos
- **THEN** o broker SHALL publicar na fila SQS configurada uma mensagem JSON com `event_type: "notification.email.requested"` e o payload completo conforme contrato da RFC

#### Scenario: Falha na publicação SQS
- **WHEN** a publicação no SQS falha por erro de rede ou permissão
- **THEN** `PublishEmailRequested` SHALL retornar erro para o serviço de domínio tratar (NACK da mensagem `payment.requested`)

### Requirement: Publicar payment.approved após webhook com status aprovado
Quando o status do pagamento retornado pelo MP for `approved`, o serviço SHALL publicar o evento `payment.approved`.

#### Scenario: Pagamento aprovado
- **WHEN** `PaymentGateway.GetPaymentStatus` retorna `approved` para um `payment_id`
- **THEN** `MessageBroker.PublishPaymentApproved` SHALL publicar na fila SQS o evento `payment.approved` com `order_id`, `payment_id`, `preference_id`, `amount`, `currency` e `approved_at`

### Requirement: Publicar payment.failed após webhook com status de falha
Quando o status retornado for `rejected` ou `cancelled`, o serviço SHALL publicar o evento `payment.failed`.

#### Scenario: Pagamento rejeitado
- **WHEN** `PaymentGateway.GetPaymentStatus` retorna `rejected` para um `payment_id`
- **THEN** `MessageBroker.PublishPaymentFailed` SHALL publicar o evento `payment.failed` com `reason` e `failed_at`

#### Scenario: Pagamento cancelado
- **WHEN** `PaymentGateway.GetPaymentStatus` retorna `cancelled` para um `payment_id`
- **THEN** `MessageBroker.PublishPaymentFailed` SHALL publicar o evento `payment.failed` com `reason: "cancelled"` e `failed_at`

#### Scenario: Status pending recebido no webhook
- **WHEN** `PaymentGateway.GetPaymentStatus` retorna `pending`
- **THEN** o serviço SHALL registrar o status localmente sem publicar evento SAGA de resultado (aguardar novo webhook)

### Requirement: Formato de eventos conforme contrato da RFC
Todos os eventos publicados SHALL seguir a estrutura: `event_type`, `event_id` (UUID v4), `timestamp` (RFC3339) e `payload` com os campos definidos na RFC.

#### Scenario: Geração de event_id único
- **WHEN** qualquer evento é publicado
- **THEN** o campo `event_id` SHALL ser um UUID v4 gerado no momento da publicação, único por evento
