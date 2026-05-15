## ADDED Requirements

### Requirement: Registrar transação ao criar preferência de pagamento
O repositório SHALL persistir um registro de transação no momento em que a preferência é criada no Mercado Pago, antes de publicar o evento `notification.email.requested`.

#### Scenario: Nova transação criada
- **WHEN** `PaymentRepository.Save` é invocado após `CreatePreference` retornar com sucesso
- **THEN** o repositório SHALL persistir o registro com: `order_id`, `preference_id`, `transaction_amount` (valor bruto do pedido), `status: PENDING_CUSTOMER_ACTION` e `created_at`
- **THEN** os campos `payment_id`, `net_amount` e `updated_at` SHALL permanecer nulos/zerados até o webhook ser recebido

### Requirement: Atualizar transação ao receber webhook do Mercado Pago
Quando o webhook do Mercado Pago for processado, o repositório SHALL atualizar o registro existente com os dados de pagamento retornados pela Payment API.

#### Scenario: Pagamento aprovado — atualização completa
- **WHEN** `PaymentRepository.UpdatePayment` é invocado após `GetPaymentStatus` retornar `approved`
- **THEN** o repositório SHALL atualizar o registro com: `payment_id`, `net_amount` (valor líquido após taxas do MP), `status: APPROVED` e `updated_at`

#### Scenario: Pagamento rejeitado ou cancelado — atualização de status
- **WHEN** `PaymentRepository.UpdatePayment` é invocado após `GetPaymentStatus` retornar `rejected` ou `cancelled`
- **THEN** o repositório SHALL atualizar o registro com: `payment_id`, `status: FAILED` e `updated_at`
- **THEN** `net_amount` SHALL ser registrado como zero

#### Scenario: Registro não encontrado para o order_id
- **WHEN** `PaymentRepository.UpdatePayment` é invocado com `order_id` não presente no banco
- **THEN** o repositório SHALL retornar erro indicando registro não encontrado

### Requirement: Consultar transação por payment_id para garantia de idempotência
O repositório SHALL permitir buscar uma transação pelo `payment_id` para que o webhook handler verifique se o pagamento já foi processado antes de chamar a Payment API do Mercado Pago.

#### Scenario: Pagamento já em estado final
- **WHEN** `PaymentRepository.FindByPaymentID` é invocado com um `payment_id` cujo status já é `APPROVED` ou `FAILED`
- **THEN** o repositório SHALL retornar o registro com o status atual, permitindo que o serviço retorne imediatamente sem reprocessar

#### Scenario: Pagamento ainda não em estado final
- **WHEN** `PaymentRepository.FindByPaymentID` é invocado com um `payment_id` com status `PENDING_CUSTOMER_ACTION`
- **THEN** o repositório SHALL retornar o registro indicando que o processamento pode continuar

#### Scenario: payment_id não encontrado no banco
- **WHEN** `PaymentRepository.FindByPaymentID` é invocado com um `payment_id` ausente
- **THEN** o repositório SHALL retornar erro de "não encontrado", e o serviço SHALL prosseguir com a consulta ao MP normalmente
