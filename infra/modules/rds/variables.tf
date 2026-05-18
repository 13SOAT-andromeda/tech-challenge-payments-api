variable "db_name" {
  description = "Name of the database and RDS identifier"
  type        = string
  default     = "payments"
}

variable "db_user" {
  description = "Master username for the RDS instance"
  type        = string
  sensitive   = true
}

variable "db_pass" {
  description = "Master password for the RDS instance"
  type        = string
  sensitive   = true
}

variable "engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15"
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "allocated_storage" {
  description = "Storage allocated to the RDS instance in GB"
  type        = number
  default     = 20
}

variable "vpc_id" {
  description = "VPC ID where the RDS instance will be deployed"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for the RDS subnet group"
  type        = list(string)
}

variable "eks_cluster_security_group_id" {
  description = "Security group ID of the EKS cluster, allowed to access RDS"
  type        = string
}
