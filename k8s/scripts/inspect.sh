#!/usr/bin/env bash

JOURNALCTL_LIMIT=100000

INSPECT_DUMP=$(pwd)/inspection-report

SVC_ARGS_DIR=/var/snap/k8s/common/args
K8SD_STATE_DIR=/var/snap/k8s/common/var/lib/k8sd/state
K8SD_BIN=/snap/k8s/current/bin/k8sd
SBOM_FILE=/snap/k8s/current/bom.json

function log_success {
    printf -- '\033[32m SUCCESS: \033[0m %s\n' "$1"
}

function log_failure {
    printf -- '\033[31m FAIL: \033[0m %s\n' "$1"
}

function log_info {
	printf -- '\033[34m INFO: \033[0m %s\n' "$1"
}

function collect_args {
	local service="$1"
	mkdir -p "$INSPECT_DUMP"/"$service"
	local args_file="${SVC_ARGS_DIR}/${service#k8s.}" # Strip k8s prefix because args dir doesn't have it

	log_info "Copy $service args to the final report tarball"
	cat "$args_file" &> "$INSPECT_DUMP"/"$service"/args
}

function collect_cluster_info {
	log_info "Copy k8s cluster-info dump to the final report tarball"
	k8s kubectl cluster-info dump &> "$INSPECT_DUMP"/cluster-info
}

function collect_sbom {
	log_info "Copy SBOM to the final report tarball"
	cat $SBOM_FILE &> "$INSPECT_DUMP"/sbom.json
}

function collect_diagnostics {
	log_info "Copy uname to the final report tarball"
	uname -a &> "$INSPECT_DUMP"/uname

	log_info "Copy snap diagnostics to the final report tarball"
	snap version &> "$INSPECT_DUMP"/snap-version
	snap list k8s &> "$INSPECT_DUMP"/snap-list-k8s
	snap services k8s &> "$INSPECT_DUMP"/snap-services-k8s
	snap logs k8s -n 10000 &> "$INSPECT_DUMP"/snap-logs-k8s

	log_info "Copy k8s version and status to the final report tarball"
	k8s version &> "$INSPECT_DUMP"/k8s-version
	k8s status &> "$INSPECT_DUMP"/k8s-status
}

function collect_microcluster_db {
	log_info "Copy k8sd database dump to the final report tarball"

	$K8SD_BIN sql .dump --state-dir $K8SD_STATE_DIR &> "$INSPECT_DUMP"/k8sd-db.sql
}

function check_service {
	local service=$1
	mkdir -p "$INSPECT_DUMP"/"$service"

	local status_file="$INSPECT_DUMP/$service/systemctl.log"

    systemctl status "snap.$service" &> "$status_file"	
	
    if grep -q "active (running)" "$status_file"; then
        log_info "Service $service is running"
    else
        log_info "Service $service is not running"
    fi

	journalctl -n $JOURNALCTL_LIMIT -u "snap.$service" &> "$INSPECT_DUMP/$service/journal.log"
}

function build_report_tarball {
  # Tar and gz the report
  local now_is=$(date +"%Y%m%d_%H%M%S")
  tar -C $(pwd) -cf "$(pwd)/inspection-report-${now_is}.tar" "inspection-report" &> /dev/null
  gzip "$(pwd)/inspection-report-${now_is}.tar"
  log_success "Report tarball is at $(pwd)/inspection-report-$now_is.tar.gz"
}

if [ "$EUID" -ne 0 ]; then
	log_failure "Elevated permissions are needed for this command. Please use sudo."
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
for service in "${services[@]}"; do
	collect_args "$service"
done

printf -- 'Collecting k8s cluster-info\n'
collect_cluster_info

printf -- 'Collecting SBOM\n'
collect_sbom

printf -- 'Collecting k8sd database\n'
collect_microcluster_db

printf -- 'Gathering system information\n'
collect_diagnostics

printf -- 'Building the report tarball\n'
build_report_tarball

exit