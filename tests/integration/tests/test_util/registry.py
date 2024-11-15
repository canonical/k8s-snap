#
# Copyright 2024 Canonical, Ltd.
#
import logging
from string import Template
from typing import List

from test_util import config
from test_util.harness import Harness, Instance
from test_util.util import get_default_ip

LOG = logging.getLogger(__name__)


class Mirror:
    def __init__(self, name: str, port: int, remote: str):
        """
        Initialize the Mirror object.

        Args:
            name (str): The name of the mirror.
            port (int): The port of the mirror.
            remote (str): The remote URL of the upstream registry.
        """
        self._name = name
        self._port = port
        self._remote = remote

    @property
    def name(self) -> str:
        """
        Get the name of the mirror.

        Returns:
            str: The name of the mirror.
        """
        return self._name

    @property
    def port(self) -> int:
        """
        Get the port of the mirror.

        Returns:
            int: The port of the mirror.
        """
        return self._port

    @property
    def remote(self) -> str:
        """
        Get the remote URL of the upstream registry.

        Returns:
            str: The remote URL of the upstream registry.
        """
        return self._remote


class Registry:

    def __init__(self, h: Harness):
        """
        Initialize the Registry object.

        Args:
            h (Harness): The test harness object.
        """
        self.registry_url = config.REGISTRY_URL
        self.registry_version = config.REGISTRY_VERSION
        self.instance: Instance = None
        self.harness: Harness = h
        self._mirrors: List[Mirror] = [
            Mirror("ghcr.io", 5000, "https://ghcr.io"),
            Mirror("docker.io", 5001, "https://registry-1.docker.io"),
        ]
        self.instance = self.harness.new_instance()

        arch = self.instance.arch
        self.instance.exec(
            [
                "curl",
                "-L",
                f"{self.registry_url}/{self.registry_version}/registry_{self.registry_version[1:]}_linux_{arch}.tar.gz",
                "-o",
                f"/tmp/registry_{self.registry_version}_linux_{arch}.tar.gz",
            ]
        )

        self.instance.exec(
            [
                "tar",
                "xzvf",
                f"/tmp/registry_{self.registry_version}_linux_{arch}.tar.gz",
                "-C",
                "/bin/",
                "registry",
            ],
        )

        self._ip = get_default_ip(self.instance)

        self.add_mirrors()

    def add_mirrors(self):
        for mirror in self._mirrors:
            self.add_mirror(mirror)

    def add_mirror(self, mirror: Mirror):
        substitutes = {
            "NAME": mirror.name,
            "PORT": mirror.port,
            "REMOTE": mirror.remote,
        }

        self.instance.exec(["mkdir", "-p", "/etc/distribution"])
        self.instance.exec(["mkdir", "-p", f"/var/lib/registry/{mirror.name}"])

        with open(
            config.REGISTRY_DIR / "registry-config.yaml", "r"
        ) as registry_template:
            src = Template(registry_template.read())
            self.instance.exec(
                ["dd", f"of=/etc/distribution/{mirror.name}.yaml"],
                input=str.encode(src.substitute(substitutes)),
            )

        with open(config.REGISTRY_DIR / "registry.service", "r") as registry_template:
            src = Template(registry_template.read())
            self.instance.exec(
                ["dd", f"of=/etc/systemd/system/registry-{mirror.name}.service"],
                input=str.encode(src.substitute(substitutes)),
            )

        self.instance.exec(["systemctl", "daemon-reload"])
        self.instance.exec(["systemctl", "enable", f"registry-{mirror.name}.service"])
        self.instance.exec(["systemctl", "start", f"registry-{mirror.name}.service"])

    @property
    def mirrors(self) -> List[Mirror]:
        """
        Get the list of mirrors in the registry.

        Returns:
            List[Mirror]: The list of mirrors.
        """
        return self._mirrors

    @property
    def ip(self) -> str:
        """
        Get the IP address of the registry.

        Returns:
            str: The IP address of the registry.
        """
        return self._ip
