#
# Copyright 2024 Canonical, Ltd.
#
import logging
import os
import shlex
import subprocess
from pathlib import Path

from e2e_util import config
from e2e_util.harness import Harness, HarnessError
from e2e_util.util import run

LOG = logging.getLogger(__name__)


class MultipassHarness(Harness):
    """A Harness that creates a Multipass VM for each instance."""

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
        self.instances = set()

        LOG.debug("Configured Multipass substrate (image %s)", self.image)

    def new_instance(self) -> str:
        instance_id = f"k8s-e2e-{os.urandom(3).hex()}-{self.next_id()}"

        LOG.debug("Creating instance %s with image %s", instance_id, self.image)
        try:
            run(
                [
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
            )
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"Failed to create multipass VM {instance_id}") from e

        self.instances.add(instance_id)

        self.exec(instance_id, ["snap", "wait", "system", "seed.loaded"])
        return instance_id

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
            run(["multipass", "transfer", source, f"{instance_id}:{destination}"])
        except subprocess.CalledProcessError as e:
            raise HarnessError("lxc file push command failed") from e

    def pull_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(source).is_absolute():
            raise HarnessError(f"path {source} must be absolute")

        LOG.debug(
            "Copying file %s from instance %s to %s", source, instance_id, destination
        )
        try:
            run(["multipass", "transfer", f"{instance_id}:{source}", destination])
        except subprocess.CalledProcessError as e:
            raise HarnessError("lxc file push command failed") from e

    def exec(self, instance_id: str, command: list, **kwargs):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.debug("Execute command %s in instance %s", command, instance_id)
        return run(
            [
                "multipass",
                "exec",
                instance_id,
                "--",
                "sudo",
                "bash",
                "-c",
                shlex.join(command),
            ],
            **kwargs,
        )

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
