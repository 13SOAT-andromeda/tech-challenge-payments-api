## ADDED Requirements

### Requirement: Handler processa webhook do tipo topic_merchant_order_wh
O endpoint `POST /webhooks/mercadopago` SHALL aceitar e processar notificações do tipo `topic_merchant_order_wh`, consultando a API de merchant orders do Mercado Pago para obter o `payment_id` associado e delegando ao fluxo existente de processamento de pagamento.

#### Scenario: Merchant order com pagamento aprovado é processada
- **WHEN** uma notificação `type: topic_merchant_order_wh` chega com `id` da merchant order
- **THEN** o handler SHALL consultar `/merchant_orders/{id}`, extrair o `payment_id` do primeiro pagamento com `status: approved` e chamar `ProcessWebhook(paymentID)`

#### Scenario: Merchant order sem pagamento aprovado é ignorada
- **WHEN** uma notificação `type: topic_merchant_order_wh` chega e a merchant order não possui pagamentos com `status: approved`
- **THEN** o handler SHALL retornar HTTP 200 sem processar e logar aviso em nível WARN

#### Scenario: ID de merchant order inválido retorna 200
- **WHEN** o campo `id` do body não é um número inteiro válido
- **THEN** o handler SHALL retornar HTTP 200 e logar erro em nível ERROR

---

### Requirement: PaymentGateway expõe consulta de merchant order
A interface `PaymentGateway` SHALL incluir o método `GetMerchantOrderPaymentID` para que o handler possa obter o `payment_id` a partir do ID de uma merchant order.

#### Scenario: Merchant order com pagamento aprovado retorna payment_id
- **WHEN** `GetMerchantOrderPaymentID(ctx, merchantOrderID)` é chamado e a merchant order possui pelo menos um pagamento com `status: approved`
- **THEN** o método SHALL retornar o ID do primeiro pagamento aprovado como string

#### Scenario: Merchant order sem pagamento aprovado retorna erro
- **WHEN** `GetMerchantOrderPaymentID(ctx, merchantOrderID)` é chamado e nenhum pagamento tem `status: approved`
- **THEN** o método SHALL retornar erro `ErrNoApprovedPayment`
