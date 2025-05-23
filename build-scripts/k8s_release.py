#!/usr/bin/env python3

import argparse
import json
import logging
import os
import re
import subprocess
import datetime
import json
from typing import List, Optional, Dict, Any, Union
from dateutil.relativedelta import relativedelta

import requests
from packaging.version import Version, InvalidVersion

K8S_TAGS_URL = "https://api.github.com/repos/kubernetes/kubernetes/tags"
EXEC_TIMEOUT = 60

LOG = logging.getLogger(__name__)


def _url_get(url: str) -> str:
    """Make a GET request to the given URL and return the response text."""
    response = requests.get(url, timeout=5)
    response.raise_for_status()
    return response.text


def is_stable_release(release: str) -> bool:
    """Check if a Kubernetes release tag is stable (no pre-release suffix).

    Args:
        release: A Kubernetes release tag (e.g. 'v1.30.1', 'v1.30.0-alpha.1').

    Returns:
        True if the release is stable, False otherwise.
    """
    return "-" not in release


def get_k8s_tags() -> List[str]:
    """Retrieve semantically ordered Kubernetes release tags from GitHub.

    Returns:
        A list of release tag strings sorted from newest to oldest.

    Raises:
        ValueError: If no tags are retrieved.
    """
    response = _url_get(K8S_TAGS_URL)
    tags_json = json.loads(response)
    if not tags_json:
        raise ValueError("No k8s tags retrieved.")
    tag_names = [tag["name"] for tag in tags_json]
    tag_names.sort(key=lambda x: Version(x), reverse=True)
    return tag_names


def get_latest_stable() -> str:
    """Get the latest stable Kubernetes release tag.

    Returns:
        The latest stable release tag string (e.g., 'v1.30.1').

    Raises:
        ValueError: If no stable release is found.
    """
    for tag in get_k8s_tags():
        if is_stable_release(tag):
            return tag
    raise ValueError("Couldn't find a stable release.")


def get_latest_releases_by_minor() -> Dict[str, str]:
    """Map each minor Kubernetes version to its latest release tag.

    Returns:
        A dictionary mapping minor versions (e.g. '1.30') to the
        latest (pre-)release tag (e.g. 'v1.30.1').
    """
    latest_by_minor: Dict[str, str] = {}

    for tag in get_k8s_tags():
        # Strip leading 'v' if present since Version expects numeric first char
        version = Version(tag.lstrip("v"))
        key = f"{version.major}.{version.minor}"
        if key not in latest_by_minor:
            latest_by_minor[key] = tag

    return latest_by_minor


def get_outstanding_prereleases(as_git_branch: bool = False) -> List[str]:
    """Return outstanding K8s pre-releases.

    Args:
        as_git_branch: If True, return the git branch name for the pre-release.
    """
    latest_release = get_latest_releases_by_minor()
    prereleases = []
    for tag in latest_release.values():
        if not is_stable_release(tag):
            prereleases.append(tag)

    if as_git_branch:
        return [get_prerelease_git_branch(tag) for tag in prereleases]

    return prereleases


def get_obsolete_prereleases() -> List[str]:
    """Return obsolete K8s pre-releases.

    We only keep the latest pre-release(s) if there is no corresponding stable
    release. All previous pre-releases are discarded.
    """
    k8s_tags = get_k8s_tags()
    seen_stable_minors = set()
    obsolete = []

    for tag in k8s_tags:
        if is_stable_release(tag):
            version = Version(tag.lstrip("v"))
            seen_stable_minors.add((version.major, version.minor))
        else:
            version = Version(tag.lstrip("v").split("-")[0])
            if (version.major, version.minor) in seen_stable_minors:
                obsolete.append(tag)

    return obsolete


def _exec(*args, **kwargs) -> tuple[str, str]:
    """Run the specified command and return the stdout/stderr output as a tuple."""
    kwargs.setdefault("text", True)
    kwargs.setdefault("check", True)
    kwargs.setdefault("timeout", EXEC_TIMEOUT)
    LOG.debug("Executing: %s, args: %s, kwargs: %s", cmd, args, kwargs)
    proc = subprocess.run(*args, **kwargs)
    return proc.stdout, proc.stderr


