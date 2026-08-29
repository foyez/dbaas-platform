#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="${ROOT_DIR}/tf"
ANSIBLE_DIR="${ROOT_DIR}/ansible"
INVENTORY="${ANSIBLE_DIR}/inventory.ini"
PRIVATE_KEY="${HOME}/.ssh/id_ed25519"

export OS_CLIENT_CONFIG_FILE="${TF_DIR}/clouds.yaml"
export ANSIBLE_CONFIG="${ANSIBLE_DIR}/ansible.cfg"

echo "=================================="
echo " OpenStack Kubernetes Deployment"
echo "=================================="

echo "==> Initialising Terraform"
terraform -chdir="$TF_DIR" init

echo "==> Planning infrastructure"
terraform -chdir="$TF_DIR" plan -out=tfplan

echo "==> Applying infrastructure"
terraform -chdir="$TF_DIR" apply -auto-approve tfplan

echo "==> Retrieving control plane floating IP"
FLOATING_IP="$(terraform -chdir="$TF_DIR" output -raw instance_floating_ip)"

echo "==> Waiting for SSH on ${FLOATING_IP}"
until ssh \
    -i "$PRIVATE_KEY" \
    -o BatchMode=yes \
    -o ConnectTimeout=5 \
    -o StrictHostKeyChecking=no \
    -o UserKnownHostsFile=/dev/null \
    "ubuntu@${FLOATING_IP}" \
    "true" >/dev/null 2>&1
do
    echo "Waiting for SSH..."
    sleep 5
done
echo "SSH is available."

echo "==> Testing Ansible connection"
ansible \
    -i "$INVENTORY" \
    kubernetes \
    -m ping

echo "==> Installing Kubernetes with Ansible"
ansible-playbook \
    -i "$INVENTORY" \
    "${ANSIBLE_DIR}/playbook.yaml"

echo
echo "=================================="
echo " Kubernetes deployment completed"
echo "=================================="
