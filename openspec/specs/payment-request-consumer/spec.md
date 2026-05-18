# payment-request-consumer Specification

## Purpose
TBD - created by archiving change mercadopago-integration-go. Update Purpose after archive.
## Requirements
### Requirement: Consumir evento payment.requested do SQS
O serviço SHALL consumir continuamente a fila SQS configurada e processar mensagens com `event_type: "payment.requested"` em uma goroutine dedicada.

#### Scenario: Mensagem válida recebida
- **WHEN** uma mensagem com `event_type: "payment.requested"` é recebida na fila SQS
- **THEN** o serviço SHALL desserializar o payload JSON para `PaymentRequestedEvent` e invocar `PaymentService.ProcessPaymentRequest`

#### Scenario: Mensagem com formato inválido
- **WHEN** uma mensagem com JSON inválido ou campos obrigatórios ausentes (`order_id`, `customer_email`, `amount`, `items`) é recebida
- **THEN** o serviço SHALL fazer NACK da mensagem (não deletar do SQS) e registrar o erro em log, permitindo que o SQS direcione para a DLQ após esgotar as tentativas

#### Scenario: Falha ao processar a mensagem (erro externo)
- **WHEN** o processamento falha por erro no Mercado Pago ou SQS publisher
- **THEN** o serviço SHALL fazer NACK da mensagem para reprocessamento, respeitando o visibility timeout do SQS

### Requirement: Operação concorrente com servidor HTTP
O consumer SQS SHALL operar em goroutine separada, sem bloquear o servidor HTTP do webhook.

#### Scenario: Inicialização do serviço
- **WHEN** o serviço é iniciado
- **THEN** o consumer SQS e o servidor HTTP SHALL estar ambos operacionais simultaneamente

### Requirement: Graceful shutdown do consumer
O consumer SHALL encerrar somente após finalizar o processamento da mensagem em curso ao receber sinal de encerramento.

#### Scenario: Sinal SIGINT ou SIGTERM recebido durante processamento
- **WHEN** o serviço recebe `SIGINT` ou `SIGTERM` enquanto uma mensagem está sendo processada
- **THEN** o consumer SHALL concluir o processamento da mensagem atual antes de encerrar
- **THEN** nenhuma mensagem em processamento SHALL ser perdida ou deixada visível na fila sem resolução

#### Scenario: Sinal SIGINT ou SIGTERM recebido sem processamento ativo
- **WHEN** o serviço recebe `SIGINT` ou `SIGTERM` sem mensagem em processamento
- **THEN** o consumer SHALL encerrar imediatamente

