variable "aws_region" {
  description = "AWS region to deploy all resources"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name used for naming and tagging all resources"
  type        = string
  default     = "tech-challenge-payments"
}

variable "lab_role_arn" {
  description = "IAM role ARN for EKS cluster and node group (FIAP lab role)"
  type        = string
}

variable "db_user" {
  description = "Master username for the RDS PostgreSQL instance"
  type        = string
  sensitive   = true
}

variable "db_pass" {
  description = "Master password for the RDS PostgreSQL instance"
  type        = string
  sensitive   = true
}
