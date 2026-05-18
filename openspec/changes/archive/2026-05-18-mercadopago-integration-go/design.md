## Context

A `PaymentsAPI` é um novo microsserviço Go que integra um sistema de oficina mecânica com o Mercado Pago (Checkout Pro). Ele opera como participante da SAGA Coreografada, comunicando-se assincronamente via AWS SQS com os serviços `OrderAPI` e `NotificationAPI`. Não há codebase existente — este é um projeto greenfield seguindo Clean/Hexagonal Architecture.

Restrições conhecidas:
- Mercado Pago garante entrega *at-least-once* para webhooks — idempotência é obrigatória
- O handler do webhook deve responder em < 500ms ao Mercado Pago para evitar reenvios
- Credenciais do Mercado Pago são sensíveis e devem ser injetadas via variáveis de ambiente

## Goals / Non-Goals

**Goals:**
- Consumir evento `payment.requested` do SQS e criar preferência de pagamento no Mercado Pago
- Expor `POST /webhooks/mercadopago` com processamento idempotente
- Publicar eventos SAGA: `notification.email.requested`, `payment.approved`, `payment.failed`
- Configurar credenciais via arquivo `.env` com `MERCADOPAGO_PUBLIC_KEY` e `MERCADOPAGO_ACCESS_TOKEN`
- Estrutura de projeto com Clean/Hexagonal Architecture para isolamento de adaptadores externos
- Graceful shutdown capturando `SIGINT`/`SIGTERM`

**Non-Goals:**
- UI de pagamento ou frontend próprio (link é gerado pelo Mercado Pago — Checkout Pro hosted)
- Implementação dos serviços `OrderAPI` ou `NotificationAPI`
- Cobrança recorrente, split de pagamentos ou outros métodos além de Checkout Pro
- Autenticação/autorização do endpoint de webhook além da validação do payload do MP
- Relatórios ou dashboard de pagamentos

## Decisions

### 1. Clean/Hexagonal Architecture com separação `cmd/`, `internal/core/`, `internal/adapters/`

**Decisão:** Separar domain, ports e use cases em `internal/core/` e os adaptadores externos (SQS, Mercado Pago, HTTP) em `internal/adapters/`.

**Rationale:** Permite trocar o broker de mensagens (ex: SQS → Kafka) ou o gateway de pagamento sem alterar a lógica de negócio. Facilita testes unitários mockando as interfaces (ports).

**Alternativa considerada:** Estrutura flat com pacotes por tipo (`handlers/`, `services/`, `models/`) — descartada por criar acoplamento entre lógica e infraestrutura.

### 2. Consumer SQS como goroutine dedicada com `context.Context` para graceful shutdown

**Decisão:** O consumer SQS roda em goroutine separada; o servidor HTTP roda concorrentemente. Ambos escutam um `ctx` cancelado ao receber `SIGINT`/`SIGTERM`.

**Rationale:** Atende ao requisito da RFC de operar consumer e webhook simultaneamente sem bloquear. Graceful shutdown garante que mensagens em processamento sejam concluídas antes do encerramento.

**Alternativa considerada:** Dois binários separados (um para HTTP, outro para consumer) — descartada por complexidade de deploy desnecessária nesta fase.

### 3. Idempotência via repositório local com verificação de estado final antes de reprocessar

**Decisão:** Antes de processar um webhook, consultar o repositório pelo `payment_id`. Se já em estado final (`approved`/`failed`), retornar `HTTP 200` imediatamente.

**Rationale:** O Mercado Pago reenvia webhooks em falha (*at-least-once*). Sem idempotência, pagamentos aprovados poderiam gerar múltiplos eventos `payment.approved` na SAGA.

**Alternativa considerada:** Idempotência baseada em TTL com Redis — descartada por adicionar dependência de infraestrutura desnecessária; estado local no banco já é suficiente.

### 4. Cliente HTTP nativo `net/http` para Mercado Pago

