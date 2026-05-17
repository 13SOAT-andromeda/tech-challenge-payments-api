## ADDED Requirements

### Requirement: Consumer loga recebimento e resultado de cada mensagem SQS com campos estruturados
O consumer SQS SHALL registrar log estruturado ao receber cada mensagem, ao rotear por `event_type` e ao concluir o processamento (sucesso ou erro), incluindo `msg_id`, `event_type`, `order_id`, `correlation_id` e `duration_ms`.

#### Scenario: Log de mensagem recebida
- **WHEN** uma mensagem é retirada da fila SQS
- **THEN** o consumer SHALL logar `msg_id` e `event_type` em nível INFO antes de iniciar o processamento

#### Scenario: Log de mensagem processada com sucesso
- **WHEN** o processamento de uma mensagem conclui sem erro
- **THEN** o consumer SHALL logar `msg_id`, `order_id`, `correlation_id` e `duration_ms` em nível INFO

#### Scenario: Log de mensagem com erro (mantida na fila)
- **WHEN** o processamento de uma mensagem retorna erro
- **THEN** o consumer SHALL logar `msg_id`, `event_type`, `error` e `duration_ms` em nível ERROR, indicando que a mensagem será retida para retry

#### Scenario: Log de event_type desconhecido
- **WHEN** uma mensagem chega com `event_type` não reconhecido
- **THEN** o consumer SHALL logar `msg_id` e `event_type` em nível WARN e ignorar a mensagem
