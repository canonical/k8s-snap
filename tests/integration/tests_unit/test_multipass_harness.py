#
# Copyright 2026 Canonical, Ltd.
#
import logging
from unittest.mock import MagicMock, patch

import pytest
from test_util.harness import HarnessError
from test_util.harness.multipass import MultipassHarness


def test_multipass_harness_cloud_init_secrets_redacted(monkeypatch, caplog):
    caplog.set_level(logging.DEBUG)
    monkeypatch.setenv("TEST_SECRET_TOKEN", "super-secret-token-12345")
    monkeypatch.setenv("SECOND_SECRET", "second-secret-value-67890")

    template_content = (
        "token: ${TEST_SECRET_TOKEN}\nsecond: $SECOND_SECRET\nplain: unchanged\n"
    )

    harness = MultipassHarness()
    harness.cloud_init = "custom_cloud_init.yaml"

    with (
        patch(
            "test_util.harness.multipass.Path.read_text",
            return_value=template_content,
        ),
        patch("test_util.harness.multipass.run") as mock_run,
        patch("test_util.harness.multipass.stubbornly"),
    ):
        instance = harness.new_instance()

    assert instance is not None
    mock_run.assert_called_once()
    _, kwargs = mock_run.call_args
    assert kwargs.get("sensitive_kwargs") is True
    assert kwargs.get("input") == (
        b"token: super-secret-token-12345\nsecond: second-secret-value-67890\nplain: unchanged\n"
    )

    assert "super-secret-token-12345" not in caplog.text
    assert "second-secret-value-67890" not in caplog.text
    assert "environ({" not in caplog.text
    assert harness.cloud_init in caplog.text


@pytest.mark.parametrize(
    "network_type,expects_ipv6",
    [
        ("IPv4", False),
        ("ipv4", False),
        ("IPv6", True),
        ("ipv6", True),
        ("dualstack", True),
        ("Dualstack", True),
    ],
)
def test_multipass_harness_accepts_case_insensitive_network_types(
    network_type, expects_ipv6
):
    harness = MultipassHarness()
    harness.cloud_init = None
    harness.exec = MagicMock()

    with (
        patch("test_util.harness.multipass.run"),
        patch("test_util.harness.multipass.stubbornly"),
    ):
        instance = harness.new_instance(network_type=network_type)

    assert instance is not None
    if expects_ipv6:
        assert harness.exec.called
    else:
        assert not harness.exec.called


def test_multipass_harness_rejects_unsupported_network_type():
    harness = MultipassHarness()
    with pytest.raises(
        HarnessError,
        match=r"^Unknown network type: unsupported; supported types: IPv4, IPv6, dualstack$",
    ):
        harness.new_instance(network_type="unsupported")
