terraform {
  required_version = ">= 1.5"

  backend "s3" {
    bucket = "tech-challenge-13-soat-tfstate"
    key    = "payments/terraform.tfstate"
    region = "us-east-1"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

module "network" {
  source = "./modules/network"

  project_name       = var.project_name
  availability_zones = ["${var.aws_region}a", "${var.aws_region}b"]
}

module "eks" {
  source = "./modules/eks"

  cluster_name       = var.project_name
  lab_role_arn       = var.lab_role_arn
  vpc_id             = module.network.vpc_id
  public_subnet_ids  = module.network.public_subnet_ids
  private_subnet_ids = module.network.private_subnet_ids
}

module "rds" {
  source = "./modules/rds"

  db_name                       = "payments"
  db_user                       = var.db_user
  db_pass                       = var.db_pass
  vpc_id                        = module.network.vpc_id
  private_subnet_ids            = module.network.private_subnet_ids
  eks_cluster_security_group_id = module.eks.cluster_security_group_id
}

output "eks_cluster_name" {
  value = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "rds_endpoint" {
  value     = module.rds.db_endpoint
  sensitive = true
}
