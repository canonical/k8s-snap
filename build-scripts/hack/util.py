import contextlib
import logging
import subprocess
import tempfile
import time
from pathlib import Path
from typing import Any, Generator
from urllib.request import urlopen

LOG = logging.getLogger(__name__)


@contextlib.contextmanager
def git_repo(
    repo_url: str,
    repo_tag: str,
    shallow: bool = True,
    retry_n: int = 5,
    retry_delay: int = 5,
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
        for attempt in range(0, retry_n):
            try:
                parse_output(cmd)
                break
            except subprocess.CalledProcessError as e:
                if attempt == retry_n - 1:
                    raise e
                LOG.warning(
                    f"Failed to clone {repo_url} @ {repo_tag}, retrying in {retry_delay}s"
                )
                time.sleep(retry_delay)
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
