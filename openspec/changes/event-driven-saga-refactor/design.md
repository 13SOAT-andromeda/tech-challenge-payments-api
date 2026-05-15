## Context

A Payment API existe como serviço Go com clean architecture (portas + adaptadores). Atualmente publica eventos diretamente em filas SQS e consome mensagens em formato próprio (não envelopado por SNS). O domínio tem um único campo `Status PaymentStatus` que mistura estado de negócio e estado de orquestração. O banco de dados é SQLite, inadequado para produção em ambiente AWS.

O ecossistema de microsserviços usa o tópico SNS `payment.topic` como barramento central de eventos. Cada serviço tem uma fila SQS própria inscrita nesse tópico. A Payment API consome da fila `payment-order-events-queue` (inscrita no `payment.topic`) e publica eventos de volta no mesmo tópico `payment.topic`.

## Goals / Non-Goals

**Goals:**
- Consumir eventos do SNS `payment.topic` via fila SQS `payment-order-events-queue` com unwrap do envelope SNS
- Publicar `PaymentCheckoutCreated`, `PaymentApproved` e `PaymentFailed` no SNS `payment.topic`
- Separar `BusinessStatus` e `SagaStatus` no domínio para suporte a Saga externa
- Garantir idempotência no processamento de webhooks via `providerPaymentId`
- Migrar publisher de SQS direto para SNS
- Configurar tratamento de erros com DLQ no SQS consumer

**Non-Goals:**
- Implementar o orquestrador Saga (responsabilidade de outro serviço)
- Suporte a múltiplos provedores de pagamento simultaneamente nesta iteração
- Migração de dados históricos do SQLite para PostgreSQL
- Interface de administração ou dashboard

## Decisions

### 1. SNS + SQS para consumo (não SNS puro)

**Decisão**: A Payment API consome via SQS inscrita no SNS, não diretamente via SNS.

**Rationale**: SNS puro é push-based — se a API estiver indisponível, mensagens são perdidas. SQS provê buffer durável, retry automático via `VisibilityTimeout`, e Dead Letter Queue. Para pagamentos, perda de evento equivale a perda de receita.

**Alternativa descartada**: Lambda inscrita diretamente no SNS — fora do escopo (é serviço HTTP, não serverless).

---

### 2. Unwrap do envelope SNS no consumer

**Decisão**: O consumer SQS detecta e extrai o campo `Message` quando a mensagem tem `Type: "Notification"` (envelope SNS).

**Rationale**: Quando SNS publica em SQS, o body da mensagem SQS é um JSON envelope com `Type`, `MessageId`, `TopicArn`, `Message` (payload real como string JSON). O consumer deve fazer double-unmarshal: primeiro o envelope, depois o payload do evento.

```json
{
  "Type": "Notification",
  "TopicArn": "arn:aws:sns:...:payment.topic",
  "Message": "{\"event_type\":\"payment.requested\", \"payload\":{...}}"
}
```

---

### 3. Separação BusinessStatus / SagaStatus

**Decisão**: Dois campos distintos no domínio `Payment`.

| Campo | Visibilidade | Valores |
|---|---|---|
| `BusinessStatus` | Externo (API, eventos) | `PENDING`, `APPROVED`, `FAILED` |
| `SagaStatus` | Interno (orquestração) | `STARTED`, `AWAITING_PAYMENT`, `PAYMENT_CONFIRMED`, `FAILED` |

**Rationale**: Status de negócio é contrato com o cliente; status Saga é detalhe interno de orquestração. Misturá-los cria acoplamento com o orquestrador e vaza lógica de compensação para consumidores externos.

---

### 4. Publisher SNS (substituindo publisher SQS)

**Decisão**: Criar `internal/adapters/out/sns/publisher.go` usando `github.com/aws/aws-sdk-go-v2/service/sns`.

**Rationale**: SNS permite fanout — múltiplos serviços podem subscrever `payment.topic` sem a Payment API conhecê-los. Com SQS direto, cada novo consumidor exigiria mudança no publisher.

**Interface de porta** (sem mudança no nome, apenas nos métodos):
```go
type EventPublisher interface {
    PublishPaymentCheckoutCreated(ctx context.Context, event PaymentCheckoutCreatedEvent) error
    PublishPaymentApproved(ctx context.Context, event PaymentApprovedEvent) error
    PublishPaymentFailed(ctx context.Context, event PaymentFailedEvent) error
}
```

---

### 5. Idempotência por providerPaymentId + transição de status

**Decisão**: Antes de processar webhook, verificar se `providerPaymentId` já está registrado e se o `BusinessStatus` atual já é final (`APPROVED` ou `FAILED`). Se já final, ignorar silenciosamente.

**Rationale**: Webhooks de provedores de pagamento são entregues at-least-once. Reprocessar um pagamento já aprovado poderia duplicar eventos no SNS e inconsistências no estado.

---

### 6. PostgreSQL como banco de dados

**Decisão**: Migrar de SQLite para PostgreSQL.

**Rationale**: SQLite não suporta concorrência de escrita adequada para ambiente de produção com múltiplas instâncias. PostgreSQL é compatível com RDS/Aurora no ecossistema AWS.

**Driver**: `github.com/jackc/pgx/v5` (melhor suporte a features PostgreSQL que `lib/pq`).

---

### 7. CorrelationID propagado em todos os eventos

**Decisão**: `CorrelationID` é campo obrigatório no domínio `Payment` e em todos os eventos publicados.

**Rationale**: Permite rastreamento distribuído da transação Saga através de múltiplos serviços sem dependência de tracing centralizado.

## Risks / Trade-offs

- **[Risco] Envelope SNS pode mudar formato** → Mitigation: testar com `aws sns publish` local via LocalStack; validar campo `Type` antes de fazer unwrap
- **[Risco] DLQ sem monitoramento** → Mitigation: documentar ARN do DLQ e criar alarme CloudWatch (fora do escopo desta iteração, registrado como dívida técnica)
- **[Trade-off] PostgreSQL adiciona dependência de infraestrutura** → Mitigation: manter suporte a SQLite via interface de repositório para testes locais
- **[Risco] Idempotência baseada em status final pode ignorar webhooks legítimos de correção** → Mitigation: logar todos os webhooks ignorados com `providerPaymentId` e status atual para auditoria

## Migration Plan

1. Adicionar campos `BusinessStatus`, `SagaStatus`, `CorrelationID`, `Provider` no banco (migration SQL)
2. Deploy com feature flag de publisher SNS desabilitado — validar consumer funcionando
3. Habilitar publisher SNS — validar eventos chegando em `payment.topic`
4. Remover publisher SQS antigo e queues de `queuePaymentApproved` / `queuePaymentFailed` / `queueNotificationEmail`
5. Migrar banco para PostgreSQL em ambiente de staging antes de produção

**Rollback**: manter publisher SQS comentado por 1 sprint após migração; reverter env var `EVENT_PUBLISHER=sqs` se necessário.

## Open Questions

- O DLQ deve ser criado via Terraform no mesmo repositório de infra ou via CDK/SAM separado?
- O `expiresAt` do checkout Mercado Pago deve ser calculado como `now + 30min` (padrão do provider) ou configurável via env?
- O evento `PaymentCheckoutCreated` deve incluir o `checkoutUrl` completo ou apenas o `preferenceId` para que outros serviços gerem a URL?
