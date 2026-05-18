## ADDED Requirements

### Requirement: Criar preferência de pagamento no Mercado Pago
O serviço SHALL chamar `POST https://api.mercadopago.com/checkout/preferences` com o `ACCESS_TOKEN` para gerar um link de pagamento Checkout Pro.

#### Scenario: Criação de preferência bem-sucedida
- **WHEN** `PaymentGateway.CreatePreference` é invocado com `order_id`, `customer_email`, `amount`, `currency`, `items` e `webhook_url` válidos
- **THEN** o cliente HTTP SHALL enviar `POST /checkout/preferences` com header `Authorization: Bearer {ACCESS_TOKEN}` e `Content-Type: application/json`
- **THEN** o cliente SHALL retornar `CreatePreferenceResponse` com `PreferenceID` e `CheckoutURL` (campo `init_point` da resposta do MP)

#### Scenario: Mercado Pago retorna erro HTTP (4xx ou 5xx)
- **WHEN** o Mercado Pago retorna status de erro na criação da preferência
- **THEN** `CreatePreference` SHALL retornar um erro descritivo contendo o status HTTP e o corpo da resposta

#### Scenario: Timeout na chamada ao Mercado Pago
- **WHEN** a chamada HTTP excede 10 segundos sem resposta
- **THEN** o cliente SHALL cancelar a requisição e retornar erro de timeout

### Requirement: Consultar status de pagamento no Mercado Pago
O serviço SHALL chamar `GET https://api.mercadopago.com/v1/payments/{id}` para obter o status de um pagamento após receber webhook.

#### Scenario: Consulta de status bem-sucedida
- **WHEN** `PaymentGateway.GetPaymentStatus` é invocado com um `payment_id` válido
- **THEN** o cliente SHALL retornar o `PaymentStatus` correspondente (`approved`, `rejected`, `cancelled`, `pending`)

#### Scenario: Payment ID não encontrado no Mercado Pago
- **WHEN** o Mercado Pago retorna HTTP 404 para o `payment_id`
- **THEN** `GetPaymentStatus` SHALL retornar erro indicando pagamento não encontrado

### Requirement: Configuração via variáveis de ambiente
O cliente Mercado Pago SHALL ler `MERCADOPAGO_ACCESS_TOKEN` e `MERCADOPAGO_PUBLIC_KEY` exclusivamente de variáveis de ambiente.

#### Scenario: Variáveis de ambiente definidas
- **WHEN** `MERCADOPAGO_ACCESS_TOKEN` e `MERCADOPAGO_PUBLIC_KEY` estão definidos no ambiente
- **THEN** o cliente SHALL usar esses valores para autenticação sem hardcode

#### Scenario: ACCESS_TOKEN ausente na inicialização
- **WHEN** `MERCADOPAGO_ACCESS_TOKEN` não está definido no ambiente ao iniciar o serviço
- **THEN** o serviço SHALL encerrar com erro fatal descritivo indicando a variável ausente

### Requirement: Arquivo .env para desenvolvimento local
O projeto SHALL incluir arquivo `.env` com as entradas `MERCADOPAGO_PUBLIC_KEY` e `MERCADOPAGO_ACCESS_TOKEN` para facilitar o desenvolvimento local.

#### Scenario: Carregamento do .env no entrypoint
- **WHEN** o serviço é iniciado em ambiente de desenvolvimento com arquivo `.env` presente
- **THEN** as variáveis `MERCADOPAGO_PUBLIC_KEY` e `MERCADOPAGO_ACCESS_TOKEN` SHALL ser carregadas do arquivo `.env` antes da inicialização das dependências
