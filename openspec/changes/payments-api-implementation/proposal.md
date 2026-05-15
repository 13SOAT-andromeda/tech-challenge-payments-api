## Why

A RFC_PaymentsAPI.md define os contratos autoritativos de eventos, tópicos e comportamentos da PaymentsAPI. A implementação atual diverge da RFC em nomes de event_type (PascalCase no código vs snake.case na RFC), referências de tópicos SNS desatualizadas nos specs existentes, e ausência de validação HMAC no webhook — requisito de segurança explícito na RFC seções 7.3 e 9.

## What Changes

- Alinhar `event_type` dos eventos SNS publicados com a convenção RFC: `payment.checkout_created`, `payment.approved`, `payment.failed`
- Implementar validação de assinatura HMAC no endpoint `POST /webhooks/mercadopago` (rejeitar com HTTP 400 se inválida)
- Corrigir referência de tópico de entrada no spec `order-event-consumer`: `payment.topic` → `order.events`
- Corrigir referência de tópico de saída no spec `sns-event-publisher`: `payment.topic` → `payment.events` e alinhar nomes de event_type
- Atualizar comentário no `main.go` com nomenclatura correta dos tópicos

## Capabilities

### New Capabilities
- `webhook-hmac-validation`: Validação de assinatura HMAC-SHA256 no endpoint de webhook do Mercado Pago, usando a chave secreta configurada via variável de ambiente

### Modified Capabilities
- `order-event-consumer`: Atualização da referência do tópico SNS de subscrição de `payment.topic` para `order.events`
- `sns-event-publisher`: Atualização do tópico de publicação de `payment.topic` para `payment.events` e alinhamento dos nomes de `event_type` com a RFC

## Impact

- `internal/adapters/out/sns/publisher.go`: renomear strings de event_type nas chamadas `publish()`
- `internal/adapters/in/http/webhook_handler.go`: adicionar validação HMAC antes de processar o payload
- `cmd/api/main.go`: atualizar comentário de nomenclatura e adicionar `requireEnv("MERCADOPAGO_WEBHOOK_SECRET")`
- Specs `order-event-consumer` e `sns-event-publisher`: atualização de requisitos de tópico e convenção de nomes
