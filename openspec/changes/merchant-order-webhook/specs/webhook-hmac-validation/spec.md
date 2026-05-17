## ADDED Requirements

### Requirement: Handler roteia por tipo de webhook
O handler SHALL inspecionar o campo `type` do body para rotear entre `payment` e `topic_merchant_order_wh`, retornando HTTP 200 para tipos desconhecidos sem processamento.

#### Scenario: Tipo payment é roteado ao fluxo de pagamento
- **WHEN** uma notificação chega com `type: payment`
- **THEN** o handler SHALL usar `data.id` do body como `payment_id` e chamar `ProcessWebhook`

#### Scenario: Tipo topic_merchant_order_wh é roteado ao fluxo de merchant order
- **WHEN** uma notificação chega com `type: topic_merchant_order_wh`
- **THEN** o handler SHALL usar o campo `id` do body como merchant order ID e chamar `GetMerchantOrderPaymentID` antes de `ProcessWebhook`

#### Scenario: Tipo desconhecido é ignorado com HTTP 200
- **WHEN** uma notificação chega com `type` diferente de `payment` ou `topic_merchant_order_wh`
- **THEN** o handler SHALL retornar HTTP 200 sem processar
