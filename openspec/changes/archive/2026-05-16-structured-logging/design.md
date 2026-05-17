## Context

O projeto usa Go 1.25 com `log/slog` disponível nativamente. O fluxo principal envolve: Consumer SQS → PaymentService → MercadoPago Client → GORM Repository → SNS Publisher. Cada etapa tem pontos de falha distintos e precisa de campos contextuais para correlacionar logs.

## Goals / Non-Goals

**Goals:**
- Logger JSON global via `slog.SetDefault` — sem passar logger por parâmetro em toda a codebase
- Campos contextuais padronizados: `order_id`, `correlation_id`, `payment_id`, `event_type`, `duration_ms`, `error`
- Log de entrada (campos do request) e saída (resultado/erro + duração) em cada etapa
- Nível `INFO` para fluxo normal, `WARN` para situações recuperáveis, `ERROR` para falhas

**Non-Goals:**
- Tracing distribuído (OpenTelemetry) — fora de escopo
- Log de corpo HTTP completo ou dados PII (email, valores) além do necessário para diagnóstico
- Rotação ou envio de logs — responsabilidade da infraestrutura (CloudWatch/stdout)
- Configuração de nível de log via env var — pode ser adicionado depois

## Decisions

### 1. `log/slog` com JSONHandler global

**Escolha:** `slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))` em `main.go`, usar `slog.Info/Warn/Error` com atributos nomeados no restante.

**Alternativas consideradas:**
- `zerolog` ou `zap`: mais performáticos, mas adicionam dependência externa desnecessária para este volume
- Logger por struct/injeção: mais testável mas exige refatoração invasiva em todas as structs

**Rationale:** `slog` stdlib elimina dependência externa e é suficiente para JSON estruturado; o logger global evita refatoração invasiva mantendo compatibilidade com o padrão atual.

### 2. Campos obrigatórios por camada

| Camada | Campos chave |
|---|---|
| SQS Consumer | `msg_id`, `event_type`, `order_id`, `correlation_id`, `duration_ms`, `error` |
| MercadoPago Client | `op` (CreatePreference/GetPaymentStatus), `order_id`, `preference_id`, `payment_id`, `duration_ms`, `error` |
| PaymentService | `op`, `order_id`, `correlation_id`, `payment_id`, `duration_ms`, `error` |
| SNS Publisher | `event_type`, `order_id`, `correlation_id`, `duration_ms`, `error` |
| Webhook Handler | `payment_id`, `mp_status`, `duration_ms`, `error` |

### 3. Duração via `time.Since`

Cada etapa captura `start := time.Now()` no início e loga `duration_ms: time.Since(start).Milliseconds()` na saída. Isso permite identificar gargalos sem tracing.

### 4. Sem propagação de logger via contexto

Usar `slog.Default()` implicitamente — todos os pacotes chamam `slog.Info/Warn/Error` sem receber logger. Simples e sem refatoração de assinaturas de função.

### 5. Não logar dados sensíveis

- **Não logar**: token de acesso, chave secreta, email completo do cliente, valor da transação
- **Logar**: `order_id`, `correlation_id`, `preference_id`, `payment_id`, status, durations, erros

## Risks / Trade-offs

- **[Risco] Performance:** `slog` JSONHandler aloca mais que `zap`. → Aceitável; volume de requisições é baixo para este serviço.
- **[Trade-off] Logger global:** Dificulta testes unitários que verificam logs. → Aceitável; testes existentes não verificam output de logs.
- **[Trade-off] Sem correlation via contexto:** Logs de etapas diferentes do mesmo pedido não são automaticamente correlacionados. → Mitigado pelos campos `order_id`/`correlation_id` em cada log.
