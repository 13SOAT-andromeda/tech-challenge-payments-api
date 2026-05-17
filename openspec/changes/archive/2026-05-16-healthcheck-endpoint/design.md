## Context

A aplicação já possui clientes para PostgreSQL (GORM), SQS e SNS inicializados em `main.go`. O healthcheck deve reutilizar esses clientes sem criar novas conexões, verificando cada dependência com uma operação leve e com timeout curto para não bloquear o endpoint.

## Goals / Non-Goals

**Goals:**
- Endpoint `GET /health` com verificações reais (não apenas "API respondeu")
- Resposta JSON com status por componente para facilitar diagnóstico
- Timeout por verificação para evitar que uma dependência lenta bloqueie o health check inteiro
- HTTP 200 quando tudo saudável, HTTP 503 quando qualquer check falha

**Non-Goals:**
- Autenticação no endpoint de healthcheck
- Métricas de latência ou histórico (isso é responsabilidade do sistema de monitoramento)
- Verificação de filas/tópicos além dos já configurados por variável de ambiente

## Decisions

### 1. Verificações por dependência

| Componente | Método de verificação |
|---|---|
| PostgreSQL | `db.DB().PingContext(ctx)` via GORM — operação leve, usa conexão existente do pool |
| SQS | `sqs.GetQueueAttributes` com `AttributeNames: ["ApproximateNumberOfMessages"]` — confirma acesso à fila |
| SNS | `sns.GetTopicAttributes` com o ARN do tópico — confirma acesso ao tópico |

### 2. Timeout por check

Cada verificação roda com `context.WithTimeout(ctx, 3s)`. Se o timeout estourar, o componente é marcado como `unhealthy` sem derrubar os outros checks. O timeout de 3s é suficientemente curto para não travar um readiness probe típico (que costuma ter timeout de 5–10s).

### 3. Checks em paralelo

As três verificações rodam em goroutines simultâneas com `sync.WaitGroup` e um canal para coletar resultados, minimizando a latência total do endpoint (dominada pelo check mais lento, não pela soma).

### 4. Estrutura da resposta

```json
{
  "status": "healthy" | "unhealthy",
  "checks": {
    "database": { "status": "healthy" },
    "sqs":      { "status": "unhealthy", "error": "connection refused" },
    "sns":      { "status": "healthy" }
  }
}
```

### 5. Handler recebe interfaces, não tipos concretos

`HealthHandler` recebe `*gorm.DB`, `*sqs.Client` e `*sns.Client` diretamente — esses são os tipos concretos já disponíveis em `main.go`. Não há necessidade de criar interfaces adicionais para esse caso de uso.

## Risks / Trade-offs

- **[Risco] Falso positivo no SNS/SQS:** `GetTopicAttributes`/`GetQueueAttributes` verifica acesso mas não garante que mensagens estão sendo entregues. → Aceitável para o objetivo de healthcheck de infraestrutura.
- **[Risco] Custo de chamadas AWS:** Cada requisição ao `/health` gera 2 chamadas à API AWS. → Mitigação: documentar que o endpoint não deve ser chamado com frequência inferior a 10s por monitoramento agressivo.
- **[Trade-off] Checks paralelos:** Aumenta complexidade do handler mas reduz latência p99 significativamente em ambientes com latência AWS variável.
