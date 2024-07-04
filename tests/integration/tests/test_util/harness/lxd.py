#
# Copyright 2024 Canonical, Ltd.
#
import logging
import os
import shlex
import subprocess
from pathlib import Path

from test_util import config
from test_util.harness import Harness, HarnessError, Instance
from test_util.util import run, stubbornly, configure_lxd_profile

LOG = logging.getLogger(__name__)


class LXDHarness(Harness):
    """A Harness that creates an LXD container for each instance."""

    name = "lxd"

    def next_id(self) -> int:
        self._next_id += 1
        return self._next_id

    def __init__(self):
        super(LXDHarness, self).__init__()

        self._next_id = 0

        self.profile = config.LXD_PROFILE_NAME
        self.dualstack_profile = None
        self.sideload_images_dir = config.LXD_SIDELOAD_IMAGES_DIR
        self.image = config.LXD_IMAGE
        self.instances = set()

        configure_lxd_profile(self.profile, config.LXD_PROFILE)

        LOG.debug(
            "Configured LXD substrate (profile %s, image %s)", self.profile, self.image
        )

    def new_instance(self, dualstack: bool = False) -> Instance:
        instance_id = f"k8s-integration-{os.urandom(3).hex()}-{self.next_id()}"

        LOG.debug("Creating instance %s with image %s", instance_id, self.image)
        launch_lxd_command = [
            "lxc",
            "launch",
            self.image,
            instance_id,
            "-p",
            "default",
            "-p",
            self.profile,
        ]

        if dualstack:
            # Setup dualstack profile once
            if self.dualstack_profile is None:
                self.dualstack_profile = config.LXD_DUALSTACK_PROFILE_NAME
                configure_lxd_profile(
                    self.dualstack_profile,
                    config.LXD_DUALSTACK_PROFILE,
                    template_overwrites={
                        "LXD_DUALSTACK_NETWORK": config.LXD_DUALSTACK_NETWORK
                    },
                )
            launch_lxd_command.extend(["-p", self.dualstack_profile])

        try:
            stubbornly(retries=3, delay_s=1).exec(launch_lxd_command)
            self.instances.add(instance_id)

            if self.sideload_images_dir:
                stubbornly(retries=3, delay_s=1).exec(
                    [
                        "lxc",
                        "config",
                        "device",
                        "add",
                        instance_id,
                        "k8s-e2e-images",
                        "disk",
                        f"source={self.sideload_images_dir}",
                        "path=/mnt/images",
                        "readonly=true",
                    ]
                )

                self.exec(
                    instance_id,
                    ["mkdir", "-p", "/var/snap/k8s/common"],
                )
                self.exec(
                    instance_id,
                    ["cp", "-rv", "/mnt/images", "/var/snap/k8s/common/images"],
                )
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"Failed to create LXD container {instance_id}") from e

        # TODO: why does this block/timeout?
        self.exec(instance_id, ["snap", "wait", "system", "seed.loaded"])
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
                capture_output=True,
            )
            run(
                ["lxc", "file", "push", source, f"{instance_id}{destination}"],
                capture_output=True,
            )
        except subprocess.CalledProcessError as e:
            LOG.error("command {e.cmd} failed")
            LOG.error(f"  {e.returncode=}")
            LOG.error(f"  {e.stdout.decode()=}")
            LOG.error(f"  {e.stderr.decode()=}")
            raise HarnessError("failed to push file") from e

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
                ["lxc", "file", "pull", f"{instance_id}{source}", destination],
                stdout=subprocess.DEVNULL,
            )
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
