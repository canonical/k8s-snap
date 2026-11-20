#
# Copyright 2026 Canonical, Ltd.
#
import subprocess
from unittest.mock import MagicMock

import pytest
from tenacity import RetryError
from test_util import util


def test_wait_for_services_stopped_polls_until_inactive():
    instance = MagicMock()
    instance.exec.side_effect = [
        subprocess.CompletedProcess(
            args=["snap", "services", "k8s"],
            returncode=0,
            stdout=(
                "Service        Startup  Current   Notes\n"
                "k8s.k8sd       enabled  active    -\n"
                "k8s.kubelet    enabled  active    -\n"
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["snap", "services", "k8s"],
            returncode=0,
            stdout=(
                "Service        Startup  Current   Notes\n"
                "k8s.k8sd       enabled  active    -\n"
                "k8s.kubelet    enabled  inactive  -\n"
            ),
            stderr="",
        ),
    ]

    util.wait_for_services_stopped(
        instance, exclude_services=["k8sd"], retries=2, delay_s=0
    )

    assert instance.exec.call_count == 2


def test_wait_for_services_stopped_raises_when_still_active():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["snap", "services", "k8s"],
        returncode=0,
        stdout=(
            "Service        Startup  Current   Notes\n"
            "k8s.kubelet    enabled  active    -\n"
        ),
        stderr="",
    )

    with pytest.raises(RetryError):
        util.wait_for_services_stopped(instance, retries=2, delay_s=0)


def test_wait_for_ports_available_polls_until_unbound():
    instance = MagicMock()
    instance.exec.side_effect = [
        subprocess.CompletedProcess(
            args=["ss", "-tuln"],
            returncode=0,
            stdout=(
                "Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port\n"
                "tcp LISTEN 0 4096 0.0.0.0:6443 0.0.0.0:*\n"
            ),
            stderr="",
        ),
        subprocess.CompletedProcess(
            args=["ss", "-tuln"],
            returncode=0,
            stdout="Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port\n",
            stderr="",
        ),
    ]

    util.wait_for_ports_available(instance, ports=[6443], retries=2, delay_s=0)

    assert instance.exec.call_count == 2


def test_wait_for_ports_available_raises_when_still_bound():
    instance = MagicMock()
    instance.exec.return_value = subprocess.CompletedProcess(
        args=["ss", "-tuln"],
        returncode=0,
        stdout=(
            "Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port\n"
            "tcp LISTEN 0 4096 [::]:6443 [::]:*\n"
        ),
        stderr="",
    )

    with pytest.raises(RetryError):
        util.wait_for_ports_available(instance, ports=[6443], retries=2, delay_s=0)
