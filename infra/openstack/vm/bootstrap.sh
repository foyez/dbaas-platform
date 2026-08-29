#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="${ROOT_DIR}/tf/openstack"

export OS_CLIENT_CONFIG_FILE="${TF_DIR}/clouds.yaml"

echo "=================================="
echo " OpenStack VM Provision           "
echo "=================================="

echo "==> Initialising Terraform"
terraform -chdir="$TF_DIR" init

echo "==> Planning infrastructure"
terraform -chdir="$TF_DIR" plan -out=tfplan

echo "==> Applying infrastructure"
terraform -chdir="$TF_DIR" apply -auto-approve tfplan

