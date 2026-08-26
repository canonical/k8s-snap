#
# Copyright 2026 Canonical, Ltd.
#
"""Thin GitHub boundary for the triage FSM.

All network side effects live here so the handlers can be driven fully offline
by substituting a fake client (see ``tests/doubles.py``). ``dry_run`` turns
every write (labels, comments, PRs, branch deletes) into a recorded no-op.

Labels are the bot's persistent state, so ``swap_label`` is the core
transition: remove the old triage label (ignoring a 404) then add the new one.
"""

from __future__ import annotations

import json
import os
import subprocess
import tempfile
from typing import Optional
from urllib.parse import quote

REPO = "canonical/k8s-snap"


def _write_askpass_helper() -> str:
    """A throwaway script that hands ``git push`` the token via env, not argv.

    Embedding the token in the remote URL puts it in this process's own argv
    for as long as the push runs -- readable by anything else on the host via
    ``ps`` or ``/proc/<pid>/cmdline``. ``GIT_ASKPASS`` has git invoke this
    script as a *separate* child process per credential prompt; the token
    reaches it only through ``TRIAGE_PUSH_TOKEN`` in that child's env, which
    ``ps``/``cmdline`` do not expose. Caller owns deleting the returned path.
    """
    fd, path = tempfile.mkstemp(prefix="triage-askpass-", suffix=".sh")
    with os.fdopen(fd, "w") as fh:
        fh.write(
            "#!/bin/sh\n"
            'case "$1" in\n'
            '*[Pp]assword*) printf %s "$TRIAGE_PUSH_TOKEN" ;;\n'
            '*) printf %s "x-access-token" ;;\n'
            "esac\n"
        )
    # 0o500: read+execute for the owner only. No write bit -- the content is
    # already final by this point, and no group/other access is ever needed
    # (git invokes this file directly, in-process, as GIT_ASKPASS).
    os.chmod(path, 0o500)
    return path


