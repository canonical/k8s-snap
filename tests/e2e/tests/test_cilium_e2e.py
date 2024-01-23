#
# Copyright 2024 Canonical, Ltd.
#
import logging
import platform
from pathlib import Path

import pytest
from e2e_util import config, harness, util

LOG = logging.getLogger(__name__)

ARCH = platform.machine()
CILIUM_CLI_ARCH_MAP = {"aarch64": "arm64", "x86_64": "amd64"}
CILIUM_CLI_VERSION = "v0.15.19"
CILIUM_CLI_TAR_GZ = f"https://github.com/cilium/cilium-cli/releases/download/{CILIUM_CLI_VERSION}/cilium-linux-{CILIUM_CLI_ARCH_MAP.get(ARCH)}.tar.gz"  # noqa


@pytest.mark.skipif(
    ARCH not in CILIUM_CLI_ARCH_MAP, reason=f"Platform {ARCH} not supported"
)
def test_cilium_e2e(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info("Create instance")
    instance_id = h.new_instance()

    util.setup_k8s_snap(h, instance_id, snap_path)
    h.exec(instance_id, ["k8s", "bootstrap"])
    util.setup_network(h, instance_id)

    h.exec(instance_id, ["bash", "-c", "mkdir -p ~/.kube"])
    h.exec(instance_id, ["bash", "-c", "k8s config > ~/.kube/config"])

    # TODO(neoaggelos): this is temporary until "k8s enable dns" is ready
    h.exec(
        instance_id,
        [
            "k8s",
            "helm",
            "install",
            "coredns",
            "coredns",
            "-n",
            "kube-system",
            "--repo",
            "https://coredns.github.io/helm",
            "--set",
            "service.clusterIP=10.152.183.10",
        ],
    )
    h.exec(
        instance_id,
        [
            "bash",
            "-c",
            "echo --cluster-dns=10.152.183.10 >> /var/snap/k8s/current/args/kubelet",
        ],
    )
    h.exec(
        instance_id,
        [
            "bash",
            "-c",
            "echo --cluster-domain=cluster.local >> /var/snap/k8s/current/args/kubelet",
        ],
    )
    h.exec(instance_id, ["snap", "restart", "k8s.kubelet"])

    # Download cilium-cli
    h.exec(instance_id, ["curl", "-L", CILIUM_CLI_TAR_GZ, "-o", "cilium.tar.gz"])
    h.exec(instance_id, ["tar", "xvzf", "cilium.tar.gz"])
    h.exec(instance_id, ["./cilium", "version", "--client"])

    # TODO(neoaggelos): replace with "k8s status --wait-ready"
    util.stubbornly(retries=15, delay_s=5).until(
        lambda p: "OK" == p.stdout.decode().strip()
    ).exec(
        "k8s kubectl exec -it ds/cilium -n kube-system -c cilium-agent -- cilium status --brief",
        h,
        instance_id,
    )

    # Run cilium e2e tests
    e2e_args = []
    if config.SUBSTRATE == "lxd":
        # NOTE(neoaggelos): disable "no-unexpected-packet-drops" on LXD as it fails:
        # [=] Test [no-unexpected-packet-drops] [1/61]
        #   [-] Scenario [no-unexpected-packet-drops/no-unexpected-packet-drops]
        #       Found unexpected packet drops:
        # {
        #   "labels": {
        #     "direction": "INGRESS",
        #     "reason": "VLAN traffic disallowed by VLAN filter"
        #   },
        #   "name": "cilium_drop_count_total",
        #   "value": 4
        # }
        e2e_args.extend(["--test", "!no-unexpected-packet-drops"])

    h.exec(instance_id, ["./cilium", "connectivity", "test", *e2e_args])

    h.cleanup()
