## Context

A aplicação é um serviço Go que depende de três externos: PostgreSQL (estado local), AWS SQS (consumo de eventos de pedido) e AWS SNS (publicação de eventos de pagamento). Localmente, não existe nenhum desses serviços configurados de forma reproduzível.

Os recursos AWS são **reais** (conta `639415499031`, região `us-east-1`) — não há intenção de emular com LocalStack. O PostgreSQL, por ser um serviço stateful sem custo de acesso externo, roda como container local.

## Goals / Non-Goals

**Goals:**
- `docker compose up` sobe PostgreSQL + a aplicação com zero configuração manual
- Credenciais AWS injetadas via `.env` (não versionadas)
- Imagem Go enxuta via build multi-stage (builder + runtime mínimo)
- `.env.example` versionado como documentação de todas as variáveis

**Non-Goals:**
- Emulação de SQS/SNS local (LocalStack) — usa AWS real
- Configuração de produção / CI/CD
- Hot-reload em desenvolvimento

## Decisions

### 1. Dockerfile multi-stage

**Decisão:** Dois stages — `builder` (golang:1.25-alpine) compila o binário; `runtime` (alpine:3.20) apenas executa.

**Motivo:** Imagem final sem toolchain Go (~15 MB vs ~300 MB). Alpine reduz superfície de ataque.

**Alternativa descartada:** Imagem única `golang:alpine` — carrega compilador e cache de módulos desnecessariamente em runtime.

---

### 2. PostgreSQL como serviço Docker; SQS/SNS reais

**Decisão:** `docker-compose.yml` tem apenas dois serviços: `postgres` e `app`. AWS é acessada diretamente via credenciais no `.env`.

**Motivo:** SQS e SNS já existem provisionados na conta AWS (`639415499031`). Replicar localmente com LocalStack adicionaria complexidade sem ganho — o projeto já conecta em recursos reais.

**Alternativa descartada:** LocalStack para SQS/SNS — desnecessário pois os recursos reais já estão disponíveis.

---

### 3. `DATABASE_URL` apontando para o serviço `postgres` do compose

**Decisão:** No `docker-compose.yml`, `DATABASE_URL=postgres://payments:payments@postgres:5432/payments?sslmode=disable`. O hostname `postgres` é resolvido pela rede interna do Docker Compose.

**Motivo:** Isola o banco do host, sem necessidade de expor a porta 5432 externamente (opcional).

---

### 4. Healthcheck no postgres antes de subir o app

**Decisão:** `app` depende de `postgres` com `condition: service_healthy`, usando `pg_isready` como healthcheck.

**Motivo:** Evita race condition onde a aplicação tenta conectar antes do PostgreSQL aceitar conexões.

---

### 5. `.env` não versionado; `.env.example` versionado

**Decisão:** `.env` fica no `.gitignore`; `.env.example` com todos os campos (sem valores sensíveis) é commitado.

**Motivo:** Credenciais AWS (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) e tokens MercadoPago não devem ir para o repositório.

## Risks / Trade-offs

- **Credenciais AWS no `.env`** → Se `.env` for commitado acidentalmente, credenciais são expostas. Mitigação: garantir `.env` no `.gitignore` antes de qualquer commit.
- **SQS/SNS reais em dev** → Ações locais consomem mensagens e publicam eventos reais. Mitigação: usar conta/ambiente de teste do MP e fila SQS de test no AWS.
- **`app` reinicia se o banco demora** → `restart: on-failure` no compose garante retry automático caso o healthcheck não passe a tempo.

## Open Questions

*(nenhuma)*
