## ADDED Requirements

### Requirement: SetupRouter centraliza o registro de rotas via Gin
O sistema SHALL expor uma função `SetupRouter(webhook *WebhookHandler, health *HealthHandler) *gin.Engine` em `internal/adapters/in/http/routes.go` que registra todas as rotas HTTP da aplicação e retorna o engine configurado.

#### Scenario: Rotas registradas corretamente
- **WHEN** `SetupRouter` é chamado com os handlers injetados
- **THEN** o engine retornado SHALL ter `POST /webhooks/mercadopago`, `GET /health` e `GET /docs/*any` registrados

#### Scenario: Engine usado como Handler do http.Server
- **WHEN** o `main.go` monta o `http.Server`
- **THEN** o campo `Handler` SHALL ser o `*gin.Engine` retornado por `SetupRouter`, sem uso de `http.NewServeMux`

---

### Requirement: Handlers stdlib adaptados com gin.WrapF
Os handlers existentes de assinatura `http.HandlerFunc` SHALL ser adaptados para Gin usando `gin.WrapF` sem alterar a lógica interna dos handlers.

#### Scenario: WebhookHandler adaptado
- **WHEN** a rota `POST /webhooks/mercadopago` recebe uma requisição
- **THEN** o Gin SHALL delegar para `gin.WrapF(webhookHandler.Handle)` preservando headers e body

#### Scenario: HealthHandler adaptado
- **WHEN** a rota `GET /health` recebe uma requisição
- **THEN** o Gin SHALL delegar para `gin.WrapF(healthHandler.Handle)` preservando a resposta JSON

---

### Requirement: Swagger UI acessível via gin-swagger
A rota `GET /docs/*any` SHALL ser registrada com `ginSwagger.WrapHandler(swaggerFiles.Handler)` para servir a Swagger UI.

#### Scenario: Swagger UI carrega corretamente
- **WHEN** `GET /docs/index.html` é acessado
- **THEN** o servidor SHALL retornar HTTP 200 com a Swagger UI renderizada a partir do `swagger.json` já gerado
