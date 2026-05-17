## MODIFIED Requirements

### Requirement: Consumer processa eventos da fila payment-order-events-queue via SNS→SQS
O consumer SQS da fila `payment-order-events-queue` (inscrita no tópico `order.events`) SHALL detectar automaticamente mensagens envelopadas pelo SNS (campo `Type: "Notification"`) e extrair o payload real do campo `Message` antes de processar.

#### Scenario: Mensagem SNS envelopada é unwrapped corretamente
- **WHEN** uma mensagem chega na SQS com body contendo `{"Type":"Notification","Message":"{...}"}`
- **THEN** o consumer SHALL extrair o conteúdo de `Message` e fazer parse como evento

#### Scenario: Mensagem não-envelopada é processada diretamente
- **WHEN** uma mensagem chega na SQS com body sendo JSON plano (sem campo `Type`)
- **THEN** o consumer SHALL processar o body diretamente como evento sem double-unmarshal
