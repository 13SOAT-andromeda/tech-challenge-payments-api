variable "api_name" {
  description = "Name of the API Gateway"
  type        = string
}

variable "subnet_ids" {
  description = "Private subnet IDs for the VPC Link"
  type        = list(string)
}

variable "security_group_ids" {
  description = "Security group IDs for the VPC Link"
  type        = list(string)
}

variable "alb_listener_arn" {
  description = "ARN of the internal ALB listener to integrate with"
  type        = string
}
