#!/usr/bin/env python3

USAGE = "Generate an SBOM for Canonical Kubernetes"

DESCRIPTION = """
The resulting SBOM includes a 'manifest.json' with all top-level dependencies,
as well as detailed reference for all transitive dependencies. We try to
automate much of the sbom generation as much as possible by parsing the source
directory.
"""

import argparse
import contextlib
import json
import logging
import subprocess
import sys
import tarfile
import tempfile
import yaml
from pathlib import Path
from typing import Any, Generator

logging.basicConfig(level=logging.INFO)

LOG = logging.getLogger(__name__)

DIR = Path(__file__).absolute().parent

SNAPCRAFT_YAML = yaml.safe_load(Path(DIR / "../../snap/snapcraft.yaml").read_text())

# FIXME: This information should not be hardcoded here
CILIUM_ROCK_REPO = "https://github.com/canonical/cilium-rocks"
CILIUM_ROCK_TAG = "main"
COREDNS_ROCK_REPO = "https://github.com/canonical/coredns-rock"
COREDNS_ROCK_TAG = "main"
METRICS_SERVER_ROCK_REPO = "https://github.com/canonical/metrics-server-rock"
METRICS_SERVER_ROCK_TAG = "main"
RAWFILE_LOCALPV_REPO = "https://github.com/canonical/rawfile-localpv"
RAWFILE_LOCALPV_TAG = "rockcraft"
SNAPCRAFT_C_COMPONENTS = ["libmnl", "libnftnl", "iptables"]
SNAPCRAFT_GO_COMPONENTS = ["runc", "containerd", "cni", "helm", "kubernetes"]
K8S_DIR = DIR / "../../src/k8s"


@contextlib.contextmanager
def _git_repo(repo_url: str, repo_tag: str) -> Generator[Path, Any, Any]:
    """
    Clone a git repository on a temporary directory and return the directory.

    Example usage:

    ```
    with _git_repo("https://github.com/canonical/k8s-snap", "master") as dir:
        print("Repo cloned at", dir)
    ```

    """
    with tempfile.TemporaryDirectory() as tmpdir:
        LOG.info("Cloning %s @ %s", repo_url, repo_tag)
        _parse_output(["git", "clone", repo_url, tmpdir, "-b", repo_tag, "--depth=1"])
        yield Path(tmpdir)


def _parse_output(*args, **kwargs):
    return (
        subprocess.run(*args, capture_output=True, check=True, **kwargs)
        .stdout.decode()
        .strip()
    )


def _read_file(path: Path) -> str:
    return path.read_text().strip()


def c_components_from_snapcraft(manifest, extra_files):
    for component in SNAPCRAFT_C_COMPONENTS:
        LOG.info("Generating SBOM info for C component %s", component)
        manifest["snap"]["external"][component] = {
            "language": "c",
            "source": {
                "type": "file",
                "url": SNAPCRAFT_YAML["parts"][component]["source"],
            },
        }


def go_components_external(manifest, extra_files):
    for component in SNAPCRAFT_GO_COMPONENTS:
        LOG.info("Generating SBOM info for Go component %s", component)
        repo_url = _read_file(DIR / "../components" / component / "repository")
        repo_tag = _parse_output([DIR / "../components" / component / "version.sh"])

        go_mod_name = f"{component}/go.mod"
        go_sum_name = f"{component}/go.sum"

        with _git_repo(repo_url, repo_tag) as dir:
            repo_commit = _parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
            extra_files[go_mod_name] = _read_file(Path(dir) / "go.mod")
            extra_files[go_sum_name] = _read_file(Path(dir) / "go.sum")

        manifest["snap"]["external"][component] = {
            "language": "go",
            "details": [go_sum_name, go_mod_name],
            "source": {
                "type": "git",
                "repo": repo_url,
                "tag": repo_tag,
                "revision": repo_commit,
            },
        }


