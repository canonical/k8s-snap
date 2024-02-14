#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import shlex
import subprocess
from pathlib import Path

from e2e_util import config
from e2e_util.harness import Harness, HarnessError, Instance
from e2e_util.util import run

LOG = logging.getLogger(__name__)


class JujuHarness(Harness):
    """A Harness that creates an Juju machine for each instance."""

    name = "juju"

    def __init__(self):
        super(JujuHarness, self).__init__()

        self.model = config.JUJU_MODEL
        if not self.model:
            raise HarnessError("Set JUJU_MODEL to the Juju model to use")

        if config.JUJU_CONTROLLER:
            self.model = f"{config.JUJU_CONTROLLER}:{self.model}"

        self.constraints = config.JUJU_CONSTRAINTS
        self.base = config.JUJU_BASE
        self.existing_machines = {}
        self.instances = set()

        if config.JUJU_MACHINES:
            self.existing_machines = {
                instance_id.strip(): False
                for instance_id in config.JUJU_MACHINES.split()
            }
            LOG.debug(
                "Configured Juju substrate (model %s, machines %s)",
                self.model,
                config.JUJU_MACHINES,
            )

        else:
            LOG.debug(
                "Configured Juju substrate (model %s, base %s, constraints %s)",
                self.model,
                self.base,
                self.constraints,
            )

    def new_instance(self) -> Instance:
        for instance_id in self.existing_machines:
            if not self.existing_machines[instance_id]:
                LOG.debug("Reusing existing machine %s", instance_id)
                self.existing_machines[instance_id] = True
                self.instances.add(instance_id)
                return Instance(self, instance_id)

        LOG.debug("Creating instance with constraints %s", self.constraints)
        try:
            p = run(
                [
                    "juju",
                    "add-machine",
                    "-m",
                    self.model,
                    "--constraints",
                    self.constraints,
                    "--base",
                    self.base,
                ],
                capture_output=True,
            )

            output = p.stderr.decode().strip()
            if not output.startswith("created machine "):
                raise HarnessError(f"failed to parse output from juju add-machine {p=}")

            instance_id = output.split(" ")[2]
        except subprocess.CalledProcessError as e:
            raise HarnessError("Failed to create Juju machine") from e

        self.instances.add(instance_id)

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
            )
            run(["juju", "scp", source, f"{instance_id}:{destination}"])
        except subprocess.CalledProcessError as e:
            raise HarnessError("juju scp command failed") from e

    def pull_file(self, instance_id: str, source: str, destination: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if not Path(source).is_absolute():
            raise HarnessError(f"path {source} must be absolute")

        LOG.debug(
            "Copying file %s from instance %s to %s", source, instance_id, destination
        )
        try:
            run(["juju", "scp", f"{instance_id}:{source}", destination])
        except subprocess.CalledProcessError as e:
            raise HarnessError("juju scp command failed") from e

    def exec(self, instance_id: str, command: list, **kwargs):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        LOG.debug("Execute command %s in instance %s", command, instance_id)
        capture_output = kwargs.pop("capture_output", False)
        check = kwargs.pop("check", True)
        stdout = kwargs.pop("stdout", None)
        stderr = kwargs.pop("stderr", None)
        input = f" <<EOF\n{b.decode()}\nEOF" if (b := kwargs.pop("input", None)) else ""
        s_result = run(
            [
                "juju",
                "exec",
                "-m",
                self.model,
                "--machine",
                instance_id,
                "--format",
                "json",
                "--wait",
                "30m",
                "--",
                "sudo",
                "-E",
                "bash",
                "-c",
                shlex.join(command) + input,
            ],
            capture_output=True,
            check=False,
            **kwargs,
        )
        if check:
            s_result.check_returncode()
        juju_response = json.loads(s_result.stdout.decode())
        results = juju_response[instance_id]["results"]
        stdout = (
            b.encode()
            if (b := results.get("stdout"))
            and (capture_output or stdout == subprocess.PIPE)
            else None
        )
        stderr = (
            b.encode()
            if (b := results.get("stderr"))
            and (capture_output or stderr != subprocess.DEVNULL)
            else None
        )
        completed = subprocess.CompletedProcess(
            command, results["return-code"], stdout, stderr
        )
        if check:
            completed.check_returncode()
        return completed

    def delete_instance(self, instance_id: str):
        if instance_id not in self.instances:
            raise HarnessError(f"unknown instance {instance_id}")

        if self.existing_machines.get(instance_id):
            # For existing machines, simply mark it as free
            LOG.debug("No longer using machine %s", instance_id)
            self.existing_machines[instance_id] = False
        else:
            # Remove the machine
            LOG.debug("Removing machine %s", instance_id)
            try:
                run(["juju", "remove-machine", instance_id, "--force", "--no-prompt"])
            except subprocess.CalledProcessError as e:
                raise HarnessError(f"failed to delete instance {instance_id}") from e

        self.instances.discard(instance_id)

    def cleanup(self):
        for instance_id in self.instances.copy():
            self.delete_instance(instance_id)