def _branch_exists(
    branch_name: str, remote=True, project_basedir: Optional[str] = None
):
    cmd = ["git", "branch"]
    if remote:
        cmd += ["-r"]

    stdout, _ = _exec(cmd, cwd=project_basedir, capture_output=True)
    return branch_name in stdout


def get_prerelease_git_branch(prerelease: str):
    """Retrieve the name of the k8s-snap git branch for a given k8s pre-release."""
    prerelease_re = r"v\d+\.\d+\.\d-(?:alpha|beta|rc)\.\d+"
    if not re.match(prerelease_re, prerelease):
        raise ValueError("Unexpected k8s pre-release name: %s", prerelease)

    # Use a single branch for all pre-releases of a given risk level,
    # e.g. v1.33.0-alpha.0 -> autoupdate/v1.33.0-alpha
    branch = f"autoupdate/{prerelease}"
    return re.sub(r"(-[a-zA-Z]+)\.[0-9]+", r"\1", branch)


def _update_prerelease_k8s_component(project_basedir: str, k8s_version: str):
    if not project_basedir:
        raise ValueError("Project base directory unspecified.")
    k8s_component_path = os.path.join(
        project_basedir, "build-scripts", "components", "kubernetes", "version"
    )
    with open(k8s_component_path, "w") as f:
        f.write(k8s_version)


def prepare_prerelease_git_branches(project_basedir: str, remote: str = "origin"):
    prereleases = get_outstanding_prereleases()
    if not prereleases:
        LOG.info("No outstanding k8s pre-releases.")
        return

    for prerelease in prereleases:
        branch = get_prerelease_git_branch(str(prerelease))
        LOG.info("Preparing pre-release branch: %s", branch)

        # Reset branch to remote main
        _exec(
            ["git", "fetch", remote],
            cwd=project_basedir,
            capture_output=False,
        )
        _exec(
            ["git", "checkout", "-B", branch, f"{remote}/main"],
            cwd=project_basedir,
            capture_output=False,
        )

        # Update the k8s version and commit
        _update_prerelease_k8s_component(project_basedir, str(prerelease))
        _exec(
            ["git", "add", "./build-scripts/components/kubernetes/version"],
            cwd=project_basedir,
            capture_output=False,
        )

        # Only commit if there are actual changes
        result = _exec(
            ["git", "status", "--porcelain"],
            cwd=project_basedir,
            capture_output=True,
        )
        if result[0]:
            _exec(
                ["git", "commit", "-m", f"Update k8s version to {prerelease}"],
                cwd=project_basedir,
                capture_output=False,
            )
        else:
            LOG.info("Nothing to commit for %s", branch)

        # Force-push branch to remote
        _exec(
            ["git", "push", "-u", remote, branch, "--force"],
            cwd=project_basedir,
            capture_output=False,
        )


def clean_obsolete_git_branches(project_basedir: str, remote="origin"):
    """Remove obsolete pre-release git branches.

    All risk levels will be removed once the latest release is stable.
    """
    obsolete_prereleases = get_obsolete_prereleases()
    for prerelease in obsolete_prereleases:
        branch = get_prerelease_git_branch(prerelease)
        LOG.info("Checking for obsolete pre-release %s branch: %s", prerelease, branch)
        if _branch_exists(
            f"{remote}/{branch}", remote=True, project_basedir=project_basedir
        ):
            LOG.info("Cleaning up obsolete pre-release branch: %s", branch)
            _exec(["git", "push", remote, "--delete", branch], cwd=project_basedir)
        else:
            LOG.debug("Obsolete branch not found, skipping: %s", branch)

def fetch_kubernetes_releases() -> List[Dict]:
    """
    Fetches Kubernetes release information from endoflife.date API.

    Returns:
        List of release dictionaries containing version info and support status.
    """
    url = "https://endoflife.date/api/v1/products/kubernetes"

    try:
        response = requests.get(url, headers={"Accept": "application/json"}, timeout=10)
        response.raise_for_status()
    except requests.RequestException as e:
        LOG.error("Failed to fetch Kubernetes EOL data: %s", e)
        raise RuntimeError(f"Failed to fetch Kubernetes EOL data: {e}")

    try:
        data = response.json()
    except ValueError as e:
        LOG.error("Failed to decode JSON response: %s", e)
        raise RuntimeError(f"Failed to decode JSON response: {e}")

    # The API returns a dictionary with 'result' containing the data
    if not isinstance(data, dict):
        LOG.error("Expected a dictionary from the API, got %s", type(data))
        raise RuntimeError("Unexpected API response format")

    # Get the releases list from the result
    releases = data.get('result', {}).get('releases', [])
    if not releases:
        LOG.warning("No releases found in the API response")
        return []

    LOG.debug("Found %d releases in API response", len(releases))
    return releases


