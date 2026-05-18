# k8s-local-overlay Specification

## Purpose
TBD - created by archiving change add-k8s-infrastructure. Update Purpose after archive.
## Requirements
### Requirement: Local kustomization.yaml
The system SHALL provide `k8s/overlays/local/kustomization.yaml` that extends the base, sets `imagePullPolicy: Never` (so locally-built images are used), and generates a ConfigMap from `k8s/overlays/local/config.yaml`.

#### Scenario: Local overlay kustomize build succeeds
- **WHEN** `kustomize build k8s/overlays/local/` is executed on a machine with a locally-built `payments-api:latest` image
- **THEN** manifests are rendered with `imagePullPolicy: Never` and ConfigMap sourced from the local config file

### Requirement: Local ConfigMap config file
The system SHALL provide `k8s/overlays/local/config.yaml` containing non-secret environment variables for local development: `AWS_REGION`, `SQS_QUEUE_URL_ORDER_EVENTS`, `SNS_TOPIC_ARN_PAYMENT`, `DATABASE_URL` (pointing to a local postgres), and `PORT=8080`.

#### Scenario: ConfigMap values are injected into the pod
- **WHEN** a pod starts under the local overlay
- **THEN** the container environment contains the values from `config.yaml` via the generated ConfigMap

### Requirement: Local NGINX Ingress
The system SHALL provide `k8s/overlays/local/ingress.yaml` using `ingressClassName: nginx` that routes HTTP traffic on the cluster's ingress IP to the `payments-api-svc` service on port 80.

#### Scenario: API is reachable via ingress locally
- **WHEN** NGINX ingress controller is enabled and the overlay is applied
- **THEN** an HTTP request to the ingress host reaches the payments-api pod

### Requirement: Local image build and deploy workflow
The system SHALL be deployable locally by: building the Docker image with `docker build -t payments-api:latest .`, loading it into the cluster with `minikube image load payments-api:latest` (or equivalent for kind), and applying `kubectl apply -k k8s/overlays/local/`.

#### Scenario: End-to-end local deploy
- **WHEN** the three commands (build, load, apply) are executed on a running minikube cluster with NGINX ingress addon
- **THEN** the payments-api pod reaches Running state and the `/health` endpoint responds 200 via the ingress

