## Context

O servidor HTTP é inicializado em `cmd/api/main.go` usando `http.NewServeMux` com rotas registradas inline, junto à composição de todas as dependências da aplicação. Isso mistura responsabilidades no `main.go` e torna difícil adicionar middlewares por grupo de rota ou evoluir o roteamento sem mexer no bootstrap da aplicação.

## Goals / Non-Goals

**Goals:**
- Centralizar o roteamento em `internal/adapters/in/http/routes.go` via função `SetupRouter`
- Adotar Gin como framework HTTP, substituindo `http.NewServeMux`
- Manter os mesmos paths e verbos HTTP (`POST /webhooks/mercadopago`, `GET /health`, `/docs/`)
- Manter a Swagger UI funcionando via `ginSwagger`

**Non-Goals:**
- Adicionar novas rotas ou endpoints neste change
- Introduzir middlewares de autenticação ou rate-limiting
- Alterar a lógica dos handlers existentes (`webhook_handler.go`, `health_handler.go`)

## Decisions

### Gin como engine HTTP

**Decisão**: adotar `github.com/gin-gonic/gin`.

**Rationale**: Gin é o framework HTTP mais adotado no ecossistema Go, com suporte nativo a grupos de rotas, middlewares encadeados e integração oficial com Swaggo (`gin-swagger`). O `http.NewServeMux` da stdlib não oferece agrupamento nem middleware por grupo.

**Alternativa considerada**: `chi` — mais leve e próximo da stdlib, mas sem integração oficial com Swaggo e menos adotado na equipe.

### SetupRouter como função pura

**Decisão**: `SetupRouter` recebe os handlers como parâmetros e retorna `*gin.Engine`, sem estado global.

**Rationale**: Facilita testes unitários do roteamento sem subir servidor real, e mantém o `main.go` como único ponto de composição de dependências.

### Adaptar handlers ao contrato do Gin

**Decisão**: os handlers atuais (`WebhookHandler.Handle`, `HealthHandler.Handle`) têm assinatura `http.HandlerFunc`. O Gin aceita `gin.HandlerFunc` (`func(*gin.Context)`). Usaremos `gin.WrapF` para adaptar sem reescrever os handlers.

**Rationale**: Minimiza o escopo do change — os handlers não precisam ser alterados agora. A migração para `gin.Context` pode ser feita em um change dedicado.

### Swagger via gin-swagger

**Decisão**: substituir `httpSwagger.WrapHandler` por `ginSwagger.WrapHandler(swaggerFiles.Handler)` registrado em `GET /docs/*any`.

**Rationale**: `gin-swagger` é o wrapper oficial do Swaggo para Gin e mantém compatibilidade com o `swagger.json` já gerado.

## Risks / Trade-offs

- **Gin adiciona ~5 MB ao binário** → aceitável dado o ganho em ergonomia de roteamento
- **`gin.WrapF` adiciona uma camada de adaptação** → risco baixo; é o padrão recomendado pelo Gin para migração gradual de handlers stdlib

## Migration Plan

1. Adicionar dependências (`gin`, `gin-swagger`) via `go get`
2. Criar `routes.go` com `SetupRouter`
3. Atualizar `main.go`: remover `http.NewServeMux`, chamar `SetupRouter`, usar engine como handler do `http.Server`
4. Verificar Swagger UI em `GET /docs/index.html`
5. Smoke test: `curl POST /webhooks/mercadopago` e `curl GET /health`
