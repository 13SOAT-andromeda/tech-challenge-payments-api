## ADDED Requirements

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
