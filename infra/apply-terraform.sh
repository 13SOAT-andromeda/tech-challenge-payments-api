#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "==> Initializing Terraform..."
terraform init

echo "==> Applying Terraform..."
terraform apply -auto-approve

echo "==> Done. Updating kubeconfig..."
CLUSTER_NAME=$(terraform output -raw eks_cluster_name)
aws eks update-kubeconfig --name "$CLUSTER_NAME" --region "${AWS_DEFAULT_REGION:-us-east-1}"
echo "==> kubeconfig updated for cluster: $CLUSTER_NAME"
