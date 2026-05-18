## Why

The payments API has no automated CI/CD path — every deploy is manual. Following the pattern established by the sibling orders service (`tech-challenge-s1`), we need GitHub Actions workflows for automated testing, security scanning, and EKS deployment. A key addition over the reference: a pre-deploy infrastructure check that fails fast if required AWS resources (ECR repository, RDS instance) are not yet provisioned, preventing broken deploys.

## What Changes

- Add `.github/workflows/deploy.yml`: triggered on push to `main`; runs infrastructure pre-check, builds and pushes Docker image to ECR, then deploys to EKS via kustomize
- Add `.github/workflows/security.yml`: triggered on PRs; runs `gosec` SAST and `govulncheck` vulnerability scan, uploads SARIF reports as artifacts
- Add `.github/workflows/sonar.yml`: triggered on PRs; runs Go tests with coverage and uploads results to SonarCloud
- The `deploy.yml` includes a **pre-deploy infra check** job (not present in the reference) that uses the AWS CLI to assert: ECR repository exists and RDS instance is in `available` state — blocking the deploy if either is missing

## Capabilities

### New Capabilities

- `ci-security-scan`: PR workflow running gosec and govulncheck against Go source, producing SARIF artifacts
- `ci-sonar-scan`: PR workflow running Go tests with coverage and uploading to SonarCloud
- `cd-deploy`: Push-to-main workflow with build, ECR push, and EKS kustomize deploy
- `cd-infra-precheck`: Pre-deploy job that validates ECR repository and RDS instance are provisioned before attempting the deploy

### Modified Capabilities

<!-- No existing spec-level behavior changes — this is purely new CI/CD infrastructure -->

## Impact

- **New files**: `.github/workflows/deploy.yml`, `.github/workflows/security.yml`, `.github/workflows/sonar.yml`
- **GitHub Secrets required**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`, `MERCADOPAGO_ACCESS_TOKEN`, `MERCADOPAGO_WEBHOOK_SECRET`, `MERCADOPAGO_PUBLIC_KEY`, `AWS_RDS_DB_PASSWORD`, `SONAR_TOKEN`
- **GitHub Variables required**: `AWS_ECR_REPOSITORY`, `AWS_EKS_CLUSTER_NAME`, `AWS_RDS_INSTANCE_ID`
- **External services**: SonarCloud (project must be configured), ECR repository must exist before first deploy
- **No changes** to Go source code, k8s manifests, or Terraform
