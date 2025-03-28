import contextlib
import logging
import subprocess
import tempfile
from pathlib import Path
from typing import Any, Generator
from urllib.request import urlopen

from tenacity import retry, retry_if_exception_type, stop_after_attempt, wait_fixed

LOG = logging.getLogger(__name__)


@contextlib.contextmanager
@retry(
    retry=retry_if_exception_type(subprocess.CalledProcessError),
    wait=wait_fixed(5),
    stop=stop_after_attempt(15),
)
def git_repo(
    repo_url: str,
    repo_tag: str,
    shallow: bool = True,
) -> Generator[Path, Any, Any]:
    """
    Clone a git repository on a temporary directory and return the directory.

    Example usage:

    ```
    with git_repo("https://github.com/canonical/k8s-snap", "master") as dir:
        print("Repo cloned at", dir)
    ```

    """

    with tempfile.TemporaryDirectory() as tmpdir:
        cmd = ["git", "clone", repo_url, tmpdir, "-b", repo_tag]
        if shallow:
            cmd.extend(["--depth", "1"])
        LOG.info("Cloning %s @ %s (shallow=%s)", repo_url, repo_tag, shallow)
        parse_output(cmd)
        yield Path(tmpdir)


def parse_output(*args, **kwargs):
    return (
        subprocess.run(*args, capture_output=True, check=True, **kwargs)
        .stdout.decode()
        .strip()
    )


def read_file(path: Path) -> str:
    return path.read_text().strip()


def read_url(url: str) -> str:
    return urlopen(url).read().decode().strip()


def helm_pull(chart, repo_url: str, version: str, destination: Path) -> None:
    parse_output(
        [
            "helm",
            "pull",
            chart,
            "--repo",
            repo_url,
            "--version",
            version,
            "--destination",
            destination,
        ]
    )

    LOG.info(
        "Pulled helm chart %s @ %s as %s to %s", chart, version, repo_url, destination
    )
