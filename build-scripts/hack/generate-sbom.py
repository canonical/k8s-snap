#!/usr/bin/env python3

USAGE = "Generate an SBOM for Canonical Kubernetes"

DESCRIPTION = """
The resulting SBOM includes a 'manifest.json' with all top-level dependencies,
as well as detailed reference for all transitive dependencies. We try to
automate much of the sbom generation as much as possible by parsing the source
directory.
"""

import argparse
import json
import logging
import sys
import tarfile
import tempfile
import yaml
from pathlib import Path
import util

logging.basicConfig(level=logging.INFO)

LOG = logging.getLogger(__name__)

DIR = Path(__file__).absolute().parent

SNAPCRAFT_YAML = yaml.safe_load(util.read_file(DIR / "../../snap/snapcraft.yaml"))

# FIXME: This information should not be hardcoded here
CILIUM_ROCK_REPO = "https://github.com/canonical/cilium-rocks"
CILIUM_ROCK_TAG = "main"
COREDNS_ROCK_REPO = "https://github.com/canonical/coredns-rock"
COREDNS_ROCK_TAG = "main"
METRICS_SERVER_ROCK_REPO = "https://github.com/canonical/metrics-server-rock"
METRICS_SERVER_ROCK_TAG = "main"
RAWFILE_LOCALPV_REPO = "https://github.com/canonical/rawfile-localpv"
RAWFILE_LOCALPV_TAG = "main"
SNAPCRAFT_C_COMPONENTS = ["libmnl", "libnftnl", "iptables"]
SNAPCRAFT_GO_COMPONENTS = [
    "runc",
    "containerd",
    "cni",
    "helm",
    "kubernetes",
    "k8s-dqlite",
]
K8S_DIR = DIR / "../../src/k8s"


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
        repo_url = util.read_file(DIR / "../components" / component / "repository")
        repo_tag = util.read_file(DIR / "../components" / component / "version")

        go_mod_name = f"{component}/go.mod"
        go_sum_name = f"{component}/go.sum"

        with util.git_repo(repo_url, repo_tag) as dir:
            repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
            extra_files[go_mod_name] = util.read_file(Path(dir) / "go.mod")
            extra_files[go_sum_name] = util.read_file(Path(dir) / "go.sum")

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
    extra_files["k8s-snap/go.mod"] = util.read_file(K8S_DIR / "go.mod")
    extra_files["k8s-snap/go.sum"] = util.read_file(K8S_DIR / "go.sum")
    manifest["snap"]["k8s-snap"]["k8s-snap"] = {
        "language": "go",
        "details": ["k8s-snap/go.mod", "k8s-snap/go.sum"],
        "source": {
            "type": "git",
            "repo": util.parse_output(["git", "remote", "get-url", "origin"]),
            "tag": util.parse_output(["git", "rev-parse", "--abbrev-ref", "HEAD"]),
            "revision": util.parse_output(["git", "rev-parse", "HEAD"]),
        },
    }


def k8s_snap_c_dqlite_components(manifest, extra_files):
    LOG.info("Generating SBOM info for dqlite components")

    repos = {}
    tags = {}
    # attempt to parse repos and tags from dqlite_version.sh
    for line in (K8S_DIR / "hack/env.sh").read_text().split():
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
        with util.git_repo(repo_url, repo_tag) as dir:
            repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=dir)

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

    with util.git_repo(CILIUM_ROCK_REPO, CILIUM_ROCK_TAG) as d:
        rock_repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=d)
        rockcraft = (d / "cilium/rockcraft.yaml").read_text()
        operator_rockcraft = (d / "cilium-operator-generic/rockcraft.yaml").read_text()

        extra_files["cilium/rockcraft.yaml"] = rockcraft
        extra_files["cilium-operator-generic/rockcraft.yaml"] = operator_rockcraft

        rockcraft_yaml = yaml.safe_load(rockcraft)
        repo_url = rockcraft_yaml["parts"]["cilium"]["source"]
        repo_tag = rockcraft_yaml["parts"]["cilium"]["source-tag"]

    with util.git_repo(repo_url, repo_tag) as dir:
        repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
        extra_files["cilium/go.mod"] = util.read_file(dir / "go.mod")
        extra_files["cilium/go.sum"] = util.read_file(dir / "go.sum")

        extra_files["cilium-operator-generic/go.mod"] = util.read_file(dir / "go.mod")
        extra_files["cilium-operator-generic/go.sum"] = util.read_file(dir / "go.sum")

    # NOTE: this silently assumes that cilium and cilium-operator-generic rocks are in sync
    manifest["rocks"]["cilium"] = {
        "rock-source": {
            "type": "git",
            "repo": CILIUM_ROCK_REPO,
            "tag": CILIUM_ROCK_TAG,
            "revision": rock_repo_commit,
        },
        "language": "go",
        "details": [
            "cilium/rockcraft.yaml",
            "cilium/go.mod",
            "cilium/go.sum",
            "cilium-operator-generic/rockcraft.yaml",
            "cilium-operator-generic/go.mod",
            "cilium-operator-generic/go.sum",
        ],
        "source": {
            "type": "git",
            "repo": repo_url,
            "tag": repo_tag,
            "revision": repo_commit,
        },
    }


