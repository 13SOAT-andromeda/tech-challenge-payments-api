## Why

Quando o Mercado Pago envia um webhook `topic_merchant_order_wh` para um pagamento rejeitado (sem pagamento aprovado), o sistema registra apenas um aviso e retorna HTTP 200 sem publicar nenhum evento no SNS. Isso impede que outros serviços (ex: orders-api) saibam que o pagamento falhou e tomem as ações compensatórias necessárias na saga.

## What Changes

- `GetMerchantOrderPaymentID` no port `PaymentGateway` passa a retornar um struct `MerchantOrderResult` contendo `OrderID` (ExternalReference) e `PaymentID`, para que o `OrderID` esteja disponível mesmo quando `ErrNoApprovedPayment` é retornado
- Nova operação `PublishPaymentRejected` adicionada ao port `MessageBroker` e ao adapter SNS, publicando o evento `payment.rejected`
- Novo método `ProcessPaymentRejected(ctx, orderID)` no `PaymentService` que atualiza o status do pagamento e publica `payment.rejected`
- O webhook handler passa a chamar `ProcessPaymentRejected` quando `ErrNoApprovedPayment`, em vez de retornar 200 silenciosamente

## Capabilities

### New Capabilities

- `payment-rejected-publisher`: Publicação do evento `payment.rejected` no SNS quando o merchant order não possui pagamento aprovado, incluindo atualização do status do pagamento no banco de dados

### Modified Capabilities

- `sns-event-publisher`: Adição do novo tipo de evento `payment.rejected` ao publisher SNS existente
- `merchant-order-webhook`: Alteração do comportamento do handler para publicar `payment.rejected` ao receber `ErrNoApprovedPayment`, em vez de retornar silenciosamente

## Impact

- `internal/core/ports/payment_gateway.go`: assinatura de `GetMerchantOrderPaymentID` **BREAKING** (novo tipo de retorno `MerchantOrderResult`)
- `internal/core/ports/message_broker.go`: novo método `PublishPaymentRejected` na interface
- `internal/adapters/out/mercadopago/client.go`: implementação atualizada para retornar `MerchantOrderResult`
- `internal/adapters/out/sns/publisher.go`: novo método `PublishPaymentRejected`
- `internal/core/services/payment_service.go`: novo método `ProcessPaymentRejected`
- `internal/adapters/in/http/webhook_handler.go`: lógica do case `topic_merchant_order_wh` atualizada
