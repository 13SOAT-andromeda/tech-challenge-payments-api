## Why

O ecossistema de processamento de pedidos requer um serviço dedicado (`PaymentsAPI`) para desacoplar o fluxo de criação de pedidos do fluxo financeiro, integrando com o Mercado Pago via Checkout Pro e participando da SAGA Coreografada via AWS SQS. A ausência deste serviço impede a cobrança automática dos clientes e o rastreamento de status de pagamento distribuído.

## What Changes

- Novo serviço `PaymentsAPI` em Go, operando como participante da SAGA Coreografada
- Consumidor SQS do evento `payment.requested` emitido pela `OrderAPI`
- Integração com Mercado Pago Checkout Pro (Preferences API) para geração de link de pagamento
- Publicação do evento `notification.email.requested` com o `checkout_url` para a `NotificationAPI`
- Endpoint `POST /webhooks/mercadopago` para recebimento de status de pagamento assíncrono
- Publicação dos eventos SAGA `payment.approved` e `payment.failed` para `OrderAPI` e `NotificationAPI`
- Arquivo `.env` com variáveis `MERCADOPAGO_PUBLIC_KEY` e `MERCADOPAGO_ACCESS_TOKEN`
- Persistência local de correlação `order_id` ↔ `preference_id` ↔ `payment_id`
- Estrutura de projeto seguindo Clean/Hexagonal Architecture

## Capabilities

### New Capabilities

- `payment-request-consumer`: Consome o evento `payment.requested` do SQS e inicia o fluxo de pagamento via Mercado Pago
- `mercadopago-checkout`: Integração com a Preferences API do Mercado Pago para criação de preferência e obtenção do link de checkout (`init_point`)
- `webhook-handler`: Endpoint HTTP que recebe notificações de status de pagamento do Mercado Pago e publica eventos SAGA correspondentes
- `saga-event-publisher`: Publicação dos eventos `notification.email.requested`, `payment.approved` e `payment.failed` no SQS
- `payment-persistence`: Repositório local para correlação e idempotência de pagamentos

### Modified Capabilities

<!-- Nenhuma capability existente sendo modificada — este é um novo serviço -->

## Impact

- **Novo serviço:** `PaymentsAPI` — serviço Go independente
- **Infraestrutura:** Requer filas SQS para `payment.requested`, `notification.email.requested`, `payment.approved` e `payment.failed`
- **Credenciais externas:** Mercado Pago `ACCESS_TOKEN` (Preferences API e Payment API) e `PUBLIC_KEY` (frontend/webhooks)
- **Dependências externas:** `github.com/aws/aws-sdk-go-v2` (SQS), `net/http` padrão (cliente Mercado Pago)
- **Banco de dados:** Armazenamento local (SQLite ou PostgreSQL) para correlação de pagamentos e garantia de idempotência
- **Sistemas downstream:** `OrderAPI` (consome `payment.approved`/`payment.failed`), `NotificationAPI` (consome `notification.email.requested` e eventos de resultado)
