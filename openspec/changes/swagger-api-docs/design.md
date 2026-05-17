## Context

A API tem atualmente apenas o endpoint `POST /webhooks/mercadopago`. A documentação é gerada via annotations Go usando a lib `swaggo/swag`, que produz um pacote `docs/` estático incluído no binário. A Swagger UI é servida pelo handler `http-swagger` diretamente pela aplicação, sem servidor separado.

## Goals / Non-Goals

**Goals:**
- Documentar o endpoint `POST /webhooks/mercadopago` com request body, headers e responses
- Servir Swagger UI em `GET /docs/*` integrada ao servidor HTTP existente
- Geração de docs reproduzível via `swag init` (CI/Dockerfile)

**Non-Goals:**
- Autenticação/proteção da rota `/docs/*` (sem dados sensíveis expostos)
- Documentar endpoints internos ou SQS
- Migração para OpenAPI 3.0 (swaggo/swag gera OpenAPI 2.0)

## Decisions

### 1. `swaggo/swag` como gerador

**Escolha:** `github.com/swaggo/swag` + `github.com/swaggo/http-swagger`

**Alternativas consideradas:**
- Escrever `openapi.yaml` manualmente: mais flexível, mas sem validação automática contra o código
- `ogen` (geração de código a partir de spec): inversão de abordagem, requer reescrever os handlers

**Rationale:** Swaggo é o padrão de facto em Go para annotation-based docs, compatível com `net/http` padrão via `http-swagger`, sem necessidade de mudar a estrutura dos handlers.

### 2. Import anônimo do pacote `docs/`

O `swag init` gera `docs/docs.go` com um `func init()` que registra a spec no runtime do swaggo. O import anônimo `_ "github.com/gedanmx/payments-api/docs"` em `main.go` garante que a spec seja embutida no binário sem dependências globais em tempo de compilação.

### 3. Rota `/docs/*` com `http-swagger`

`httpSwagger.WrapHandler` retorna um `http.Handler` compatível com o mux padrão (`net/http`). A rota será registrada como `GET /docs/{wildcard...}` usando a sintaxe do Go 1.22+ pattern matching.

### 4. `swag init` no Dockerfile

O `swag` CLI precisa ser instalado na imagem builder antes do `go build`. A imagem `golang:1.25-alpine` permite instalar via `go install github.com/swaggo/swag/cmd/swag@latest`. A pasta `docs/` gerada é committed para que a imagem de desenvolvimento local funcione sem precisar do CLI.

## Risks / Trade-offs

- **Docs desatualizados:** Se annotations mudarem sem `swag init` ser reexecutado, a UI fica desatualizada. Mitigação: adicionar `swag init` como passo explícito no Dockerfile e no README.
- **Tamanho do binário:** O pacote `docs/` adiciona ~50KB ao binário. Impacto negligenciável.
- **OpenAPI 2.0 vs 3.0:** swaggo/swag gera 2.0. Ferramentas modernas aceitam ambas; não é bloqueante para o objetivo do change.
