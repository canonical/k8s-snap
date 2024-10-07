#!/bin/bash
set -e

function download() {
    local architecture=$1
    local release=$(curl --silent -m 10 --connect-timeout 5 "https://api.github.com/repos/vmware-tanzu/sonobuoy/releases/latest")
    local tag=$(echo "$release" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    curl -L -o sonobuoy.tar.gz https://github.com/vmware-tanzu/sonobuoy/releases/download/${tag}/sonobuoy_${tag//v}_linux_${architecture}.tar.gz
    tar -xf sonobuoy.tar.gz
    rm sonobuoy.tar.gz
}

function create_container() {
  local os=$1
  lxc profile create k8s-e2e
  cat tests/integration/lxd-profile.yaml | lxc profile edit k8s-e2e
  lxc launch -p default -p k8s-e2e ${os} k8s-e2e
  lxc config device add k8s-e2e repo disk source=${PWD} path=/repo/
}

function setup_k8s() {
  lxc exec k8s-e2e -- service snapd start
  lxc exec k8s-e2e -- snap install /repo/k8s.snap --dangerous --classic
  lxc exec k8s-e2e -- k8s bootstrap
  lxc exec k8s-e2e -- k8s status --wait-ready
  mkdir -p ~/.kube
  lxc exec k8s-e2e -- k8s config > ~/.kube/config
}

function run_e2e() {
  ./sonobuoy run --plugin e2e --wait
  ./sonobuoy retrieve -f sonobuoy_e2e.tar.gz
  ./sonobuoy results sonobuoy_e2e.tar.gz
  tar -xf sonobuoy_e2e.tar.gz --one-top-level
}

function main() {
    local command=$1
    shift
    case $command in
        "download")
            download "${@}"
            ;;
        "create_container")
            create_container "${@}"
            ;;
        "setup_k8s")
            setup_k8s "${@}"
            ;;
        "run_e2e")
            run_e2e "${@}"
            ;;
        *)
            cat << EOF
Unknown command: $1

usage: $0 <command>

Commands:
    download <amd64|arm64|386|ppc64le|s390x>   Download sonobuoy for given architecture.

    create_container <os>                      Creates lxd container for given operating system.

    setup_k8s                                  Install k8s in k8s lxd container.

    run_e2e                                    Runs sonobuoy end-to-end tests and saves results in
                                               sonobuoy_e2e.tar.gz and sonobuoy_e2e directory
EOF
            ;;
    esac
}

if [[ $sourced -ne 1 ]]; then
    main $@
fi
