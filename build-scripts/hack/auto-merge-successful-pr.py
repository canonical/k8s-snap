#!/bin/env python3

import shlex
import subprocess
import json
import sys

LABEL = "automerge"


def sh(cmd: str) -> str:
    """Run a shell command and return its output."""
    _pipe = subprocess.PIPE
    result = subprocess.run(shlex.split(cmd), stdout=_pipe, stderr=_pipe, text=True)
    if result.returncode != 0:
        raise Exception(f"Error running command: {cmd}\nError: {result.stderr}")
    return result.stdout.strip()


def get_pull_requests() -> list:
    """Fetch open pull requests matching some label."""
    prs_json = sh("gh pr list --state open --json number,labels")
    prs = json.loads(prs_json)
    return [pr for pr in prs if any(label["name"] == LABEL for label in pr["labels"])]


def check_pr_passed(pr_number) -> bool:
    """Check if all status checks passed for the given PR."""
    # gh pr checks exits non-zero when checks are pending or failed, but still
    # emits valid JSON to stdout. Avoid raising so pending PRs are quietly
    # skipped rather than counted as processing errors.
    _pipe = subprocess.PIPE
    result = subprocess.run(
        shlex.split(f"gh pr checks {pr_number} --json bucket"),
        stdout=_pipe, stderr=_pipe, text=True
    )
    if not result.stdout.strip():
        return False
    checks = json.loads(result.stdout.strip())
    return all(check["bucket"] in ["pass", "skipping"] for check in checks)


def merge_pr(pr_number) -> None:
    """Merge the PR using admin bypass, no review step required."""
    print(f"All status checks passed for PR #{pr_number}. Proceeding with merge...")
    sh(f"gh pr merge {pr_number} --admin --squash")


def process_pull_requests():
    """Process the PRs and merge if checks have passed."""
    prs = get_pull_requests()
    failed = []

    for pr in prs:
        pr_number: int = pr["number"]
        try:
            if check_pr_passed(pr_number):
                merge_pr(pr_number)
            else:
                print(f"Status checks have not passed for PR #{pr_number}. Skipping merge.")
        except Exception as e:
            print(f"Failed to merge PR #{pr_number}: {e}")
            failed.append(pr_number)

    if failed:
        print(f"The following PRs failed to merge: {failed}")
        sys.exit(1)


if __name__ == "__main__":
    process_pull_requests()
