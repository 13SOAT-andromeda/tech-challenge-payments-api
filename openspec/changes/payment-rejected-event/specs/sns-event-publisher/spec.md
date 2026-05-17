## ADDED Requirements

### Requirement: payment.rejected é publicado ao rejeitar pagamento via merchant order
O sistema SHALL publicar o evento com `event_type = "payment.rejected"` quando o webhook `topic_merchant_order_wh` indicar que o pagamento foi recusado (nenhum pagamento com `status: approved` na merchant order).

#### Scenario: Evento payment.rejected publicado com campos obrigatórios
- **WHEN** `PublishPaymentRejected` é chamado com um `PaymentFailedEvent` válido
- **THEN** o publisher SHALL publicar no SNS uma mensagem com `event_type = "payment.rejected"` contendo `order_id`, `payment_id`, `preference_id`, `amount`, `currency`, `reason` e `failed_at`

#### Scenario: MessageAttribute identifica payment.rejected
- **WHEN** `PublishPaymentRejected` publica no SNS
- **THEN** a mensagem SHALL conter `MessageAttribute` `event_type` com valor `"payment.rejected"`