def k8s_snap_go_components(manifest, extra_files):
    LOG.info("Generating SBOM info for k8s-snap")
    extra_files["k8s-snap/go.mod"] = _read_file(K8S_DIR / "go.mod")
    extra_files["k8s-snap/go.sum"] = _read_file(K8S_DIR / "go.sum")
    manifest["snap"]["k8s-snap"]["k8s-snap"] = {
        "language": "go",
        "details": ["k8s-snap/go.mod", "k8s-snap/go.sum"],
        "source": {
            "type": "git",
            "repo": _parse_output(["git", "remote", "get-url", "origin"]),
            "tag": _parse_output(["git", "rev-parse", "--abbrev-ref", "HEAD"]),
            "revision": _parse_output(["git", "rev-parse", "HEAD"]),
        },
    }


def k8s_snap_c_dqlite_components(manifest, extra_files):
    LOG.info("Generating SBOM info for k8s-snap dqlite components")

    repos = {}
    tags = {}
    # attempt to parse repos and tags from dqlite_version.sh
    for line in (K8S_DIR / "cmd/k8s-dqlite/dqlite_version.sh").read_text().split():
        # parse(REPO_DQLITE="https://github.com/ref") ==> repos["dqlite"] = "https://github.com/ref"
        if line.startswith("REPO_"):
            key, value = line.split("=")
            repos[key[len("REPO_") :].lower()] = value.strip('"')

        # parse(TAG_DQLITE="v1.1.3") ==> tags["dqlite"] = "v1.1.3"
        if line.startswith("TAG_"):
            key, value = line.split("=")
            tags[key[len("TAG_") :].lower()] = value.strip('"')

    for component in repos:
        repo_url = repos[component]
        repo_tag = tags[component]
        with _git_repo(repo_url, repo_tag) as dir:
            repo_commit = _parse_output(["git", "rev-parse", "HEAD"], cwd=dir)

        manifest["snap"]["k8s-snap"][component] = {
            "language": "c",
            "source": {
                "type": "git",
                "repo": repo_url,
                "tag": repo_tag,
                "revision": repo_commit,
            },
        }


def rock_cilium(manifest, extra_files):
    LOG.info("Generating SBOM info for Cilium rocks")

    with _git_repo(CILIUM_ROCK_REPO, CILIUM_ROCK_TAG) as d:
        rockcraft = (d / "cilium/rockcraft.yaml").read_text()
        operator_rockcraft = (d / "cilium-operator-generic/rockcraft.yaml").read_text()

        extra_files["cilium/rockcraft.yaml"] = rockcraft
        extra_files["cilium-operator-generic/rockcraft.yaml"] = operator_rockcraft

        rockcraft_yaml = yaml.safe_load(rockcraft)
        repo_url = rockcraft_yaml["parts"]["cilium"]["source"]
        repo_tag = rockcraft_yaml["parts"]["cilium"]["source-tag"]

    with _git_repo(repo_url, repo_tag) as dir:
        repo_commit = _parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
        extra_files["cilium/go.mod"] = _read_file(dir / "go.mod")
        extra_files["cilium/go.sum"] = _read_file(dir / "go.sum")

        extra_files["cilium-operator-generic/go.mod"] = _read_file(dir / "go.mod")
        extra_files["cilium-operator-generic/go.sum"] = _read_file(dir / "go.sum")

    # NOTE: this silently assumes that cilium and cilium-operator-generic rocks are in sync
    manifest["rocks"]["cilium"] = {
        "language": "go",
        "details": ["cilium/rockcraft.yaml", "cilium-operator-generic/rockcraft.yaml"],
        "source": {
            "type": "git",
            "repo": repo_url,
            "tag": repo_tag,
            "revision": repo_commit,
        },
    }


def rock_coredns(manifest, extra_files):
    LOG.info("Generating SBOM info for CoreDNS rock")

    with _git_repo(COREDNS_ROCK_REPO, COREDNS_ROCK_TAG) as d:
        rockcraft = (d / "rockcraft.yaml").read_text()

        extra_files["coredns/rockcraft.yaml"] = rockcraft

        rockcraft_yaml = yaml.safe_load(rockcraft)
        repo_url = rockcraft_yaml["parts"]["coredns"]["source"]
        repo_tag = rockcraft_yaml["parts"]["coredns"]["source-tag"]

    with _git_repo(repo_url, repo_tag) as dir:
        repo_commit = _parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
        extra_files["coredns/go.mod"] = _read_file(dir / "go.mod")
        extra_files["coredns/go.sum"] = _read_file(dir / "go.sum")

    manifest["rocks"]["coredns"] = {
        "language": "go",
        "details": ["coredns/rockcraft.yaml", "coredns/go.mod", "coredns/go.sum"],
        "source": {
            "type": "git",
            "repo": repo_url,
            "tag": repo_tag,
            "revision": repo_commit,
        },
    }


