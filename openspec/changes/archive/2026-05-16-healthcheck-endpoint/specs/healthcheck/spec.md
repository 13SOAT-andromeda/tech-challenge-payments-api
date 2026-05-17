## ADDED Requirements

### Requirement: GET /health retorna status agregado da API e suas dependências
O sistema SHALL expor o endpoint `GET /health` que verifica e reporta o status de saúde do banco de dados PostgreSQL, da fila SQS e do tópico SNS, retornando uma resposta JSON estruturada com o status global e o status individual de cada componente.

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

---

### Requirement: Checks executados com timeout independente por componente
Cada verificação de dependência SHALL ser executada com timeout de 3 segundos independente das demais, de forma que a falha ou lentidão de um componente não bloqueie ou atrase a verificação dos outros.

#### Scenario: Timeout em um check não atrasa os demais
- **WHEN** uma dependência não responde dentro de 3 segundos
- **THEN** esse componente SHALL ser marcado como `"unhealthy"` com erro de timeout, e os demais checks SHALL completar normalmente

#### Scenario: Checks executados em paralelo
- **WHEN** o endpoint recebe uma requisição
- **THEN** as verificações de banco, SQS e SNS SHALL ser iniciadas simultaneamente e o endpoint SHALL responder após o término de todos eles

---

### Requirement: Resposta JSON estruturada com campo status global
O endpoint SHALL retornar `Content-Type: application/json` com um objeto JSON contendo `status` (string: `"healthy"` ou `"unhealthy"`) e `checks` (objeto com um campo por componente verificado).

#### Scenario: Campo status reflete o pior estado
- **WHEN** qualquer componente retorna `"unhealthy"`
- **THEN** o campo `status` raiz SHALL ser `"unhealthy"`

#### Scenario: Campo error presente apenas em falhas
- **WHEN** um componente está saudável
- **THEN** seu objeto de check SHALL conter apenas `{"status":"healthy"}` sem o campo `error`

---

### Requirement: Endpoint documentado no Swagger
O handler `GET /health` SHALL ter annotations Swaggo completas para ser exibido na Swagger UI.

#### Scenario: Rota aparece na Swagger UI
- **WHEN** `swag init` é executado após a implementação
- **THEN** o `swagger.json` SHALL conter o endpoint `GET /health` com responses 200 e 503 documentadas
