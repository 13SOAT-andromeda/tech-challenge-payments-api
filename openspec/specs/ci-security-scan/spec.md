# ci-security-scan Specification

## Purpose
TBD - created by archiving change github-actions-pipeline. Update Purpose after archive.
## Requirements
### Requirement: PR security scan trigger
The security workflow SHALL trigger on `pull_request` events targeting `main`, `develop`, `release/**`, and `feature/**` branches.

#### Scenario: PR opened against main triggers security scan
- **WHEN** a pull request is opened, synchronized, or reopened targeting `main`
- **THEN** the security workflow starts in GitHub Actions

### Requirement: gosec SAST scan
The security job SHALL install and run `gosec` against the entire Go codebase, output results in SARIF format to `gosec-report.sarif`, and upload it as a workflow artifact.

#### Scenario: gosec scan completes
- **WHEN** gosec is executed against the repository
- **THEN** a SARIF report is produced and uploaded as artifact `gosec-report.sarif`, regardless of whether vulnerabilities are found

### Requirement: govulncheck vulnerability scan
The security job SHALL install and run `govulncheck ./...` against the Go module, output results to `vulncheck-report.json`, and upload it as a workflow artifact. The step SHALL use `continue-on-error: true` so the workflow uploads the report even if vulnerabilities are found.

#### Scenario: govulncheck scan completes
- **WHEN** govulncheck is executed
- **THEN** a JSON report is produced and uploaded as artifact `vulncheck-report.json`

### Requirement: Go build before scan
The security job SHALL run `go build ./...` before security tools are executed to ensure the codebase compiles. A build failure SHALL abort the security scan.

#### Scenario: Build fails before scan
- **WHEN** `go build ./...` exits with non-zero
- **THEN** the job fails and security scan steps are skipped