def supported_upstream_releases() -> List[str]:
    """
    Returns a list of currently supported Kubernetes minor versions as strings.
    Uses data from https://endoflife.date/api/v1/products/kubernetes.
    """
    releases = fetch_kubernetes_releases()
    now = datetime.datetime.now(datetime.timezone.utc).date()
    supported = []

    for release in releases:
        if not isinstance(release, dict):
            LOG.warning("Expected dictionary for release, got %s", type(release))
            continue

        # Get the version name (e.g., '1.33')
        version = release.get('name')

        # Check if the version is EOL
        is_eol = release.get('isEol', True)  # Default to True if missing
        eol_date_str = release.get('eolFrom')

        if not version or not eol_date_str:
            LOG.debug("Skipping release with missing version or EOL date: %s", release)
            continue

        try:
            eol_date = datetime.datetime.strptime(eol_date_str, "%Y-%m-%d").date()
        except ValueError as e:
            LOG.warning("Invalid EOL date format for %s: %s", version, e)
            continue

        # Consider a version supported if it's not EOL and its EOL date is in the future
        if not is_eol and eol_date > now:
            LOG.info("Kubernetes %s is supported until %s", version, eol_date_str)
            supported.append(version)

    # Sort versions in ascending order (oldest first)
    supported.sort()

    if supported:
        LOG.info("Found %d supported Kubernetes versions: %s", len(supported), ", ".join(supported))
    else:
        LOG.warning("No supported Kubernetes versions found!")

    return supported

def is_lts_release(version_str: str) -> bool:
    """
    Determines if a Kubernetes version is an LTS release.
    The first LTS is 1.32. Then we align with the Ubuntu LTS releases.
    Hence, the initial LTS with the "normal" release cadence is 1.36,
    then every 6th minor release (every 2years - 1.42, 1.48, etc.)

    Args:
        version_str: Kubernetes version string (e.g., "1.36")

    Returns:
        True if the version is an LTS release, False otherwise
    """
    version = Version(version_str)
    if version < Version("1.32"):
        return False

    if version == Version("1.32"):
        return True

    # Check if it's 1.36 or a later version that's every 6th release after 1.36
    first_lts = 36
    if version.minor >= first_lts and (version.minor - first_lts) % 6 == 0:
        return True

    return False

