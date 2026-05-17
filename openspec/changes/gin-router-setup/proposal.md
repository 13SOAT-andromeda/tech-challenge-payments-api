## Why

O servidor HTTP atual registra rotas inline no `main.go` usando `http.NewServeMux`, misturando bootstrap de infraestrutura com definição de rotas e dificultando a evolução do roteamento. Adotar o Gin centraliza o roteamento em um `routes.go` dedicado, habilita middlewares por grupo de rota e mantém o `main.go` restrito à composição de dependências.

## What Changes

- Adicionar `github.com/gin-gonic/gin` como dependência do módulo
- Criar `internal/adapters/in/http/routes.go` com função `SetupRouter` que registra todas as rotas
- Remover o `http.NewServeMux` e o registro inline de rotas do `main.go`
- Substituir `http.Server.Handler` pelo engine do Gin retornado por `SetupRouter`
- Manter compatibilidade com o Swagger via middleware `ginSwagger`

## Capabilities

### New Capabilities

- `gin-router`: Arquivo `routes.go` com `SetupRouter` centralizando registro de rotas, grupos e middlewares via Gin

### Modified Capabilities

- `healthcheck`: rota `GET /health` passa a ser registrada pelo Gin em vez de `http.NewServeMux`
- `webhook-hmac-validation`: rota `POST /webhooks/mercadopago` passa a ser registrada pelo Gin

## Impact

- **`cmd/api/main.go`**: remove `http.NewServeMux` e registro de rotas; passa a chamar `SetupRouter` e usa o engine como `Handler`
- **`internal/adapters/in/http/`**: novo arquivo `routes.go`
- **`go.mod` / `go.sum`**: adição de `github.com/gin-gonic/gin` e `github.com/swaggo/gin-swagger`
- Sem breaking changes de API — paths e verbos HTTP permanecem os mesmos
