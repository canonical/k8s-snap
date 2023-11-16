#
# Copyright 2023 Canonical, Ltd.
#
import logging
import shlex
import shutil
import socket
import subprocess
import time
from pathlib import Path

import config
import pytest

LOG = logging.getLogger(__name__)


def run(command: list, **kwargs) -> subprocess.CompletedProcess:
    """Log and run command."""
    kwargs.setdefault("check", True)

    LOG.debug("Execute command %s (kwargs=%s)", shlex.join(command), kwargs)
    return subprocess.run(command, **kwargs)


class HarnessError(Exception):
    """Base error for all our harness failures"""

    pass


class Harness:
    """Abstract how e2e tests can start and manage multiple machines. This allows writing
    e2e tests that can run on the local machine, LXD, or Multipass with minimum effort.
    """

    def new_instance(self) -> str:
        """
        creates a new instance on the infrastructure and returns an ID that
        can be used to interact with it.

        :raise: an exception in case the operation failed
        """
        raise NotImplementedError

    def send_file(self, instance_id: str, source: str, destination: str):
        """send a local file to the instance.

        :param instance_id: The instance_id, as returned by new_instance()
        :param source: Path to the file that will be copied to the instance
        :param destination: Path in the instance where the file will be copied.
                                 This must always be an absolute path.

        :raise: an exception in case the operation failed
        """
        raise NotImplementedError

    def pull_file(self, instance_id: str, source: str, destination: str):
        """pull a file from the instance and save it on the local machine

        :param instance_id: The instance_id, as returned by new_instance()
        :param source: Path to the file that will be copied from the instance.
                            This must always be an absolute path.
        :param destination: Path on the local machine the file will be saved.

        :raise: an exception in case the operation failed
        """
        raise NotImplementedError

    def exec(
        self, instance_id: str, command: list, **kwargs
    ) -> subprocess.CompletedProcess:
        """run a command as root on the instance.

        :param instance_id: The instance_id, as returned by new_instance()
        :param command: Command for subprocess.run()
        :param kwargs: Keyword args compatible with subprocess.run()

        :raise: an exception in case the operation failed
        """
        raise NotImplementedError

    def delete_instance(self, instance_id: str):
        """delete a previously created instance.

        :param instance_id: The instance_id, as returned by new_instance()

        :raise: an exception in case the operation failed
        """
        raise NotImplementedError

    def cleanup(self):
        """delete any leftover resources after the tests are done, e.g. delete any instances
        that might still be running"""
        raise NotImplementedError


class LXDHarness(Harness):
    """A Harness that creates an LXD container for each instance."""

    def __init__(self):
        super(LXDHarness, self).__init__()

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
            subprocess.run(
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
        instance_id = f"k8s-e2e-{int(time.time())}-{len(self.instances)}"

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


class LocalHarness(Harness):
    """A Harness that uses the local machine. Asking for more than 1 instance will fail."""

    def __init__(self):
        super(LocalHarness, self).__init__()
        self.initialized = False
        self.hostname = socket.gethostname().lower()

        LOG.debug("Configured local substrate")

    def new_instance(self) -> str:
        if self.initialized:
            raise HarnessError("local substrate only supports up to one instance")

        self.initialized = True
        LOG.debug("Initializing instance")
        try:
            self.exec(self.hostname, ["snap", "wait", "system", "seed.loaded"])
        except subprocess.CalledProcessError as e:
            raise HarnessError("failed to wait for snapd seed") from e

        return self.hostname

    def send_file(self, _: str, source: str, destination: str):
        if not self.initialized:
            raise HarnessError("no instance initialized")

        if not Path(destination).is_absolute():
            raise HarnessError(f"path {destination} must be absolute")

        LOG.debug("Copying file %s to %s", source, destination)
        try:
            shutil.copy(source, destination)
        except shutil.SameFileError:
            pass

    def pull_file(self, _: str, source: str, destination: str):
        return self.send_file(_, destination, source)

    def exec(self, _: str, command: list, **kwargs):
        if not self.initialized:
            raise HarnessError("no instance initialized")

        LOG.debug("Executing command %s on %s", command, self.hostname)
        return run(["sudo", "-E", "bash", "-c", shlex.join(command)], **kwargs)

    def delete_instance(self, _: str):
        LOG.debug("Stopping instance")
        self.initialized = False

    def cleanup(self):
        LOG.debug("Stopping instance")
        self.initialized = False


@pytest.fixture(scope="session")
def h() -> Harness:
    LOG.debug("Create harness for %s", config.SUBSTRATE)
    if config.SUBSTRATE == "local":
        harness = LocalHarness()
    elif config.SUBSTRATE == "lxd":
        harness = LXDHarness()
    else:
        raise HarnessError("TEST_SUBSTRATE must be one of: lxd, multipass, local")

    yield harness

    if config.SKIP_CLEANUP:
        return

    LOG.debug("Cleanup")
    harness.cleanup()
