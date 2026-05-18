## 1. Security Scan Workflow

- [x] 1.1 Create `.github/workflows/security.yml` with `pull_request` trigger on `main`, `develop`, `release/**`, `feature/**`
- [x] 1.2 Add `setup-go@v5` step (Go 1.25) and `go build ./...` step to verify compilation
- [x] 1.3 Add step to install and run `gosec` with SARIF output to `gosec-report.sarif`
- [x] 1.4 Add step to install and run `govulncheck ./...` with JSON output to `vulncheck-report.json` (continue-on-error: true)
- [x] 1.5 Add `actions/upload-artifact` steps for both `gosec-report.sarif` and `vulncheck-report.json`

## 2. SonarCloud Scan Workflow

- [x] 2.1 Create `.github/workflows/sonar.yml` with `pull_request` trigger on `main`, `develop`, `release/**`, `feature/**`
- [x] 2.2 Add `actions/cache` step for Go module cache (keyed on `go.sum` hash)
- [x] 2.3 Add `setup-go@v5` and `go mod download` steps
- [x] 2.4 Add `go test ./... -coverprofile=coverage.out` step
- [x] 2.5 Add `SonarSource/sonarcloud-github-action` step using `SONAR_TOKEN` secret and pointing coverage to `coverage.out`
- [x] 2.6 Create `sonar-project.properties` file at repo root with project key, organization, and coverage report path

## 3. Deploy Workflow — Infrastructure Pre-check Job

- [x] 3.1 Create `.github/workflows/deploy.yml` with `push` trigger on `main` and `workflow_dispatch`
- [x] 3.2 Add `infra-check` job with `aws-actions/configure-aws-credentials@v4` using `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION` secrets
- [x] 3.3 Add ECR check step: run `aws ecr describe-repositories --repository-names ${{ vars.AWS_ECR_REPOSITORY }}` and fail with descriptive message if repository not found
- [x] 3.4 Add RDS check step: run `aws rds describe-db-instances --db-instance-identifier ${{ vars.AWS_RDS_INSTANCE_ID }}`, extract `DBInstanceStatus`, and fail if status is not `available`

## 4. Deploy Workflow — Build Job

- [x] 4.1 Add `build` job (parallel to `infra-check`) with `aws-actions/configure-aws-credentials@v4`
- [x] 4.2 Add `aws-actions/amazon-ecr-login@v2` step to authenticate to ECR
- [x] 4.3 Add Docker build step: `docker build -t $ECR_REGISTRY/${{ vars.AWS_ECR_REPOSITORY }}:latest .`
- [x] 4.4 Add Docker push step: `docker push $ECR_REGISTRY/${{ vars.AWS_ECR_REPOSITORY }}:latest`

## 5. Deploy Workflow — Deploy Job

- [x] 5.1 Add `deploy` job with `needs: [infra-check, build]` and `aws-actions/configure-aws-credentials@v4`
- [x] 5.2 Add `aws eks update-kubeconfig` step using `vars.AWS_EKS_CLUSTER_NAME`
- [x] 5.3 Add step to replace `ECR_IMAGE` placeholder in `k8s/overlays/aws/kustomization.yaml` with the actual ECR URI using `sed`
- [x] 5.4 Add step to generate `k8s/overlays/aws/.env.host` from GitHub Variables (`AWS_REGION`, `AWS_RDS_INSTANCE_ID` for endpoint lookup, `SQS_QUEUE_URL_ORDER_EVENTS`, `SNS_TOPIC_ARN_PAYMENT`)
- [x] 5.5 Add step to generate `k8s/overlays/aws/.env.secrets` from GitHub Secrets (`MERCADOPAGO_ACCESS_TOKEN`, `MERCADOPAGO_WEBHOOK_SECRET`, `MERCADOPAGO_PUBLIC_KEY`, `AWS_RDS_DB_PASSWORD`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- [x] 5.6 Add `kubectl apply -k k8s/overlays/aws/` step
- [x] 5.7 Add `kubectl rollout status deployment/payments-api --timeout=120s` step to verify rollout succeeds

## 6. Documentation

- [x] 6.1 Add a `## CI/CD` section to `README.md` documenting the three workflows, required GitHub Secrets/Variables, and the pre-deploy infra check behavior
- [x] 6.2 Document the one-time setup steps: SonarCloud project creation, ECR repository creation, and how to refresh `AWS_SESSION_TOKEN` for lab accounts
