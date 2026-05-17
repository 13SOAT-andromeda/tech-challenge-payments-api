## Why

A API não possui documentação interativa, dificultando o consumo e o teste dos endpoints por outros times e ferramentas. Implementar Swagger/OpenAPI expõe a especificação e uma UI navegável sem necessidade de ferramentas externas.

## What Changes

- Adicionar dependências `github.com/swaggo/swag` (gerador de docs via annotations) e `github.com/swaggo/http-swagger` (handler de UI)
- Adicionar anotações Swaggo no `cmd/api/main.go` (metadados gerais da API) e no `WebhookHandler` (endpoint documentado)
- Gerar o pacote `docs/` com `swag init` e incluí-lo no binário via import anônimo
- Registrar rota `GET /docs/*` no mux para servir a Swagger UI

## Capabilities

### New Capabilities

- `swagger-ui`: Endpoint `GET /docs/*` que serve a Swagger UI interativa e o arquivo `swagger.json` gerado a partir de annotations Go

### Modified Capabilities

- `webhook-hmac-validation`: Adicionar documentação OpenAPI (request body, responses, headers) ao endpoint `POST /webhooks/mercadopago`

## Impact

- **Dependências novas:** `github.com/swaggo/swag` (CLI + runtime), `github.com/swaggo/http-swagger`
- **Geração de código:** pasta `docs/` gerada pelo `swag init` — deve ser mantida no repositório e regerada sempre que as annotations mudarem
- **Rota nova:** `GET /docs/*` — somente para uso interno/desenvolvimento; não expõe dados sensíveis
- **Build:** adicionar passo `swag init` antes de `go build` no Dockerfile e no Makefile (se houver)
