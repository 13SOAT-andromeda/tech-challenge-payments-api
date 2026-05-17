## 1. Inicialização do Logger

- [x] 1.1 Adicionar `slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))` como primeira instrução de `main()` em `cmd/api/main.go`
- [x] 1.2 Substituir todos os `log.Printf` / `log.Fatalf` de inicialização em `main.go` por `slog.Info` / `slog.Error` com atributos nomeados

## 2. Consumer SQS

- [x] 2.1 Substituir o log de início do consumer em `Start()` por `slog.Info("sqs consumer iniciado")`
- [x] 2.2 Adicionar `start := time.Now()` no início de `process()` e logar `msg_id` e `event_type` ao receber cada mensagem (`slog.Info`)
- [x] 2.3 Substituir o log de `event_type` desconhecido por `slog.Warn` com campos `msg_id` e `event_type`
- [x] 2.4 Substituir o log de erro em `poll()` por `slog.Error` com campo `error`
- [x] 2.5 Logar resultado de cada mensagem: sucesso com `order_id`, `correlation_id`, `duration_ms` (`slog.Info`); erro com `msg_id`, `error`, `duration_ms` (`slog.Error`)

## 3. PaymentService

- [x] 3.1 Adicionar `start := time.Now()` e log de entrada em `ProcessPaymentRequest`: `slog.Info` com `op`, `order_id`, `correlation_id`, `event_type`
- [x] 3.2 Logar saída de `ProcessPaymentRequest`: sucesso com `preference_id`, `checkout_url`, `duration_ms`; erro com `error`, `duration_ms`
- [x] 3.3 Adicionar `start := time.Now()` e log de entrada em `ProcessWebhook`: `slog.Info` com `op`, `payment_id`
- [x] 3.4 Logar saída de `ProcessWebhook`: sucesso com `mp_status`, `business_status`, `duration_ms`; erro com `error`, `duration_ms`
- [x] 3.5 Substituir os `log.Printf` existentes em `ProcessWebhook` (idempotência, status pendente, aviso de repositório) por `slog.Info` / `slog.Warn` com campos estruturados

## 4. MercadoPago Client

- [x] 4.1 Adicionar `start := time.Now()` e `slog.Info("mp.CreatePreference inicio", "order_id", req.OrderID)` em `CreatePreference`
- [x] 4.2 Logar saída de `CreatePreference`: sucesso com `preference_id`, `duration_ms`; erro com `error`, `duration_ms`
- [x] 4.3 Adicionar `start := time.Now()` e `slog.Info("mp.GetPaymentStatus inicio", "payment_id", paymentID)` em `GetPaymentStatus`
- [x] 4.4 Logar saída de `GetPaymentStatus`: sucesso com `mp_status`, `duration_ms`; erro com `error`, `duration_ms`

## 5. SNS Publisher

- [x] 5.1 Adicionar `start := time.Now()` no método `publish()` e logar `event_type`, `order_id`, `correlation_id` antes de publicar
- [x] 5.2 Logar saída de `publish()`: sucesso com `duration_ms`; erro com `error`, `duration_ms` em nível ERROR

## 6. Webhook Handler

- [x] 6.1 Adicionar `start := time.Now()` no início de `Handle()` e logar recebimento com `payment_id` (de `data.id` query param) e `mp_type`
- [x] 6.2 Substituir log de `x-signature ausente` e `assinatura inválida` por `slog.Warn` com campos `remote_addr` e `error`
- [x] 6.3 Logar saída de `Handle()`: sucesso com `payment_id`, `duration_ms`; erro com `payment_id`, `error`, `duration_ms`

## 7. Verificação

- [x] 7.1 Executar `go build ./...` e confirmar que compila sem erros
- [ ] 7.2 Enviar uma mensagem de teste na fila e confirmar que os logs aparecem em JSON com os campos esperados
