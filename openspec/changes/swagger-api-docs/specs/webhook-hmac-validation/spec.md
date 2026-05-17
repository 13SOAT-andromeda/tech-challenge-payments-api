## ADDED Requirements

### Requirement: Endpoint POST /webhooks/mercadopago documentado com annotations Swaggo
O handler `WebhookHandler.Handle` SHALL conter annotations Swaggo completas descrevendo o request body, os headers obrigatórios e os possíveis códigos de resposta, de forma que a Swagger UI exiba a documentação correta do endpoint.

#### Scenario: Annotation @Summary e @Description presentes
- **WHEN** `swag init` é executado
- **THEN** o `swagger.json` gerado SHALL conter o endpoint `POST /webhooks/mercadopago` com `summary` e `description` descrevendo a notificação de pagamento do Mercado Pago

#### Scenario: Headers x-signature e x-request-id documentados
- **WHEN** `swag init` é executado
- **THEN** o `swagger.json` SHALL listar `x-signature` e `x-request-id` como parâmetros do tipo `header` com `required: true`

#### Scenario: Request body documentado com struct de exemplo
- **WHEN** `swag init` é executado
- **THEN** o `swagger.json` SHALL conter a definição do body com campos `type` (string) e `data.id` (string)

#### Scenario: Responses 200, 400 e 500 documentados
- **WHEN** `swag init` é executado
- **THEN** o `swagger.json` SHALL listar as respostas `200 OK`, `400 Bad Request` e `500 Internal Server Error` para o endpoint