def supported_canonical_k8s_releases(as_channel: bool = False) -> List[str]:
    """
    Returns a list of Canonical Kubernetes versions that are still supported.
    Support is based on a 12-year cycle for LTS releases and a 2-year cycle
    for intermediate releases, calculated from their actual release date.
    Only versions 1.32 and newer are considered.

    Note that the support duration for intermediate releases is not entirely
    accurate, since intermediate releases are supported until one year after
    the next LTS release. This means that releases right before an LTS release
    only have 1 year of support, while releases right after an LTS release have
    2 years of support. This is a simplification that is good enough for our
    purposes.

    Args:
        as_channel: If True, returns versions in channel format (e.g., "1.33-classic/stable")
                    instead of just the version number.
    """
    min_version = Version("1.32")
    support_durations = {
        "LTS": 12,  # 12 years for LTS
        "INTERMEDIATE": 2,  # 2 years for Intermediate (not entirely accurate but good enough)
    }

    all_releases_data = fetch_kubernetes_releases()
    now = datetime.datetime.now(datetime.timezone.utc).date()
    candidate_versions = []

    if all_releases_data:
        for release_info in all_releases_data:
            version_str = release_info.get('name')
            release_date_str = release_info.get('releaseDate')

            if not version_str or not release_date_str:
                LOG.debug("Skipping release with missing version or release date: %s", release_info)
                continue

            try:
                current_version = Version(version_str)
                parsed_release_date = datetime.datetime.strptime(release_date_str, "%Y-%m-%d").date()
            except InvalidVersion:
                LOG.warning("Skipping invalid version format from API: %s", version_str)
                continue
            except ValueError:
                LOG.warning("Skipping release with invalid date format: %s for version %s", release_date_str, version_str)
                continue

            if current_version.major != 1 or current_version < min_version:
                continue # Skip versions before 1.32 or not in 1.x series

            is_lts = is_lts_release(version_str)
            duration_years = support_durations["LTS"] if is_lts else support_durations["INTERMEDIATE"]

            # Calculate Canonical EOL
            canonical_eol_date = parsed_release_date + relativedelta(years=duration_years)

            if canonical_eol_date > now:
                candidate_versions.append(version_str)
                LOG.debug(
                    "Version %s (LTS: %s) released on %s, supported by Canonical until %s (for %d years)",
                    version_str, is_lts, parsed_release_date, canonical_eol_date, duration_years
                )
            else:
                LOG.debug(
                    "Version %s (LTS: %s) released on %s, Canonical EOL was %s (after %d years), no longer supported.",
                    version_str, is_lts, parsed_release_date, canonical_eol_date, duration_years
                )

    final_versions = sorted(list(set(candidate_versions)), key=Version)

    if final_versions:
        LOG.info(
            "Supported Canonical Kubernetes versions (based on calculated EOL): %s",
            ", ".join(final_versions)
        )
    else:
        LOG.warning("No actively supported Canonical Kubernetes versions found based on calculated EOL!")

    if as_channel:
        # TODO(ben): This will break if we have flavors in the future again.
        # The current channel format is hardcoded to "-classic/stable".
        channel_versions = [f"{v}-classic/stable" for v in final_versions]
        return channel_versions

    return final_versions

def format_output(result: Any, output_format: str) -> str:
    """
    Format the result according to the specified output format.

    Args:
        result: The result to format
        output_format: The output format, either "plain" or "json"

    Returns:
        The formatted result as a string
    """
    if output_format == "json":
        return json.dumps(result)
    else:  # plain format
        if isinstance(result, list):
            return "\n".join(str(item) for item in result)
        return str(result)

if __name__ == "__main__":
    logging.basicConfig(format="%(asctime)s %(message)s", level=logging.DEBUG)

    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--output",
        choices=["plain", "json"],
        default="plain",
        help="Output format (plain text or JSON)"
    )
    parser.add_argument(
        "--as-channel",
        action="store_true",
        help="Format versions as channels (e.g., '1.33-classic/stable' instead of just '1.33')"
    )
    subparsers = parser.add_subparsers(dest="subparser", required=True)

    cmd = subparsers.add_parser("clean_obsolete_git_branches")
    cmd.add_argument(
        "--project-basedir",
        dest="project_basedir",
        help="The k8s-snap project base directory.",
        default=os.getcwd(),
    )
    cmd.add_argument("--remote", dest="remote", help="Git remote.", default="origin")

    cmd = subparsers.add_parser("prepare_prerelease_git_branches")
    cmd.add_argument(
        "--project-basedir",
        dest="project_basedir",
        help="The k8s-snap project base directory.",
        default=os.getcwd(),
    )
    cmd.add_argument("--remote", dest="remote", help="Git remote.", default="origin")

    subparsers.add_parser("get_outstanding_prereleases")
    subparsers.add_parser("get_obsolete_prereleases")
    subparsers.add_parser("remove_obsolete_prereleases")
    subparsers.add_parser("supported_upstream_releases")
    subparsers.add_parser("supported_canonical_k8s_releases")

    kwargs = vars(parser.parse_args())
    output_format = kwargs.pop("output")
    as_channel = kwargs.pop("as_channel", False)
    subparser_name = kwargs.pop("subparser")

    # Add as_channel parameter only to canonical releases function
    if subparser_name == "supported_canonical_k8s_releases":
        kwargs["as_channel"] = as_channel
    f = locals()[subparser_name]

    out = f(**kwargs)

    if out is not None:
        # Format the output according to the specified format
        formatted_output = format_output(out, output_format)
        print(formatted_output)
    else:
        print("")
