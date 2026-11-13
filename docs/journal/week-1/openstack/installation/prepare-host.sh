#!/bin/bash
# prepare-host.sh
# Run as root (or sudo)

set -euo pipefail

STACK_USER=stack
STACK_HOME=/opt/stack

echo "▶ Installing dependencies"
apt update
apt -y install git curl sudo

if ! id "$STACK_USER" &>/dev/null; then
    echo "▶ Creating stack user"
    useradd -s /bin/bash -d $STACK_HOME -m $STACK_USER

    # Allow password less sudo
    echo "$STACK_USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$STACK_USER
    chmod 0440 /etc/sudoers.d/$STACK_USER

    # Create .ssh for $STACK_USER
    mkdir -p /opt/stack/.ssh
    cp /root/.ssh/authorized_keys /opt/stack/.ssh/
    chown -R stack:stack /opt/stack/.ssh
    chmod 700 /opt/stack/.ssh
    chmod 600 /opt/stack/.ssh/authorized_keys
fi

echo "▶ Host preparation complete"
echo "Now switch user:"
echo "sudo su - stack"