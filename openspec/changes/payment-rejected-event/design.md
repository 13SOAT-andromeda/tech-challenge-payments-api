## Context

O webhook `topic_merchant_order_wh` do Mercado Pago é enviado quando o estado de uma merchant order muda. O handler atual consulta `GetMerchantOrderPaymentID` para obter o ID do pagamento aprovado. Quando não há pagamento aprovado (`ErrNoApprovedPayment`), o sistema retorna HTTP 200 silenciosamente — sem publicar evento, sem atualizar o banco de dados.

O problema: outros serviços da saga (ex: `orders-api`) aguardam um evento de resultado do pagamento. A ausência de evento bloqueia a compensação da saga e deixa o pedido em estado pendente indefinidamente.

Estado atual do fluxo para rejeição:
```
webhook → GetMerchantOrderPaymentID → ErrNoApprovedPayment → log WARN → HTTP 200 (evento perdido)
```

## Goals / Non-Goals

**Goals:**
- Publicar `payment.rejected` no SNS quando o merchant order webhook chegar sem pagamento aprovado
- Atualizar o status do pagamento no banco para `FAILED` / `SagaStatus: FAILED`
- Preservar idempotência (não re-publicar se já estiver em status final)

**Non-Goals:**
- Alterar o fluxo de aprovação (`payment.approved`)
- Alterar o fluxo de falha técnica (`payment.failed`)
- Cobrir casos de timeout ou expiração de checkout

## Decisions

### 1. Alterar retorno de `GetMerchantOrderPaymentID` para struct `MerchantOrderResult`

**Decisão**: Mudar a assinatura de `(string, error)` para `(MerchantOrderResult, error)`, onde `MerchantOrderResult` contém `PaymentID string` e `OrderID string` (ExternalReference da merchant order).

**Rationale**: Quando `ErrNoApprovedPayment` é retornado, o caller precisa do `OrderID` para encontrar o pagamento no banco e publicar o evento de rejeição. Sem isso, seria necessária uma segunda chamada à API do MP ou uma outra abordagem de lookup. Retornar o struct mesmo em caso de erro segue o padrão Go de retornar informação parcial junto com o erro quando útil.

**Alternativas descartadas**:
- Novo método `GetMerchantOrderOrderID`: evitaria a mudança de assinatura, mas adicionaria uma segunda chamada à API MP ou repetiria lógica duplicada no cliente.
- Lookup pelo banco via `merchantOrderID`: não há campo `merchantOrderID` na tabela de pagamentos; o vínculo é via `ExternalReference` (orderID) → preference → merchant order.

### 2. Evento separado `payment.rejected` vs reutilizar `payment.failed`

**Decisão**: Adicionar `PublishPaymentRejected` no `MessageBroker` com event_type `"payment.rejected"`, reutilizando o struct `PaymentFailedEvent`.

**Rationale**: `payment.failed` é publicado para falhas técnicas (MP retornou status `rejected`/`cancelled` via webhook de pagamento direto). `payment.rejected` é semânticamente distinto: o comprador tentou pagar e o pagamento foi recusado pelo emissor/MP. Consumidores podem querer tratar esses casos de forma diferente (ex: notificar o usuário para tentar outro cartão). Reutilizar o mesmo struct `PaymentFailedEvent` evita duplicação de dados.

### 3. Novo método `ProcessPaymentRejected` no service

**Decisão**: Adicionar `ProcessPaymentRejected(ctx context.Context, orderID string) error` ao `PaymentService`.

**Rationale**: Isola a lógica de rejeição no service layer (onde vive `ProcessWebhook`), mantendo o webhook handler fino. O handler continua responsável apenas por decodificar o payload e rotear.

## Risks / Trade-offs

- **`ExternalReference` vazia**: Se uma merchant order foi criada fora deste serviço (ex: via dashboard MP), `ExternalReference` pode ser vazia. Mitigação: logar erro e retornar HTTP 200 sem publicar evento.
- **Pagamento não encontrado no banco**: Se o `orderID` retornado não existir no repositório (ex: race condition na criação), o service deve logar warn e publicar o evento com `order_id` mesmo assim, usando campos vazios para `preference_id` e `amount`.
- **Mudança BREAKING em `PaymentGateway`**: A interface muda — o teste `consumer_test.go` e qualquer mock precisam ser atualizados.
