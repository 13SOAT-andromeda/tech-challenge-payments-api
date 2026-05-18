## Why

The payments API currently runs only in Docker Compose locally and has no path to production deployment. To complete the FIAP tech-challenge delivery, we need Kubernetes manifests for both a local cluster (developer testing) and AWS EKS (production), following the same kustomize overlay pattern already adopted by the sibling orders service.

## What Changes

- Add `k8s/base/` with a single `deployment.yaml`, `service.yaml`, and `hpa.yaml` for the payments API
- Add `k8s/overlays/local/` with NGINX ingress, local image pull policy, and LocalStack-compatible env config
- Add `k8s/overlays/aws/` with ALB ingress, ECR image reference, and production ConfigMap/Secret wiring
- Add `infra/` Terraform modules (`network`, `eks`, `rds`) to provision the AWS environment (EKS cluster, PostgreSQL RDS, VPC)
- Add `apply-terraform.sh` helper script for CI/CD

## Capabilities

### New Capabilities

- `k8s-base-manifests`: Core Kubernetes manifests (Deployment, Service, HPA) shared across all environments
- `k8s-local-overlay`: Kustomize overlay for running the API on a local cluster (minikube or kind) with NGINX ingress and `imagePullPolicy: Never`
- `k8s-aws-overlay`: Kustomize overlay for AWS EKS with ALB ingress, ECR image reference, and production secrets/configmap
- `terraform-aws-infra`: Terraform modules for VPC networking, EKS cluster, and RDS PostgreSQL, with S3-backed remote state

### Modified Capabilities

<!-- No existing spec-level behavior changes — this is purely new infrastructure -->

## Impact

- **New files**: entire `k8s/` and `infra/` directory trees (no changes to Go source code)
- **Dependencies**: kubectl, kustomize CLI, terraform; AWS account with permissions for EKS/RDS/VPC/ECR/ALB
- **Secrets**: MercadoPago credentials, DB password, and AWS credentials must be supplied as Kubernetes Secrets in the AWS overlay (not committed to git)
- **Existing docker-compose.yml**: untouched; local dev workflow unchanged