class GitHubClient:
    """Wrapper over ``gh`` for the handful of calls triage needs."""

    def __init__(self, repo: str = REPO, dry_run: bool = False):
        self.repo = repo
        self.dry_run = dry_run

    def _api(
        self,
        endpoint: str,
        method: str = "GET",
        fields: Optional[dict] = None,
        raw_fields: Optional[dict] = None,
        tolerate_missing: bool = False,
    ):
        cmd = ["gh", "api", "-X", method, endpoint]
        for key, value in (fields or {}).items():
            cmd += ["-f", f"{key}={value}"]
        for key, value in (raw_fields or {}).items():
            cmd += ["-F", f"{key}={value}"]
        try:
            r = subprocess.run(
                cmd, capture_output=True, text=True, check=True, timeout=30
            )
            return json.loads(r.stdout) if r.stdout.strip() else None
        except subprocess.CalledProcessError as e:
            if tolerate_missing and "HTTP 404" in (e.stderr or ""):
                return None
            raise GitHubError(f"gh api {method} {endpoint}: {e.stderr or e}") from e
        except subprocess.TimeoutExpired as e:
            raise GitHubError(f"gh api {method} {endpoint}: {e}") from e
        except json.JSONDecodeError:
            return None

    def get_issue(self, number: int) -> dict:
        data = self._api(f"repos/{self.repo}/issues/{number}")
        if not isinstance(data, dict):
            raise GitHubError(f"issue {number}: unexpected response")
        return data

    def get_comments(self, number: int) -> list[dict]:
        data = self._api(f"repos/{self.repo}/issues/{number}/comments")
        return data if isinstance(data, list) else []

    def search_issues(self, query: str) -> list[dict]:
        endpoint = f"search/issues?q={query}"
        data = self._api(endpoint)
        if isinstance(data, dict):
            return data.get("items", [])
        return []

    def add_labels(self, number: int, labels: list[str]) -> None:
        if self.dry_run or not labels:
            return
        for label in labels:
            self._api(
                f"repos/{self.repo}/issues/{number}/labels",
                method="POST",
                fields={"labels[]": label},
            )

    def add_comment(self, number: int, body: str) -> None:
        if self.dry_run or not body:
            return
        self._api(
            f"repos/{self.repo}/issues/{number}/comments",
            method="POST",
            fields={"body": body},
        )

    def remove_label(self, number: int, label: str) -> None:
        if self.dry_run:
            return
        self._api(
            f"repos/{self.repo}/issues/{number}/labels/{quote(label, safe='')}",
            method="DELETE",
            tolerate_missing=True,
        )

    def swap_label(self, number: int, old: Optional[str], new: str) -> None:
        """Atomic triage transition: drop the old label, add the new one."""
        if old and old != new:
            self.remove_label(number, old)
        self.add_labels(number, [new])

    def list_labels(self, number: int) -> list[str]:
        data = self._api(f"repos/{self.repo}/issues/{number}/labels")
        if isinstance(data, list):
            return [item.get("name", "") for item in data if item.get("name")]
        return []

    def push_branch(self, branch: str, cwd: str) -> bool:
        """Push a locally-committed fix branch to a token-authenticated remote.

        Runs in the trusted orchestrator process (which holds GH_TOKEN), never
        the agent shell -- that is exactly why the shell env is stripped of
        credentials. Returns False (without pushing) in dry-run or when the fix
        skill did not actually leave the branch committed locally, so a missing
        branch degrades to "no PR" rather than crashing the pipeline.

        The CI checkout uses ``persist-credentials: false``, so ``origin`` holds
        no usable credential; push to ``https://x-access-token@github.com/...``
        with the token supplied via :func:`_write_askpass_helper` (never in the
        URL -- see its docstring) so the token the orchestrator already has
        authenticates the push, falling back to ``origin`` locally when no
        token is set. ``branch`` is the bot's own ``triage/fix-<n>``, so a
        force update is safe and makes a re-fix idempotent (no
        remote-tracking ref to lease against here).
        """
        if self.dry_run:
            return False
        check = subprocess.run(
            [
                "git",
                "-C",
                cwd,
                "rev-parse",
                "--verify",
                "--quiet",
                f"refs/heads/{branch}",
            ],
            capture_output=True,
            text=True,
        )
        if check.returncode != 0:
            return False
        token = os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")
        env = None
        askpass_path = None
        if token:
            remote = f"https://x-access-token@github.com/{self.repo}.git"
            askpass_path = _write_askpass_helper()
            env = {
                **os.environ,
                "GIT_ASKPASS": askpass_path,
                "GIT_TERMINAL_PROMPT": "0",
                "TRIAGE_PUSH_TOKEN": token,
            }
        else:
            remote = "origin"
        try:
            subprocess.run(
                [
                    "git",
                    "-C",
                    cwd,
                    "push",
                    "--force",
                    remote,
                    f"refs/heads/{branch}:refs/heads/{branch}",
                ],
                capture_output=True,
                text=True,
                check=True,
                timeout=120,
                env=env,
            )
        except subprocess.CalledProcessError as e:
            err = str(e.stderr or e)
            if token:
                err = err.replace(token, "***")
            raise GitHubError(f"git push {branch}: {err}") from e
        except subprocess.TimeoutExpired as e:
            raise GitHubError(f"git push {branch}: {e}") from e
        finally:
            if askpass_path:
                os.unlink(askpass_path)
        return True

    def create_pull_request(
        self, *, head: str, base: str, title: str, body: str
    ) -> Optional[dict]:
        if self.dry_run:
            return None
        data = self._api(
            f"repos/{self.repo}/pulls",
            method="POST",
            fields={"head": head, "base": base, "title": title, "body": body},
            raw_fields={"draft": "true"},
        )
        if not isinstance(data, dict):
            raise GitHubError("create pull request: unexpected response")
        return data

    def find_pull_request(self, head: str, state: str = "open") -> Optional[dict]:
        owner = self.repo.split("/")[0]
        data = self._api(
            f"repos/{self.repo}/pulls?head={quote(f'{owner}:{head}', safe=':')}"
            f"&state={state}"
        )
        if isinstance(data, list) and data:
            return data[0]
        return None

    def find_branch(self, candidates: list[str]) -> Optional[str]:
        for branch in candidates:
            data = self._api(
                f"repos/{self.repo}/git/ref/heads/{quote(branch, safe='')}",
                tolerate_missing=True,
            )
            if isinstance(data, dict) and data.get("ref") == f"refs/heads/{branch}":
                return branch
        return None

    def delete_branch(self, branch: str) -> None:
        if self.dry_run:
            return
        self._api(
            f"repos/{self.repo}/git/refs/heads/{quote(branch, safe='')}",
            method="DELETE",
            tolerate_missing=True,
        )


class GitHubError(RuntimeError):
    """Raised when a ``gh`` call fails in a way triage cannot recover from."""
