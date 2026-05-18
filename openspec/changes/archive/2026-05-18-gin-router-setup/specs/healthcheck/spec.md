## MODIFIED Requirements

### Requirement: GET /health retorna status agregado da API e suas dependências
O sistema SHALL expor o endpoint `GET /health` que verifica e reporta o status de saúde do banco de dados PostgreSQL, da fila SQS e do tópico SNS, retornando uma resposta JSON estruturada com o status global e o status individual de cada componente. A rota SHALL ser registrada pelo Gin via `SetupRouter` em vez de `http.NewServeMux`.

#### Scenario: Todos os componentes saudáveis
- **WHEN** o banco de dados, a fila SQS e o tópico SNS estão acessíveis
- **THEN** o endpoint SHALL retornar HTTP 200 com body `{"status":"healthy","checks":{"database":{"status":"healthy"},"sqs":{"status":"healthy"},"sns":{"status":"healthy"}}}`

#### Scenario: Banco de dados inacessível
- **WHEN** o banco de dados não responde ao ping
- **THEN** o endpoint SHALL retornar HTTP 503 com `checks.database.status = "unhealthy"` e `checks.database.error` contendo a mensagem de erro

#### Scenario: Fila SQS inacessível
- **WHEN** a chamada `GetQueueAttributes` retorna erro (permissão negada, fila inexistente ou timeout)
- **THEN** o endpoint SHALL retornar HTTP 503 com `checks.sqs.status = "unhealthy"` e `checks.sqs.error` contendo a mensagem de erro

#### Scenario: Tópico SNS inacessível
- **WHEN** a chamada `GetTopicAttributes` retorna erro (permissão negada, tópico inexistente ou timeout)
- **THEN** o endpoint SHALL retornar HTTP 503 com `checks.sns.status = "unhealthy"` e `checks.sns.error` contendo a mensagem de erro

#### Scenario: Falha parcial não impede reporte dos componentes saudáveis
- **WHEN** uma ou mais dependências falham mas as demais respondem com sucesso
- **THEN** o endpoint SHALL retornar HTTP 503 com o status correto (healthy/unhealthy) para cada componente individualmente
