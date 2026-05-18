## Context

The payments API is a single Go binary (`payments-api`) that exposes an HTTP server on port 8080. It consumes SQS order-events, publishes SNS payment-events, persists state in PostgreSQL, and integrates with MercadoPago for payment processing and webhooks. Currently there is no Kubernetes deployment path; the reference project (`tech-challenge-orders`) established the conventions (kustomize overlays, Terraform modules, Dockerfile build targets) that this service should adopt.

## Goals / Non-Goals

**Goals:**
- Provide runnable k8s manifests for a local cluster (minikube/kind) via kustomize overlay
- Provide runnable k8s manifests for AWS EKS via kustomize overlay
- Provide Terraform modules (network, eks, rds) for automated AWS infrastructure provisioning
- Keep the existing docker-compose.yml and Go source code untouched

**Non-Goals:**
- CI/CD pipeline configuration (GitHub Actions, ArgoCD)
- Helm chart packaging
- Multi-region or multi-AZ high-availability setup
- Monitoring stack provisioning (Datadog, Prometheus)

## Decisions

### 1. Kustomize over Helm
**Decision**: Use kustomize overlays (`k8s/base/` + `k8s/overlays/{local,aws}/`), identical to the orders service.
**Rationale**: No template language to learn; manifests are plain YAML; overlays handle all env differences via patches. Helm would add indirection without benefit for a single-service deployment.
**Alternative considered**: Helm chart — rejected for added complexity with no clear gain at this scale.

### 2. Single Deployment (no worker split)
**Decision**: One `deployment.yaml` for the API — unlike orders which has separate api and worker deployments.
**Rationale**: The payments binary is a single `payments-api` executable that handles both HTTP and SQS consumption in the same process. There is no separate worker binary.

### 3. RDS PostgreSQL for AWS, plain postgres for local
**Decision**: AWS overlay consumes DATABASE_URL pointing to RDS; local overlay uses a postgres pod or external docker-compose postgres (developer chooses).
**Rationale**: RDS provides managed backups and failover in AWS. Locally, developers already run postgres via docker-compose and can keep that or use a k8s pod.

### 4. Secrets via Kubernetes Secret (not ConfigMap)
**Decision**: Sensitive values (MercadoPago tokens, DB password, AWS credentials) are injected via a Kubernetes Secret generated from `.env.secrets` in the AWS overlay.
**Rationale**: Keeps credentials out of git. Local overlay omits secrets file; developers populate manually from `.env.example`.

### 5. HPA on CPU
**Decision**: HPA targets 70% CPU with min 1 / max 5 replicas, matching the orders service.
**Rationale**: The API is CPU-bound during payment processing. Memory-based scaling adds little value for this workload.

### 6. Terraform S3 Backend
**Decision**: Terraform state stored in S3 bucket `tech-challenge-13-soat-tfstate` (shared with orders).
**Rationale**: Same team, same AWS account — reuse existing backend to avoid bucket sprawl.

## Risks / Trade-offs

- **AWS credentials in `.env.secrets`** → Mitigation: `.gitignore` the file; use IAM roles/IRSA in production instead of static keys.
- **SQS/SNS not provisioned by Terraform here** → Mitigation: SQS queue and SNS topic are created separately (orders service owns them or they are provisioned manually); URL/ARN are passed as environment variables.
- **Local overlay needs cluster with NGINX ingress** → Mitigation: Document `minikube addons enable ingress` or `kind` with ingress-nginx in README.
- **Single replica for local** → Acceptable for dev; AWS HPA handles production scale.

## Migration Plan

1. Create `k8s/` and `infra/` directories with manifests (no code changes)
2. For local: `minikube start`, build image locally, `kubectl apply -k k8s/overlays/local/`
3. For AWS: run `./apply-terraform.sh`, push image to ECR, update `.env.host`/`.env.secrets`, `kubectl apply -k k8s/overlays/aws/`
4. Rollback: `kubectl delete -k k8s/overlays/<env>/` — stateless manifests, rollback is instant

## Open Questions

- Will the team provision SQS/SNS via Terraform in this repo or keep them external? (currently assumed external)
- Is there a shared ECR repository or does each service get its own? (assumed per-service ECR)
