# cd-infra-precheck Specification

## Purpose
TBD - created by archiving change github-actions-pipeline. Update Purpose after archive.
## Requirements
### Requirement: ECR repository existence check
The pipeline SHALL verify that the ECR repository named by the `AWS_ECR_REPOSITORY` GitHub Variable exists before any Docker build or deploy step executes. If the repository does not exist, the job SHALL fail with a descriptive error message.

#### Scenario: ECR repository exists
- **WHEN** the `infra-check` job runs and the ECR repository specified in `AWS_ECR_REPOSITORY` exists in the target AWS account and region
- **THEN** the ECR check step exits with code 0 and the job continues

#### Scenario: ECR repository does not exist
- **WHEN** the `infra-check` job runs and the ECR repository does not exist
- **THEN** the ECR check step exits with a non-zero code and the job is marked as failed, blocking the dependent deploy job

### Requirement: RDS instance availability check
The pipeline SHALL verify that the RDS instance identified by the `AWS_RDS_INSTANCE_ID` GitHub Variable exists and has `DBInstanceStatus == "available"` before deploy executes. Any other status (creating, modifying, stopped, etc.) SHALL cause the job to fail.

#### Scenario: RDS instance is available
- **WHEN** the `infra-check` job runs and the RDS instance exists with status `available`
- **THEN** the RDS check step exits with code 0 and the job continues

#### Scenario: RDS instance does not exist
- **WHEN** the `infra-check` job runs and no RDS instance with the given identifier exists
- **THEN** the RDS check step exits with a non-zero code and the job fails

#### Scenario: RDS instance exists but is not available
- **WHEN** the `infra-check` job runs and the RDS instance exists with a status other than `available` (e.g., `creating`, `stopped`)
- **THEN** the RDS check step exits with a non-zero code, prints the current status, and the job fails

### Requirement: Pre-check runs in parallel with build
The `infra-check` job SHALL have no `needs` dependency on the `build` job, allowing both to run concurrently. The `deploy` job SHALL declare `needs: [infra-check, build]` so it only proceeds when both succeed.

#### Scenario: Both infra-check and build succeed
- **WHEN** the `infra-check` job and the `build` job both complete successfully
- **THEN** the `deploy` job starts execution

#### Scenario: infra-check fails while build succeeds
- **WHEN** the `infra-check` job fails (e.g., RDS not available) even if `build` succeeds
- **THEN** the `deploy` job is skipped and the overall workflow is marked as failed

