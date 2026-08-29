#!/usr/bin/env bash

set -euo pipefail

DEVSTACK_DIR="/opt/stack/devstack"
STACK_HOME="/opt/stack"

echo "=========================================="
echo " DevStack Installation"
echo "=========================================="

# Update system
sudo apt-get update
sudo apt-get upgrade -y

# Create stack user if it doesn't exist
if ! id stack &>/dev/null; then
    echo "Creating stack user..."
    sudo useradd -s /bin/bash -d "${STACK_HOME}" -m stack
fi
# Required on Ubuntu 21.04+ and some RHEL distributions
sudo chmod +x "${STACK_HOME}"

# Configure passwordless sudo
if [ ! -f /etc/sudoers.d/stack ]; then
    echo "Granting sudo privileges..."
    echo "stack ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/stack >/dev/null
    sudo chmod 440 /etc/sudoers.d/stack
fi

# Clone DevStack
if [ ! -d "${DEVSTACK_DIR}" ]; then
    echo "Cloning DevStack..."
    sudo -u stack -i git clone https://opendev.org/openstack/devstack.git "${DEVSTACK_DIR}"
fi

# Copy local.conf
if [ ! -f "./local.conf" ]; then
    echo "ERROR: local.conf not found in current directory."
    exit 1
fi

sudo cp ./local.conf "${DEVSTACK_DIR}/local.conf"
sudo chown stack:stack "${DEVSTACK_DIR}/local.conf"

echo "Starting DevStack installation..."
echo "This may take 20-40 minutes."

sudo -u stack -i bash -c "
cd ${DEVSTACK_DIR}
./stack.sh
"

echo
echo "======================================"
echo "Installation complete!"
echo
echo "Horizon Dashboard:"
echo "http://<YOUR_VM_IP>/dashboard"
echo
echo "Admin credentials:"
echo "source /opt/stack/devstack/openrc admin admin"
echo "======================================"
