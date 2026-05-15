## ADDED Requirements

### Requirement: Domínio Payment possui BusinessStatus e SagaStatus separados
A entidade `Payment` SHALL conter dois campos de status distintos: `BusinessStatus` para comunicação externa e `SagaStatus` para controle interno de orquestração Saga.

#### Scenario: Payment criado com status iniciais corretos
- **WHEN** um novo pagamento é criado a partir de `OrderCreated`
- **THEN** `BusinessStatus` SHALL ser `PENDING` e `SagaStatus` SHALL ser `AWAITING_PAYMENT`

#### Scenario: BusinessStatus não expõe detalhes de orquestração
- **WHEN** um evento de pagamento é publicado no SNS
- **THEN** o payload SHALL conter `status` mapeado de `BusinessStatus`, nunca `SagaStatus`

---

### Requirement: BusinessStatus segue ciclo de vida definido
O `BusinessStatus` SHALL seguir exclusivamente as transições: `PENDING → APPROVED` ou `PENDING → FAILED`.

#### Scenario: Transição para APPROVED
- **WHEN** o provedor confirma pagamento via webhook
- **THEN** `BusinessStatus` SHALL ser atualizado para `APPROVED`

#### Scenario: Transição para FAILED
- **WHEN** o provedor recusa o pagamento via webhook
- **THEN** `BusinessStatus` SHALL ser atualizado para `FAILED`

#### Scenario: Status final não pode ser alterado
- **WHEN** um webhook chega para um `Payment` com `BusinessStatus = APPROVED` ou `FAILED`
- **THEN** o sistema SHALL ignorar a transição e registrar log de auditoria

---

### Requirement: SagaStatus segue ciclo de vida de orquestração
O `SagaStatus` SHALL seguir as transições: `STARTED → AWAITING_PAYMENT → PAYMENT_CONFIRMED` ou `AWAITING_PAYMENT → FAILED`.

#### Scenario: SagaStatus inicia em AWAITING_PAYMENT ao criar pagamento
- **WHEN** o pagamento é criado e o checkout é gerado no provedor
- **THEN** `SagaStatus` SHALL ser `AWAITING_PAYMENT`

#### Scenario: SagaStatus atualiza para PAYMENT_CONFIRMED ao aprovar
- **WHEN** pagamento é aprovado via webhook
- **THEN** `SagaStatus` SHALL ser atualizado para `PAYMENT_CONFIRMED`

#### Scenario: SagaStatus atualiza para FAILED ao reprovar
- **WHEN** pagamento é recusado via webhook
- **THEN** `SagaStatus` SHALL ser atualizado para `FAILED`

---

### Requirement: Payment persiste CorrelationID e Provider
A entidade `Payment` SHALL armazenar `CorrelationID` (para rastreamento distribuído Saga) e `Provider` (identificador do gateway de pagamento).

#### Scenario: CorrelationID propagado do OrderCreated
- **WHEN** um `Payment` é criado a partir de `OrderCreated`
- **THEN** `CorrelationID` SHALL ser igual ao `correlationId` do evento de entrada

#### Scenario: Provider registrado conforme evento
- **WHEN** `OrderCreated` especifica `payment.provider = MERCADO_PAGO`
- **THEN** `Payment.Provider` SHALL ser persistido como `MERCADO_PAGO`
