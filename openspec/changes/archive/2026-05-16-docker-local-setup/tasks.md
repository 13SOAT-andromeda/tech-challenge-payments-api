## 1. Dockerfile

- [x] 1.1 Criar `Dockerfile` na raiz com stage `builder` usando `golang:1.25-alpine` e `go build -o /app/payments-api ./cmd/api`
- [x] 1.2 Adicionar stage `runtime` usando `alpine:3.20`, copiar o binário e certificados CA (`ca-certificates`)
- [x] 1.3 Declarar `EXPOSE 8080` e `ENTRYPOINT ["/app/payments-api"]` no stage runtime
- [x] 1.4 Verificar que `docker build -t payments-api .` conclui sem erros

## 2. docker-compose.yml

- [x] 2.1 Criar `docker-compose.yml` na raiz com serviço `postgres` (imagem `postgres:16-alpine`, credenciais `payments/payments/payments`, healthcheck com `pg_isready`)
- [x] 2.2 Adicionar serviço `app` com `build: .`, `depends_on: postgres: condition: service_healthy` e `restart: on-failure`
- [x] 2.3 Configurar `DATABASE_URL=postgres://payments:payments@postgres:5432/payments?sslmode=disable` fixo no serviço `app`
- [x] 2.4 Repassar via `environment` as variáveis do `.env`: `MERCADOPAGO_ACCESS_TOKEN`, `MERCADOPAGO_WEBHOOK_SECRET`, `SQS_QUEUE_URL_ORDER_EVENTS`, `SNS_TOPIC_ARN_PAYMENT`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `PORT`
- [x] 2.5 Mapear porta `${PORT:-8080}:8080` para acesso do host

## 3. Variáveis de ambiente

- [x] 3.1 Atualizar `.env` com `SQS_QUEUE_URL_ORDER_EVENTS=https://sqs.us-east-1.amazonaws.com/639415499031/payment-order-events-queue`
- [x] 3.2 Atualizar `.env` com `SNS_TOPIC_ARN_PAYMENT=arn:aws:sns:us-east-1:639415499031:payment-events`
- [x] 3.3 Adicionar `AWS_ACCESS_KEY_ID=` e `AWS_SECRET_ACCESS_KEY=` no `.env`
- [x] 3.4 Criar `.env.example` com todos os campos e valores placeholder (sem segredos reais)

## 4. .gitignore

- [x] 4.1 Verificar se `.env` está no `.gitignore`; adicionar se não estiver

## 5. Verificação

- [x] 5.1 Rodar `docker compose up --build` e confirmar que `postgres` e `app` sobem sem erros
- [x] 5.2 Confirmar que a aplicação loga conexão com banco e início do consumer SQS
- [x] 5.3 Confirmar que `POST /webhooks/mercadopago` responde na porta mapeada
