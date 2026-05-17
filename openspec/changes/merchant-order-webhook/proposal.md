## Why

O Mercado Pago envia webhooks do tipo `topic_merchant_order_wh` quando uma merchant order é atualizada. O handler atual só processa `type: "payment"`, ignorando esses eventos — o que impede que pagamentos realizados via checkout pro sejam confirmados.

## What Changes

- Suporte ao tipo `topic_merchant_order_wh` no webhook handler
- Nova chamada à API de merchant orders do Mercado Pago para obter o `payment_id` associado à order
- Reutilização do fluxo existente de `ProcessWebhook(paymentID)` após obter o ID

## Capabilities

### New Capabilities

- `merchant-order-webhook`: Suporte ao tipo `topic_merchant_order_wh` no handler de webhook, incluindo consulta à API de merchant orders para extrair o payment_id e processar o status do pagamento.

### Modified Capabilities

- `webhook-hmac-validation`: Handler passa a aceitar e rotear o tipo `topic_merchant_order_wh` além do `payment`.

## Impact

- `internal/adapters/in/http/webhook_handler.go`: adiciona branch para `topic_merchant_order_wh`
- `internal/adapters/out/mercadopago/client.go`: novo método `GetMerchantOrder` que consulta `/merchant_orders/{id}`
- `internal/core/ports/payment_gateway.go`: nova assinatura `GetMerchantOrder` na interface `PaymentGateway`
- Dependência: SDK do Mercado Pago (`mercadopago/sdk-go`) — verificar se expõe merchant orders
