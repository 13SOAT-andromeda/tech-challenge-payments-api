# terraform-aws-infra Specification

## Purpose
TBD - created by archiving change add-k8s-infrastructure. Update Purpose after archive.
## Requirements
### Requirement: Root Terraform configuration
The system SHALL provide `infra/main.tf` that declares an S3 backend for remote state (`bucket = "tech-challenge-13-soat-tfstate"`, key `payments/terraform.tfstate`) and composes three child modules: `network`, `eks`, and `rds`.

#### Scenario: Terraform init and plan succeed
- **WHEN** `terraform init && terraform plan` is run from the `infra/` directory with valid AWS credentials
- **THEN** no errors occur and the plan shows the resources to be created across the three modules

### Requirement: Network Terraform module
The system SHALL provide `infra/modules/network/` with a VPC, public subnets, private subnets, internet gateway, and the necessary route tables for EKS and RDS connectivity. Output variables SHALL include `vpc_id`, `public_subnet_ids`, and `private_subnet_ids`.

#### Scenario: Network module creates VPC
- **WHEN** the network module is applied
- **THEN** a VPC with configurable CIDR, at least two public and two private subnets across multiple AZs, and an internet gateway are created in AWS

### Requirement: EKS Terraform module
The system SHALL provide `infra/modules/eks/` that provisions an EKS cluster using a `lab_role_arn` variable (FIAP lab IAM role), attached to the VPC and subnets from the network module. Output variables SHALL include `cluster_name`, `cluster_endpoint`, and `cluster_security_group_id`.

#### Scenario: EKS module creates a functional cluster
- **WHEN** the eks module is applied with a valid `lab_role_arn` and VPC inputs
- **THEN** an EKS cluster is created and `aws eks update-kubeconfig` allows `kubectl` to connect to it

### Requirement: RDS Terraform module
The system SHALL provide `infra/modules/rds/` that provisions a PostgreSQL 15 RDS instance (db.t3.micro, 20GB) in private subnets with a security group that allows access only from the EKS cluster security group. Variables SHALL include `db_name` (default: `payments`), `db_user` (sensitive), `db_pass` (sensitive), `engine_version`, `instance_class`, `allocated_storage`. Output SHALL include `db_endpoint`.

#### Scenario: RDS is reachable from EKS pods
- **WHEN** the rds module is applied and a pod inside EKS attempts to connect to the `db_endpoint` on port 5432
- **THEN** the connection succeeds (security group allows EKS → RDS traffic)

#### Scenario: RDS is not reachable from the internet
- **WHEN** an external client attempts to connect to the RDS endpoint on port 5432
- **THEN** the connection is refused (RDS is in private subnets with no public access)

### Requirement: apply-terraform.sh helper script
The system SHALL provide `infra/apply-terraform.sh` that runs `terraform init` followed by `terraform apply -auto-approve` from the `infra/` directory, for use in CI/CD pipelines.

#### Scenario: Script applies infrastructure
- **WHEN** `./apply-terraform.sh` is executed with `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_DEFAULT_REGION` set in the environment
- **THEN** Terraform provisions all modules without manual intervention

