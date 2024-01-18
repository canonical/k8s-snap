#
# Copyright 2023 Canonical, Ltd.
#
import logging
import shlex
import subprocess
import time
from pathlib import Path
from typing import Callable, List, Optional, Union

from e2e_util import config, harness

LOG = logging.getLogger(__name__)


def run(command: list, **kwargs) -> subprocess.CompletedProcess:
    """Log and run command."""
    kwargs.setdefault("check", True)

    LOG.debug("Execute command %s (kwargs=%s)", shlex.join(command), kwargs)
    return subprocess.run(command, **kwargs)


def run_with_retry(
    command,
    max_retries=3,
    delay_between_retries=1,
    exceptions: Optional[tuple] = None,
    **kwargs,
) -> subprocess.CompletedProcess:
    """
    Run a command using subprocess.run with retry logic.

    Parameters:
    - command (list): The command to be executed, as a list of strings.
    - max_retries (int): Maximum number of retries in case of failure.
    - delay_between_retries (int): Delay in seconds between retries.
    - exceptions (tuple(Exception)): Excepections that should be retried. Retry all if None or empty (default)
    - **kwargs: Additional keyword arguments to be passed to subprocess.run.

    Returns:
    - subprocess.CompletedProcess: CompletedProcess object representing the result of the command.
    """
    for attempt in range(1, max_retries + 1):
        try:
            result = run(command, **kwargs)
            return result
        except Exception as e:
            if exceptions is None or len(exceptions) == 0 or isinstance(e, exceptions):
                LOG.info(f"Attempt {attempt}/{max_retries} failed. Error: {e}")
                if attempt < max_retries:
                    LOG.info(f"Retrying in {delay_between_retries} seconds...")
                    time.sleep(delay_between_retries)
                else:
                    # If all attempts fail, raise the last error
                    raise e


# TODO: Consider how to handle more advanced retry logic
def retry_until_condition(
    h: harness.Harness,
    instance_id,
    command,
    condition: Callable[[subprocess.CompletedProcess], bool] = None,
    max_retries=15,
    delay_between_retries=5,
    exceptions: Optional[tuple] = None,
    **kwargs,
) -> subprocess.CompletedProcess:
    for attempt in range(max_retries):
        try:
            p = h.exec(instance_id, command, capture_output=True, **kwargs)
            if condition is not None:
                assert condition(p), "Failed to meet condition."
            return p
        except Exception as e:
            if (
                exceptions is None
                or len(exceptions) == 0
                or isinstance(e, exceptions)
                or isinstance(e, AssertionError)
            ):
                LOG.info(f"Attempt {attempt}/{max_retries} failed. Error: {e}")
                if attempt < max_retries:
                    LOG.info(f"Retrying in {delay_between_retries} seconds...")
                    time.sleep(delay_between_retries)
                else:
                    # If all attempts fail, raise the last error
                    raise e


def setup_network(h: harness.Harness, instance_id: str):
    h.exec(instance_id, ["/snap/k8s/current/k8s/network-requirements.sh"])

    LOG.info("Waiting for network to be enabled...")
    retry_until_condition(
        h,
        instance_id,
        ["k8s", "enable", "network"],
        condition=lambda p: "enabled" in p.stderr.decode(),
    )
    LOG.info("Network enabled.")

    LOG.info("Waiting for cilium pods to show up...")
    retry_until_condition(
        h,
        instance_id,
        ["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"],
        condition=lambda p: "cilium" in p.stdout.decode(),
    )
    LOG.info("Cilium pods showed up.")

    retry_until_condition(
        h,
        instance_id,
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "io.cilium/app=operator",
            "--timeout",
            "180s",
        ],
        max_retries=3,
        delay_between_retries=1,
    )

    retry_until_condition(
        h,
        instance_id,
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "k8s-app=cilium",
            "--timeout",
            "180s",
        ],
        max_retries=3,
        delay_between_retries=1,
    )


# Installs and setups the k8s snap on the given instance and connects the interfaces.
def setup_k8s_snap(h: harness.Harness, instance_id: str, snap_path: Path):
    LOG.info("Install snap")
    h.send_file(instance_id, config.SNAP, snap_path)
    h.exec(instance_id, ["snap", "install", snap_path, "--dangerous"])

    LOG.info("Initialize Kubernetes")
    h.exec(instance_id, ["/snap/k8s/current/k8s/connect-interfaces.sh"])


# Validates that the K8s node is in Ready state.
def wait_until_k8s_ready(h: harness.Harness, instances: Union[str, List[str]]):
    if isinstance(instances, str):
        instances = [instances]

    for instance in instances:
        hostname = (
            h.exec(instance, ["hostname"], capture_output=True).stdout.decode().strip()
        )
        result = retry_until_condition(
            h,
            instance,
            ["k8s", "kubectl", "get", "node", hostname, "--no-headers"],
            condition=lambda p: "Ready" in p.stdout.decode(),
        )
    LOG.info("Kubelet registered successfully!")
    LOG.info("%s", result.stdout.decode())
