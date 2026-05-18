## Context

The payments API is deployed to AWS EKS but has no automated CI/CD. The sibling orders service (`tech-challenge-s1`) has three GitHub Actions workflows (deploy, security, sonar) that serve as the reference implementation. The payments API has a simpler secret surface (MercadoPago credentials instead of JWT/admin/Mailtrap), no Datadog integration, and already has kustomize overlays for both local and AWS from the `add-k8s-infrastructure` change. The key differentiator from the reference is a mandatory pre-deploy job that validates required AWS infrastructure is provisioned before any build or deploy step is attempted.

## Goals / Non-Goals

**Goals:**
- Automate security and code quality checks on every PR (gosec, govulncheck, SonarCloud)
- Automate Docker build + ECR push + EKS deploy on every push to `main`
- Fail fast before deploy if ECR repository or RDS instance are not `available` — preventing bad deploys and wasted build minutes
- Mirror the reference workflow structure so the team can maintain consistency across services

**Non-Goals:**
- Multi-environment (staging/prod) pipeline — single environment only
- Helm-based deploy — kustomize only
- Automatic Terraform provisioning in the pipeline — infra is provisioned separately via `apply-terraform.sh`
- Datadog operator installation — not used in the payments service
- Rollback automation — manual rollback via `kubectl rollout undo`

## Decisions

### 1. Pre-check as a separate job, not a step
**Decision**: The infrastructure pre-check runs as its own job (`infra-check`) with `needs: []`, in parallel with the `build` job but `deploy` depends on both (`needs: [infra-check, build]`).
**Rationale**: Separates concerns cleanly — a failed pre-check shows up as a distinct red job in the UI, not buried in deploy steps. Parallel execution also saves time: image build and ECR push happen while the checks run.
**Alternative considered**: A step at the start of the deploy job — simpler but mixes infra validation with deploy logic and makes it harder to diagnose failures.

### 2. ECR check via `aws ecr describe-repositories`
**Decision**: Use `aws ecr describe-repositories --repository-names $ECR_REPO` to assert the repository exists. Exit non-zero if the repository doesn't exist, blocking the pipeline.
**Rationale**: Simple, no extra tooling. The AWS CLI is already available in the runner. The repository name is stored in a GitHub Variable (`AWS_ECR_REPOSITORY`) not a secret, so it can be referenced safely.

### 3. RDS check via `aws rds describe-db-instances`
**Decision**: Use `aws rds describe-db-instances --db-instance-identifier $RDS_INSTANCE_ID` and assert `DBInstanceStatus == "available"`. Fail if the instance doesn't exist or is in any other state.
**Rationale**: Checks both existence and readiness. An RDS instance being created or in a bad state would cause the app to crash on startup; failing the pipeline is better than deploying a broken pod. The instance ID is a GitHub Variable.

### 4. Single deploy job (no worker split)
**Decision**: One `deploy` job — unlike the orders reference which conceptually handles both API and worker. Payments is a single binary.
**Rationale**: Matches the actual application structure.

### 5. No ALB wait step
**Decision**: Omit the ALB ingress wait loop from the reference deploy workflow.
**Rationale**: The ALB controller provisions the load balancer once on first deploy. Subsequent deploys don't need to wait for ALB re-provisioning. If needed on first deploy, the operator can check manually.

### 6. AWS_SESSION_TOKEN included
**Decision**: Include `AWS_SESSION_TOKEN` as a secret alongside `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.
**Rationale**: Required for FIAP lab AWS accounts which use temporary session credentials.

## Risks / Trade-offs

- **Pre-check adds ~30s to every deploy** → Acceptable cost; the check prevents broken deploys that would take longer to debug.
- **AWS_SESSION_TOKEN expires** → Lab sessions have a limited TTL; the team must refresh secrets before deploying. No automation mitigation — this is a lab constraint.
- **SonarCloud requires external project setup** → First-time setup (organization, project key) must be done manually on sonarcloud.io before the workflow runs. Document this prerequisite.
- **ECR image tag is always `latest`** → Simplifies the workflow but means there's no rollback by image tag. Mitigation: `kubectl rollout undo` is the rollback path.
- **deploy.yml has no test gate** → The deploy runs on push to `main` regardless of test results. Tests run in `sonar.yml` only on PRs. Mitigation: enforce branch protection rules requiring sonar check to pass before merging to main.

## Migration Plan

1. Configure GitHub Secrets and Variables in the repository settings (one-time)
2. Set up SonarCloud project (one-time)
3. Provision AWS infrastructure via `./infra/apply-terraform.sh` (one-time, before first deploy)
4. Create ECR repository (one-time, can be added to Terraform or done manually)
5. Merge the workflow files to main — first deploy runs automatically
6. Rollback: `kubectl rollout undo deployment/payments-api` — no pipeline change needed

## Open Questions

- Should ECR repository creation be added to the Terraform `infra/` module? (currently assumed pre-created)
- Should we add a branch protection rule on `main` requiring the sonar/security checks to pass?
