#
# Copyright 2025 Canonical, Ltd.
#
import logging
import os
from pathlib import Path
from string import Template
from typing import List, Optional

from test_util import config
from test_util.harness import Harness, Instance
from test_util.util import get_default_ip, setup_k8s_snap

LOG = logging.getLogger(__name__)


class Mirror:
    def __init__(
        self,
        name: str,
        port: int,
        remote: str,
        username: Optional[str] = None,
        password: Optional[str] = None,
    ):
        """
        Initialize the Mirror object.

        Args:
            name (str): The name of the mirror.
            port (int): The port of the mirror.
            remote (str): The remote URL of the upstream registry.
            username (str, optional): Authentication username.
            password (str, optional): Authentication password.
        """
        self.name = name
        self.port = port
        self.remote = remote
        self.username = username
        self.password = password


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
        self._mirrors: List[Mirror] = self.get_configured_mirrors()
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

        # Setup the k8s snap on the instance.
        # Use the latest/edge/classic channel as this version is only used to collect logs.
        # This would fail if the `TEST_SNAP` environment variable is not set which however has
        # valid use cases, e.g. in the promotion scenarios.
        setup_k8s_snap(self.instance, Path("/"), "latest/edge/classic")

    def get_configured_mirrors(self) -> List[Mirror]:
        mirrors: List[Mirror] = []
        for mirror_dict in config.MIRROR_LIST:
            for field in ["name", "port", "remote"]:
                if field not in mirror_dict:
                    raise Exception(
                        f"Invalid 'TEST_MIRROR_LIST' configuration. Missing field: {field}"
                    )

            mirror = Mirror(
                mirror_dict["name"],
                mirror_dict["port"],
                mirror_dict["remote"],
                mirror_dict.get("username"),
                mirror_dict.get("password"),
            )
            mirrors.append(mirror)
        return mirrors

    def add_mirrors(self):
        for mirror in self._mirrors:
            self.add_mirror(mirror)

    def add_mirror(self, mirror: Mirror):
        substitutes = {
            "NAME": mirror.name,
            "PORT": mirror.port,
            "REMOTE": mirror.remote,
            "USERNAME": mirror.username or "",
            "PASSWORD": mirror.password or "",
        }

        self.instance.exec(["mkdir", "-p", "/etc/distribution"])
        self.instance.exec(["mkdir", "-p", f"/var/lib/registry/{mirror.name}"])

        with open(
            config.REGISTRY_DIR / "registry-config.yaml", "r"
        ) as registry_template:
            src = Template(registry_template.read())
            self.instance.exec(
                ["dd", f"of=/etc/distribution/{mirror.name}.yaml"],
                sensitive_kwargs=True,
                input=str.encode(src.substitute(substitutes)),
            )

        with open(config.REGISTRY_DIR / "registry.service", "r") as registry_template:
            src = Template(registry_template.read())
            self.instance.exec(
                ["dd", f"of=/etc/systemd/system/registry-{mirror.name}.service"],
                sensitive_kwargs=True,
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

    # Configure the specified instance to use this registry mirror.
    def apply_configuration(self, instance, containerd_basedir="/etc/containerd"):
        for mirror in self.mirrors:
            substitutes = {
                "IP": self.ip,
                "PORT": mirror.port,
            }

            mirror_dir = os.path.join(containerd_basedir, "hosts.d", mirror.name)
            instance.exec(["mkdir", "-p", mirror_dir])

            with open(config.REGISTRY_DIR / "hosts.toml", "r") as registry_template:
                src = Template(registry_template.read())
                instance.exec(
                    [
                        "dd",
                        f"of={mirror_dir}/hosts.toml",
                    ],
                    input=str.encode(src.substitute(substitutes)),
                )
