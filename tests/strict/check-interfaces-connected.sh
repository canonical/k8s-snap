#!/usr/bin/env bash

set -ue

if [ "$EUID" -ne 0 ]
then echo "Please run this script as root."
  exit 1
fi

declare -a INTERFACES=(
  docker-privileged
  kubernetes-support
  network
  network-bind
  network-control
  network-observe
  firewall-control
  process-control
  kernel-module-observe
  cilium-module-load
  mount-observe
  hardware-observe
  system-observe
  home
  opengl
  home-read-all
  login-session-observe
  log-observe
  hi
)

for if in "${INTERFACES[@]}"; do
    sudo snap run --shell k8s -c "snapctl is-connected $if"
done
