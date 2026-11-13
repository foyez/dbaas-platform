#!/bin/bash
# install-devstack.sh
# Run as: stack user

set -euo pipefail

cd ~

# Clone DevStack if not present
if [ ! -d devstack ]; then
    echo "▶ Cloning DevStack"
    git clone https://opendev.org/openstack/devstack
fi

cd devstack

echo "▶ Creating local.conf"

HOST_IP=$(hostname -I | awk '{print $1}')

cat > local.conf <<EOF
# Sample ``local.conf`` for user-configurable variables in ``stack.sh``

# NOTE: Copy this file to the root DevStack directory for it to work properly.

# ``local.conf`` is a user-maintained settings file that is sourced from ``stackrc``.
# This gives it the ability to override any variables set in ``stackrc``.
# Also, most of the settings in ``stack.sh`` are written to only be set if no
# value has already been set; this lets ``local.conf`` effectively override the
# default values.

# This is a collection of some of the settings we have found to be useful
# in our DevStack development environments. Additional settings are described
# in https://docs.openstack.org/devstack/latest/configuration.html#local-conf
# These should be considered as samples and are unsupported DevStack code.

# The ``localrc`` section replaces the old ``localrc`` configuration file.
# Note that if ``localrc`` is present it will be used in favor of this section.
[[local|localrc]]

# ─────────────────────────────
# Basic Configuration
# ─────────────────────────────

# While ``stack.sh`` is happy to run without ``localrc``, devlife is better when
# there are a few minimal variables set:

# If the ``*_PASSWORD`` variables are not set here you will be prompted to enter
# values for them by ``stack.sh``and they will be added to ``local.conf``.
DEST=/opt/stack
ADMIN_PASSWORD=devstack
DATABASE_PASSWORD=\$ADMIN_PASSWORD
RABBIT_PASSWORD=\$ADMIN_PASSWORD
SERVICE_PASSWORD=\$ADMIN_PASSWORD

# ``HOST_IP`` and ``HOST_IPV6`` should be set manually for best results if
# the NIC configuration of the host is unusual, i.e. ``eth1`` has the default
# route but ``eth0`` is the public interface.  They are auto-detected in
# ``stack.sh`` but often is indeterminate on later runs due to the IP moving
# from an Ethernet interface to a bridge on the host. Setting it here also
# makes it available for ``openrc`` to include when setting ``OS_AUTH_URL``.
# Neither is set by default.
#HOST_IP=w.x.y.z
#HOST_IPV6=2001:db8::7
HOST_IP=$HOST_IP


# ─────────────────────────────
# Logging
# ─────────────────────────────

# By default ``stack.sh`` output only goes to the terminal where it runs.  It can
# be configured to additionally log to a file by setting ``LOGFILE`` to the full
# path of the destination log file.  A timestamp will be appended to the given name.
LOGFILE=\$DEST/logs/stack.sh.log

# Old log files are automatically removed after 2 days to keep things neat.  Change
# the number of days by setting ``LOGDAYS``.
LOGDAYS=2

# Nova logs will be colorized if ``SYSLOG`` is not set; turn this off by setting
# ``LOG_COLOR`` false.
#LOG_COLOR=False

# ─────────────────────────────
# Core Services
# ─────────────────────────────
enable_service key mysql rabbit

# ─────────────────────────────
# Nova (Compute)
# ─────────────────────────────
enable_service n-api n-cond n-sch n-cpu n-novnc

# ─────────────────────────────
# Glance (Images)
# ─────────────────────────────
enable_service g-api

# ─────────────────────────────
# Neutron (OVN)
# ─────────────────────────────
enable_service neutron q-svc
enable_service ovn-northd ovn-controller q-ovn-metadata-agent
disable_service q-l3
disable_service q-dhcp

# ─────────────────────────────
# Cinder
# ─────────────────────────────
enable_service c-api c-vol c-sch

# ─────────────────────────────
# Placement
# ─────────────────────────────
enable_service placement-api

# ─────────────────────────────
# Horizon
# ─────────────────────────────
enable_service horizon

# Using milestone-proposed branches
# ---------------------------------

# Uncomment these to grab the milestone-proposed branches from the
# repos:
#CINDER_BRANCH=milestone-proposed
#GLANCE_BRANCH=milestone-proposed
#HORIZON_BRANCH=milestone-proposed
#KEYSTONE_BRANCH=milestone-proposed
#KEYSTONECLIENT_BRANCH=milestone-proposed
#NOVA_BRANCH=milestone-proposed
#NOVACLIENT_BRANCH=milestone-proposed
#NEUTRON_BRANCH=milestone-proposed
#SWIFT_BRANCH=milestone-proposed

# Using git versions of clients
# -----------------------------
# By default clients are installed from pip.  See LIBS_FROM_GIT in
# stackrc for details on getting clients from specific branches or
# revisions.  e.g.
# LIBS_FROM_GIT="python-ironicclient"
# IRONICCLIENT_BRANCH=refs/changes/44/2.../1

# Swift
# -----

# Swift is now used as the back-end for the S3-like object store. Setting the
# hash value is required and you will be prompted for it if Swift is enabled
# so just set it to something already:
SWIFT_HASH=66a3d6b56c1f479c8b4e70ab5c2000f5

# For development purposes the default of 3 replicas is usually not required.
# Set this to 1 to save some resources:
SWIFT_REPLICAS=1

# The data for Swift is stored by default in (``\$DEST/data/swift``),
# or (``\$DATA_DIR/swift``) if ``DATA_DIR`` has been set, and can be
# moved by setting ``SWIFT_DATA_DIR``. The directory will be created
# if it does not exist.
SWIFT_DATA_DIR=\$DEST/data
EOF

echo "▶ local.conf created"

echo "▶ Starting DevStack installation"
./stack.sh