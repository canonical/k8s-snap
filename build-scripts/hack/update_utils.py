import logging
from pathlib import Path
import re
import urllib.request

logging.basicConfig(level=logging.INFO)

LOG = logging.getLogger(__name__)

DIR = Path(__file__).absolute().parent
SNAPCRAFT = DIR.parent.parent / "snap/snapcraft.yaml"
COMPONENTS = DIR.parent / "components"


def update_go_version(dry_run: bool) -> str:
    k8s_version = (COMPONENTS / "kubernetes/version").read_text().strip()
    url = f"https://raw.githubusercontent.com/kubernetes/kubernetes/refs/tags/{k8s_version}/.go-version"
    with urllib.request.urlopen(url) as response:
        go_version = response.read().decode("utf-8").strip()

    LOG.info("Upstream go version is %s", go_version)

    _update_go_version_in_snapcraft(k8s_version, go_version, dry_run)

    return go_version


def _update_go_version_in_snapcraft(k8s_version: str, go_version: str, dry_run: bool):
    [k8s_major, k8s_minor] = map(int, re.match(r"v?(\d+)\.(\d+)", k8s_version).groups())
    [go_major, go_minor] = map(int, go_version.split(".")[:2])
    go_snap = f"go/{go_major}.{go_minor}-fips/stable"
    # We don't support fips for versions under 1.34
    if k8s_major == 1 and k8s_minor < 34:
        go_snap = f"go/{go_major}.{go_minor}/stable"
    snapcraft_yaml = SNAPCRAFT.read_text()
    if f"- {go_snap}" in snapcraft_yaml:
        LOG.info("snapcraft.yaml already contains go version %s", go_snap)
        return

    LOG.info("Update go snap version to %s in %s", go_snap, SNAPCRAFT)
    if not dry_run:
        updated = re.sub(
            r"- go/\d+\.\d+(?:-fips)?/stable", f"- {go_snap}", snapcraft_yaml
        )
        SNAPCRAFT.write_text(updated)
