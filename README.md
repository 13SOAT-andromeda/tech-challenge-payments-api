# Payments API

Microsserviço Go responsável pelo processamento de pagamentos em um sistema de oficina mecânica. Opera como participante da **SAGA Coreografada**, consumindo eventos de pedidos via AWS SQS, integrando com o **Mercado Pago Checkout Pro** para geração de links de pagamento e publicando eventos de resultado via **AWS SNS**.

## Visão Geral

```
OrderAPI
   │  SNS: order.events
   │  └─► SQS: order-events-queue
   │            │
   │            ▼
   │      PaymentsAPI  ◄──── POST /webhooks/mercadopago ◄──── Mercado Pago
   │            │
   │            │  SNS: payment.events
   │            ├─► PaymentCheckoutCreated → NotificationAPI (envia link ao cliente)
   │            ├─► PaymentApproved        → OrderAPI + NotificationAPI
   │            └─► PaymentFailed          → OrderAPI + NotificationAPI
```

## Fluxo de Pagamento

![Payment Flow](docs/Paymentflow.svg)

1. `OrderAPI` publica `OrderCreated` no SNS `order.events`
2. `PaymentsAPI` consome via SQS (com unwrap do envelope SNS)
3. Cria preferência de pagamento no Mercado Pago (Checkout Pro)
4. Persiste o pagamento com status `PENDING_CUSTOMER_ACTION` / `SagaStatus: AWAITING_PAYMENT`
5. Publica `PaymentCheckoutCreated` com o `checkout_url` para `NotificationAPI`
6. Mercado Pago notifica o status via webhook `POST /webhooks/mercadopago`
7. Webhook é validado via assinatura HMAC (`x-signature`) e processado idempotentemente
8. Publica `PaymentApproved` ou `PaymentFailed` no SNS `payment.events`

![Diagrama de Sequência](docs/tech_challenge_payment_sequence_diagram.png)

## Arquitetura

Projeto organizado seguindo **Clean/Hexagonal Architecture**:

```
cmd/
└── api/
    └── main.go                        # Entrypoint, wiring de dependências, graceful shutdown

internal/
├── core/
│   ├── domain/
│   │   └── payment.go                 # Entidade Payment, BusinessStatus, SagaStatus, PaymentStatus
│   ├── ports/
│   │   ├── message_broker.go          # Interface MessageBroker (SNS)
│   │   ├── payment_gateway.go         # Interface PaymentGateway (Mercado Pago)
│   │   └── payment_repository.go      # Interface PaymentRepository
│   └── services/
│       └── payment_service.go         # Casos de uso: ProcessPaymentRequest, ProcessWebhook
└── adapters/
    ├── in/
    │   ├── http/
    │   │   └── webhook_handler.go     # POST /webhooks/mercadopago
    │   └── sqs/
    │       └── consumer.go            # Consumer SQS com unwrap de envelope SNS
    └── out/
        ├── database/
        │   └── sqlite_repository.go   # Repositório SQLite (PostgreSQL em produção via DATABASE_URL)
        ├── mercadopago/
        │   └── client.go              # Cliente HTTP nativo para Preferences API e Payment API
        └── sns/
            └── publisher.go           # Publisher SNS para payment.events
```

## Eventos

### Consumidos

| Evento | Fila SQS | Origem |
|--------|----------|--------|
| `OrderCreated` | `SQS_QUEUE_URL_ORDER_EVENTS` (inscrita no SNS `order.events`) | OrderAPI |

### Publicados

| Evento | Tópico SNS | Destino |
|--------|-----------|---------|
| `PaymentCheckoutCreated` | `SNS_TOPIC_ARN_PAYMENT` | NotificationAPI |
| `PaymentApproved` | `SNS_TOPIC_ARN_PAYMENT` | OrderAPI, NotificationAPI |
| `PaymentFailed` | `SNS_TOPIC_ARN_PAYMENT` | OrderAPI, NotificationAPI |

## Endpoint HTTP

### `POST /webhooks/mercadopago`

Recebe notificações de status de pagamento do Mercado Pago.

**Headers obrigatórios:**

| Header | Descrição |
|--------|-----------|
| `x-signature` | Assinatura HMAC-SHA256 gerada pelo MP (`ts=...;v1=...`) |
| `x-request-id` | ID da requisição gerado pelo MP, usado no manifesto de validação |

**Payload:**

```json
{
  "type": "payment",
  "data": {
    "id": "<payment_id>"
  }
}
```

