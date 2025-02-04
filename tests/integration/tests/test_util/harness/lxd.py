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
from test_util.util import run, stubbornly

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

        self._configure_profile(self.profile, config.LXD_PROFILE)

        self._configure_network(
            config.LXD_DUALSTACK_NETWORK,
            "ipv4.address=auto",
            "ipv6.address=auto",
            "ipv4.nat=true",
            "ipv6.nat=true",
        )
        self.dualstack_profile = config.LXD_DUALSTACK_PROFILE_NAME
        self._configure_profile(
            self.dualstack_profile,
            config.LXD_DUALSTACK_PROFILE.replace(
                "LXD_DUALSTACK_NETWORK", config.LXD_DUALSTACK_NETWORK
            ),
        )

        self._configure_network(
            config.LXD_IPV6_NETWORK,
            "ipv4.address=none",
            "ipv6.address=auto",
            "ipv4.nat=false",
            "ipv6.nat=true",
        )
        self.ipv6_profile = config.LXD_IPV6_PROFILE_NAME
        self._configure_profile(
            self.ipv6_profile,
            config.LXD_IPV6_PROFILE.replace(
                "LXD_IPV6_NETWORK", config.LXD_IPV6_NETWORK
            ),
        )

        LOG.debug(
            "Configured LXD substrate (profile %s, image %s)", self.profile, self.image
        )

    def new_instance(self, network_type: str = "IPv4") -> Instance:
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

        if network_type.lower() not in ["ipv4", "dualstack", "ipv6"]:
            raise HarnessError(
                f"unknown network type {network_type}, need to be one of 'IPv4', 'IPv6', 'dualstack'"
            )

        if network_type.lower() == "dualstack":
            if self.dualstack_profile is None:
                raise ValueError(
                    "LXD `dualstack_profile` must be set for network_type=dualstack"
                )
            launch_lxd_command.extend(["-p", self.dualstack_profile])

        if network_type.lower() == "ipv6":
            launch_lxd_command.extend(["-p", self.ipv6_profile])

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

        self.exec(instance_id, ["snap", "wait", "system", "seed.loaded"])
        return Instance(self, instance_id)

    def _configure_profile(self, profile_name: str, profile_config: str):
        LOG.debug("Checking for LXD profile %s", profile_name)
        try:
            run(["lxc", "profile", "show", profile_name])
        except subprocess.CalledProcessError:
            try:
                LOG.debug("Creating LXD profile %s", profile_name)
                run(["lxc", "profile", "create", profile_name])

            except subprocess.CalledProcessError as e:
                raise HarnessError(
                    f"Failed to create LXD profile {profile_name}"
                ) from e

        try:
            LOG.debug("Configuring LXD profile %s", profile_name)
            run(
                ["lxc", "profile", "edit", profile_name],
                input=profile_config.encode(),
            )
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"Failed to configure LXD profile {profile_name}") from e

    def _configure_network(self, network_name: str, *network_args: str):
        LOG.debug("Checking for LXD network %s", network_name)
        try:
            run(["lxc", "network", "show", network_name])
        except subprocess.CalledProcessError:
            try:
                LOG.debug("Creating LXD network %s", network_name)
                run(["lxc", "network", "create", network_name, *network_args])

            except subprocess.CalledProcessError as e:
                raise HarnessError(
                    f"Failed to create LXD network {network_name}"
                ) from e

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
            # There are cases where the instance is not deleted properly and this command is stuck.
            # A timeout prevents this.
            # TODO(ben): This is a workaround for an issue that arises because of our use of
            # privileged containers. We eventually move away from this (not supported >24.10)
            # which should also fix this issue and make this timeout unnecessary.
            run(["lxc", "rm", instance_id, "--force"], timeout=60 * 5)
        except subprocess.CalledProcessError as e:
            raise HarnessError(f"failed to delete instance {instance_id}") from e
        except subprocess.TimeoutExpired:
            LOG.warning("LXC container removal timed out.")
            pass

        self.instances.discard(instance_id)

    def cleanup(self):
        for instance_id in self.instances.copy():
            self.delete_instance(instance_id)
