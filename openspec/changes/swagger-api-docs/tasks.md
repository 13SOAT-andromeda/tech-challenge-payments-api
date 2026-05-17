## 1. Dependências

- [x] 1.1 Adicionar `github.com/swaggo/http-swagger` ao `go.mod` via `go get`
- [x] 1.2 Instalar o CLI `swag` globalmente: `go install github.com/swaggo/swag/cmd/swag@latest`

## 2. Annotations Swaggo em main.go

- [x] 2.1 Adicionar annotations gerais da API em `cmd/api/main.go`: `@title`, `@version`, `@description`, `@host`, `@BasePath`

## 3. Annotations no WebhookHandler

- [x] 3.1 Adicionar annotations Swaggo ao método `Handle` em `internal/adapters/in/http/webhook_handler.go`: `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`
- [x] 3.2 Documentar headers obrigatórios: `@Param x-signature header string true` e `@Param x-request-id header string false`
- [x] 3.3 Documentar o request body com a struct `mpWebhookPayload` (ou inline): `@Param payload body mpWebhookPayload true`
- [x] 3.4 Documentar responses: `@Success 200`, `@Failure 400`, `@Failure 500` com `@Router /webhooks/mercadopago [post]`

## 4. Geração do pacote docs/

- [x] 4.1 Executar `swag init -g cmd/api/main.go` na raiz do projeto para gerar `docs/docs.go`, `docs/swagger.json` e `docs/swagger.yaml`
- [x] 4.2 Adicionar import anônimo `_ "github.com/gedanmx/payments-api/docs"` em `cmd/api/main.go`

## 5. Rota da Swagger UI

- [x] 5.1 Importar `httpSwagger "github.com/swaggo/http-swagger"` em `cmd/api/main.go`
- [x] 5.2 Registrar a rota `GET /docs/` no mux usando `mux.Handle("GET /docs/", httpSwagger.WrapHandler)` (ou equivalente com wildcard)

## 6. Dockerfile

- [x] 6.1 Adicionar no estágio `builder` do `Dockerfile`: `go install github.com/swaggo/swag/cmd/swag@latest` antes do `go build`
- [x] 6.2 Adicionar `RUN swag init -g cmd/api/main.go` antes do `RUN go build` no `Dockerfile`

## 7. Verificação

- [x] 7.1 Executar `go build ./...` e confirmar que compila sem erros
- [x] 7.2 Iniciar a aplicação localmente e acessar `http://localhost:8080/docs/index.html` confirmando que a UI carrega com o endpoint documentado
