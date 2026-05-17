## 1. Port PaymentGateway — MerchantOrderResult

- [x] 1.1 Adicionar struct `MerchantOrderResult` com campos `PaymentID string` e `OrderID string` em `internal/core/ports/payment_gateway.go`
- [x] 1.2 Alterar assinatura de `GetMerchantOrderPaymentID` na interface `PaymentGateway` de `(string, error)` para `(MerchantOrderResult, error)`

## 2. Adapter MercadoPago — retorno atualizado

- [x] 2.1 Em `internal/adapters/out/mercadopago/client.go`, atualizar `GetMerchantOrderPaymentID` para retornar `MerchantOrderResult{PaymentID: paymentID, OrderID: resp.ExternalReference}` quando pagamento aprovado for encontrado
- [x] 2.2 Atualizar o caminho de retorno de `ErrNoApprovedPayment` para retornar `MerchantOrderResult{OrderID: resp.ExternalReference}` junto com o erro

## 3. Port MessageBroker — PublishPaymentRejected

- [x] 3.1 Adicionar método `PublishPaymentRejected(ctx context.Context, event PaymentFailedEvent) error` à interface `MessageBroker` em `internal/core/ports/message_broker.go`

## 4. Adapter SNS — implementação de PublishPaymentRejected

- [x] 4.1 Implementar `PublishPaymentRejected` em `internal/adapters/out/sns/publisher.go` usando `p.publish(ctx, "payment.rejected", payload)` com o mesmo payload de `PublishPaymentFailed`

## 5. Service — ProcessPaymentRejected

- [x] 5.1 Adicionar método `ProcessPaymentRejected(ctx context.Context, orderID string) error` em `internal/core/services/payment_service.go`
- [x] 5.2 Em `ProcessPaymentRejected`: buscar pagamento por `orderID` com `repository.FindByOrderID`; se não encontrado, logar aviso e usar `domain.Payment{OrderID: orderID}`
- [x] 5.3 Em `ProcessPaymentRejected`: verificar idempotência — retornar `nil` se `payment.BusinessStatus.IsFinal()` for verdadeiro
- [x] 5.4 Em `ProcessPaymentRejected`: chamar `repository.UpdatePayment` com `BusinessStatus: FAILED`, `SagaStatus: FAILED` e `finalMPStatus: StatusFailed`
- [x] 5.5 Em `ProcessPaymentRejected`: chamar `broker.PublishPaymentRejected` com `PaymentFailedEvent{OrderID, PaymentID, PreferenceID, Amount, Currency, Reason: "rejected", FailedAt: now}`

## 6. Webhook Handler — roteamento de rejeição

- [x] 6.1 Em `internal/adapters/in/http/webhook_handler.go`, atualizar a chamada a `GetMerchantOrderPaymentID` para usar o novo tipo de retorno `MerchantOrderResult`
- [x] 6.2 No branch `ErrNoApprovedPayment`: verificar se `result.OrderID` está vazio — se sim, logar aviso e retornar HTTP 200 sem processar
- [x] 6.3 No branch `ErrNoApprovedPayment`: quando `result.OrderID` não vazio, chamar `h.service.ProcessPaymentRejected(ctx, result.OrderID)` e logar resultado

## 7. Compilação e verificação

- [x] 7.1 Executar `go build ./...` e confirmar compilação sem erros
- [ ] 7.2 Realizar checkout no sandbox, rejeitar o pagamento e confirmar que `payment.rejected` é publicado no SNS e o status no banco é atualizado para `FAILED`
