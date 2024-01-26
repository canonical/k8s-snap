#
# Copyright 2024 Canonical, Ltd.
#
import logging
import shlex
import subprocess
import time
from functools import partial
from pathlib import Path
from typing import Callable, List, Optional, Union

from e2e_util import config, harness
from tenacity import (
    RetryCallState,
    retry,
    retry_if_exception_type,
    stop_after_attempt,
    stop_never,
    wait_fixed,
)

LOG = logging.getLogger(__name__)


def run(command: list, **kwargs) -> subprocess.CompletedProcess:
    """Log and run command."""
    kwargs.setdefault("check", True)

    LOG.debug("Execute command %s (kwargs=%s)", shlex.join(command), kwargs)
    return subprocess.run(command, **kwargs)


def stubbornly(
    retries: Optional[int] = None,
    delay_s: Optional[Union[float, int]] = None,
    exceptions: Optional[tuple] = None,
    **retry_kds,
):
    """
    Retry a command for a while, using tenacity

    By default, retry immediately and forever until no exceptions occur.

    Some commands need to execute until they pass some condition
    > stubbornly(*retry_args).until(*some_condition).exec(*some_command)

    Some commands need to execute until they complete
    > stubbornly(*retry_args).exec(*some_command)

    : param    retries              int: convenience param to use stop=retry.stop_after_attempt(<int>)
    : param    delay_s        float|int: convenience param to use wait=retry.wait_fixed(delay_s)
    : param exceptions Tuple[Exception]: convenience param to use retry=retry.retry_if_exception_type(exceptions)
    : param retry_kds           Mapping: direct interface to all tenacity arguments for retrying
    """

    def _before_sleep(retry_state: RetryCallState):
        attempt = retry_state.attempt_number
        tries = f"/{retries}" if retries is not None else ""
        LOG.info(
            f"Attempt {attempt}{tries} failed. Error: {retry_state.outcome.exception()}"
        )
        LOG.info(f"Retrying in {delay_s} seconds...")

    _waits = wait_fixed(delay_s) if delay_s is not None else wait_fixed(0)
    _stops = stop_after_attempt(retries) if retries is not None else stop_never
    _exceptions = exceptions or (Exception,)  # default to retry on all exceptions

    _retry_args = dict(
        wait=_waits,
        stop=_stops,
        retry=retry_if_exception_type(_exceptions),
        before_sleep=_before_sleep,
    )
    # Permit any tenacity retry overrides from these ^defaults
    _retry_args.update(retry_kds)

    class Retriable:
        def __init__(self) -> None:
            self._condition = None
            self._run = subprocess.run

        @retry(**_retry_args)
        def exec(
            self,
            command_args: List[str],
            **command_kwds,
        ):
            """
            Execute a command against a harness or locally with subprocess to be retried.

            :param  List[str]        command_args: The command to be executed, as a str or list of str
            :param Mapping[str,str]      command_kwds: Additional keyword arguments to be passed to exec
            """

            try:
                resp = self._run(command_args, **command_kwds)
            except subprocess.CalledProcessError as e:
                LOG.error(f"  rc={e.returncode}")
                LOG.error(f"  stdout={e.stdout.decode()}")
                LOG.error(f"  stderr={e.stderr.decode()}")
                raise
            if self._condition:
                assert self._condition(resp), "Failed to meet condition"
            return resp

        def on(self, instance: harness.Instance) -> "Retriable":
            """
            Target the command at some instance.

            :param instance Instance: Instance on a test harness.
            """
            self._run = partial(instance.exec, capture_output=True)
            return self

        def until(
            self, condition: Callable[[subprocess.CompletedProcess], bool] = None
        ) -> "Retriable":
            """
            Test the output of the executed command against an expected response

            :param Callable condition: a callable which returns a truth about the command output
            """
            self._condition = condition
            return self

    return Retriable()


def setup_dns(instance: harness.Instance):
    LOG.info("Waiting for dns to be enabled...")
    stubbornly(retries=15, delay_s=5).on(instance).exec(
        ["k8s", "enable", "dns", "--cluster-domain=foo.local"]
    )
    LOG.info("DNS enabled.")

    LOG.info("Waiting for CoreDNS pod to show up...")
    stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "coredns" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"])
    LOG.info("CoreDNS pod showed up.")

    stubbornly(retries=3, delay_s=1).on(instance).exec(
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
        ]
    )


def setup_network(instance: harness.Instance):
    time.sleep(30)
    instance.exec(
        ["/snap/k8s/current/k8s/network-requirements.sh"], stdout=subprocess.DEVNULL
    )

    LOG.info("Waiting for network to be enabled...")
    stubbornly(retries=15, delay_s=5).on(instance).exec(["k8s", "enable", "network"])
    LOG.info("Network enabled.")

    LOG.info("Waiting for cilium pods to show up...")
    stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "cilium" in p.stdout.decode()
    ).exec(
        ["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"],
    )
    LOG.info("Cilium pods showed up.")

    stubbornly(retries=3, delay_s=1).on(instance).exec(
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
    )

    stubbornly(retries=3, delay_s=1).on(instance).exec(
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
        ]
    )


# Installs and setups the k8s snap on the given instance and connects the interfaces.
def setup_k8s_snap(i: harness.Instance, snap_path: Path):
    LOG.info("Install snap")
    i.send_file(config.SNAP, snap_path)
    i.exec(["snap", "install", snap_path, "--dangerous"])

    LOG.info("Initialize Kubernetes")
    i.exec(["/snap/k8s/current/k8s/connect-interfaces.sh"])


# Validates that the K8s node is in Ready state.
def wait_until_k8s_ready(
    control_node: harness.Instance, instances: List[harness.Instance]
):
    for instance in instances:
        host = hostname(instance)
        result = (
            stubbornly(retries=15, delay_s=5)
            .on(control_node)
            .until(lambda p: " Ready" in p.stdout.decode())
            .exec(["k8s", "kubectl", "get", "node", host, "--no-headers"])
        )
    LOG.info("Kubelet registered successfully!")
    LOG.info("%s", result.stdout.decode())


def hostname(instance: harness.Instance) -> str:
    """Return the hostname for a given instance."""
    resp = instance.exec(["hostname"], capture_output=True)
    return resp.stdout.decode().strip()
