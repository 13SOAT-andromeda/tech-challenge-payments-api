## Context

O Mercado Pago envia dois tipos de webhook para pagamentos via Checkout Pro:
- `payment`: disparado diretamente com o `payment_id` em `data.id`
- `topic_merchant_order_wh`: disparado quando a merchant order é atualizada, com o `merchant_order_id` no campo `id` do body

O handler atual trata apenas `payment`. O fluxo de checkout pro gera `topic_merchant_order_wh`, que carrega a lista de pagamentos associados (`Payments[]`) dentro da merchant order. Para obter o `payment_id`, é necessário consultar a API `/merchant_orders/{id}` e extrair o ID do pagamento aprovado.

O SDK já disponibiliza `github.com/mercadopago/sdk-go/pkg/merchantorder` com `Get(ctx, id int)` retornando `Response.Payments[]PaymentResponse`.

## Goals / Non-Goals

**Goals:**
- Rotear `type: topic_merchant_order_wh` no webhook handler
- Consultar a merchant order no MP via SDK para extrair o `payment_id`
- Reutilizar `ProcessWebhook(paymentID)` sem alteração na lógica de negócio
- Manter idempotência e validação HMAC inalteradas

**Non-Goals:**
- Suporte a merchant orders com múltiplos pagamentos parciais (split payment)
- Sincronização de itens ou shipments da merchant order
- Alteração no domínio ou nos eventos publicados no SNS

## Decisions

### 1. Extrair `payment_id` da merchant order no adapter, não no service

O `PaymentService.ProcessWebhook` recebe um `paymentID string` e já lida com toda a lógica de negócio. A resolução do `payment_id` a partir do merchant order é um detalhe de integração com o MP — pertence ao adapter, não ao domínio.

**Alternativa considerada:** criar `ProcessMerchantOrderWebhook(merchantOrderID)` no service. Rejeitado pois duplicaria responsabilidade e adicionaria complexidade desnecessária ao domínio.

### 2. Novo método `GetMerchantOrderPaymentID` no port `PaymentGateway`

A interface `PaymentGateway` recebe um novo método:
```go
GetMerchantOrderPaymentID(ctx context.Context, merchantOrderID string) (string, error)
```
Retorna o `payment_id` (string) do primeiro pagamento com `Status == "approved"` na merchant order. Se nenhum aprovado, retorna erro sentinela `ErrNoApprovedPayment`.

**Alternativa considerada:** retornar todos os payment IDs. Rejeitado pois o service não está preparado para processar múltiplos pagamentos e o caso de uso atual é single payment.

### 3. Top-level `ID` no `MPWebhookPayload`

Para `topic_merchant_order_wh`, o merchant order ID vem no campo `id` do body (não dentro de `data`). A struct `MPWebhookPayload` recebe um campo `ID string json:"id"` para capturar esse valor.

### 4. Assinatura HMAC para merchant order webhooks

O MP usa o mesmo mecanismo de assinatura para ambos os tipos. O `data.id` da query string para `topic_merchant_order_wh` contém o merchant order ID. A validação atual (`validateSignature`) não precisa de alteração.

## Risks / Trade-offs

- **Merchant order sem pagamento aprovado** → Handler retorna HTTP 200 (sem erro visível ao MP) e loga warn. O MP retentar o webhook não causaria problema pois o handler é idempotente.
- **Merchant order com múltiplos pagamentos** → Apenas o primeiro aprovado é processado. Aceitável para o fluxo de checkout pro que gera pagamento único.
- **SDK `merchantorder.Get` usa `id int`** → O `id` do webhook chega como string. Conversão com `strconv.Atoi` no adapter; se inválido, retorna erro imediatamente.

## Migration Plan

1. Adicionar `merchantorder.Client` ao `Client` struct do adapter MP
2. Implementar `GetMerchantOrderPaymentID` no `client.go`
3. Adicionar método à interface `PaymentGateway`
4. Atualizar `MPWebhookPayload` com campo `ID`
5. Adicionar branch `topic_merchant_order_wh` no `webhook_handler.go`
6. Deploy sem necessidade de migração de dados ou feature flag
