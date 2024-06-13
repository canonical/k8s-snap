import contextlib
from typing import Any, Generator
import tempfile
import subprocess
import logging
import yaml
from urllib.request import urlopen
from pathlib import Path

LOG = logging.getLogger(__name__)


@contextlib.contextmanager
def git_repo(
    repo_url: str, repo_tag: str, shallow: bool = True
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

def helm_pull(chart_name: str, repo_url: str, version: str, destination: Path) -> None:
    parse_output(["helm", "repo", "add", chart_name, repo_url])
    parse_output(["helm", "pull", f"{chart_name}/{chart_name}", "--version", version, "--destination", destination])

    LOG.info("Pulled helm repository %s @ %s as %s to %s", repo_url, version, chart_name, destination)