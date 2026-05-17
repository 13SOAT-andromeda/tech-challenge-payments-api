## MODIFIED Requirements

### Requirement: Handler processa webhook do tipo topic_merchant_order_wh
O endpoint `POST /webhooks/mercadopago` SHALL aceitar e processar notificações do tipo `topic_merchant_order_wh`, consultando a API de merchant orders do Mercado Pago para obter o `payment_id` associado e delegando ao fluxo adequado de processamento de pagamento (aprovação ou rejeição).

#### Scenario: Merchant order com pagamento aprovado é processada
- **WHEN** uma notificação `type: topic_merchant_order_wh` chega com `id` da merchant order
- **THEN** o handler SHALL consultar `/merchant_orders/{id}`, extrair o `payment_id` do primeiro pagamento com `status: approved` e chamar `ProcessWebhook(paymentID)`

#### Scenario: Merchant order sem pagamento aprovado publica payment.rejected
- **WHEN** uma notificação `type: topic_merchant_order_wh` chega e a merchant order não possui pagamentos com `status: approved`
- **THEN** o handler SHALL chamar `ProcessPaymentRejected(orderID)` com o `OrderID` retornado pelo gateway, logar aviso em nível WARN e retornar HTTP 200

#### Scenario: OrderID vazio na merchant order encerra sem publicar evento
- **WHEN** `GetMerchantOrderPaymentID` retorna `ErrNoApprovedPayment` e `MerchantOrderResult.OrderID` está vazio
- **THEN** o handler SHALL logar aviso e retornar HTTP 200 sem chamar `ProcessPaymentRejected`

#### Scenario: ID de merchant order inválido retorna 200
- **WHEN** o campo `id` do body não é um número inteiro válido
- **THEN** o handler SHALL retornar HTTP 200 e logar erro em nível ERROR