def rock_metrics_server(manifest, extra_files):
    LOG.info("Generating SBOM info for metrics-server rock")

    with _git_repo(METRICS_SERVER_ROCK_REPO, METRICS_SERVER_ROCK_TAG) as d:
        rockcraft = (d / "rockcraft.yaml").read_text()

        extra_files["metrics-server/rockcraft.yaml"] = rockcraft

        rockcraft_yaml = yaml.safe_load(rockcraft)
        repo_url = rockcraft_yaml["parts"]["metrics-server"]["source"]
        repo_tag = rockcraft_yaml["parts"]["metrics-server"]["source-tag"]

    with _git_repo(repo_url, repo_tag) as dir:
        repo_commit = _parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
        extra_files["metrics-server/go.mod"] = _read_file(dir / "go.mod")
        extra_files["metrics-server/go.sum"] = _read_file(dir / "go.sum")

    manifest["rocks"]["metrics-server"] = {
        "language": "go",
        "details": [
            "metrics-server/rockcraft.yaml",
            "metrics-server/go.mod",
            "metrics-server/go.sum",
        ],
        "source": {
            "type": "git",
            "repo": repo_url,
            "tag": repo_tag,
            "revision": repo_commit,
        },
    }


def rock_rawfile_localpv(manifest, extra_files):
    LOG.info("Generating SBOM info for rawfile-localpv rock")

    repo_url = RAWFILE_LOCALPV_REPO
    repo_tag = RAWFILE_LOCALPV_TAG

    with _git_repo(repo_url, repo_tag) as dir:
        rockcraft = (dir / "rockcraft.yaml").read_text()
        requirements = (dir / "requirements.txt").read_text()

        repo_commit = _parse_output(["git", "rev-parse", "HEAD"], cwd=dir)

        extra_files["rawfile-localpv/rockcraft.yaml"] = rockcraft
        extra_files["rawfile-localpv/requirements.txt"] = requirements

    manifest["rocks"]["rawfile-localpv"] = {
        "language": "python",
        "details": [
            "rawfile-localpv/rockcraft.yaml",
            "rawfile-localpv/requirements.txt",
        ],
        "source": {
            "type": "git",
            "repo": repo_url,
            "tag": repo_tag,
            "revision": repo_commit,
        },
    }


def generate_sbom(output):
    manifest = {
        "snap": {
            "k8s-snap": {},
            "external": {},
        },
        "rocks": {},
    }
    extra_files = {}

    c_components_from_snapcraft(manifest, extra_files)
    go_components_external(manifest, extra_files)
    k8s_snap_go_components(manifest, extra_files)
    k8s_snap_c_dqlite_components(manifest, extra_files)
    rock_cilium(manifest, extra_files)
    rock_coredns(manifest, extra_files)
    rock_rawfile_localpv(manifest, extra_files)

    # TODO(neoaggelos): enable these after we build metrics-server and CSI rocks
    # rock_metrics_server(manifest, extra_files)
    # rock_csi(manifest, extra_files)

    files = {"manifest.json": json.dumps(manifest, indent=4), **extra_files}
    LOG.info("Creating archive %s", output)
    tar = tarfile.open(output, "w:gz")
    with tempfile.TemporaryDirectory() as tmpdir:
        for name, contents in files.items():
            LOG.info("Adding %s to the archive", name)
            file = Path(tmpdir) / "sbom" / name
            file.parent.mkdir(parents=True, exist_ok=True)
            file.write_text(contents)
            tar.add(file, f"sbom/{name}")

    tar.close()

    LOG.info("Generated SBOM can be found at %s", output)


def main():
    parser = argparse.ArgumentParser(
        "generate-sbom.py", usage=USAGE, description=DESCRIPTION
    )
    parser.add_argument(
        "output", help="Path to save the sbom", default="k8s-sbom.tar.gz"
    )
    args = parser.parse_args(sys.argv[1:])

    return generate_sbom(args.output)


if __name__ == "__main__":
    main()
