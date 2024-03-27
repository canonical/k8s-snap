#
# Copyright 2024 Canonical, Ltd.
#
import subprocess
from functools import partial


class HarnessError(Exception):
    """Base error for all our harness failures"""

    pass


class Instance:
    """Reference to a harness and a given instance id.

    Provides convenience methods for an instance to call its harness' methods
    """

    def __init__(self, h: "Harness", id: str) -> None:
        self._h = h
        self._id = id

        self.send_file = partial(h.send_file, id)
        self.pull_file = partial(h.pull_file, id)
        self.exec = partial(h.exec, id)
        self.delete_instance = partial(h.delete_instance, id)

    @property
    def id(self) -> str:
        return self._id

    def __str__(self) -> str:
        return f"{self._h.name}:{self.id}"


class Harness:
    """Abstract how integration tests can start and manage multiple machines. This allows
    writing integration tests that can run on the local machine, LXD, or Multipass with minimum
    effort.
    """

    name: str

    def new_instance(self) -> Instance:
        """Creates a new instance on the infrastructure and returns an object
        which can be used to interact with it.

        If the operation fails, a HarnessError is raised.
        """
        raise NotImplementedError

    def send_file(self, instance_id: str, source: str, destination: str):
        """Send a local file to the instance.

        :param instance_id: The instance_id, as returned by new_instance()
        :param source: Path to the file that will be copied to the instance
        :param destination: Path in the instance where the file will be copied.
                                 This must always be an absolute path.


        If the operation fails, a HarnessError is raised.
        """
        raise NotImplementedError

    def pull_file(self, instance_id: str, source: str, destination: str):
        """Pull a file from the instance and save it on the local machine

        :param instance_id: The instance_id, as returned by new_instance()
        :param source: Path to the file that will be copied from the instance.
                            This must always be an absolute path.
        :param destination: Path on the local machine the file will be saved.

        If the operation fails, a HarnessError is raised.
        """
        raise NotImplementedError

    def exec(
        self, instance_id: str, command: list, **kwargs
    ) -> subprocess.CompletedProcess:
        """Run a command as root on the instance.

        :param instance_id: The instance_id, as returned by new_instance()
        :param command: Command for subprocess.run()
        :param kwargs: Keyword args compatible with subprocess.run()

        If the operation fails, a subprocesss.CalledProcessError is raised.
        """
        raise NotImplementedError

    def delete_instance(self, instance_id: str):
        """Delete a previously created instance.

        :param instance_id: The instance_id, as returned by new_instance()

        If the operation fails, a HarnessError is raised.
        """
        raise NotImplementedError

    def cleanup(self):
        """Delete any leftover resources after the tests are done, e.g. delete any
        instances that might still be running.

        If the operation fails, a HarnessError is raised.
        """
        raise NotImplementedError
