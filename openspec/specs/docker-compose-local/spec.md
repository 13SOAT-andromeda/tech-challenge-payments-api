## ADDED Requirements

### Requirement: docker compose up sobe postgres e app sem configuração manual
O `docker-compose.yml` SHALL definir os serviços `postgres` e `app` de forma que `docker compose up` inicie o ambiente completo sem passos adicionais além de preencher o `.env`.

#### Scenario: Ambiente sobe com um único comando
- **WHEN** o `.env` está preenchido e `docker compose up` é executado
- **THEN** os serviços `postgres` e `app` sobem sem erros

#### Scenario: App aguarda postgres estar pronto
- **WHEN** o serviço `postgres` ainda está inicializando
- **THEN** o serviço `app` aguarda o healthcheck do postgres passar antes de iniciar

---

### Requirement: Serviço postgres usa credenciais isoladas
O serviço `postgres` no compose SHALL usar banco `payments`, usuário `payments` e senha `payments`, sem expor credenciais de produção.

#### Scenario: Banco criado com credenciais do compose
- **WHEN** o serviço `postgres` inicia pela primeira vez
- **THEN** o banco `payments` existe e aceita conexão com usuário `payments` e senha `payments`

---

### Requirement: DATABASE_URL do app aponta para o postgres do compose
O serviço `app` SHALL usar `DATABASE_URL=postgres://payments:payments@postgres:5432/payments?sslmode=disable`, resolvendo `postgres` via rede interna do Docker Compose.

#### Scenario: App conecta ao banco no startup
- **WHEN** o serviço `app` inicia após o postgres estar healthy
- **THEN** a aplicação realiza `AutoMigrate` sem erro e loga confirmação de conexão

---

### Requirement: Credenciais AWS injetadas via .env no compose
O `docker-compose.yml` SHALL repassar `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` e `AWS_REGION` do `.env` do host para o container `app`, permitindo acesso aos recursos SQS e SNS reais.

#### Scenario: Consumer SQS conecta à fila real
- **WHEN** o container `app` inicia com credenciais AWS válidas
- **THEN** o consumer SQS conecta à fila `payment-order-events-queue` (`arn:aws:sqs:us-east-1:639415499031:payment-order-events-queue`) e começa a fazer long-polling

#### Scenario: Publisher SNS publica no tópico real
- **WHEN** um pagamento é processado com sucesso
- **THEN** o evento é publicado no tópico `arn:aws:sns:us-east-1:639415499031:payment-events`

---

### Requirement: .env.example documenta todas as variáveis obrigatórias
O `.env.example` SHALL listar todas as variáveis esperadas pela aplicação com valores placeholder, sem nenhum valor sensível real.

#### Scenario: Novo desenvolvedor consegue configurar o ambiente a partir do exemplo
- **WHEN** o desenvolvedor copia `.env.example` para `.env` e preenche os valores
- **THEN** `docker compose up` inicia o ambiente sem erros de variável ausente

---

### Requirement: .env não é versionado
O `.gitignore` SHALL incluir `.env` para impedir que credenciais sejam commitadas acidentalmente.

#### Scenario: .env ignorado pelo git
- **WHEN** `git status` é executado com um `.env` preenchido
- **THEN** o arquivo `.env` não aparece como untracked ou staged
