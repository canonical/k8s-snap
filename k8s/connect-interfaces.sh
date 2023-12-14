#!/usr/bin/env bash

set -u

if [ "$EUID" -ne 0 ]
then echo "Please run this script as root."
  exit 1
fi

for i in account-control \
         docker-privileged \
         kubernetes-support \
         k8s-journald \
         k8s-kubelet \
         k8s-kubeproxy \
         network \
         network-bind \
         network-control \
         network-observe \
         firewall-control \
         process-control \
         kernel-module-observe \
         mount-observe \
         hardware-observe \
         system-observe \
         home \
         opengl \
         home-read-all \
         login-session-observe \
         log-observe
do
  snap connect k8s:$i
done
