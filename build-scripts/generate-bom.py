#!/usr/bin/env python3

import json
import os
from pathlib import Path
import subprocess
import sys
import yaml

DIR = Path(__file__).absolute().parent

SNAPCRAFT_PART_BUILD = Path(os.getenv("SNAPCRAFT_PART_BUILD", ""))
SNAPCRAFT_PART_INSTALL = Path(os.getenv("SNAPCRAFT_PART_INSTALL", ""))

BUILD_DIRECTORY = SNAPCRAFT_PART_BUILD.exists() and SNAPCRAFT_PART_BUILD or DIR / ".build"
INSTALL_DIRECTORY = SNAPCRAFT_PART_INSTALL.exists() and SNAPCRAFT_PART_INSTALL or DIR / ".install"

# List of tools used to build or bundled in the snap
TOOLS = {
    "go": ["go", "version"],
    "gcc": ["gcc", "--version"],
}

# Retrieve list of components we care about from the snapcraft.yaml file
with open(DIR / ".." / "snap" / "snapcraft.yaml") as fin:
    COMPONENTS = yaml.safe_load(fin)["parts"]["bom"]["after"]


def _parse_output(*args, **kwargs):
    return subprocess.check_output(*args, **kwargs).decode().strip()


def _read_file(path: Path) -> str:
    return path.read_text().strip()


if __name__ == "__main__":
    BOM = {
        "k8s": {
            "version": _parse_output(["git", "rev-parse", "--abbrev-ref", "HEAD"]),
            "revision": _parse_output(["git", "rev-parse", "HEAD"]),
        },
        "tools": {},
        "components": {},
    }

    for tool_name, version_cmd in TOOLS.items():
        BOM["tools"][tool_name] = _parse_output(version_cmd).split("\n")

    for component in COMPONENTS:
        component_dir = DIR / "components" / component

        try:
            version = _read_file(component_dir / "version")
            patches = _parse_output([sys.executable, DIR / "print-patches-for.py", component, version])
            clean_patches = []
            if patches:
                clean_patches = [p[p.find("build-scripts/") :] for p in patches.split("\n")]

            BOM["components"][component] = {
                "repository": _read_file(component_dir / "repository"),
                "version": version,
                "revision": _parse_output(
                    ["git", "rev-parse", f"HEAD~{len(clean_patches)}"],
                    cwd=BUILD_DIRECTORY / ".." / ".." / component / "build" / component,
                ),
                "patches": clean_patches,
            }
        except OSError as e:
            print(f"Could not get info for {component}: {e}", file=sys.stderr)

    print(json.dumps(BOM, indent=2))
