## ADDED Requirements

### Requirement: Swagger UI acessível em GET /docs/*
A aplicação SHALL servir a Swagger UI interativa e o arquivo `swagger.json` na rota `GET /docs/*` usando o handler `http-swagger`, sem autenticação, para facilitar o consumo e teste da API.

#### Scenario: UI carrega com sucesso
- **WHEN** uma requisição `GET /docs/index.html` é feita ao servidor
- **THEN** o servidor SHALL retornar HTTP 200 com o HTML da Swagger UI contendo a especificação da API carregada

#### Scenario: swagger.json retorna a especificação completa
- **WHEN** uma requisição `GET /docs/swagger.json` é feita ao servidor
- **THEN** o servidor SHALL retornar HTTP 200 com JSON válido no formato OpenAPI 2.0 contendo todos os endpoints documentados

---

### Requirement: Metadados gerais da API declarados via annotations Swaggo
O arquivo `cmd/api/main.go` SHALL conter annotations Swaggo (`@title`, `@version`, `@description`, `@host`, `@BasePath`) que descrevem a API ao nível global e são usadas pelo `swag init` para gerar o pacote `docs/`.

#### Scenario: Annotation @title presente em main.go
- **WHEN** `swag init` é executado
- **THEN** o `swagger.json` gerado SHALL conter o campo `info.title` com o nome da API

#### Scenario: @BasePath correto
- **WHEN** `swag init` é executado
- **THEN** o `swagger.json` gerado SHALL conter `basePath: "/"` e `host` configurável via annotation

---

### Requirement: Pacote docs/ gerado por swag init é incluído no binário
A pasta `docs/` gerada pelo CLI `swag init` SHALL ser comittada no repositório e importada anonimamente em `main.go` (`_ "github.com/gedanmx/payments-api/docs"`) para que a spec seja registrada no runtime sem dependência do CLI em produção.

#### Scenario: Import anônimo registra a spec no runtime
- **WHEN** o binário é iniciado
- **THEN** o pacote `docs` SHALL executar seu `init()` e registrar a spec OpenAPI antes do servidor HTTP iniciar

#### Scenario: swag init reproduz docs/ deterministicamente
- **WHEN** `swag init` é executado na raiz do projeto
- **THEN** a pasta `docs/` SHALL ser criada com `docs.go`, `swagger.json` e `swagger.yaml` compatíveis com os handlers anotados

---

### Requirement: swag init executado antes de go build no Dockerfile
O `Dockerfile` SHALL executar `go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/api/main.go` antes do `go build` para garantir que o pacote `docs/` esteja atualizado na imagem Docker.

#### Scenario: Build Docker gera docs atualizados
- **WHEN** a imagem Docker é construída via `docker build`
- **THEN** o binário SHALL conter a spec OpenAPI gerada a partir das annotations presentes no código