def rock_coredns(manifest, extra_files):
    LOG.info("Generating SBOM info for CoreDNS rock")

    with util.git_repo(COREDNS_ROCK_REPO, COREDNS_ROCK_TAG) as d:
        rock_repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=d)
        # TODO(ben): This should not be hard coded.
        rockcraft = (d / "1.11.1/rockcraft.yaml").read_text()

        extra_files["coredns/1.11.1/rockcraft.yaml"] = rockcraft

        rockcraft_yaml = yaml.safe_load(rockcraft)
        repo_url = rockcraft_yaml["parts"]["coredns"]["source"]
        repo_tag = rockcraft_yaml["parts"]["coredns"]["source-tag"]

    with util.git_repo(repo_url, repo_tag) as dir:
        repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
        extra_files["coredns/go.mod"] = util.read_file(dir / "go.mod")
        extra_files["coredns/go.sum"] = util.read_file(dir / "go.sum")

    manifest["rocks"]["coredns"] = {
        "rock-source": {
            "type": "git",
            "repo": COREDNS_ROCK_REPO,
            "tag": COREDNS_ROCK_TAG,
            "revision": rock_repo_commit,
        },
        "language": "go",
        "details": [
            "coredns/1.11.1/rockcraft.yaml",
            "coredns/go.mod",
            "coredns/go.sum",
        ],
        "source": {
            "type": "git",
            "repo": repo_url,
            "tag": repo_tag,
            "revision": repo_commit,
        },
    }


def rock_metrics_server(manifest, extra_files):
    LOG.info("Generating SBOM info for metrics-server rock")

    with util.git_repo(METRICS_SERVER_ROCK_REPO, METRICS_SERVER_ROCK_TAG) as d:
        rock_repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=d)
        # TODO(ben): This should not be hard coded.
        rockcraft = (d / "0.7.0/rockcraft.yaml").read_text()

        extra_files["metrics-server/0.7.0/rockcraft.yaml"] = rockcraft

        rockcraft_yaml = yaml.safe_load(rockcraft)
        repo_url = rockcraft_yaml["parts"]["metrics-server"]["source"]
        repo_tag = rockcraft_yaml["parts"]["metrics-server"]["source-tag"]

    with util.git_repo(repo_url, repo_tag) as dir:
        repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=dir)
        extra_files["metrics-server/go.mod"] = util.read_file(dir / "go.mod")
        extra_files["metrics-server/go.sum"] = util.read_file(dir / "go.sum")

    manifest["rocks"]["metrics-server"] = {
        "rock-source": {
            "type": "git",
            "repo": METRICS_SERVER_ROCK_REPO,
            "tag": METRICS_SERVER_ROCK_TAG,
            "revision": rock_repo_commit,
        },
        "language": "go",
        "details": [
            "metrics-server/0.7.0/rockcraft.yaml",
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

    with util.git_repo(repo_url, repo_tag) as dir:
        rockcraft = (dir / "rockcraft.yaml").read_text()
        requirements = (dir / "requirements.txt").read_text()

        repo_commit = util.parse_output(["git", "rev-parse", "HEAD"], cwd=dir)

        extra_files["rawfile-localpv/rockcraft.yaml"] = rockcraft
        extra_files["rawfile-localpv/requirements.txt"] = requirements

    manifest["rocks"]["rawfile-localpv"] = {
        "rock-source": {
            "type": "git",
            "repo": repo_url,
            "tag": repo_tag,
            "revision": repo_commit,
        },
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
    rock_metrics_server(manifest, extra_files)

    # TODO(neoaggelos): enable these after we build metrics-server and CSI rocks
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
