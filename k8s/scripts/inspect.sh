#!/usr/bin/env bash

INSPECT_DUMP=$(pwd)/inspection-report

function log_success {
  printf -- '\033[32m SUCCESS: \033[0m %s\n' "$1"
}

function log_info {
  printf -- '\033[34m INFO: \033[0m %s\n' "$1"
}

function log_warning_red {
  printf -- '\033[31m WARNING: \033[0m %s\n' "$1"
}

function collect_args {
  log_info "Copy service args to the final report tarball"
  cp -r --no-preserve=mode,ownership /var/snap/k8s/common/args "$INSPECT_DUMP"
}

function collect_cluster_info {
  log_info "Copy k8s cluster-info dump to the final report tarball"
  k8s kubectl cluster-info dump &>"$INSPECT_DUMP"/cluster-info
}

function collect_sbom {
  log_info "Copy SBOM to the final report tarball"
  cp --no-preserve=mode,ownership /snap/k8s/current/bom.json "$INSPECT_DUMP"/sbom.json
}

function collect_diagnostics {
  log_info "Copy uname to the final report tarball"
  uname -a &>"$INSPECT_DUMP"/uname

  log_info "Copy snap diagnostics to the final report tarball"
  snap version &>"$INSPECT_DUMP"/snap-version
  snap list k8s &>"$INSPECT_DUMP"/snap-list-k8s
  snap services k8s &>"$INSPECT_DUMP"/snap-services-k8s
  snap logs k8s -n 10000 &>"$INSPECT_DUMP"/snap-logs-k8s

  log_info "Copy k8s diagnostics to the final report tarball"
  k8s version &>"$INSPECT_DUMP"/k8s-version
  k8s status &>"$INSPECT_DUMP"/k8s-status
  k8s get &>"$INSPECT_DUMP"/k8s-get
  k8s kubectl get cm k8sd-config -n kube-system -o yaml &>"$INSPECT_DUMP"/k8sd-configmap
  k8s kubectl get cm -n kube-system &>"$INSPECT_DUMP"/k8s-configmaps

  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.yaml "$INSPECT_DUMP"/k8s-dqlite-cluster.yaml
  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8s-dqlite/info.yaml "$INSPECT_DUMP"/k8s-dqlite-info.yaml
  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8sd/state/database/cluster.yaml "$INSPECT_DUMP"/k8sd-cluster.yaml
  cp --no-preserve=mode,ownership /var/snap/k8s/common/var/lib/k8sd/state/database/info.yaml "$INSPECT_DUMP"/k8sd-info.yaml
}

function check_service {
  local service=$1
  mkdir -p "$INSPECT_DUMP"/"$service"

  local status_file="$INSPECT_DUMP/$service/systemctl.log"

  systemctl status "snap.$service" &>"$status_file"

  if grep -q "active (running)" "$status_file"; then
    log_info "Service $service is running"
  else
    log_info "Service $service is not running"
  fi

  journalctl -n 100000 -u "snap.$service" &>"$INSPECT_DUMP/$service/journal.log"
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

declare -a services=("k8s.containerd" "k8s.k8s-apiserver-proxy" "k8s.k8s-dqlite" "k8s.k8sd" "k8s.kube-apiserver" "k8s.kube-controller-manager" "k8s.kube-proxy" "k8s.kube-scheduler" "k8s.kubelet")

printf -- 'Inspecting services\n'
for service in "${services[@]}"; do
  check_service "$service"
done

printf -- 'Collecting service arguments\n'
collect_args

printf -- 'Collecting k8s cluster-info\n'
collect_cluster_info

printf -- 'Collecting SBOM\n'
collect_sbom

printf -- 'Gathering system information\n'
collect_diagnostics

matches=$(grep -rlEi "BEGIN CERTIFICATE|PRIVATE KEY" inspection-report)
if [ -n "$matches" ]; then
  matches_comma_separated=$(echo "$matches" | tr '\n' ',')
  log_warning_red 'Unexpected private key or certificate found in the report:'
  log_warning_red "Found in the following files: ${matches_comma_separated%,}"
  log_warning_red 'Please remove the private key or certificate from the report before sharing.'
fi

printf -- 'Building the report tarball\n'
build_report_tarball
