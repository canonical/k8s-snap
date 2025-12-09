#
# Copyright 2025 Canonical, Ltd.
#

from test_util import config

ETCD_PORTS = [
    2379,  # etcd client
    2380,  # etcd peer
]

K8S_DQLITE_PORTS = [
    9000,
]

# K8S_CORE_PORTS are ports used by core Kubernetes components.
# These are always required regardless of datastore choice.
K8S_CORE_PORTS = [
    6443,  # kube-apiserver
    10250,  # kubelet
    10257,  # kube-controller-manager
    10259,  # kube-scheduler
    4240,  # cilium health
    8472,  # VXLAN overlay
]

# DEFAULT_OPEN_PORTS is the complete set of ports that will be opened in the firewall for every test.
# Tests can specify additional ports via the @pytest.mark.required_ports() decorator.
DEFAULT_OPEN_PORTS = K8S_CORE_PORTS + (
    K8S_DQLITE_PORTS if config.DATASTORE == "k8s-dqlite" else ETCD_PORTS
)
