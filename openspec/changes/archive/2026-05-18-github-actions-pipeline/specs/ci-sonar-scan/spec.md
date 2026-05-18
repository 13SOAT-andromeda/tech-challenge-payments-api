## ADDED Requirements

### Requirement: PR SonarCloud trigger
The sonar workflow SHALL trigger on `pull_request` events targeting `main`, `develop`, `release/**`, and `feature/**` branches.

#### Scenario: PR triggers SonarCloud scan
- **WHEN** a pull request is opened, synchronized, or reopened targeting any of the configured branches
- **THEN** the sonar workflow starts in GitHub Actions

### Requirement: Go test coverage report
The sonar job SHALL run `go test ./... -coverprofile=coverage.out` and produce a `coverage.out` file that is passed to the SonarCloud scanner.

#### Scenario: Tests pass and coverage is collected
- **WHEN** `go test ./...` exits with code 0
- **THEN** a `coverage.out` file exists and is available to the SonarCloud scan step

#### Scenario: Tests fail
- **WHEN** `go test ./...` exits with a non-zero code
- **THEN** the job fails and SonarCloud scan is skipped

### Requirement: SonarCloud scan
The sonar job SHALL run the SonarCloud GitHub Action (`SonarSource/sonarcloud-github-action`) with the `SONAR_TOKEN` secret, pointing coverage to `coverage.out`. The SonarCloud organization SHALL be configurable via the workflow (not hardcoded to the orders service).

#### Scenario: SonarCloud scan completes successfully
- **WHEN** the SonarCloud action runs with a valid `SONAR_TOKEN` and configured project
- **THEN** analysis results appear in the SonarCloud dashboard and the step exits with code 0

### Requirement: Go module cache
The sonar job SHALL cache Go module dependencies between runs using `actions/cache` to reduce build time.

#### Scenario: Cache hit on repeated runs
- **WHEN** the sonar job runs a second time with no go.mod/go.sum changes
- **THEN** the cache is restored and `go mod download` is skipped or faster
