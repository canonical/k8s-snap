#
# Copyright 2025 Canonical, Ltd.
#
import base64
import logging
import os
import shlex
import subprocess
from pathlib import Path

from test_util import config
from test_util.harness import Harness, HarnessError, Instance
from test_util.util import run

LOG = logging.getLogger(__name__)


class MultipassHarness(Harness):
    """A Harness that creates a Multipass VM for each instance."""

    name = "multipass"

    def next_id(self) -> int:
        self._next_id += 1
        return self._next_id

    def __init__(self):
        super(MultipassHarness, self).__init__()

        self._next_id = 0

        self.image = config.MULTIPASS_IMAGE
        self.cpus = config.MULTIPASS_CPUS
        self.memory = config.MULTIPASS_MEMORY
        self.disk = config.MULTIPASS_DISK
        self.cloud_init = base64.b64decode(config.MULTIPASS_CLOUD_INIT_BASE64)
        self.instances = set()

        LOG.debug("Configured Multipass substrate (image %s)", self.image)

    def new_instance(
        self, network_type: str = "IPv4", name_suffix: str = ""
    ) -> Instance:
        if network_type not in ("IPv4", "IPv6"):
            raise HarnessError("Currently only IPv4 is supported by Multipass harness")

        instance_id = (
            f"k8s-integration-{self.next_id()}-{os.urandom(3).hex()}{name_suffix}"
        )

        LOG.debug("Creating instance %s with image %s", instance_id, self.image)
        try:
            cmd = [
                "sudo",
                "multipass",
                "launch",
                self.image,
                "--name",
                instance_id,
                "--cpus",
                self.cpus,
                "--memory",
                self.memory,
                "--disk",
                self.disk,
            ]

            if self.cloud_init:
                LOG.info("Using cloud-init: %s", self.cloud_init)
                # Increase timeout to 15 minutes since custom setup steps, e.g. FIPS, may take a while.
                run(
                    cmd + ["--cloud-init", "-", "--timeout", "900"],
                    input=self.cloud_init,
                )
            else:
                run(cmd)

        except subprocess.CalledProcessError as e:
            raise HarnessError(f"Failed to create multipass VM {instance_id}") from e

        self.instances.add(instance_id)

        self.exec(instance_id, ["snap", "wait", "system", "seed.loaded"])
        if network_type == "IPv6":
            LOG.debug("Enabling IPv6 support in instance %s", instance_id)
            try:
                self.exec(
                    instance_id, ["sysctl", "-w", "net.ipv6.conf.all.disable_ipv6=0"]
                )
                self.exec(
                    instance_id,
                    ["sysctl", "-w", "net.ipv6.conf.default.disable_ipv6=0"],
                )
                self.exec(
                    instance_id, ["sysctl", "-w", "net.ipv6.conf.lo.disable_ipv6=0"]
                )
                self.exec(instance_id, ["ip", "-6", "addr"])
            except subprocess.CalledProcessError as e:
                raise HarnessError(
                    f"Failed to configure IPv6 in instance {instance_id}"
                ) from e

        return Instance(self, instance_id)

    def send_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(destination).is_absolute():
            raise HarnessError(f"path {destination} must be absolute")

        LOG.debug(
            "Copying file %s to instance %s at %s", source, instance_id, destination
        )
        try:
            self.exec(
                instance_id,
                ["mkdir", "-m=0777", "-p", Path(destination).parent.as_posix()],
            )
            run(
                [
                    "sudo",
                    "multipass",
                    "transfer",
                    source,
                    f"{instance_id}:{destination}",
                ]
            )
        except subprocess.CalledProcessError as e:
            raise HarnessError("multipass file push command failed") from e

    def pull_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(source).is_absolute():
            raise HarnessError(f"path {source} must be absolute")

        LOG.debug(
            "Copying file %s from instance %s to %s", source, instance_id, destination
        )
        try:
            run(
                [
                    "sudo",
                    "multipass",
                    "transfer",
                    f"{instance_id}:{source}",
                    destination,
                ]
            )
        except subprocess.CalledProcessError as e:
            raise HarnessError("multipass file pull command failed") from e

    def exec(self, instance_id: str, command: list, **kwargs):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.debug("Execute command %s in instance %s", command, instance_id)

        if ">" in " ".join(command):
            command_str = " ".join(command)
        else:
            command_str = shlex.join(command)

        return run(
            [
                "sudo",
                "multipass",
                "exec",
                instance_id,
                "--",
                "sudo",
                "bash",
                "-c",
                command_str,
            ],
            **kwargs,
        )

    def delete_instance(self, instance_id: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        try:
            run(["sudo", "multipass", "delete", instance_id])
            run(["sudo", "multipass", "purge"])
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"failed to delete instance {instance_id}") from e

        self.instances.discard(instance_id)

    def cleanup(self):
        for instance_id in self.instances.copy():
            self.delete_instance(instance_id)
