## ADDED Requirements

### Requirement: MerchantOrderResult transporta OrderID junto com PaymentID
A interface `PaymentGateway` SHALL expor `GetMerchantOrderPaymentID` com retorno `(MerchantOrderResult, error)`, onde `MerchantOrderResult` contém os campos `PaymentID string` (ID do pagamento aprovado) e `OrderID string` (ExternalReference da merchant order).

#### Scenario: Merchant order com pagamento aprovado retorna PaymentID e OrderID
- **WHEN** `GetMerchantOrderPaymentID` é chamado e há pagamento com `status: approved`
- **THEN** o método SHALL retornar `MerchantOrderResult{PaymentID: "<id>", OrderID: "<externalRef>"}` e `nil` error

#### Scenario: Merchant order sem pagamento aprovado retorna OrderID com ErrNoApprovedPayment
- **WHEN** `GetMerchantOrderPaymentID` é chamado e nenhum pagamento tem `status: approved`
- **THEN** o método SHALL retornar `MerchantOrderResult{OrderID: "<externalRef>"}` junto com `ErrNoApprovedPayment`

#### Scenario: ExternalReference vazia é permitida mas logada
- **WHEN** a merchant order não possui `ExternalReference`
- **THEN** o método SHALL retornar `MerchantOrderResult{OrderID: ""}` sem erro adicional, e o caller SHALL logar aviso e encerrar sem publicar evento

---

### Requirement: PaymentService expõe ProcessPaymentRejected
O `PaymentService` SHALL implementar o método `ProcessPaymentRejected(ctx context.Context, orderID string) error` que atualiza o status do pagamento no banco e publica `payment.rejected` no SNS.

#### Scenario: Pagamento encontrado é atualizado e evento publicado
- **WHEN** `ProcessPaymentRejected` é chamado com um `orderID` existente no repositório
- **THEN** o service SHALL atualizar o pagamento para `BusinessStatus: FAILED`, `SagaStatus: FAILED` e publicar o evento `payment.rejected`

#### Scenario: Pagamento não encontrado publica evento com campos disponíveis
- **WHEN** `ProcessPaymentRejected` é chamado com `orderID` que não existe no repositório
- **THEN** o service SHALL logar aviso e publicar `payment.rejected` com o `order_id` fornecido e campos opcionais vazios

#### Scenario: Idempotência: status final não re-publica
- **WHEN** `ProcessPaymentRejected` é chamado para um pagamento já em status final (`APPROVED` ou `FAILED`)
- **THEN** o service SHALL retornar `nil` sem atualizar o banco ou publicar evento
