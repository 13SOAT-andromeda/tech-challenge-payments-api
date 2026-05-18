## ADDED Requirements

### Requirement: Trigger on push to main
The deploy workflow SHALL trigger automatically on every push to the `main` branch and support manual triggering via `workflow_dispatch`.

#### Scenario: Code merged to main triggers deploy
- **WHEN** a commit is pushed to the `main` branch (including PR merges)
- **THEN** the deploy workflow starts within GitHub Actions

### Requirement: Docker build and ECR push
The `build` job SHALL build the Docker image using the existing `Dockerfile`, tag it as `latest`, authenticate to ECR, and push the image. The ECR registry URL SHALL be derived from the AWS account ID and region.

#### Scenario: Successful image build and push
- **WHEN** the `build` job runs and AWS credentials are valid and the ECR repository exists
- **THEN** the image is built, tagged as `<ecr-registry>/<repo>:latest`, and pushed to ECR

### Requirement: EKS kustomize deploy
The `deploy` job SHALL update the kubeconfig for the EKS cluster, replace the `ECR_IMAGE` placeholder in `k8s/overlays/aws/kustomization.yaml` with the actual ECR URI, populate `.env.host` and `.env.secrets` from GitHub Secrets, and apply `kubectl apply -k k8s/overlays/aws/`.

#### Scenario: Successful EKS deployment
- **WHEN** the `deploy` job runs after successful `build` and `infra-check` jobs
- **THEN** the kustomize overlay is applied to EKS, and the payments-api pods are updated with the new image

#### Scenario: Deployment rollout verification
- **WHEN** `kubectl apply -k` completes
- **THEN** the job runs `kubectl rollout status deployment/payments-api` to confirm pods become ready before marking the workflow as successful

### Requirement: GitHub Secrets and Variables for deploy
The deploy workflow SHALL require the following GitHub Secrets: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`, `MERCADOPAGO_ACCESS_TOKEN`, `MERCADOPAGO_WEBHOOK_SECRET`, `MERCADOPAGO_PUBLIC_KEY`, `AWS_RDS_DB_PASSWORD`. And the following GitHub Variables: `AWS_ECR_REPOSITORY`, `AWS_EKS_CLUSTER_NAME`, `AWS_RDS_INSTANCE_ID`.

#### Scenario: Required secrets are configured
- **WHEN** all required secrets and variables are set in the repository settings
- **THEN** the workflow executes without credential errors
