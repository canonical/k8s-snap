#!/usr/bin/env bash

JOURNALCTL_LIMIT=100000

INSPECT_DUMP=$(pwd)/inspection-report

SVC_ARGS_DIR=/var/snap/k8s/common/args
K8SD_STATE_DIR=/var/snap/k8s/common/var/lib/k8sd/state
SBOM_FILE=/snap/k8s/current/bom.json

function collect_args {
	local service=$1

	mkdir -p $INSPECT_DUMP/$service

	if [ -e $SVC_ARGS_DIR/${service#k8s.} ]; then
		# Strip k8s. prefix if present because args directories _are not_ created with k8s. prefix
		cat $SVC_ARGS_DIR/${service#k8s.} &> $INSPECT_DUMP/$service/args
		printf -- '\033[32m SUCCESS: \033[0m Found arguments for %s\n' "$service"
	else
		printf -- '\033[31m FAIL: \033[0m Arguments for %s not found\n' "$service"
	fi
}

function collect_cluster_info {
	mkdir -p $INSPECT_DUMP

	k8s kubectl cluster-info dump &> $INSPECT_DUMP/cluster-info

	printf -- '\033[32m SUCCESS: \033[0m Collected k8s cluster-info %s\n'
}

function collect_sbom {
	mkdir -p $INSPECT_DUMP

	cat $SBOM_FILE &> $INSPECT_DUMP/sbom.json

	printf -- '\033[32m SUCCESS: \033[0m Collected SBOM\n'
}

function collect_microcluster_db {
	mkdir -p $INSPECT_DUMP

	/snap/k8s/current/bin/k8sd sql .dump --state-dir $k8sd_state_dir &> $INSPECT_DUMP/k8sd-db.sql

	printf -- '\033[32m SUCCESS: \033[0m Collected k8sd database dump\n'
}

function check_service {
	local service=$1
	mkdir -p $INSPECT_DUMP/$service

	status="inactive"
	
	journalctl -n $JOURNALCTL_LIMIT -u snap.$service &> $INSPECT_DUMP/$service/journal.log
	systemctl status snap.$service &> $INSPECT_DUMP/$service/systemctl.log
	if systemctl status snap.$service &> /dev/null; then
		status="active"
	fi

	if [ "$status" == "active" ]; then
	  printf -- '\033[32m SUCCESS: \033[0m Service %s is running\n' "$service"
	else
	  printf -- '\033[31m FAIL: \033[0m Service %s is not running\n' "$service"
	  printf -- 'For more details look at: sudo journalctl -u snap.%s\n' "$service"
	fi
}

# Source: https://github.com/canonical/microk8s/blob/master/microk8s-resources/actions/common/utils.sh#L1272
# test if we run with sudo
if [ "$EUID" -ne 0 ]; then
	echo "Elevated permissions are needed for this command. Please use sudo."
	exit 1
fi

rm -rf $INSPECT_DUMP
mkdir -p $INSPECT_DUMP

svc_containerd='k8s.containerd'
svc_api_server_proxy='k8s.k8s-apiserver-proxy'
svc_k8s_dqlite='k8s.k8s-dqlite'
svc_k8sd='k8s.k8sd'
svc_kube_apiserver='k8s.kube-apiserver'
svc_kube_controller_manager='k8s.kube-controller-manager'
svc_kube_proxy='k8s.kube-proxy'
svc_kube_scheduler='k8s.kube-scheduler'
svc_kubelet='k8s.kubelet'

printf -- 'Inspecting services\n'
check_service $svc_containerd
check_service $svc_api_server_proxy
check_service $svc_k8s_dqlite
check_service $svc_k8sd
check_service $svc_kube_apiserver
check_service $svc_kube_controller_manager
check_service $svc_kube_proxy
check_service $svc_kube_scheduler
check_service $svc_kubelet
printf -- '\n'

printf -- 'Collecting arguments\n'
if [ ! -d $SVC_ARGS_DIR ]; then
	printf -- '\033[31m FAIL: \033[0m Arguments directory not found.\n'
else
	collect_args $svc_containerd
	collect_args $svc_api_server_proxy
	collect_args $svc_k8s_dqlite
	collect_args $svc_k8sd
	collect_args $svc_kube_apiserver
	collect_args $svc_kube_controller_manager
	collect_args $svc_kube_proxy
	collect_args $svc_kube_scheduler
	collect_args $svc_kubelet
fi
printf -- '\n'


printf -- 'Collecting k8s cluster-info\n'
if ! k8s &> /dev/null; then
	printf -- '\033[31m FAIL: \033[0m k8s command not found.\n'
else
	collect_cluster_info
fi
printf -- '\n'

printf -- 'Collecting SBOM\n'
if [ ! -f $SBOM_FILE ]; then
	printf -- '\033[31m FAIL: \033[0m SBOM file not found.\n'
else
	collect_sbom
fi
printf -- '\n'

printf -- 'Collecting k8sd database dump\n'
if [ ! -d $K8SD_STATE_DIR ]; then
	printf -- '\033[31m FAIL: \033[0m k8sd state directory not found.\n'
else
	collect_microcluster_db
fi
printf -- '\n'

exit