## ADDED Requirements

### Requirement: Build multi-stage produz binário Go enxuto
O `Dockerfile` SHALL usar dois stages: `builder` compila o binário com `go build`; `runtime` copia apenas o binário para uma imagem Alpine mínima.

#### Scenario: Build bem-sucedido gera imagem executável
- **WHEN** `docker build -t payments-api .` é executado na raiz do projeto
- **THEN** a imagem é criada sem erros e o binário `/app/payments-api` está presente no container

#### Scenario: Imagem final não contém toolchain Go
- **WHEN** a imagem final é inspecionada
- **THEN** não há compilador Go, módulos de cache ou código-fonte — apenas o binário e certificados CA

---

### Requirement: Aplicação escuta na porta configurável via PORT
O container SHALL expor a porta definida pela variável `PORT` (padrão `8080`) e o `Dockerfile` SHALL declarar `EXPOSE` correspondente.

#### Scenario: Container inicia e responde na porta correta
- **WHEN** o container é iniciado com `PORT=8080`
- **THEN** o endpoint `POST /webhooks/mercadopago` está acessível na porta 8080

---

### Requirement: Variáveis de ambiente injetadas em runtime
O `Dockerfile` SHALL NOT embutir valores de variáveis de ambiente na imagem. Todas as configurações SHALL ser fornecidas via `--env-file` ou `environment` no compose.

#### Scenario: Imagem sem credenciais embutidas
- **WHEN** a imagem é construída sem `.env`
- **THEN** nenhuma variável sensível está presente nas layers da imagem (`docker inspect`)
