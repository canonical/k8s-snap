#
# Copyright 2026 Canonical, Ltd.
#
import logging
import shlex
import shutil
import subprocess
from pathlib import Path
from typing import List

from test_util import config
from test_util.harness import Harness, HarnessError, Instance
from test_util.util import run

LOG = logging.getLogger(__name__)

SSH_OPTS = [
    "-o",
    "StrictHostKeyChecking=no",
    "-o",
    "UserKnownHostsFile=/dev/null",
    "-o",
    "LogLevel=ERROR",
]


class BareMetalHarness(Harness):
    """A Harness that runs commands on an existing machine.

    Unlike LXD or Multipass, this harness does not create or destroy machines.
    It treats the target host as a single node.

    When the host is 'localhost', commands are run directly via subprocess
    (no SSH overhead). Otherwise, commands are run over SSH.

    This is useful for testflinger environments where the device under test
    (DUT) has hardware (e.g. GPUs) that cannot be virtualized.
    """

    name = "bare_metal"

    def __init__(self):
        super().__init__()

        self.ssh_host = config.BARE_METAL_SSH_HOST
        self.ssh_user = config.BARE_METAL_SSH_USER
        self._is_local = self.ssh_host in ("localhost", "127.0.0.1", "")

        if not self._is_local and not self.ssh_host:
            raise HarnessError(
                "TEST_BARE_METAL_SSH_HOST must be set when using the bare_metal substrate"
            )

        self.instances = set()
        self._instance_id = f"bare-metal-{self.ssh_host or 'localhost'}"

        LOG.debug(
            "Configured bare_metal substrate (host=%s, user=%s, local=%s)",
            self.ssh_host,
            self.ssh_user,
            self._is_local,
        )

    def _ssh_target(self) -> str:
        return f"{self.ssh_user}@{self.ssh_host}"

    def new_instance(
        self, network_type: str = "IPv4", name_suffix: str = ""
    ) -> Instance:
        if self.instances:
            raise HarnessError(
                "bare_metal harness only supports a single instance. "
                "Set @pytest.mark.node_count(1) on your test."
            )

        instance_id = self._instance_id
        self.instances.add(instance_id)

        if not self._is_local:
            # Verify SSH connectivity before proceeding.
            try:
                run(
                    ["ssh", *SSH_OPTS, self._ssh_target(), "echo", "ok"],
                    capture_output=True,
                    timeout=30,
                )
            except (subprocess.CalledProcessError, subprocess.TimeoutExpired) as e:
                self.instances.discard(instance_id)
                raise HarnessError(
                    f"Cannot connect to bare_metal host {self.ssh_host} via SSH"
                ) from e

        LOG.info("Connected to bare_metal instance %s", instance_id)
        instance = Instance(self, instance_id)

        # Wait for snapd to be seeded (same as LXD harness).
        try:
            instance.exec(
                ["snap", "wait", "system", "seed.loaded"], capture_output=True
            )
        except subprocess.CalledProcessError:
            LOG.warning("snap wait seed.loaded failed, continuing anyway")

        return instance

    def send_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(destination).is_absolute():
            raise HarnessError(f"path {destination} must be absolute")

        LOG.debug(
            "Copying file %s to %s at %s", source, instance_id, destination
        )
        try:
            self.exec(
                instance_id,
                ["mkdir", "-m=0777", "-p", Path(destination).parent.as_posix()],
                capture_output=True,
            )
            if self._is_local:
                src = Path(source).resolve()
                dst = Path(destination).resolve()
                if src != dst:
                    shutil.copy2(source, destination)
            else:
                run(
                    ["scp", *SSH_OPTS, source, f"{self._ssh_target()}:{destination}"],
                    capture_output=True,
                )
        except subprocess.CalledProcessError as e:
            raise HarnessError(
                f"failed to send file {source} to {destination}"
            ) from e

    def pull_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(source).is_absolute():
            raise HarnessError(f"path {source} must be absolute")

        LOG.debug(
            "Pulling file %s from %s to %s", source, instance_id, destination
        )
        try:
            if self._is_local:
                src = Path(source).resolve()
                dst = Path(destination).resolve()
                if src != dst:
                    shutil.copy2(source, destination)
            else:
                run(
                    ["scp", *SSH_OPTS, f"{self._ssh_target()}:{source}", destination],
                    capture_output=True,
                )
        except subprocess.CalledProcessError as e:
            raise HarnessError(
                f"failed to pull file {source} from {instance_id}"
            ) from e

    def exec(
        self, instance_id: str, command: list, **kwargs
    ) -> subprocess.CompletedProcess:
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.debug("Execute command %s on %s", command, instance_id)

        if self._is_local:
            return run(["sudo", *command], **kwargs)
        else:
            command_str = shlex.join(command)
            return run(
                ["ssh", *SSH_OPTS, self._ssh_target(), "--", "sudo", "bash", "-c",
                 command_str],
                **kwargs,
            )

    def restart_instance(self, instance_id: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.info("Rebooting bare_metal instance %s", instance_id)

        if self._is_local:
            raise HarnessError(
                "Cannot reboot localhost — would kill the test process"
            )

        try:
            run(
                ["ssh", *SSH_OPTS, self._ssh_target(), "sudo", "reboot"],
                check=False,
                timeout=10,
            )
        except subprocess.TimeoutExpired:
            pass

        # Wait for the machine to come back.
        import time

        for _ in range(60):
            try:
                run(
                    ["ssh", *SSH_OPTS, self._ssh_target(), "echo", "ok"],
                    capture_output=True,
                    timeout=10,
                )
                LOG.info("Instance %s is back after reboot", instance_id)
                return
            except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
                time.sleep(5)

        raise HarnessError(f"Instance {instance_id} did not come back after reboot")

    def open_ports(self, instance_id: str, ports: List[int]):
        # No firewall management on bare metal - ports are open by default.
        pass

    def delete_instance(self, instance_id: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.info("Cleaning up bare_metal instance %s (removing k8s snap)", instance_id)
        try:
            self.exec(
                instance_id,
                ["snap", "remove", config.SNAP_NAME, "--purge"],
                check=False,
                capture_output=True,
            )
        except subprocess.CalledProcessError:
            LOG.warning("Failed to remove k8s snap during cleanup")

        self.instances.discard(instance_id)

    def log_environment_info(self):
        """Log relevant environment information."""
        try:
            if self.instances:
                instance_id = next(iter(self.instances))
                LOG.info("Bare metal host info:")
                result = self.exec(
                    instance_id, ["uname", "-a"], capture_output=True, text=True
                )
                LOG.info("  %s", result.stdout.strip())

                result = self.exec(
                    instance_id, ["df", "-h", "/"], capture_output=True, text=True
                )
                LOG.info("  Disk: %s", result.stdout.strip())
        except Exception:
            LOG.exception("Failed to obtain environment info")

    def cleanup(self):
        for instance_id in self.instances.copy():
            self.delete_instance(instance_id)
