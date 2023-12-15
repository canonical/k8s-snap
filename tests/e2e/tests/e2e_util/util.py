#
# Copyright 2023 Canonical, Ltd.
#
import logging
import shlex
import subprocess
import time
from pathlib import Path
from typing import Optional

import pytest
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


# Installs and setups the k8s snap on the given instance and connects the interfaces.
def setup_k8s_snap(h: harness.Harness, instance_id: str, snap_path: Path):
    LOG.info("Install snap")
    h.send_file(instance_id, config.SNAP, snap_path)
    h.exec(instance_id, ["snap", "install", snap_path, "--dangerous"])

    LOG.info("Initialize Kubernetes")
    h.exec(instance_id, ["/snap/k8s/current/k8s/connect-interfaces.sh"])


# Validates that the K8s node is in Ready state.
def wait_until_k8s_ready(h: harness.Harness, instance_id):
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
