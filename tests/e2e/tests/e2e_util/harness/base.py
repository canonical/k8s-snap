#
# Copyright 2023 Canonical, Ltd.
#
import subprocess


class HarnessError(Exception):
    """Base error for all our harness failures"""

    pass


class Harness:
    """Abstract how e2e tests can start and manage multiple machines. This allows
    writing e2e tests that can run on the local machine, LXD, or Multipass with minimum
    effort.
    """

    def new_instance(self) -> str:
        """Creates a new instance on the infrastructure and returns an ID that
        can be used to interact with it.

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
