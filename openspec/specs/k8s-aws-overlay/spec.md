# k8s-aws-overlay Specification

## Purpose
TBD - created by archiving change add-k8s-infrastructure. Update Purpose after archive.
## Requirements
### Requirement: AWS kustomization.yaml
The system SHALL provide `k8s/overlays/aws/kustomization.yaml` that extends the base, patches the image to use an ECR URI (placeholder `ECR_IMAGE`), sets `imagePullPolicy: Always`, generates a ConfigMap from `.env.host` (merge behavior), and generates a Secret from `.env.secrets` (replace behavior).

#### Scenario: AWS overlay kustomize build succeeds
- **WHEN** `kustomize build k8s/overlays/aws/` is run with `.env.host` and `.env.secrets` present
- **THEN** manifests are rendered with the ECR image reference, `imagePullPolicy: Always`, and the ConfigMap and Secret populated from the env files

### Requirement: AWS ConfigMap host env file
The system SHALL provide `k8s/overlays/aws/.env.host` (not committed to git; template committed as `.env.host.example`) containing: `DB_HOST`, `DB_USER`, `DATABASE_URL`, `SQS_QUEUE_URL_ORDER_EVENTS`, `SNS_TOPIC_ARN_PAYMENT`, `AWS_REGION`, and `PORT=8080`.

#### Scenario: ConfigMap is generated from .env.host
- **WHEN** kustomize processes the AWS overlay with a populated `.env.host`
- **THEN** a ConfigMap `payments-api-config` is created with the key-value pairs from the file

### Requirement: AWS Secret env file
The system SHALL provide `k8s/overlays/aws/.env.secrets` (never committed to git) containing sensitive values: `MERCADOPAGO_ACCESS_TOKEN`, `MERCADOPAGO_WEBHOOK_SECRET`, `MERCADOPAGO_PUBLIC_KEY`, `DB_PASSWORD`, `AWS_ACCESS_KEY_ID`, and `AWS_SECRET_ACCESS_KEY`.

#### Scenario: Secret is generated from .env.secrets
- **WHEN** kustomize processes the AWS overlay with a populated `.env.secrets`
- **THEN** a Secret `payments-api-secrets` is created with base64-encoded values, not visible in plain-text in any committed file

### Requirement: ALB Ingress
The system SHALL provide `k8s/overlays/aws/ingress.yaml` using `ingressClassName: alb` with annotations for internet-facing scheme and IP target type, routing HTTP traffic to `payments-api-svc` on port 80.

#### Scenario: API is reachable via ALB in AWS
- **WHEN** the AWS overlay is applied to an EKS cluster with the AWS Load Balancer Controller installed
- **THEN** an ALB is provisioned and HTTP requests to the ALB DNS name reach the payments-api pods

### Requirement: ECR image placeholder and replacement instructions
The system SHALL include comments or a README note explaining how to replace the `ECR_IMAGE` placeholder with the actual ECR URI before applying the AWS overlay (e.g., `sed -i 's|ECR_IMAGE|<account>.dkr.ecr.us-east-1.amazonaws.com/payments-api:latest|g' k8s/overlays/aws/kustomization.yaml`).

#### Scenario: Operator replaces ECR image placeholder
- **WHEN** the operator follows the documented sed command with their ECR URI
- **THEN** `kustomize build k8s/overlays/aws/` renders a valid image reference that EKS can pull from ECR

