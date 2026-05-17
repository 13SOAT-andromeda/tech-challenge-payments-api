## Why

Sem um endpoint de healthcheck, não há como verificar programaticamente se a API está operacional, se o banco de dados está acessível e se a conexão com os recursos AWS (SQS e SNS) está funcional — dificultando diagnósticos, readiness probes de orquestradores e monitoramento.

## What Changes

- Adicionar endpoint `GET /health` que retorna o status individual de cada dependência: banco de dados (PostgreSQL), fila SQS (`payment-order-events-queue`) e tópico SNS (`payment-events`)
- O endpoint retorna `HTTP 200` quando todos os componentes estão saudáveis e `HTTP 503` quando qualquer dependência falha
- Resposta JSON estruturada com campo `status` global e `checks` detalhando cada componente
- Novo handler `HealthHandler` em `internal/adapters/in/http/health_handler.go`
- Rota `GET /health` registrada no mux em `cmd/api/main.go`

## Capabilities

### New Capabilities

- `healthcheck`: Endpoint `GET /health` que verifica e reporta o status da API, banco de dados PostgreSQL, fila SQS e tópico SNS com resposta JSON estruturada

### Modified Capabilities

<!-- Nenhuma capability existente sendo modificada -->

## Impact

- **Novo arquivo:** `internal/adapters/in/http/health_handler.go`
- **Alterado:** `cmd/api/main.go` — registrar rota e injetar dependências no handler
- **Dependências novas:** nenhuma — usa os clientes AWS SDK e GORM já presentes
- **Infraestrutura:** nenhuma mudança; verifica recursos AWS já existentes via `GetQueueAttributes` (SQS) e `GetTopicAttributes` (SNS)
