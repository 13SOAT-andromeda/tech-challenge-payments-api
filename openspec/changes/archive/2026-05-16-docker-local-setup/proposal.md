## Why

Atualmente não existe forma padronizada de rodar o projeto localmente — o desenvolvedor precisa instalar e configurar PostgreSQL manualmente e preencher variáveis de ambiente na mão. Um `Dockerfile` + `docker-compose.yml` empacota a aplicação e o banco em um único comando (`docker compose up`), eliminando atrito no onboarding e garantindo paridade de configuração entre membros do time.

## What Changes

- Adição de `Dockerfile` multi-stage para build e execução da aplicação Go
- Adição de `docker-compose.yml` com serviços `app` e `postgres`
- Atualização do `.env` com os valores reais de SQS e SNS já conhecidos:
  - `SQS_QUEUE_URL_ORDER_EVENTS` = `https://sqs.us-east-1.amazonaws.com/639415499031/payment-order-events-queue`
  - `SNS_TOPIC_ARN_PAYMENT` = `arn:aws:sns:us-east-1:639415499031:payment-events`
- Adição de `.env.example` como referência versionável (sem valores sensíveis)
- Adição de `AWS_ACCESS_KEY_ID` e `AWS_SECRET_ACCESS_KEY` no `.env` (não versionados) para autenticação local com SQS/SNS reais da AWS

## Capabilities

### New Capabilities

- `dockerfile`: Empacotamento da aplicação Go em imagem Docker via build multi-stage
- `docker-compose-local`: Orquestração local com PostgreSQL e a aplicação, com variáveis de ambiente injetadas via `.env`

### Modified Capabilities

*(nenhuma spec existente muda de requisito)*

## Impact

- **Novos arquivos**: `Dockerfile`, `docker-compose.yml`, `.env.example`
- **`.env`**: adição de `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` e preenchimento de `SQS_QUEUE_URL_ORDER_EVENTS` e `SNS_TOPIC_ARN_PAYMENT`
- **`.gitignore`**: garantir que `.env` está ignorado (credenciais AWS)
- **Sem alteração** em código Go, interfaces ou domínio
