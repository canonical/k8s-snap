#!/usr/bin/env bash

JOURNALCTL_LIMIT=100000
RETURN_CODE=0

source utils.sh

INSPECT_DUMP=${SNAP_DATA}/inspection-report

function collect_args {
	mkdir -p $INSPECT_DUMP/$service

	# Strip k8s prefix if present because args directories _are not_ created with k8s prefix
	cp -r $SNAP_COMMON/args/${service#k8s.} $INSPECT_DUMP/$service/args
}

function check_service {
	local service=$1
	mkdir -p $INSPECT_DUMP/$service

	status="inactive"
	
	journalctl -n $JOURNALCTL_LIMIT -u snap.$service &> $INSPECT_DUMP/$service/journal.log
	systemctl status snap.$service &> $INSPECT_DUMP/$service/systemctl.log
	if systemctl status snap.$service &> /dev/null
	then
		status="active"
	fi

	if [ "$status" == "active" ]
	then
	  printf -- '  Service %s is running\n' "$service"
	else
	  printf -- '\033[31m FAIL: \033[0m Service %s is not running\n' "$service"
	  printf -- 'For more details look at: sudo journalctl -u snap.%s\n' "$service"
	  RETURN_CODE=1
	fi
}

rm -rf $INSPECT_DUMP
mkdir -p $INSPECT_DUMP

printf -- 'Inspecting services\n'
svc_containerd='k8s.containerd'
svc_api_server_proxy='k8s.k8s-apiserver-proxy'
svc_k8s_dqlite='k8s.k8s-dqlite'
svc_k8sd='k8s.k8sd'
svc_kube_apiserver='k8s.kube-apiserver'
svc_kube_controller_manager='k8s.kube-controller-manager'
svc_kube_proxy='k8s.kube-proxy'
svc_kube_scheduler='k8s.kube-scheduler'
svc_kubelet='k8s.kubelet'

check_service $svc_containerd
check_service $svc_api_server_proxy
check_service $svc_k8s_dqlite
check_service $svc_k8sd
check_service $svc_kube_apiserver
check_service $svc_kube_controller_manager
check_service $svc_kube_proxy
check_service $svc_kube_scheduler
check_service $svc_kubelet
