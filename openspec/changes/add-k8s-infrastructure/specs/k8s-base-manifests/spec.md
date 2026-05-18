## ADDED Requirements

### Requirement: API Deployment manifest
The system SHALL provide a Kubernetes Deployment manifest at `k8s/base/deployment.yaml` that runs the `payments-api` container on port 8080, reads all configuration from a ConfigMap (`payments-api-config`) and a Secret (`payments-api-secrets`), and sets resource requests/limits suitable for the workload.

#### Scenario: Deployment is applied to a cluster
- **WHEN** `kubectl apply -k k8s/base/` is executed
- **THEN** a Deployment named `payments-api` with 1 replica is created, with the container image `payments-api:latest`, port 8080 exposed, and env sourced from ConfigMap `payments-api-config` and Secret `payments-api-secrets`

#### Scenario: Resource limits are defined
- **WHEN** the Deployment manifest is inspected
- **THEN** the container SHALL have CPU request 50m / limit 200m and memory request 64Mi / limit 128Mi

### Requirement: ClusterIP Service manifest
The system SHALL provide a Kubernetes Service manifest at `k8s/base/service.yaml` of type ClusterIP that routes port 80 to container port 8080.

#### Scenario: Service routes traffic to the API pod
- **WHEN** a pod inside the cluster sends HTTP to `payments-api-svc:80`
- **THEN** the request is forwarded to the payments-api pod on port 8080

### Requirement: HPA manifest
The system SHALL provide a HorizontalPodAutoscaler manifest at `k8s/base/hpa.yaml` targeting the `payments-api` Deployment with min 1, max 5 replicas, scaling at 70% average CPU utilization.

#### Scenario: CPU exceeds threshold
- **WHEN** average CPU utilization of payments-api pods exceeds 70%
- **THEN** the HPA scales the Deployment up (max 5 replicas)

#### Scenario: CPU drops below threshold
- **WHEN** average CPU utilization falls below 70%
- **THEN** the HPA scales the Deployment down (min 1 replica)

### Requirement: Base kustomization.yaml
The system SHALL provide `k8s/base/kustomization.yaml` listing all base resources (deployment, service, hpa) so overlays can extend them.

#### Scenario: Kustomize build from base
- **WHEN** `kustomize build k8s/base/` is run
- **THEN** all three manifests (Deployment, Service, HPA) are rendered without errors
