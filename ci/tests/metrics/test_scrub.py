#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for the secret scrubber.

The scrubber gates whether any log-derived text may be persisted, so these
tests are the review artifact the plan requires before the first excerpt is
written. They are split into two halves that pull in opposite directions:

* :class:`TestSecretsAreRemoved` -- nothing sensitive survives.
* :class:`TestUsefulTextSurvives` -- the excerpt is still worth reading.

A scrubber that passes only the first half is trivially satisfiable by
returning the empty string, which is why the second half exists.
"""

import re

from metrics.scrub import DROPPED, REDACTED, find_secrets, scrub, scrub_line


class TestSecretsAreRemoved:
    def test_github_token(self):
        line = (
            "fatal: could not read Password for 'https://ghp_"
            + "a" * 36
            + "@github.com'"
        )
        assert "ghp_" not in scrub_line(line)

    def test_github_fine_grained_pat(self):
        assert "github_pat_" not in scrub_line("token=github_pat_" + "b" * 40)

    def test_jwt(self):
        jwt = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dBjftJeZ4CVPmB92K27uhbUJU1p1r"
        assert "eyJ" not in scrub_line(f"Bearer {jwt} rejected")

    def test_aws_access_key(self):
        assert "AKIA" not in scrub_line("aws key AKIAIOSFODNN7EXAMPLE failed")

    def test_kubernetes_bootstrap_token(self):
        # Real join tokens are minted at runtime, so GitHub never masks them.
        out = scrub_line("failed to join with token abcdef.0123456789abcdef")
        assert "abcdef.0123456789abcdef" not in out

    def test_url_userinfo(self):
        out = scrub_line("cloning https://user:hunter2@git.launchpad.net/foo")
        assert "hunter2" not in out
        assert (
            "git.launchpad.net" in out
        ), "host should survive; only the credential goes"

    def test_authorization_header(self):
        assert "abc123def456" not in scrub_line("Authorization: abc123def456")

    def test_env_assignment_by_name(self):
        # The value looks innocuous; only the *name* marks it sensitive.
        out = scrub_line("UBUNTU_PRO_TOKEN=C1a2b3c4d5")
        assert "C1a2b3c4d5" not in out

    def test_json_assignment(self):
        out = scrub_line('{"client_secret": "s3cr3tvalue", "user": "bob"}')
        assert "s3cr3tvalue" not in out
        assert "bob" in out

    def test_cli_flag(self):
        assert "swordfish" not in scrub_line("k8s bootstrap --token swordfish")

    def test_private_key_block_drops_the_line(self):
        assert scrub_line("-----BEGIN RSA PRIVATE KEY-----") == DROPPED

    def test_kubeconfig_material_drops_the_line(self):
        assert scrub_line("    client-key-data: LS0tLS1CRUdJTg==") == DROPPED

    def test_long_base64_blob_is_redacted_in_place(self):
        """The blob is the payload boundary, so the line can survive it.

        Dropping the whole line cost us the command that failed and left the
        failure unclassifiable; the bytes removed are identical either way.
        """
        assert scrub_line("data: " + "QUJDRA" * 12) == "data: " + REDACTED

    def test_real_join_token_is_removed_but_the_command_survives(self):
        """Regression from real CI output (job 83117743428).

        ``k8s join-cluster`` takes a base64 token that embeds a cluster secret
        and fingerprint. It is minted at runtime, so GitHub never masks it --
        this scrubber is the only thing standing between it and a durable
        artifact. 281 excerpts in the 90-day backfill contained one.
        """
        token = (
            "eyJzZWNyZXQiOiJkNTYzZGI0NDJiY2ZlM2YyYTY1NzViZjljYjNhOGJiZjZmZWI1"
            "MmZkYmFkZDgxNTJmNmMyYTU1NGViZTkyYTQwIiwiZmluZ2VycHJpbnQiOiI3Zjhj"
        )
        line = "CalledProcessError: Command '['k8s', 'join-cluster', '{}']'".format(
            token
        )
        out = scrub_line(line)
        assert token not in out
        assert "d563db442bcfe3f2" not in out
        assert "join-cluster" in out, "context needed for classification was lost"
        assert find_secrets(out) == []

    def test_multiline_excerpt(self):
        text = "\n".join(
            [
                "AssertionError: Service kube-proxy should be active",
                "export TEST_ADMIN_PASSWORD=letmein",
                "-----BEGIN CERTIFICATE-----",
            ]
        )
        out = scrub(text)
        assert "letmein" not in out
        assert DROPPED in out
        assert "kube-proxy" in out


class TestUsefulTextSurvives:
    """A scrubber that eats the diagnosis is as useless as one that leaks."""

    def test_real_assertion_is_untouched(self):
        # Verbatim from nightly run 34172139128, job 101894429657.
        line = "AssertionError: Service kube-proxy should be active, but it is inactive"
        assert scrub_line(line) == line

    def test_traceback_frame_is_untouched(self):
        line = '  File "/home/ubuntu/actions-runner/_work/k8s-snap/tests/util.py", line 362, in wait_until_k8s_ready'
        assert scrub_line(line) == line

    def test_tenacity_retry_error_is_untouched(self):
        line = "tenacity.RetryError: RetryError[<Future at 0x736ed8a7faa0 state=finished raised AssertionError>]"
        assert scrub_line(line) == line

    def test_pytest_summary_is_untouched(self):
        line = "FAILED tests/test_version_upgrades.py::test_version_downgrades_with_rollback"
        assert scrub_line(line) == line

    def test_lxc_instance_name_is_untouched(self):
        # Volatile, but not secret. Stabilising it is the normaliser's job.
        line = "Execute command lxc rm k8s-integration-1-9e535e-registry --force"
        assert scrub_line(line) == line

    def test_snap_channel_is_untouched(self):
        line = "Installing k8s from channel 1.35-classic/edge"
        assert scrub_line(line) == line

    def test_words_containing_token_do_not_nuke_the_line(self):
        line = "tokenizer produced 12 tokens"
        assert "tokenizer" in scrub_line(line)

    def test_github_masked_values_are_left_alone(self):
        line = "Setting up with ***"
        assert scrub_line(line) == line


class TestAuditPath:
    """find_secrets is the gate check; it must be independent of scrub()."""

    def test_detects_a_secret_in_unscrubbed_text(self):
        findings = find_secrets("token=ghp_" + "c" * 36)
        assert findings
        assert any(name == "github-token" for _, name in findings)

    def test_scrubbed_text_is_clean(self):
        text = "\n".join(
            [
                "ghp_" + "d" * 36,
                "PASSWORD=hunter2",
                "https://u:p@example.com",
            ]
        )
        assert find_secrets(scrub(text)) == []

    def test_reports_line_numbers(self):
        text = "clean line\nghp_" + "e" * 36
        findings = find_secrets(text)
        assert findings[0][0] == 2

    def test_redaction_marker_is_not_itself_a_finding(self):
        assert find_secrets(f"Authorization: {REDACTED}") == []


class TestBase64UrlPayloads:
    """base64url payloads must not survive in fragments.

    Redacting in place rather than dropping the line is only safe if the
    match covers the whole payload. A character class missing `-` and `_`
    splits a urlsafe payload into runs and leaves every run under the
    threshold verbatim -- so the narrower class leaks *more*.
    """

    def test_urlsafe_payload_is_removed_whole(self):
        import base64

        token = base64.urlsafe_b64encode(bytes(range(200))[:200]).decode()
        assert "-" in token or "_" in token, "fixture must exercise urlsafe chars"

        cleaned = scrub_line("k8s join-cluster " + token)

        # No fragment of the payload survives. Checking against every
        # separator-delimited run catches the partial-redaction case that a
        # simple `token not in cleaned` would pass.
        for run in re.split(r"[-_]", token):
            if len(run) >= 8:
                assert run not in cleaned, f"leaked {len(run)} chars of payload"

    def test_the_surrounding_command_still_survives(self):
        """The whole point of in-place redaction: the failure stays classifiable."""
        import base64

        token = base64.urlsafe_b64encode(bytes(range(200))[:200]).decode()

        cleaned = scrub_line("k8s join-cluster " + token)

        assert "k8s join-cluster" in cleaned

    def test_audit_gate_agrees_with_the_scrubber(self):
        """find_secrets is the release gate; it must not pass what scrub leaves."""
        import base64

        token = base64.urlsafe_b64encode(bytes(range(200))[:200]).decode()

        assert find_secrets("k8s join-cluster " + token), "gate missed a raw payload"
        assert not find_secrets(scrub_line("k8s join-cluster " + token))

    def test_long_runner_paths_are_not_mistaken_for_payloads(self):
        """A path must not redact, or the same failure hashes by runner type.

        The self-hosted work directory clears 60 characters where the
        GitHub-hosted one does not, so over-redaction here would split one
        signature in two along an axis that has nothing to do with the bug.
        """
        selfhosted = (
            '  File "/home/ubuntu/actions-runner/_work/k8s-snap/k8s-snap'
            '/tests/integration/tests/test_version_upgrades.py", line 231'
        )
        hosted = (
            '  File "/home/runner/work/k8s-snap/k8s-snap'
            '/tests/integration/tests/test_version_upgrades.py", line 231'
        )

        assert scrub_line(selfhosted) == selfhosted
        assert scrub_line(hosted) == hosted
