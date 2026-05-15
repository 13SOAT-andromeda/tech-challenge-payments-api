## ADDED Requirements

### Requirement: Webhook handler verifica idempotência antes de processar
O endpoint de webhook SHALL verificar se o `providerPaymentId` já foi processado e se o `BusinessStatus` atual já é final (`APPROVED` ou `FAILED`) antes de aplicar qualquer atualização.

#### Scenario: Webhook duplicado com status final é ignorado silenciosamente
- **WHEN** um webhook chega para um `providerPaymentId` cujo `Payment` já tem `BusinessStatus = APPROVED` ou `FAILED`
- **THEN** o handler SHALL retornar HTTP 200 sem aplicar nenhuma alteração e SHALL registrar log de auditoria com `providerPaymentId` e status atual

#### Scenario: Primeiro webhook processado normalmente
- **WHEN** um webhook chega para um `providerPaymentId` ainda com `BusinessStatus = PENDING`
- **THEN** o handler SHALL processar normalmente, atualizar status e publicar evento SNS

---

### Requirement: Webhook de aprovação atualiza status e publica PaymentApproved
Ao receber confirmação de pagamento aprovado, o sistema SHALL atualizar ambos os status e publicar o evento correspondente de forma atômica.

#### Scenario: Aprovação processa atualização de status e publicação
- **WHEN** o payload do webhook indica pagamento aprovado
- **THEN** o sistema SHALL atualizar `BusinessStatus = APPROVED` e `SagaStatus = PAYMENT_CONFIRMED` e publicar `PaymentApproved` no SNS `payment.events`

#### Scenario: Falha na publicação SNS reverte atualização de status
- **WHEN** o status é atualizado mas a publicação SNS falha
- **THEN** o sistema SHALL retornar HTTP 500 para que o provedor reenvie o webhook (não confirmar recebimento)

---

### Requirement: Webhook de recusa atualiza status e publica PaymentFailed
Ao receber notificação de pagamento recusado, o sistema SHALL atualizar ambos os status e publicar o evento de falha.

#### Scenario: Recusa processa atualização de status e publicação
- **WHEN** o payload do webhook indica pagamento recusado
- **THEN** o sistema SHALL atualizar `BusinessStatus = FAILED` e `SagaStatus = FAILED` e publicar `PaymentFailed` no SNS `payment.events`

---

### Requirement: Webhook retorna HTTP 200 apenas após processamento completo
O endpoint SHALL retornar HTTP 200 somente após persistir as atualizações de status e publicar o evento SNS com sucesso, garantindo entrega at-least-once por parte do provedor.

#### Scenario: Resposta 200 após processamento bem-sucedido
- **WHEN** status atualizado e evento SNS publicado com sucesso
- **THEN** handler retorna HTTP 200 com body mínimo

#### Scenario: Resposta 500 em caso de falha parcial
- **WHEN** qualquer etapa do processamento falha (banco ou SNS)
- **THEN** handler retorna HTTP 500 para forçar reenvio pelo provedor
