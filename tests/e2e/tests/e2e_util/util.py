#
# Copyright 2023 Canonical, Ltd.
#
import logging
import shlex
import subprocess
import time
from pathlib import Path

import pytest
from e2e_util import config, harness

LOG = logging.getLogger(__name__)


def run(command: list, **kwargs) -> subprocess.CompletedProcess:
    """Log and run command."""
    kwargs.setdefault("check", True)

    LOG.debug("Execute command %s (kwargs=%s)", shlex.join(command), kwargs)
    return subprocess.run(command, **kwargs)


# Installs and setups the k8s snap on the given instance.
# If wait_ready is set, it is validated that the K8s node is in Ready state.
def setup_k8s_snap(
    h: harness.Harness, instance_id: str, snap_path: Path, wait_ready: bool = True
):
    LOG.info("Install snap")
    h.send_file(instance_id, config.SNAP, snap_path)
    h.exec(instance_id, ["snap", "install", snap_path, "--dangerous"])

    LOG.info("Initialize Kubernetes")
    h.exec(instance_id, ["/snap/k8s/current/k8s/init.sh"])
    h.exec(instance_id, ["k8s", "init"])

    LOG.info("Start Kubernetes")
    h.exec(instance_id, ["k8s", "start"])

    if not wait_ready:
        return

    hostname = (
        h.exec(instance_id, ["hostname"], capture_output=True).stdout.decode().strip()
    )
    success = False
    for attempt in range(30):
        try:
            LOG.info("(attempt %d) Waiting for Kubelet to register", attempt)
            p = h.exec(
                instance_id,
                ["k8s", "kubectl", "get", "node", hostname, "--no-headers"],
                capture_output=True,
            )

            if "NotReady" in p.stdout.decode():
                continue

            success = True
            LOG.info("Kubelet registered successfully!")
            LOG.info("%s", p.stdout.decode())
            break
        except subprocess.CalledProcessError:
            time.sleep(5)

    if not success:
        pytest.fail("Kubelet node did not register")
