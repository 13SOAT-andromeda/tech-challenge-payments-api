## MODIFIED Requirements

### Requirement: Webhook valida assinatura HMAC-SHA256 antes de processar
O endpoint `POST /webhooks/mercadopago` SHALL verificar a assinatura do header `x-signature` enviado pelo Mercado Pago usando HMAC-SHA256 com a chave `MERCADOPAGO_WEBHOOK_SECRET` antes de qualquer processamento do payload. A rota SHALL ser registrada pelo Gin via `SetupRouter` em vez de `http.NewServeMux`.

#### Scenario: Assinatura válida permite processamento
- **WHEN** a requisição chega com header `x-signature` cujo campo `v1` corresponde ao HMAC-SHA256 calculado sobre `id:<data.id>;request-id:<x-request-id>;ts:<ts>`
- **THEN** o handler SHALL prosseguir com o processamento normal e retornar HTTP 200

#### Scenario: Assinatura inválida é rejeitada
- **WHEN** a requisição chega com `x-signature` cujo campo `v1` não corresponde ao HMAC esperado
- **THEN** o handler SHALL retornar HTTP 400 sem processar o payload e SHALL registrar log de aviso com o IP de origem

#### Scenario: Header x-signature ausente é rejeitado
- **WHEN** a requisição chega sem o header `x-signature`
- **THEN** o handler SHALL retornar HTTP 400
