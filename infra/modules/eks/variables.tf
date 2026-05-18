variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "lab_role_arn" {
  description = "IAM role ARN for the EKS cluster and node group (FIAP lab role)"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID where the EKS cluster will be deployed"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for the EKS cluster"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for EKS worker nodes"
  type        = list(string)
}
