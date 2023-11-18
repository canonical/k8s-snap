#
# Copyright 2023 Canonical, Ltd.
#
import logging
import shlex
import subprocess
import time
from pathlib import Path

import config
from harness_base import Harness, HarnessError
from util import run

LOG = logging.getLogger(__name__)


class LXDHarness(Harness):
    """A Harness that creates an LXD container for each instance."""

    def next_id(self) -> int:
        self._next_id += 1
        return self._next_id

    def __init__(self):
        super(LXDHarness, self).__init__()

        self._next_id = 0

        self.profile = config.LXD_PROFILE_NAME
        self.image = config.LXD_IMAGE
        self.instances = set()

        LOG.debug("Checking for LXD profile %s", self.profile)
        try:
            run(["lxc", "profile", "show", self.profile])
        except subprocess.CalledProcessError:
            try:
                LOG.debug("Creating LXD profile %s", self.profile)
                run(["lxc", "profile", "create", self.profile])

            except subprocess.CalledProcessError as e:
                raise HarnessError(
                    f"Failed to create LXD profile {self.profile}"
                ) from e

        try:
            LOG.debug("Configuring LXD profile %s", self.profile)
            run(
                ["lxc", "profile", "edit", self.profile],
                input=config.LXD_PROFILE.encode(),
            )
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"Failed to configure LXD profile {self.profile}") from e

        LOG.debug(
            "Configured LXD substrate (profile %s, image %s)", self.profile, self.image
        )

    def new_instance(self) -> str:
        # TODO(neoaggelos): make this unique
        instance_id = f"k8s-e2e-{int(time.time())}-{self.next_id()}"

        LOG.debug("Creating instance %s with image %s", instance_id, self.image)
        try:
            run(
                [
                    "lxc",
                    "launch",
                    self.image,
                    instance_id,
                    "-p",
                    "default",
                    "-p",
                    self.profile,
                ]
            )
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"Failed to create LXD container {instance_id}") from e

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
            run(["lxc", "file", "push", source, f"{instance_id}{destination}"])
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
            run(["lxc", "file", "pull", f"{instance_id}{source}", destination])
        except subprocess.CalledProcessError as e:
            raise HarnessError("lxc file push command failed") from e

    def exec(self, instance_id: str, command: list, **kwargs):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.debug("Execute command %s in instance %s", command, instance_id)
        return run(
            ["lxc", "shell", instance_id, "--", "bash", "-c", shlex.join(command)],
            **kwargs,
        )

    def delete_instance(self, instance_id: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        try:
            run(["lxc", "rm", instance_id, "--force"])
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"failed to delete instance {instance_id}") from e

        self.instances.discard(instance_id)

    def cleanup(self):
        for instance_id in self.instances.copy():
            self.delete_instance(instance_id)
