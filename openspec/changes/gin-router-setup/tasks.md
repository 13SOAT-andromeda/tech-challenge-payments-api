## 1. Dependências

- [x] 1.1 Adicionar `github.com/gin-gonic/gin` via `go get`
- [x] 1.2 Adicionar `github.com/swaggo/gin-swagger` e `github.com/swaggo/files` via `go get`

## 2. Router

- [x] 2.1 Criar `internal/adapters/in/http/routes.go` com função `SetupRouter(webhook *WebhookHandler, health *HealthHandler) *gin.Engine`
- [x] 2.2 Registrar `POST /webhooks/mercadopago` usando `gin.WrapF(webhook.Handle)`
- [x] 2.3 Registrar `GET /health` usando `gin.WrapF(health.Handle)`
- [x] 2.4 Registrar `GET /docs/*any` com `ginSwagger.WrapHandler(swaggerFiles.Handler)`

## 3. Main

- [x] 3.1 Remover `http.NewServeMux` e registro inline de rotas do `main.go`
- [x] 3.2 Chamar `SetupRouter` passando `webhookHandler` e `healthHandler`
- [x] 3.3 Usar o `*gin.Engine` retornado como `Handler` do `http.Server`
- [x] 3.4 Remover import `httpSwagger "github.com/swaggo/http-swagger"` e adicionar imports do gin-swagger

## 4. Verificação

- [x] 4.1 Executar `go build ./...` sem erros
- [x] 4.2 Executar `go test ./...` sem regressões
- [x] 4.3 Smoke test: `curl -X POST http://localhost:8080/webhooks/mercadopago` retorna 400 (sem assinatura)
- [x] 4.4 Smoke test: `curl http://localhost:8080/health` retorna JSON de health
- [x] 4.5 Smoke test: `curl http://localhost:8080/docs/index.html` retorna Swagger UI
