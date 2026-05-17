## 1. Handler de Healthcheck

- [x] 1.1 Criar `internal/adapters/in/http/health_handler.go` com a struct `HealthHandler` recebendo `*gorm.DB`, `*sqs.Client`, `*sns.Client`, `sqsQueueURL string` e `snsTopicARN string`
- [x] 1.2 Definir as structs de resposta JSON: `HealthResponse{Status string; Checks map[string]CheckResult}` e `CheckResult{Status string; Error string,omitempty}`
- [x] 1.3 Implementar o método `checkDatabase(ctx) CheckResult`: chama `db.DB().PingContext` com timeout de 3s
- [x] 1.4 Implementar o método `checkSQS(ctx) CheckResult`: chama `sqs.GetQueueAttributes` com `QueueUrl` e timeout de 3s
- [x] 1.5 Implementar o método `checkSNS(ctx) CheckResult`: chama `sns.GetTopicAttributes` com `TopicArn` e timeout de 3s
- [x] 1.6 Implementar o método `Handle(w, r)`: executa os três checks em goroutines paralelas com `sync.WaitGroup`, agrega os resultados, define `status` global como `"unhealthy"` se qualquer check falhar e retorna HTTP 200 ou 503

## 2. Annotations Swagger

- [x] 2.1 Adicionar annotations Swaggo ao método `Handle` do `HealthHandler`: `@Summary`, `@Description`, `@Tags`, `@Produce`, `@Success 200 {object} HealthResponse`, `@Failure 503 {object} HealthResponse`, `@Router /health [get]`

## 3. Wiring em main.go

- [x] 3.1 Instanciar `HealthHandler` em `cmd/api/main.go` passando `repo.DB()`, `sqsClient`, `snsClient`, `SQS_QUEUE_URL_ORDER_EVENTS` e `SNS_TOPIC_ARN_PAYMENT`
- [x] 3.2 Registrar a rota `GET /health` no mux: `mux.HandleFunc("GET /health", healthHandler.Handle)`

## 4. Expor DB do repositório

- [x] 4.1 Adicionar método `DB() *gorm.DB` ao `GORMRepository` em `internal/adapters/out/database/gorm_repository.go` para permitir que o handler de health acesse o `*gorm.DB` sem quebrar o encapsulamento da porta

## 5. Regenerar docs Swagger

- [x] 5.1 Executar `swag init -g cmd/api/main.go` para atualizar `docs/` com o novo endpoint `/health`

## 6. Verificação

- [x] 6.1 Executar `go build ./...` e confirmar que compila sem erros
- [x] 6.2 Subir a aplicação e fazer `curl http://localhost:8080/health` confirmando resposta JSON com status de cada componente
