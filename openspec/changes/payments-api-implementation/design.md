## Context

A PaymentsAPI está funcional em termos de fluxo (consumer SQS → Mercado Pago → SNS → webhook), mas diverge da RFC em dois pontos concretos:

1. **Nomes de event_type**: o publisher SNS usa `PaymentCheckoutCreated`, `PaymentApproved`, `PaymentFailed` (PascalCase). A RFC define `payment.checkout_created`, `payment.approved`, `payment.failed` (snake.case com prefixo de domínio).
2. **HMAC webhook ausente**: o endpoint `POST /webhooks/mercadopago` não valida a assinatura `x-signature` enviada pelo Mercado Pago. A RFC (seções 7.3 e 9) exige rejeição com HTTP 400.

Os specs existentes (`order-event-consumer`, `sns-event-publisher`) ainda referenciam `payment.topic` onde deveriam referenciar `order.events` e `payment.events` respectivamente.

## Goals / Non-Goals

**Goals:**
- Alinhar `event_type` dos eventos SNS publicados com a RFC
- Implementar validação HMAC-SHA256 do header `x-signature` do Mercado Pago
- Corrigir referências de tópicos SNS nos specs afetados

**Non-Goals:**
- Alterar o fluxo de negócio ou as structs de domínio
- Migrar banco de dados
- Modificar o contrato de entrada (`payment.requested`)

## Decisions

### 1. Renomear event_type de PascalCase para snake.case

A mudança é aplicada **somente nas strings** passadas para `publisher.publish()` em `sns/publisher.go`. Não há alteração em structs Go nem em interfaces — apenas as constantes literais de `event_type` mudam.

**Impacto em subscriptions SNS**: se existirem filter policies baseadas em `event_type = "PaymentCheckoutCreated"`, precisam ser atualizadas para `payment.checkout_created`. Em ambiente de desenvolvimento/testes, isso não há impacto imediato; em produção, coordenar com consumidores upstream antes do deploy.

**Alternativa considerada**: manter PascalCase e atualizar a RFC — rejeitada porque a RFC é o contrato autoritativo e outros serviços (OrderAPI, NotificationAPI) já foram alinhados a ela.

### 2. Validação HMAC do webhook Mercado Pago

O Mercado Pago envia o header `x-signature` com formato `ts=<timestamp>,v1=<hmac_sha256>`.

O HMAC é calculado sobre a string `id:<notification_id>;request-id:<x-request-id>;ts:<timestamp>` usando HMAC-SHA256 com a chave secreta configurada no painel do MP (env var `MERCADOPAGO_WEBHOOK_SECRET`).

**Fluxo de validação**:
1. Extrair `ts` e `v1` do header `x-signature`
2. Montar a string de manifesto com os valores dos headers `x-request-id` e o campo `data.id` do body
3. Calcular HMAC-SHA256 com `MERCADOPAGO_WEBHOOK_SECRET`
4. Comparar via `hmac.Equal()` (constant-time) com `v1`
5. Rejeitar com HTTP 400 se inválido; prosseguir se válido

**Alternativa considerada**: validar apenas por IP allowlist — rejeitada por ser menos robusta e exigir manutenção de lista de IPs do MP.

### 3. Atualização dos specs (sem alteração de código)

Os specs `order-event-consumer` e `sns-event-publisher` referenciam `payment.topic` em textos descritivos. A correção é puramente documental nos arquivos `.md` — o código já usa os ARNs corretos via variáveis de ambiente.

## Risks / Trade-offs

- **Quebra de consumers no SNS por renome de event_type**: qualquer subscriber com filter policy em `event_type` deixará de receber mensagens. → Mitigação: coordenar deploy simultâneo com OrderAPI e NotificationAPI ou usar período de transição publicando os dois nomes.
- **HMAC secret ausente em ambiente local**: sem `MERCADOPAGO_WEBHOOK_SECRET`, todos os webhooks serão rejeitados. → Mitigação: em modo sandbox, usar um secret configurável; adicionar ao `.env.example`.
- **`x-request-id` ausente em testes manuais**: o Mercado Pago sempre envia esse header, mas ferramentas de teste (curl, Postman) podem omiti-lo. → Mitigação: tratar ausência do header como string vazia na montagem do manifesto (comportamento documentado pelo MP).

## Migration Plan

1. Atualizar `sns/publisher.go`: trocar strings de event_type
2. Atualizar `webhook_handler.go`: adicionar validação HMAC
3. Atualizar `main.go`: adicionar `requireEnv("MERCADOPAGO_WEBHOOK_SECRET")` e corrigir comentário
4. Atualizar specs `order-event-consumer` e `sns-event-publisher`
5. Deploy coordenado com serviços consumidores (se em produção)

**Rollback**: reverter os literais de event_type e remover a validação HMAC são mudanças isoladas e facilmente revertíveis sem impacto em banco ou infraestrutura.

## Open Questions

- Os serviços consumidores (OrderAPI, NotificationAPI) já esperam `payment.checkout_created` (snake.case) ou ainda consomem `PaymentCheckoutCreated` (PascalCase)?
- Há ambientes de staging onde a filter policy do SNS precisa ser atualizada antes do deploy?
