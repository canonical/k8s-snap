#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import os
import re
import shlex
import subprocess
from pathlib import Path
from typing import List

from test_util import config
from test_util.harness import Harness, HarnessError, Instance
from test_util.util import run, stubbornly

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
        self.cloud_init = config.MULTIPASS_CLOUD_INIT
        self.instances = set()

        LOG.debug("Configured Multipass substrate (image %s)", self.image)

    def new_instance(
        self, network_type: str = "IPv4", name_suffix: str = ""
    ) -> Instance:
        if network_type not in ("IPv4", "IPv6", "dualstack"):
            raise HarnessError(
                "Unknown network type: %s; supported types: IPv4, IPv6, dualstack",
                network_type,
            )

        instance_id = (
            f"k8s-integration-{self.next_id()}-{os.urandom(3).hex()}{name_suffix}"
        )

        LOG.debug("Creating instance %s with image %s", instance_id, self.image)
        try:
            cmd = [
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

                cloud_init_content = Path(
                    config.CLOUD_INIT_DIR / self.cloud_init
                ).read_text()

                # Replace environment variables in the format ${VAR} or $VAR
                def replace_env_var(match):
                    LOG.info(match)
                    var_name = match.group(1) or match.group(2)
                    LOG.info(var_name)
                    return os.environ.get(var_name, match.group(0))

                LOG.info(os.environ)
                cloud_init_content = re.sub(
                    r"\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)",
                    replace_env_var,
                    cloud_init_content,
                )

                LOG.info("Using cloud-init: %s", cloud_init_content)
                # Note(ben): Multipass does not handle restarts in
                # cloud-init very well and the command times out even if
                # the underlying machine works just fine.
                # See https://github.com/canonical/multipass/issues/4199
                # Hence, we don't fail on the timeout of this command, and manually wait until
                # the cloud-init is done.
                run(
                    cmd + ["--cloud-init", "-"],
                    input=cloud_init_content.encode(),
                    sensitive_kwargs=True,
                    check=False,
                )
                stubbornly(retries=200, delay_s=10).until(
                    lambda p: json.loads(p.stdout).get("status") == "done"
                ).exec(
                    [
                        "multipass",
                        "exec",
                        instance_id,
                        "--",
                        "cloud-init",
                        "status",
                        "--format",
                        "json",
                    ],
                    capture_output=True,
                    # cloud-init returns 2 even when it is done.
                    check=False,
                    text=True,
                    timeout=20,
                )

            else:
                run(cmd)

        except subprocess.CalledProcessError as e:
            raise HarnessError(f"Failed to create multipass VM {instance_id}") from e

        self.instances.add(instance_id)

        instance = Instance(self, instance_id)
        stubbornly(retries=5, delay_s=5).on(instance).exec(
            ["snap", "wait", "system", "seed.loaded"]
        )
        if network_type in ("IPv6", "dualstack"):
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

        return instance

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

    def restart_instance(self, instance_id):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        try:
            stubbornly(delay_s=5, retries=20).exec(
                ["multipass", "restart", instance_id], timeout=60 * 5
            )
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"failed to restart instance {instance_id}") from e

    def open_ports(self, instance_id: str, ports: List[int]):
        """Open ports on the instance.

        :param instance_id: The instance_id, as returned by new_instance()
        :param ports: List of ports to open on the instance.

        Ports will be opened on a best effort basis. If the port is already open,
        or UFW is not installed, no error will be raised.
        """
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        for port in ports:
            LOG.debug("Opening port %s on instance %s", port, instance_id)
            # UFW might not be installed, if this is the case, then no firewall
            # is active and nothing needs to be done.
            self.exec(instance_id, ["sudo", "ufw", "allow", str(port)], check=False)

    def delete_instance(self, instance_id: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        try:
            run(["multipass", "delete", instance_id])
            run(["multipass", "purge"])
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"failed to delete instance {instance_id}") from e

        self.instances.discard(instance_id)

    def cleanup(self):
        for instance_id in self.instances.copy():
            self.delete_instance(instance_id)
