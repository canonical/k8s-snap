#!/bin/bash
set -ex
# /home/maciek/canon/k8s-snap/build-scripts/hack/sonobuoy.sh sdf ubuntu:24.04

function download() {
    local architecture=$1
    local release=$(curl --silent -m 10 --connect-timeout 5 "https://api.github.com/repos/vmware-tanzu/sonobuoy/releases/latest")
    local tag=$(echo "$release" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    curl -L -o sonobuoy.tar.gz https://github.com/vmware-tanzu/sonobuoy/releases/download/${tag}/sonobuoy_${tag//v}_linux_${architecture}.tar.gz
    tar -xf sonobuoy.tar.gz
    rm sonobuoy.tar.gz
}

function create_container() {
    local container_name=$1
    local os=$2
#    lxc profile create ${container_name}
#    cat tests/integration/lxd-profile.yaml | lxc profile edit ${container_name}
    lxc launch -p default -p ${container_name} ${os} ${container_name}
    lxc config device add ${container_name} repo disk source=${PWD} path=/repo/
}

function setup_k8s() {
    local container_name=$1
    lxc exec ${container_name} -- service snapd start
    lxc exec ${container_name} -- snap install /repo/k8s.snap --dangerous --classic
    lxc exec ${container_name} -- k8s bootstrap
    lxc exec ${container_name} -- k8s status --wait-ready
    mkdir -p ~/.kube
    lxc exec ${container_name} -- k8s config >> ~/.kube/config
    export token=$(lxc exec ${container_name} -- k8s get-join-token second)
}

function add_k8s_node() {
    local container_name=$1
    local token1=$2
    lxc exec ${container_name} -- service snapd start
    lxc exec ${container_name} -- snap install /repo/k8s.snap --dangerous --classic
    lxc exec ${container_name} -- k8s join-cluster ${token1}
    lxc exec ${container_name} -- k8s status --wait-ready
}

function run_e2e() {
    ./sonobuoy run --plugin e2e --wait
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
    local main="main"
    local fallow="fallow"
#    download ${architecture}
    for container_name in  ${main} ${fallow}; do
        create_container ${container_name} ${os} &
#        create_container ${container_name} ${os}
    done
    wait
    setup_k8s ${main}
    echo "token: $token"
    add_k8s_node ${fallow} ${token}
#    run_e2e
#    exit $?
}

if [[ $sourced -ne 1 ]]; then
    main $@
fi
