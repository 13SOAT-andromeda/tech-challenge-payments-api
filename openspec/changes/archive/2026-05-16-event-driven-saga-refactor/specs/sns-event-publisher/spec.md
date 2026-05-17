## ADDED Requirements

### Requirement: Publisher publica eventos no SNS payment.topic
O sistema SHALL publicar todos os eventos de pagamento no tópico SNS `payment.topic` (env var `SNS_TOPIC_ARN_PAYMENT`) usando envelope padronizado com `event_type`, `event_id`, `timestamp` e `payload`.

#### Scenario: Evento publicado com envelope correto
- **WHEN** qualquer método de publicação é chamado com payload válido
- **THEN** o sistema SHALL publicar no SNS uma mensagem com body JSON contendo `event_type`, `event_id` (UUID único), `timestamp` (RFC3339 UTC) e `payload`

#### Scenario: Falha na publicação SNS retorna erro
- **WHEN** o AWS SDK retorna erro ao publicar no SNS
- **THEN** o publisher SHALL retornar o erro encapsulado com contexto do tipo de evento

---

### Requirement: PaymentCheckoutCreated é publicado após geração do checkout
O sistema SHALL publicar o evento `PaymentCheckoutCreated` imediatamente após o provedor retornar a URL de checkout com sucesso.

#### Scenario: Evento publicado com todos os campos obrigatórios
- **WHEN** o checkout é gerado com sucesso pelo provedor
- **THEN** o sistema SHALL publicar `PaymentCheckoutCreated` com `correlationId`, `orderId`, `paymentId`, `checkoutUrl`, `expiresAt` e `status = PENDING`

---

### Requirement: PaymentApproved é publicado ao confirmar pagamento
O sistema SHALL publicar `PaymentApproved` no SNS quando o webhook indicar aprovação do pagamento.

#### Scenario: Evento PaymentApproved publicado com sagaStatus
- **WHEN** o webhook do provedor confirma aprovação e o status é atualizado
- **THEN** o sistema SHALL publicar `PaymentApproved` com `correlationId`, `orderId`, `paymentId`, `amount`, `approvedAt` e `sagaStatus = PAYMENT_CONFIRMED`

---

### Requirement: PaymentFailed é publicado ao reprovar pagamento
O sistema SHALL publicar `PaymentFailed` no SNS quando o webhook indicar recusa ou falha do pagamento.

#### Scenario: Evento PaymentFailed publicado com motivo
- **WHEN** o webhook do provedor indica recusa
- **THEN** o sistema SHALL publicar `PaymentFailed` com `correlationId`, `orderId`, `paymentId`, `reason` e `sagaStatus = FAILED`

---

### Requirement: MessageAttributes identificam o tipo de evento no SNS
O publisher SHALL incluir `MessageAttribute` com chave `event_type` em cada publicação SNS para permitir filtragem por subscription filter policy.

#### Scenario: MessageAttribute event_type presente em todas as publicações
- **WHEN** qualquer evento é publicado no SNS
- **THEN** a mensagem SHALL conter `MessageAttribute` `event_type` do tipo `String` com o nome do evento (ex: `PaymentCheckoutCreated`)
