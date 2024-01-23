#
# Copyright 2024 Canonical, Ltd.
#
import logging
import shlex
import subprocess
from pathlib import Path
from typing import Callable, List, Optional, Union

from e2e_util import config, harness
from tenacity import (
    RetryCallState,
    retry,
    retry_if_exception_type,
    stop_after_attempt,
    wait_fixed,
)

LOG = logging.getLogger(__name__)


def run(command: list, **kwargs) -> subprocess.CompletedProcess:
    """Log and run command."""
    kwargs.setdefault("check", True)

    LOG.debug("Execute command %s (kwargs=%s)", shlex.join(command), kwargs)
    return subprocess.run(command, **kwargs)


def stubbornly(retries=3, delay_s=1, exceptions: Optional[tuple] = None):
    """
    Some commands need to execute until they pass some condition
    > stubbornly(*retry_args).until(*some_condition).exec(*some_command)

    Some commands need to execute until they complete
    > stubbornly(*retry_args).exec(*some_command)
    """

    def _before_sleep(retry_state: RetryCallState):
        attempt = retry_state.attempt_number
        LOG.info(
            f"Attempt {attempt}/{retries} failed. Error: {retry_state.outcome.exception()}"
        )
        LOG.info(f"Retrying in {delay_s} seconds...")

    _exceptions = exceptions or (Exception,)  # default to retry on all exceptions

    class Retriable:
        def __init__(self) -> None:
            self.condition = None

        @retry(
            wait=wait_fixed(delay_s),
            stop=stop_after_attempt(retries),
            retry=retry_if_exception_type(_exceptions),
            before_sleep=_before_sleep,
        )
        def exec(
            self,
            command_args: Union[str, List[str]],
            harness: Optional[harness.Harness] = None,
            instance: str = "",
            **command_kwds,
        ):
            """
            Execute a command against a harness or locally with subprocess to be retried.

            :param str | List[str]       command_args: The command to be executed, as a str or list of str
            :param Map[str,str]          command_kwds: Additional keyword arguments to be passed to exec
            :param Optional[Harness]          harness: test Harness object, to run the command on
            :param str                       instance: Instance id in the test harness.
            """
            if isinstance(command_args, str):
                # Safely split the string command into command arguments
                command_args = shlex.split(command_args)

            try:
                if harness is not None:
                    resp = harness.exec(
                        instance, command_args, capture_output=True, **command_kwds
                    )
                else:
                    resp = subprocess.run(command_args, **command_kwds)
            except subprocess.CalledProcessError as e:
                LOG.error(f"  rc={e.returncode}")
                LOG.error(f"  stdout={e.stdout.decode()}")
                LOG.error(f"  stderr={e.stderr.decode()}")
                raise
            if self.condition:
                assert self.condition(resp), "Failed to meet condition"

        def until(
            self, condition: Callable[[subprocess.CompletedProcess], bool] = None
        ):
            """
            Test the output of the executed command against an expected response

            :param Callable condition: a callable which returns a truth about the command output
            """
            self.condition = condition
            return self

    return Retriable()


def setup_dns(h: harness.Harness, instance_id: str):
    LOG.info("Waiting for dns to be enabled...")
    retry_until_condition(
        h,
        instance_id,
        ["k8s", "enable", "dns", "--cluster-domain=foo.local"],
        condition=lambda p: p.returncode == 0,
    )
    LOG.info("DNS enabled.")

    LOG.info("Waiting for CoreDNS pod to show up...")
    retry_until_condition(
        h,
        instance_id,
        ["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"],
        condition=lambda p: "coredns" in p.stdout.decode(),
    )
    LOG.info("CoreDNS pod showed up.")

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
            "app.kubernetes.io/name=coredns",
            "--timeout",
            "180s",
        ],
        max_retries=3,
        delay_between_retries=1,
    )


def setup_network(h: harness.Harness, instance_id: str):
    h.exec(instance_id, ["/snap/k8s/current/k8s/network-requirements.sh"])

    LOG.info("Waiting for network to be enabled...")
    stubbornly(retries=15, delay_s=5).until(
        lambda p: "enabled" in p.stdout.decode()
    ).exec(["k8s", "enable", "network"], h, instance_id)
    LOG.info("Network enabled.")

    LOG.info("Waiting for cilium pods to show up...")
    stubbornly(retries=15, delay_s=5).until(
        lambda p: "cilium" in p.stdout.decode()
    ).exec(
        ["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"],
        h,
        instance_id,
    )
    LOG.info("Cilium pods showed up.")

    stubbornly().exec(
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
        h,
        instance_id,
    )

    stubbornly().exec(
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
        h,
        instance_id,
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
        result = (
            stubbornly(retries=15, delay_s=5)
            .until(lambda p: "Ready" in p.stdout.decode())
            .exec(f"k8s kubectl get node {hostname} --no-headers")
        )
    LOG.info("Kubelet registered successfully!")
    LOG.info("%s", result.stdout.decode())