**Comportamento:**
- Retorna `HTTP 200` imediatamente; processamento ocorre em goroutine para atender ao SLA de < 500ms do Mercado Pago
- Requisições sem `x-signature` ou com assinatura inválida retornam `HTTP 400`
- Idempotente: webhooks para pagamentos já em estado final (`APPROVED`/`FAILED`) são ignorados

## Variáveis de Ambiente

| Variável | Obrigatória | Padrão | Descrição |
|----------|-------------|--------|-----------|
| `MERCADOPAGO_ACCESS_TOKEN` | Sim | — | Token de acesso da conta Mercado Pago |
| `MERCADOPAGO_PUBLIC_KEY` | Sim | — | Chave pública do Mercado Pago |
| `MERCADOPAGO_WEBHOOK_SECRET` | Sim | — | Secret para validação HMAC dos webhooks |
| `SQS_QUEUE_URL_ORDER_EVENTS` | Sim | — | URL da fila SQS de eventos de pedidos |
| `SNS_TOPIC_ARN_PAYMENT` | Sim | — | ARN do tópico SNS de eventos de pagamento |
| `DATABASE_URL` | Não | `payments.db` | DSN do banco. SQLite por padrão; PostgreSQL em produção |
| `AWS_REGION` | Não | `us-east-1` | Região AWS |
| `PORT` | Não | `8080` | Porta do servidor HTTP |
| `WEBHOOK_BASE_URL` | Não | — | URL base pública do serviço (usada na `notification_url` enviada ao MP) |
| `BACK_URL_SUCCESS` | Não | — | URL de redirecionamento após pagamento aprovado |
| `BACK_URL_FAILURE` | Não | — | URL de redirecionamento após pagamento recusado |
| `BACK_URL_PENDING` | Não | — | URL de redirecionamento para pagamento pendente |

## Rodando Localmente

### Pré-requisitos

