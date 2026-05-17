## ADDED Requirements

### Requirement: Webhook valida assinatura HMAC-SHA256 antes de processar
O endpoint `POST /webhooks/mercadopago` SHALL verificar a assinatura do header `x-signature` enviado pelo Mercado Pago usando HMAC-SHA256 com a chave `MERCADOPAGO_WEBHOOK_SECRET` antes de qualquer processamento do payload.

#### Scenario: Assinatura válida permite processamento
- **WHEN** a requisição chega com header `x-signature` cujo campo `v1` corresponde ao HMAC-SHA256 calculado sobre `id:<data.id>;request-id:<x-request-id>;ts:<ts>`
- **THEN** o handler SHALL prosseguir com o processamento normal e retornar HTTP 200

#### Scenario: Assinatura inválida é rejeitada
- **WHEN** a requisição chega com `x-signature` cujo campo `v1` não corresponde ao HMAC esperado
- **THEN** o handler SHALL retornar HTTP 400 sem processar o payload e SHALL registrar log de aviso com o IP de origem

#### Scenario: Header x-signature ausente é rejeitado
- **WHEN** a requisição chega sem o header `x-signature`
- **THEN** o handler SHALL retornar HTTP 400

---

### Requirement: Comparação HMAC usa tempo constante
A comparação entre o HMAC recebido e o calculado SHALL usar `hmac.Equal()` para evitar timing attacks.

#### Scenario: Comparação via hmac.Equal
- **WHEN** os dois valores HMAC são comparados
- **THEN** o sistema SHALL usar `hmac.Equal(expected, received)` e nunca comparação direta com `==` ou `bytes.Equal`

---

### Requirement: Segredo HMAC é obrigatório na inicialização
A aplicação SHALL exigir a variável de ambiente `MERCADOPAGO_WEBHOOK_SECRET` na inicialização e falhar com erro fatal se ausente.

#### Scenario: Ausência do secret impede inicialização
- **WHEN** `MERCADOPAGO_WEBHOOK_SECRET` não está definida
- **THEN** a aplicação SHALL encerrar com `log.Fatalf` antes de aceitar conexões

#### Scenario: Secret presente permite inicialização normal
- **WHEN** `MERCADOPAGO_WEBHOOK_SECRET` está definida com valor não-vazio
- **THEN** a aplicação SHALL inicializar normalmente e passar o secret ao webhook handler

---

### Requirement: Webhook handler loga recebimento e resultado do processamento com campos estruturados
O handler `POST /webhooks/mercadopago` SHALL registrar log estruturado ao receber a notificação, ao validar a assinatura e ao concluir o processamento, incluindo `payment_id`, `mp_type`, resultado da validação e `duration_ms`.

#### Scenario: Log de webhook recebido
- **WHEN** uma requisição chega em `POST /webhooks/mercadopago`
- **THEN** o handler SHALL logar `payment_id` (de `data.id` da query string) e `mp_type` (campo `type` do body) em nível INFO

#### Scenario: Log de assinatura inválida
- **WHEN** a validação HMAC falha
- **THEN** o handler SHALL logar `remote_addr` e `error="assinatura inválida"` em nível WARN

#### Scenario: Log de webhook processado com sucesso
- **WHEN** o processamento conclui sem erro
- **THEN** o handler SHALL logar `payment_id` e `duration_ms` em nível INFO

#### Scenario: Log de erro no processamento
- **WHEN** `ProcessWebhook` retorna erro
- **THEN** o handler SHALL logar `payment_id`, `error` e `duration_ms` em nível ERROR
