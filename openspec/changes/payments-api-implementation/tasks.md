## 1. Alinhar event_type dos eventos SNS

- [x] 1.1 Em `internal/adapters/out/sns/publisher.go`, alterar a string `"PaymentCheckoutCreated"` para `"payment.checkout_created"` na chamada `publish()`
- [x] 1.2 Em `internal/adapters/out/sns/publisher.go`, alterar a string `"PaymentApproved"` para `"payment.approved"` na chamada `publish()`
- [x] 1.3 Em `internal/adapters/out/sns/publisher.go`, alterar a string `"PaymentFailed"` para `"payment.failed"` na chamada `publish()`
- [x] 1.4 Verificar que o `MessageAttribute` `event_type` também usa o novo nome snake.case em cada método de publicação

## 2. Implementar validação HMAC no webhook

- [x] 2.1 Adicionar campo `webhookSecret string` à struct `WebhookHandler` em `internal/adapters/in/http/webhook_handler.go`
- [x] 2.2 Atualizar `NewWebhookHandler` para receber `webhookSecret string` como parâmetro
- [x] 2.3 Implementar função `validateSignature(secret, xSignature, xRequestID, notificationID string) bool` que extrai `ts` e `v1` do header `x-signature`, monta o manifesto `id:<notificationID>;request-id:<xRequestID>;ts:<ts>` e compara com `hmac.Equal()` usando HMAC-SHA256
- [x] 2.4 Chamar `validateSignature` no início do método `Handle` antes de decodificar o body; retornar HTTP 400 se inválida
- [x] 2.5 Registrar log de aviso com IP de origem quando a assinatura for inválida

## 3. Atualizar configuração de inicialização

- [x] 3.1 Em `cmd/api/main.go`, adicionar `requireEnv("MERCADOPAGO_WEBHOOK_SECRET")` junto aos outros `requireEnv`
- [x] 3.2 Passar `os.Getenv("MERCADOPAGO_WEBHOOK_SECRET")` para `inhttp.NewWebhookHandler`
- [x] 3.3 Corrigir comentário do consumer SQS em `main.go`: `payment.topic` → `order.events`

## 4. Atualizar specs existentes

- [x] 4.1 Em `openspec/specs/order-event-consumer/spec.md`, alterar referência `payment.topic` para `order.events` no requisito do consumer
- [x] 4.2 Em `openspec/specs/sns-event-publisher/spec.md`, substituir o conteúdo pelo spec atualizado da change (tópico `payment.events`, event_types em snake.case)

## 5. Validação

- [x] 5.1 Executar `go build ./...` e garantir que não há erros de compilação
- [x] 5.2 Executar `go test ./...` e garantir que todos os testes passam
- [x] 5.3 Verificar manualmente que os event_types nos logs de publicação SNS aparecem como `payment.checkout_created`, `payment.approved`, `payment.failed`