- Go 1.24+
- GCC (necessário para `go-sqlite3`)
- [AWS CLI](https://aws.amazon.com/cli/) configurado (ou LocalStack para desenvolvimento local)
- Conta Mercado Pago (sandbox disponível em [developers.mercadopago.com](https://developers.mercadopago.com))
- [ngrok](https://ngrok.com/) ou similar para expor o webhook publicamente

### Configuração

```bash
cp .env.example .env
# Preencha as variáveis no .env
```

Exemplo de `.env`:

```env
MERCADOPAGO_ACCESS_TOKEN=TEST-xxxx
MERCADOPAGO_PUBLIC_KEY=TEST-xxxx
MERCADOPAGO_WEBHOOK_SECRET=seu_secret_aqui

SQS_QUEUE_URL_ORDER_EVENTS=https://sqs.us-east-1.amazonaws.com/000000000000/order-events-queue
SNS_TOPIC_ARN_PAYMENT=arn:aws:sns:us-east-1:000000000000:payment-events

WEBHOOK_BASE_URL=https://seu-ngrok.ngrok.io
DATABASE_URL=payments.db
```

### Executar

```bash
go run ./cmd/api
```

### Testes

```bash
go test ./...
```

## Infraestrutura AWS

Filas e tópicos necessários:

| Recurso | Tipo | Descrição |
|---------|------|-----------|
| `order-events-queue` | SQS | Inscrita no tópico SNS `order.events` da OrderAPI |
| `order-events-queue-dlq` | SQS | DLQ para mensagens que excedem retries |
| `payment-events` | SNS | Tópico de saída dos eventos de pagamento |

> Recomendado configurar DLQ em todas as filas com política de redrive após 3 tentativas.

## Modelo de Dados

O domínio `Payment` separa estado de negócio de estado de orquestração:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `BusinessStatus` | `PENDING / APPROVED / FAILED` | Status externo publicado nos eventos SAGA |
| `SagaStatus` | `STARTED / AWAITING_PAYMENT / PAYMENT_CONFIRMED / FAILED` | Estado interno de orquestração |
| `Status` | `PENDING_CUSTOMER_ACTION / APPROVED / FAILED / CANCELLED / PENDING` | Status bruto retornado pelo Mercado Pago |
| `TransactionAmount` | `float64` | Valor cobrado do cliente |
| `NetAmount` | `float64` | Valor líquido recebido (após taxas do MP) |
| `CorrelationID` | `string` | ID de correlação da SAGA distribuída |
| `Provider` | `string` | Gateway de pagamento (`MERCADO_PAGO`) |

## CI/CD

O projeto tem três GitHub Actions workflows em `.github/workflows/`:

| Workflow | Trigger | O que faz |
|----------|---------|-----------|
| `security.yml` | Pull Request | Compila o projeto, roda `gosec` (SAST) e `govulncheck`, publica relatórios SARIF/JSON como artifacts |
| `sonar.yml` | Pull Request | Roda testes com cobertura e envia resultado ao SonarCloud |
| `deploy.yml` | Push em `main` / Manual | Verifica infra (ECR + RDS), faz build e push para ECR, deploya no EKS via kustomize |

### Fluxo do deploy

```
push to main
    ├── infra-check (paralelo)   ← falha se ECR não existe ou RDS não está "available"
    ├── build (paralelo)         ← docker build + push para ECR
    └── deploy (aguarda os dois)
            ├── aws eks update-kubeconfig
            ├── sed ECR_IMAGE → URI do ECR
            ├── gera .env.host e .env.secrets
            ├── kubectl apply -k k8s/overlays/aws/
            └── kubectl rollout status deployment/payments-api
```

### Segredos e Variáveis necessários

Configure no repositório GitHub em **Settings → Secrets and variables → Actions**.

**Secrets:**

| Nome | Descrição |
|------|-----------|
| `AWS_ACCESS_KEY_ID` | Credencial AWS |
| `AWS_SECRET_ACCESS_KEY` | Credencial AWS |
| `AWS_SESSION_TOKEN` | Token de sessão (obrigatório para contas FIAP lab) |
| `AWS_REGION` | Região AWS (ex: `us-east-1`) |
| `MERCADOPAGO_ACCESS_TOKEN` | Token de acesso do Mercado Pago |
| `MERCADOPAGO_WEBHOOK_SECRET` | Secret HMAC dos webhooks do MP |
| `MERCADOPAGO_PUBLIC_KEY` | Chave pública do Mercado Pago |
| `AWS_RDS_DB_PASSWORD` | Senha master do RDS PostgreSQL |
| `SONAR_TOKEN` | Token do SonarCloud (apenas para o workflow sonar) |

**Variables (não sensíveis):**

| Nome | Exemplo |
|------|---------|
| `AWS_ECR_REPOSITORY` | `payments-api` |
| `AWS_EKS_CLUSTER_NAME` | `tech-challenge-payments` |
| `AWS_RDS_INSTANCE_ID` | `payments` |
| `SQS_QUEUE_URL_ORDER_EVENTS` | `https://sqs.us-east-1.amazonaws.com/123456789/payment-order-events-queue` |
| `SNS_TOPIC_ARN_PAYMENT` | `arn:aws:sns:us-east-1:123456789:payment-events` |

### Setup único (first time)

**1. SonarCloud:**
- Acesse [sonarcloud.io](https://sonarcloud.io) e crie o projeto `tech-challenge-payments-api` na organização `13soat-andromeda`
- Copie o token gerado e configure como secret `SONAR_TOKEN`
- O arquivo `sonar-project.properties` na raiz do repositório já está configurado

**2. ECR Repository:**
```bash
aws ecr create-repository \
  --repository-name payments-api \
  --region us-east-1
```
> Ou adicione um recurso `aws_ecr_repository` ao módulo `infra/modules/eks/` no Terraform.

**3. Infraestrutura AWS (EKS + RDS):**
```bash
cd infra
export TF_VAR_lab_role_arn="arn:aws:iam::<account>:role/LabRole"
export TF_VAR_db_user="payments"
export TF_VAR_db_pass="<senha>"
./apply-terraform.sh
```

**4. Atualizar `AWS_SESSION_TOKEN` (contas FIAP lab):**

As credenciais do laboratório FIAP expiram. Antes de cada deploy, atualize os três secrets
(`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`) com as credenciais
geradas no portal do laboratório.

---

## Kubernetes Deployment

### Pré-requisitos

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) (ou `kubectl apply -k`)
- Docker

---

### Deploy Local (kind)

**1. Crie o cluster com mapeamento de portas para o ingress:**

Crie o arquivo `kind-config.yaml` na raiz do projeto:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
```

```bash
kind create cluster --name payments --config kind-config.yaml
```

**2. Instale o ingress-nginx:**

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Aguarde o controller ficar pronto
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

**3. Build da imagem e carregamento no cluster:**

```bash
docker build -t payments-api:latest .
kind load docker-image payments-api:latest --name payments
```

**4. Aplique o overlay local:**

```bash
kubectl apply -k k8s/overlays/local/
```

**5. Verifique o deploy:**

```bash
kubectl get pods
kubectl get ingress
# Acesse: http://localhost/health
```

> O overlay local usa `imagePullPolicy: Never` e valores de desenvolvimento no ConfigMap.
> Para credenciais reais, edite `k8s/overlays/local/kustomization.yaml` na seção `secretGenerator`.

---

### Deploy AWS (EKS)

**1. Faça build e push da imagem para o ECR:**

```bash
AWS_ACCOUNT_ID=<seu-account-id>
AWS_REGION=us-east-1

aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

docker build -t payments-api:latest .
docker tag payments-api:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/payments-api:latest
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/payments-api:latest
```

**2. Configure os arquivos de ambiente:**

```bash
# Copie o template e preencha com valores reais
cp k8s/overlays/aws/.env.host.example k8s/overlays/aws/.env.host

# Crie o arquivo de segredos (nunca comite este arquivo)
cat > k8s/overlays/aws/.env.secrets <<EOF
MERCADOPAGO_ACCESS_TOKEN=<token>
MERCADOPAGO_WEBHOOK_SECRET=<secret>
MERCADOPAGO_PUBLIC_KEY=<public-key>
DB_PASSWORD=<db-password>
AWS_ACCESS_KEY_ID=<key-id>
AWS_SECRET_ACCESS_KEY=<secret-key>
EOF
```

**3. Substitua o placeholder da imagem ECR:**

```bash
sed -i 's|ECR_IMAGE|'$AWS_ACCOUNT_ID'.dkr.ecr.'$AWS_REGION'.amazonaws.com/payments-api:latest|g' \
  k8s/overlays/aws/kustomization.yaml
```

**4. Aplique o overlay AWS:**

```bash
kubectl apply -k k8s/overlays/aws/
```

**5. Verifique o deploy:**

```bash
kubectl get pods
kubectl get ingress  # anote o ADDRESS do ALB
curl http://<alb-address>/health
```

---

## Infrastructure (Terraform)

O diretório `infra/` provisionam a infraestrutura AWS (VPC, EKS, RDS) usando Terraform.

### Pré-requisitos

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.5
- AWS CLI configurado com credenciais válidas
- Bucket S3 `tech-challenge-13-soat-tfstate` existente (compartilhado com o projeto orders)

### Módulos

| Módulo | Descrição |
|--------|-----------|
| `modules/network` | VPC, subnets públicas/privadas, NAT Gateway, Route Tables |
| `modules/eks` | Cluster EKS com managed node group (t3.medium, 1–4 nós) |
| `modules/rds` | PostgreSQL 15, db.t3.micro, 20GB, em subnets privadas |

### Variáveis obrigatórias

| Variável | Descrição |
|----------|-----------|
| `lab_role_arn` | ARN do IAM role do laboratório FIAP (ex: `arn:aws:iam::<account>:role/LabRole`) |
| `db_user` | Usuário master do RDS |
| `db_pass` | Senha master do RDS |

### Aplicar infraestrutura

```bash
cd infra

# Exportar variáveis sensíveis
export TF_VAR_lab_role_arn="arn:aws:iam::<account>:role/LabRole"
export TF_VAR_db_user="payments"
export TF_VAR_db_pass="<senha-segura>"
export AWS_DEFAULT_REGION="us-east-1"

# Aplicar (provisiona VPC + EKS + RDS)
./apply-terraform.sh
```

O script executa `terraform init`, `terraform apply -auto-approve` e atualiza automaticamente o kubeconfig local com `aws eks update-kubeconfig`.

### Destruir infraestrutura

```bash
cd infra && terraform destroy
```

---

## Decisões de Design

- **`net/http` nativo** em vez do SDK oficial do Mercado Pago — controle total sobre timeouts e retry, sem dependência extra para dois endpoints simples
- **SQLite em desenvolvimento, PostgreSQL em produção** — mesma interface `database/sql`; troca via `DATABASE_URL` sem alterar lógica
- **Idempotência por estado final** — antes de processar webhook, verifica `BusinessStatus`; se já em estado final, retorna 200 sem reprocessar
- **Resposta imediata no webhook** — `HTTP 200` retornado antes do processamento (goroutine em background) para atender ao SLA de 500ms do Mercado Pago
- **Graceful shutdown** — captura `SIGINT`/`SIGTERM`; aguarda até 10s para o servidor HTTP finalizar requisições em andamento
