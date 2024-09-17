#!/bin/env python3

import shlex
import subprocess
import json

LABEL = "automerge"
APPROVE_MSG = "All status checks passed for PR #{}."


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
    checks_json = sh(f"gh pr checks {pr_number} --json bucket")
    checks = json.loads(checks_json)
    return all(check["bucket"] == "pass" for check in checks)


def approve_and_merge_pr(pr_number) -> None:
    """Approve and merge the PR."""
    print(APPROVE_MSG.format(pr_number) + "Proceeding with merge...")
    sh(f'gh pr review {pr_number} --approve -b "{APPROVE_MSG.format(pr_number)}"')
    sh(f"gh pr merge {pr_number} --auto --squash")


def process_pull_requests():
    """Process the PRs and merge if checks have passed."""
    prs = get_pull_requests()

    for pr in prs:
        pr_number: int = pr["number"]

        if check_pr_passed(pr_number):
            approve_and_merge_pr(pr_number)
        else:
            print(f"Status checks have not passed for PR #{pr_number}. Skipping merge.")


if __name__ == "__main__":
    process_pull_requests()
