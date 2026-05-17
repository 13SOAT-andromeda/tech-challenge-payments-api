## 1. Port PaymentGateway

- [x] 1.1 Adicionar `ErrNoApprovedPayment` como variável de erro sentinela em `internal/core/ports/payment_gateway.go`
- [x] 1.2 Adicionar método `GetMerchantOrderPaymentID(ctx context.Context, merchantOrderID string) (string, error)` à interface `PaymentGateway`

## 2. Adapter MercadoPago

- [x] 2.1 Adicionar campo `merchantOrderClient merchantorder.Client` na struct `Client` em `internal/adapters/out/mercadopago/client.go`
- [x] 2.2 Inicializar `merchantOrderClient` em `NewClient` com `merchantorder.NewClient(cfg)`
- [x] 2.3 Implementar `GetMerchantOrderPaymentID`: converter `merchantOrderID` para `int` com `strconv.Atoi`, chamar `merchantOrderClient.Get`, iterar `Payments` e retornar o ID do primeiro com `Status == "approved"` como string; retornar `ErrNoApprovedPayment` se nenhum encontrado

## 3. Webhook Handler

- [x] 3.1 Adicionar campo `ID string json:"id"` na struct `MPWebhookPayload` para capturar o merchant order ID do body
- [x] 3.2 Adicionar campo `merchantOrderClient` do tipo `ports.PaymentGateway` ao `WebhookHandler` e atualizar `NewWebhookHandler` para recebê-lo (ou reutilizar o service existente via nova assinatura no service)
- [x] 3.3 Em `Handle()`, após decodificar o payload, adicionar branch `case "topic_merchant_order_wh"`: converter `payload.ID` para int, chamar `gateway.GetMerchantOrderPaymentID`, logar warn se `ErrNoApprovedPayment`, chamar `service.ProcessWebhook(paymentID)` com o ID retornado
- [x] 3.4 Atualizar a condição de roteamento existente para cobrir `payment` e `topic_merchant_order_wh`, retornando HTTP 200 para outros tipos

## 4. Injeção de Dependência

- [x] 4.1 Em `cmd/api/main.go`, passar o `mpClient` (que implementa `PaymentGateway`) ao `NewWebhookHandler`

## 5. Verificação

- [x] 5.1 Executar `go build ./...` e confirmar compilação sem erros
- [ ] 5.2 Realizar checkout no sandbox, aguardar webhook `topic_merchant_order_wh` e confirmar que o pagamento é processado e o status atualizado no banco
