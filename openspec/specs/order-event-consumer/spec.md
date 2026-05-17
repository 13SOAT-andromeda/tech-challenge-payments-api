## MODIFIED Requirements

### Requirement: Consumer processa eventos da fila payment-order-events-queue via SNS→SQS
O consumer SQS da fila `payment-order-events-queue` (inscrita no tópico `order.events`) SHALL detectar automaticamente mensagens envelopadas pelo SNS (campo `Type: "Notification"`) e extrair o payload real do campo `Message` antes de processar.

#### Scenario: Mensagem SNS envelopada é unwrapped corretamente
- **WHEN** uma mensagem chega na SQS com body contendo `{"Type":"Notification","Message":"{...}"}`
- **THEN** o consumer SHALL extrair o conteúdo de `Message` e fazer parse como evento

#### Scenario: Mensagem não-envelopada é processada diretamente
- **WHEN** uma mensagem chega na SQS com body sendo JSON plano (sem campo `Type`)
- **THEN** o consumer SHALL processar o body diretamente como evento sem double-unmarshal

---

### Requirement: OrderCreated cria registro de pagamento no banco
Ao receber evento `OrderCreated`, o sistema SHALL criar um registro de `Payment` no banco de dados com os campos obrigatórios mapeados do evento.

#### Scenario: Pagamento criado com sucesso a partir de OrderCreated
- **WHEN** um evento `OrderCreated` válido é consumido com `orderId`, `correlationId`, `amount`, `payment.method = CHECKOUT`, `payment.provider = MERCADO_PAGO`
- **THEN** o sistema SHALL persistir um `Payment` com `BusinessStatus = PENDING`, `SagaStatus = AWAITING_PAYMENT`, `CorrelationID` e `Provider` preenchidos

#### Scenario: Evento com campos obrigatórios ausentes é rejeitado
- **WHEN** um evento `OrderCreated` chega sem `orderId` ou `correlationId`
- **THEN** o consumer SHALL registrar erro de validação e NÃO deletar a mensagem da fila (deixar para retry/DLQ)

---

### Requirement: Consumer integra com provedor após criar pagamento
Após persistir o `Payment`, o sistema SHALL chamar o provedor de pagamento para gerar a URL de checkout.

#### Scenario: Checkout gerado com sucesso
- **WHEN** o pagamento é persistido e a chamada ao provedor retorna `checkoutUrl` e `expiresAt`
- **THEN** o sistema SHALL atualizar o `Payment` com `CheckoutURL`, `ExpiresAt` e `SagaStatus = AWAITING_PAYMENT` e publicar o evento `PaymentCheckoutCreated`

#### Scenario: Falha na chamada ao provedor
- **WHEN** o provedor retorna erro ou timeout
- **THEN** o sistema SHALL NÃO deletar a mensagem da SQS, permitindo retry via `VisibilityTimeout`

---

### Requirement: Mensagens com falha permanente vão para DLQ
O consumer SHALL configurar política de retry na SQS: após N tentativas falhas, a mensagem SHALL ser movida automaticamente para a Dead Letter Queue associada.

#### Scenario: Mensagem excede tentativas máximas
- **WHEN** uma mensagem falha no processamento por `maxReceiveCount` vezes consecutivas
- **THEN** a SQS SHALL mover automaticamente a mensagem para a DLQ configurada

#### Scenario: Mensagem processada com sucesso é deletada da fila
- **WHEN** o consumer processa uma mensagem sem erro
- **THEN** o consumer SHALL chamar `DeleteMessage` na SQS antes de passar para a próxima mensagem

---

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
