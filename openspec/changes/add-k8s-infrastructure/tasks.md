## 1. K8s Base Manifests

- [x] 1.1 Create `k8s/base/deployment.yaml` for the `payments-api` container (port 8080, envFrom ConfigMap + Secret, resource limits: 50m/200m CPU, 64Mi/128Mi memory)
- [x] 1.2 Create `k8s/base/service.yaml` ClusterIP service named `payments-api-svc` mapping port 80 → 8080
- [x] 1.3 Create `k8s/base/hpa.yaml` targeting `payments-api` deployment, min 1 / max 5 replicas, 70% CPU threshold
- [x] 1.4 Create `k8s/base/kustomization.yaml` listing deployment, service, and hpa resources

## 2. Local Kustomize Overlay

- [x] 2.1 Create `k8s/overlays/local/config.yaml` with local non-secret env vars (DATABASE_URL pointing to local postgres, SQS/SNS LocalStack or placeholder URLs, AWS_REGION, PORT=8080)
- [x] 2.2 Create `k8s/overlays/local/ingress.yaml` using `ingressClassName: nginx` routing to `payments-api-svc:80`
- [x] 2.3 Create `k8s/overlays/local/kustomization.yaml` referencing base, patching `imagePullPolicy: Never`, and generating ConfigMap from `config.yaml`
- [x] 2.4 Verify `kustomize build k8s/overlays/local/` renders all manifests without errors

## 3. AWS Kustomize Overlay

- [x] 3.1 Create `k8s/overlays/aws/.env.host.example` template with placeholder values for DB_HOST, DB_USER, DATABASE_URL, SQS_QUEUE_URL_ORDER_EVENTS, SNS_TOPIC_ARN_PAYMENT, AWS_REGION, PORT
- [x] 3.2 Add `k8s/overlays/aws/.env.host` and `k8s/overlays/aws/.env.secrets` to `.gitignore`
- [x] 3.3 Create `k8s/overlays/aws/ingress.yaml` using `ingressClassName: alb` with internet-facing and IP target type annotations
- [x] 3.4 Create `k8s/overlays/aws/kustomization.yaml` referencing base, patching image to `ECR_IMAGE`, setting `imagePullPolicy: Always`, generating ConfigMap from `.env.host` (merge), and generating Secret from `.env.secrets` (replace)
- [x] 3.5 Verify `kustomize build k8s/overlays/aws/` renders correctly when `.env.host` and `.env.secrets` are present

## 4. Terraform Network Module

- [x] 4.1 Create `infra/modules/network/main.tf` with VPC, public/private subnets (2 AZs), internet gateway, and route tables
- [x] 4.2 Create `infra/modules/network/variables.tf` with inputs: `vpc_cidr`, `availability_zones`, `project_name`
- [x] 4.3 Create `infra/modules/network/outputs.tf` exposing `vpc_id`, `public_subnet_ids`, `private_subnet_ids`

## 5. Terraform EKS Module

- [x] 5.1 Create `infra/modules/eks/main.tf` provisioning an EKS cluster and managed node group using `lab_role_arn`
- [x] 5.2 Create `infra/modules/eks/variables.tf` with inputs: `cluster_name`, `lab_role_arn`, `vpc_id`, `public_subnet_ids`, `private_subnet_ids`
- [x] 5.3 Create `infra/modules/eks/outputs.tf` exposing `cluster_name`, `cluster_endpoint`, `cluster_security_group_id`

## 6. Terraform RDS Module

- [x] 6.1 Create `infra/modules/rds/main.tf` provisioning a PostgreSQL 15 db.t3.micro RDS instance in private subnets with a security group allowing only EKS cluster SG access on port 5432
- [x] 6.2 Create `infra/modules/rds/variables.tf` with inputs: `db_name` (default: payments), `db_user` (sensitive), `db_pass` (sensitive), `engine_version` (default: 15), `instance_class` (default: db.t3.micro), `allocated_storage` (default: 20), `vpc_id`, `private_subnet_ids`, `eks_cluster_security_group_id`
- [x] 6.3 Create `infra/modules/rds/outputs.tf` exposing `db_endpoint`

## 7. Terraform Root and Scripts

- [x] 7.1 Create `infra/main.tf` with S3 backend (`bucket = "tech-challenge-13-soat-tfstate"`, key `payments/terraform.tfstate`) and compose network, eks, rds modules
- [x] 7.2 Create `infra/variables.tf` with top-level variables: `aws_region`, `project_name`, `lab_role_arn`, `db_user`, `db_pass`
- [x] 7.3 Create `infra/apply-terraform.sh` (chmod +x) running `terraform init && terraform apply -auto-approve`
- [x] 7.4 Add `infra/.terraform/`, `infra/*.tfstate`, `infra/*.tfstate.backup`, and `infra/.terraform.lock.hcl` to `.gitignore`

## 8. Documentation

- [x] 8.1 Add a `## Kubernetes Deployment` section to `README.md` documenting the local workflow (minikube, image build/load, kustomize apply) and the AWS workflow (ECR push, env file setup, kustomize apply)
- [x] 8.2 Add a `## Infrastructure (Terraform)` section to `README.md` documenting prerequisites and how to run `apply-terraform.sh`
