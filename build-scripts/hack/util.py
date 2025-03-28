import contextlib
import logging
import subprocess
import tempfile
from pathlib import Path
from typing import Any, Generator
from urllib.request import urlopen

from tenacity import (
    before_sleep_log,
    retry,
    retry_if_exception_type,
    stop_after_attempt,
    wait_exponential,
    wait_random,
)

LOG = logging.getLogger(__name__)


@contextlib.contextmanager
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
        _clone_with_retry(cmd)
        yield Path(tmpdir)


def _clone_with_retry(cmd: list[str]):
    @retry(
        retry=retry_if_exception_type(subprocess.CalledProcessError),
        wait=wait_exponential(multiplier=1, min=1, max=60) + wait_random(0, 3),
        stop=stop_after_attempt(15),
        before_sleep=before_sleep_log(LOG, logging.WARNING),
    )
    def _run():
        parse_output(cmd)

    _run()


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
