#
# Copyright 2024 Canonical, Ltd.
#

import logging
import shlex
import shutil
import socket
import subprocess
from pathlib import Path
from typing import Set

from test_util.harness import Harness, HarnessError, Instance
from test_util.util import run

LOG = logging.getLogger(__name__)


class LocalHarness(Harness):
    """A Harness that uses the local machine. Asking for more than 1 instance will fail."""

    name = "local"

    def __init__(self):
        super(LocalHarness, self).__init__()
        self.instance = None
        self.hostname = socket.gethostname().lower()

        LOG.debug("Configured local substrate")

    def new_instance(self, dualstack: bool = False) -> Instance:
        if self.instance is not None:
            raise HarnessError("local substrate only supports up to one instance")

        if dualstack:
            raise HarnessError("Dualstack is currently not supported by Local harness")

        LOG.debug("Initializing instance")
        try:
            self.exec(self.hostname, ["snap", "wait", "system", "seed.loaded"])
        except subprocess.CalledProcessError as e:
            raise HarnessError("failed to wait for snapd seed") from e

        self.instance = self.hostname
        return Instance(self, self.hostname)


    def get_instances(self) -> Set[str]:
        if self.instance is None:
            return set()
        return set(self.instance)

    def send_file(self, _: str, source: str, destination: str):
        if not self.initialized:
            raise HarnessError("no instance initialized")

        if not Path(destination).is_absolute():
            raise HarnessError(f"path {destination} must be absolute")

        LOG.debug("Copying file %s to %s", source, destination)
        try:
            self.exec(
                _, ["mkdir", "-m=0777", "-p", Path(destination).parent.as_posix()]
            )
            shutil.copy(source, destination)
        except subprocess.CalledProcessError as e:
            raise HarnessError("failed to copy file") from e
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
