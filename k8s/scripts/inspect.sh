#!/usr/bin/env bash

JOURNALCTL_LIMIT=100000

INSPECT_DUMP=$(pwd)/inspection-report

SVC_ARGS_DIR=/var/snap/k8s/common/args
K8SD_STATE_DIR=/var/snap/k8s/common/var/lib/k8sd/state
K8SD_BIN=/snap/k8s/current/bin/k8sd
SBOM_FILE=/snap/k8s/current/bom.json

function log_success() {
    printf -- '\033[32m SUCCESS: \033[0m %s\n' "$1"
}

function log_failure() {
    printf -- '\033[31m FAIL: \033[0m %s\n' "$1"
}

function collect_args {
	local service="$1"
	mkdir -p "$INSPECT_DUMP"/"$service"
	local args_file="${SVC_ARGS_DIR}/${service#k8s.}" # Strip k8s prefix because args dir doesn't have it

	if [ -e "$args_file" ]; then
		cat "$args_file" &> "$INSPECT_DUMP"/"$service"/args
		log_success "Found arguments for $service"
	else
		log_failure "Arguments for $service not found"
	fi
}

function collect_cluster_info {
	k8s kubectl cluster-info dump &> "$INSPECT_DUMP"/cluster-info

	log_success "Collected k8s cluster-info"
}

function collect_sbom {
	cat $SBOM_FILE &> "$INSPECT_DUMP"/sbom.json

	log_success "Collected SBOM"
}

function collect_diagnostics {
	snap version &> "$INSPECT_DUMP"/snap-version
	uname -a &> "$INSPECT_DUMP"/uname
	snap list k8s &> "$INSPECT_DUMP"/snap-list-k8s
	snap services k8s &> "$INSPECT_DUMP"/snap-services-k8s
	snap logs k8s -n 10000 &> "$INSPECT_DUMP"/snap-logs-k8s
	k8s status &> "$INSPECT_DUMP"/k8s-status
}

function collect_microcluster_db {
	$K8SD_BIN sql .dump --state-dir $K8SD_STATE_DIR &> "$INSPECT_DUMP"/k8sd-db.sql

	log_success "Collected k8sd database dump"
}

function check_service {
	local service=$1
	mkdir -p "$INSPECT_DUMP"/"$service"

	local status_file="$INSPECT_DUMP/$service/systemctl.log"

    systemctl status "snap.$service" &> "$status_file"	
	
    if grep -q "active (running)" "$status_file"; then
        log_success "Service $service is running"
    else
        log_failure "Service $service is not running"
        echo "For more details look at: sudo journalctl -u snap.$service"
    fi

	journalctl -n $JOURNALCTL_LIMIT -u "snap.$service" &> "$INSPECT_DUMP/$service/journal.log"
}

if [ "$EUID" -ne 0 ]; then
	echo "Elevated permissions are needed for this command. Please use sudo."
	exit 1
fi

rm -rf "$INSPECT_DUMP"
mkdir -p "$INSPECT_DUMP"

declare -a services=("k8s.containerd" "k8s.k8s-apiserver-proxy" "k8s.k8s-dqlite" "k8s.k8sd" "k8s.kube-apiserver" "k8s.kube-controller-manager" "k8s.kube-proxy" "k8s.kube-scheduler" "k8s.kubelet")

printf -- 'Inspecting services\n'
for service in "${services[@]}"; do
    check_service "$service"
done
printf -- '\n'

printf -- 'Collecting arguments\n'
if [ ! -d $SVC_ARGS_DIR ]; then
	log_failure "Arguments directory not found"
else
	for service in "${services[@]}"; do
		collect_args "$service"
	done
fi
printf -- '\n'

printf -- 'Collecting k8s cluster-info\n'
if ! k8s &> /dev/null; then
	log_failure "k8s command not found"
else
	collect_cluster_info
fi
printf -- '\n'

printf -- 'Collecting SBOM\n'
if [ ! -f $SBOM_FILE ]; then
	log_failure "SBOM file not found"
else
	collect_sbom
fi
printf -- '\n'

printf -- 'Collecting k8sd database dump\n'
if [ ! -d $K8SD_STATE_DIR ]; then
	log_failure "k8sd state directory not found"
elif [ ! -f $K8SD_BIN ]; then
	log_failure "k8sd binary not found"
else
	collect_microcluster_db
fi
printf -- '\n'

printf -- 'Collecting general diagnostics\n'
if ! k8s &> /dev/null; then
	log_failure "k8s command not found"
elif ! snap version &> /dev/null; then
	log_failure "snap command not found"
else
	collect_diagnostics
fi
printf -- '\n'

printf -- 'Creating inspection tarball\n'
if [ ! -d "$INSPECT_DUMP" ]; then
	log_failure "No inspection-dump folder found. Nothing to tarball."
else
	tar -Pczf inspection_dump.tar.gz "$INSPECT_DUMP"
	log_success "Tarball inspection_dump.tar.gz created"
fi

exit