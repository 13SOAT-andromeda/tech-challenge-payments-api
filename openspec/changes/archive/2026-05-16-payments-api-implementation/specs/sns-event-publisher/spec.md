## MODIFIED Requirements

### Requirement: Publisher publica eventos no SNS payment.events
O sistema SHALL publicar todos os eventos de pagamento no tópico SNS `payment.events` (env var `SNS_TOPIC_ARN_PAYMENT`) usando envelope padronizado com `event_type`, `event_id`, `timestamp` e `payload`.

#### Scenario: Evento publicado com envelope correto
- **WHEN** qualquer método de publicação é chamado com payload válido
- **THEN** o sistema SHALL publicar no SNS uma mensagem com body JSON contendo `event_type`, `event_id` (UUID único), `timestamp` (RFC3339 UTC) e `payload`

#### Scenario: Falha na publicação SNS retorna erro
- **WHEN** o AWS SDK retorna erro ao publicar no SNS
- **THEN** o publisher SHALL retornar o erro encapsulado com contexto do tipo de evento

---

### Requirement: payment.checkout_created é publicado após geração do checkout
O sistema SHALL publicar o evento com `event_type = "payment.checkout_created"` imediatamente após o provedor retornar a URL de checkout com sucesso.

#### Scenario: Evento publicado com todos os campos obrigatórios
- **WHEN** o checkout é gerado com sucesso pelo provedor
- **THEN** o sistema SHALL publicar `payment.checkout_created` com `order_id`, `payment_id`, `preference_id`, `checkout_url` e `expires_at`

---

### Requirement: payment.approved é publicado ao confirmar pagamento
O sistema SHALL publicar o evento com `event_type = "payment.approved"` quando o webhook indicar aprovação do pagamento.

#### Scenario: Evento payment.approved publicado com campos obrigatórios
- **WHEN** o webhook do provedor confirma aprovação e o status é atualizado
- **THEN** o sistema SHALL publicar `payment.approved` com `order_id`, `payment_id`, `preference_id`, `amount`, `currency` e `approved_at`

---

### Requirement: payment.failed é publicado ao reprovar pagamento
O sistema SHALL publicar o evento com `event_type = "payment.failed"` quando o webhook indicar recusa ou falha do pagamento.

#### Scenario: Evento payment.failed publicado com motivo
- **WHEN** o webhook do provedor indica recusa ou cancelamento
- **THEN** o sistema SHALL publicar `payment.failed` com `order_id`, `payment_id`, `preference_id`, `amount`, `currency`, `reason` e `failed_at`

---

### Requirement: MessageAttributes identificam o tipo de evento no SNS
O publisher SHALL incluir `MessageAttribute` com chave `event_type` em cada publicação SNS para permitir filtragem por subscription filter policy.

#### Scenario: MessageAttribute event_type presente em todas as publicações
- **WHEN** qualquer evento é publicado no SNS
- **THEN** a mensagem SHALL conter `MessageAttribute` `event_type` do tipo `String` com o nome do evento em snake.case (ex: `payment.checkout_created`)
