#!/usr/bin/env bash

INSPECT_DUMP=$(pwd)/inspection-report

function log_success {
  printf -- '\033[32m SUCCESS: \033[0m %s\n' "$1"
}

function log_info {
  printf -- '\033[34m INFO: \033[0m %s\n' "$1"
}

function log_warning() {
  printf -- '\033[33m WARNING: \033[0m %s\n' "$1"
}

function log_warning_red {
  printf -- '\033[31m WARNING: \033[0m %s\n' "$1"
}

function is_control_plane_node {
  k8s local-node-status | grep -q "control-plane"
}

function is_service_active {
  local service
  service=$1

  systemctl status "snap.$service" | grep -q "active (running)"
}

function collect_args {
  log_info "Copy service args to the final report tarball"
  cp -r --no-preserve=mode,ownership /var/snap/k8s/common/args "$INSPECT_DUMP"
}

function collect_cluster_info {
  log_info "Copy k8s cluster-info dump to the final report tarball"
  k8s kubectl cluster-info dump --output-directory "$INSPECT_DUMP/cluster-info" &>/dev/null
}

function collect_sbom {
  log_info "Copy SBOM to the final report tarball"
  cp --no-preserve=mode,ownership /snap/k8s/current/bom.json "$INSPECT_DUMP"/sbom.json
}

function collect_k8s_diagnostics {
  log_info "Copy uname to the final report tarball"
  uname -a &>"$INSPECT_DUMP/uname.log"

  log_info "Copy snap diagnostics to the final report tarball"
  snap version &>"$INSPECT_DUMP/snap-version.log"
  snap list k8s &>"$INSPECT_DUMP/snap-list-k8s.log"
  snap services k8s &>"$INSPECT_DUMP/snap-services-k8s.log"
  snap logs k8s -n 10000 &>"$INSPECT_DUMP/snap-logs-k8s.log"

  log_info "Copy k8s diagnostics to the final report tarball"
  k8s kubectl version &>"$INSPECT_DUMP/k8s-version.log"
  k8s status &>"$INSPECT_DUMP/k8s-status.log"
  k8s get &>"$INSPECT_DUMP/k8s-get.log"
  k8s kubectl get cm k8sd-config -n kube-system -o yaml &>"$INSPECT_DUMP/k8s.k8sd/k8sd-configmap.log"
  k8s kubectl get cm -n kube-system &>"$INSPECT_DUMP/k8s-configmaps.log"

  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml "$INSPECT_DUMP/k8s.k8s-dqlite/k8s-dqlite-cluster.yaml"
  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml "$INSPECT_DUMP/k8s.k8s-dqlite/k8s-dqlite-info.yaml"
  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8sd/state/database/cluster.yaml "$INSPECT_DUMP/k8s.k8sd/k8sd-cluster.yaml"
  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8sd/state/database/info.yaml "$INSPECT_DUMP/k8s.k8sd/k8sd-info.yaml"

  ls -la /var/snap/k8s/common/var/lib/k8s-dqlite &>"$INSPECT_DUMP/k8s.k8s-dqlite/k8s-dqlite-files.log"
  ls -la /var/snap/k8s/common/var/lib/k8sd &>"$INSPECT_DUMP/k8s.k8sd/k8sd-files.log"
}

function collect_service_diagnostics {
  local service
  service=$1

  mkdir -p "$INSPECT_DUMP/$service"

  local status_file
  status_file="$INSPECT_DUMP/$service/systemctl.log"

  systemctl status "snap.$service" &>"$status_file"

  local n_restarts
  n_restarts=$(systemctl show "snap.$service" -p NRestarts | cut -d'=' -f2) 

  printf -- "%s -> %s\n" "$service" "$n_restarts" >> "$INSPECT_DUMP/nrestarts.log"

  if [ "$n_restarts" -gt 0 ]; then
    log_warning "Service $service has restarted $n_restarts times due to errors"
  fi

  journalctl -n 100000 -u "snap.$service" &>"$INSPECT_DUMP/$service/journal.log"
}

function collect_network_diagnostics {
  log_info "Copy network diagnostics to the final report tarball"
  ip a &>"$INSPECT_DUMP/ip-a.log"
  ip r &>"$INSPECT_DUMP/ip-r.log"
  iptables-save &>"$INSPECT_DUMP/iptables.log"
  ss -plnt &>"$INSPECT_DUMP/ss-plnt.log"
}

function check_expected_services {
  local services
  services=("$@")

  for service in "${services[@]}"; do
    collect_service_diagnostics "$service"
    if ! is_service_active "$service"; then
      log_info "Service $service is not running"
      log_warning "Service $service should be running on this node"
    else
      log_info "Service $service is running"
    fi
  done
}

function build_report_tarball {
  local now_is
  now_is=$(date +"%Y%m%d_%H%M%S")

  tar -C "$(pwd)" -cf "$(pwd)/inspection-report-${now_is}.tar" inspection-report &>/dev/null
  gzip "$(pwd)/inspection-report-${now_is}.tar"
  log_success "Report tarball is at $(pwd)/inspection-report-$now_is.tar.gz"
}

if [ "$EUID" -ne 0 ]; then
  printf -- "Elevated permissions are needed for this command. Please use sudo."
  exit 1
fi

rm -rf "$INSPECT_DUMP"
mkdir -p "$INSPECT_DUMP"

printf -- 'Collecting service information\n'

if is_control_plane_node; then
  printf -- 'Running inspection on a control-plane node\n'
  printf -- 'Inspection ran on a control plane node.' >"$INSPECT_DUMP/is-control-plane-node"
else
  printf -- 'Running inspection on a worker node\n'
  printf -- 'Inspection ran on a worker node.' >"$INSPECT_DUMP/is-worker-node"
fi

control_plane_services=("k8s.containerd" "k8s.kube-proxy" "k8s.k8s-dqlite" "k8s.k8sd" "k8s.kube-apiserver" "k8s.kube-controller-manager" "k8s.kube-scheduler" "k8s.kubelet")
worker_services=("k8s.containerd" "k8s.k8s-apiserver-proxy" "k8s.kubelet" "k8s.k8sd" "k8s.kube-proxy")

if is_control_plane_node; then
  check_expected_services "${control_plane_services[@]}"
else
  check_expected_services "${worker_services[@]}"
fi

printf -- 'Collecting service arguments\n'
collect_args

printf -- 'Collecting k8s cluster-info\n'
collect_cluster_info

printf -- 'Collecting SBOM\n'
collect_sbom

printf -- 'Collecting system information\n'
collect_k8s_diagnostics

printf -- 'Collecting networking information\n'
collect_network_diagnostics

matches=$(grep -rlEi "BEGIN CERTIFICATE|PRIVATE KEY" inspection-report)
if [ -n "$matches" ]; then
  matches_comma_separated=$(echo "$matches" | tr '\n' ',')
  log_warning_red 'Unexpected private key or certificate found in the report:'
  log_warning_red "Found in the following files: ${matches_comma_separated%,}"
  log_warning_red 'Please remove the private key or certificate from the report before sharing.'
fi

printf -- 'Building the report tarball\n'
build_report_tarball