**Decisão:** Usar `net/http` com `http.Client` configurado com timeout de 10s, sem SDK externo do Mercado Pago.

**Rationale:** Reduz dependências externas, controle total sobre headers e retry logic. A API do MP é REST simples — uma SDK não adiciona valor significativo aqui.

**Alternativa considerada:** SDK oficial `github.com/mercadopago/sdk-go` — descartada por ser wrapper adicional sem benefício claro para os dois endpoints utilizados.

### 5. Arquivo `.env` carregado com `godotenv` no entrypoint

**Decisão:** Usar `github.com/joho/godotenv` para carregar `.env` no `cmd/api/main.go` antes da inicialização.

**Rationale:** Padrão consolidado em projetos Go para desenvolvimento local. Em produção, as variáveis são injetadas diretamente pelo ambiente (ECS, EKS, etc.) sem necessidade do `.env`.

### 6. Banco de dados para registro financeiro de transações

**Decisão:** Persistir duas fases da transação: (1) ao criar a preferência — registrar `order_id`, `preference_id`, `transaction_amount` e status `PENDING_CUSTOMER_ACTION`; (2) ao processar o webhook — atualizar com `payment_id`, `net_amount` (valor líquido após taxas do MP) e status final.

**Rationale:** O Mercado Pago cobra tarifas sobre o valor bruto — o valor líquido (`net_amount`) retornado na Payment API é diferente do valor cobrado do cliente. Persistir ambos permite reconciliação financeira (quanto foi cobrado vs. quanto entrou no caixa). Sem isso, essa informação se perde após o processamento do webhook.

Implementação: SQLite via `database/sql` + `github.com/mattn/go-sqlite3` para desenvolvimento. Variável `DATABASE_URL` define o driver em runtime, permitindo PostgreSQL em produção sem alterar a lógica.

## Risks / Trade-offs

- **[Risco] Webhook recebido antes do banco persistir a preferência** → Mitigação: Verificar se `preference_id` está presente antes de processar; caso não, retornar `HTTP 200` e aguardar reenvio do MP (que reprocessará em minutos).
- **[Risco] Falha na publicação SQS após consultar o MP no webhook** → Mitigação: Usar transação no repositório: só marcar estado final após publicação bem-sucedida no SQS. Em caso de falha, estado permanece pendente e o webhook será reenviado.
- **[Risco] Rate limiting da Preferences API do MP** → Mitigação: HTTP client com retry exponencial (máx 3 tentativas) + DLQ no SQS para `payment.requested` não processados.
- **[Trade-off] Processamento síncrono do webhook** → Para atingir < 500ms de resposta, o processamento crítico (consulta MP + publicação SQS) deve ser rápido. Em caso de lentidão, processar assincronamente via goroutine e responder 200 imediatamente, aceitando risco de perda em crash.

## Migration Plan

1. Provisionar filas SQS: `payment-requested`, `notification-email-requested`, `payment-approved`, `payment-failed` (+ DLQs para cada)
2. Configurar `.env` com credenciais reais do Mercado Pago (sandbox primeiro)
3. Configurar `notification_url` do MP apontando para o endpoint público `/webhooks/mercadopago` (usar ngrok em dev)
4. Deploy do serviço com variáveis de ambiente injetadas (sem `.env` em produção)
5. Validar fluxo completo em sandbox antes de ativar `live_mode`

**Rollback:** Como é serviço novo, rollback = desligar o serviço. Eventos não consumidos permanecem na fila SQS por até 14 dias.

## Open Questions

- Qual banco de dados será usado em produção — PostgreSQL (RDS) ou DynamoDB?
- O endpoint `/webhooks/mercadopago` precisa de validação de assinatura HMAC (X-Signature header)? O MP suporta isso na versão configurada?
- Qual o tempo de retenção e número de tentativas configurado nas filas SQS antes de mover para DLQ?
