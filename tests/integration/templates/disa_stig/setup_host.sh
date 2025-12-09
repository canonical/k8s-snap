#!/bin/bash
#
# Copyright 2025 Canonical, Ltd.
#
# DISA-STIG Host Setup Script
#
# This script configures a host with DISA-STIG security hardening by:
# - Installing and configuring Ubuntu Pro
# - Applying USG (Ubuntu Security Guide) fixes with custom tailoring
# - Configuring kernel parameters for Kubernetes
# - Setting up firewall rules
#
# Usage: setup_host.sh
#
# Required Environment Variables:
#   UBUNTU_PRO_TOKEN: Ubuntu Pro token for USG access
#   TAILORING_FILE: Path to the USG tailoring XML file
#   FIREWALL_PORTS: Space-separated list of ports to open in the firewall
#   UBUNTU_PRO_CONTRACT_URL: Ubuntu Pro contract server URL

set -euo pipefail

# Validate required environment variables
if [ -z "${UBUNTU_PRO_TOKEN:-}" ]; then
    echo "Error: UBUNTU_PRO_TOKEN environment variable is required" >&2
    exit 1
fi

if [ -z "${TAILORING_FILE:-}" ]; then
    echo "Error: TAILORING_FILE environment variable is required" >&2
    exit 1
fi

if [ -z "${FIREWALL_PORTS:-}" ]; then
    echo "Error: FIREWALL_PORTS environment variable is required" >&2
    exit 1
fi

if [ -z "${UBUNTU_PRO_CONTRACT_URL:-}" ]; then
    echo "Error: UBUNTU_PRO_CONTRACT_URL environment variable is required" >&2
    exit 1
fi

# Parse ports into array
read -ra PORTS <<< "${FIREWALL_PORTS}"

echo "==> Starting DISA-STIG host setup"

echo "==> Installing ubuntu-pro-client"
apt-get update -qq
apt-get install -y ubuntu-pro-client

echo "==> Configuring Ubuntu Pro contract URL"
echo "contract_url: ${UBUNTU_PRO_CONTRACT_URL}" > /etc/ubuntu-advantage/uaclient.conf

echo "==> Attaching to Ubuntu Pro"
pro attach "${UBUNTU_PRO_TOKEN}" --no-auto-enable

echo "==> Configuring kernel parameters"
cat > /etc/sysctl.d/99-kubelet.conf <<EOF
vm.overcommit_memory=1
vm.panic_on_oom=0
kernel.keys.root_maxbytes=25000000
kernel.keys.root_maxkeys=1000000
kernel.panic=10
kernel.panic_on_oops=1
EOF

echo "==> Configuring IP forwarding"
sed -i '/^net.ipv4.ip_forward/d' /etc/sysctl.conf
echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf
sysctl -w net.ipv4.ip_forward=1

echo "==> Setting up sudo access (DISA STIG requires password)"
echo 'ubuntu:mynewpassword' | chpasswd
echo 'ubuntu ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/ubuntu_nopasswd
chmod 0440 /etc/sudoers.d/ubuntu_nopasswd

echo "==> Deploying USG tailoring file"
mkdir -p /etc/usg
cp "${TAILORING_FILE}" /etc/usg/tailoring.xml
chown root:root /etc/usg/tailoring.xml
chmod 755 /etc/usg/tailoring.xml

echo "==> Enabling USG"
pro enable usg --assume-yes

echo "==> Applying USG fixes with tailoring file"
set +e
usg fix --tailoring-file /etc/usg/tailoring.xml
USG_FIX_RC=$?
echo "usg fix exited with code ${USG_FIX_RC}"
set -e

echo "==> Configuring UFW firewall"
sed -i 's/^DEFAULT_FORWARD_POLICY=.*/DEFAULT_FORWARD_POLICY="ACCEPT"/' /etc/default/ufw

# Open required ports
for port in "${PORTS[@]}"; do
    echo "==> Opening port ${port}"
    ufw allow "${port}"
done

# Allow SSH
echo "==> Allowing SSH access"
ufw allow in ssh
ufw allow in 22

# Enable UFW
echo "==> Enabling UFW"
ufw --force enable

echo "==> DISA-STIG host setup completed successfully"
