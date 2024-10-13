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
#    ./sonobuoy run --plugin e2e --wait
    ./sonobuoy run --plugin e2e --wait --mode quick
    ./sonobuoy retrieve -f sonobuoy_e2e.tar.gz
    ./sonobuoy results sonobuoy_e2e.tar.gz
    tar -xf sonobuoy_e2e.tar.gz --one-top-level
    set +e
    ./sonobuoy results sonobuoy_e2e.tar.gz | grep -E "^Failed: 0$"
    return $?
}

function main() {
    if [ "$#" -ne 2 ]; then
      cat << EOF
Expected 2 arguments, provided: $@

usage: $0 <architecture> <os>
EOF
      exit 255
    fi

    local architecture=$1
    local os=$2

    download ${architecture}
    create_container ${os}
    setup_k8s
    run_e2e
    exit $?

}

if [[ $sourced -ne 1 ]]; then
    main $@
fi
