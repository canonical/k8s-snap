#
# Copyright 2025 Canonical, Ltd.
#
import time
from typing import List

import pytest
from test_util import config, harness
from test_util import registry as reg
from test_util import tags, util

REGISTRY_PORT = 5000


def setup_proxy(proxy: harness.Instance):
    """Installs and configures Squid proxy on the given instance."""
    proxy.exec("apt update".split())
    proxy.exec("apt install squid --yes".split())
    proxy.exec("echo 'http_access allow all' >> /etc/squid/conf.d/allow.conf".split())
    time.sleep(1)
    proxy.exec("systemctl restart squid.service".split())


def configure_proxy_env(
    instance: harness.Instance, proxy_ip: str, extra_no_proxy: str = ""
):
    """Sets proxy environment variables on the instance."""
    no_proxy = (
        f"localhost,127.0.0.1,{extra_no_proxy}"
        if extra_no_proxy
        else "localhost,127.0.0.1"
    )
    proxy_settings = f"""
http_proxy="http://{proxy_ip}:3128"
https_proxy="http://{proxy_ip}:3128"
no_proxy="{no_proxy}"
HTTP_PROXY="http://{proxy_ip}:3128"
HTTPS_PROXY="http://{proxy_ip}:3128"
NO_PROXY="{no_proxy}"
"""
    instance.exec("tee /etc/environment".split(), input=proxy_settings.encode())


def restrict_network(instance: harness.Instance, allow_ports: List[int] = []):
    """Blocks all outgoing traffic except for allowed ports."""
    instance.exec("iptables -A OUTPUT -p tcp --dport 443 -j REJECT".split())
    instance.exec("iptables -A OUTPUT -p tcp --dport 80 -j REJECT".split())
    for port in allow_ports:
        instance.exec(f"iptables -A OUTPUT -p tcp --dport {port} -j ACCEPT".split())
    instance.exec(
        "iptables -A OUTPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT".split()
    )


def setup_containerd_proxy(instance: harness.Instance, proxy_ip: str):
    """Configures containerd to use the proxy."""
    config = f"""
[Service]
Environment="HTTPS_PROXY=http://{proxy_ip}:3128"
Environment="HTTP_PROXY=http://{proxy_ip}:3128"
Environment="NO_PROXY=10.1.0.0/16,10.152.183.0/24,192.168.0.0/16,127.0.0.1,172.16.0.0/12"
"""
    instance.exec("mkdir -p /etc/systemd/system/snap.k8s.containerd.service.d/".split())
    instance.exec(
        "tee /etc/systemd/system/snap.k8s.containerd.service.d/http-proxy.conf".split(),
        input=config.encode(),
    )


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="runner size too small on multipass"
)
def test_airgapped_with_proxy(instances: List[harness.Instance], datastore_type: str):
    proxy, instance = instances
    proxy_ip = util.get_default_ip(proxy)
    instance_ip = util.get_default_ip(instance)

    setup_proxy(proxy)
    configure_proxy_env(instance, proxy_ip, instance_ip)
    restrict_network(instance, allow_ports=[3128])

    # Verify connectivity without the proxy is blocked.
    assert (
        instance.exec(
            ["curl", "-I", "-4", "--noproxy", "*", "https://www.google.com"],
            check=False,
            capture_output=True,
        ).returncode
        == 7
    )

    # Export the proxy settings and verify connectivity through proxy.
    # This is required because the proxy settings are not available to the Python
    # subprocess shell that runs the connectivity test.
    instance.exec(
        [
            "export",
            "$(cat /etc/environment | xargs)",
            "&&",
            "curl",
            "-I",
            "-4",
            "https://www.google.com",
        ]
    )

    # Install and configure Kubernetes snap
    util.setup_k8s_snap(instance)
    setup_containerd_proxy(instance, proxy_ip)
    util.bootstrap(instance, datastore_type=datastore_type)
    util.wait_until_k8s_ready(instance, [instance])


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skipif(
    config.SUBSTRATE == "multipass", reason="runner size too small on multipass"
)
def test_airgapped_with_image_mirror(
    h: harness.Harness,
    instances: List[harness.Instance],
    function_scoped_registry: reg.Registry,
    datastore_type: str,
):
    proxy, instance = instances
    proxy_ip = util.get_default_ip(proxy)
    registry = function_scoped_registry
    registry_ip = util.get_default_ip(registry.instance)

    setup_proxy(proxy)
    configure_proxy_env(registry.instance, proxy_ip, registry_ip)
    restrict_network(registry.instance, allow_ports=[3128])

    # Verify connectivity without the proxy is blocked.
    assert (
        registry.exec(
            ["curl", "-I", "-4", "--noproxy", "*", "https://www.google.com"],
            check=False,
            capture_output=True,
        ).returncode
        == 7
    )
    # Verify connectivity through the proxy.
    registry.exec(
        [
            "export",
            "$(cat /etc/environment | xargs)",
            "&&",
            "curl",
            "-I",
            "-4",
            "https://www.google.com",
        ]
    )

    setup_containerd_proxy(registry.instance, proxy_ip)
    util.bootstrap(registry.instance, datastore_type=datastore_type)

    # Mirror images
    out = registry.exec(["k8s", "list-images"], capture_output=True, text=True)
    images = out.stdout.splitlines()
    for image in images:
        link = "/".join(image.split("/")[1:])
        tag = f"{registry_ip}:{REGISTRY_PORT}/{link}"
        # Pull the image from the upstream registry and push it to the local registry.
        # Pipe the pull and push output to /dev/null as ctr is very verbose.
        registry.exec(
            [
                "export",
                "$(cat /etc/environment | xargs)",
                "&&",
                "/snap/k8s/current/bin/ctr",
                "images",
                "pull",
                "--all-platforms",
                image,
                ">",
                "/dev/null",
            ]
        )
        registry.exec(
            [
                "export",
                "$(cat /etc/environment | xargs)",
                "&&",
                "/snap/k8s/current/bin/ctr",
                "images",
                "tag",
                image,
                tag,
                ">",
                "/dev/null",
            ]
        )

        # The 443 port is required to upload to the local registry. So, we need to temporarily allow it.
        registry.exec("iptables -D OUTPUT -p tcp --dport 443 -j REJECT".split())
        registry.exec("iptables -A OUTPUT -p tcp --dport 443 -j ACCEPT".split())

        registry.exec(
            [
                "bash",
                "-c",
                f"export $(grep -v '^#' /etc/environment | xargs) && "
                f"/snap/k8s/current/bin/ctr images push --plain-http {tag}",
            ]
        )

        registry.exec("iptables -D OUTPUT -p tcp --dport 443 -j ACCEPT".split())
        registry.exec("iptables -A OUTPUT -p tcp --dport 443 -j REJECT".split())

    # Simulate airgap by cutting off proxy
    registry.exec("iptables -D OUTPUT -p tcp --dport 3128 -j ACCEPT".split())
    registry.exec("iptables -A OUTPUT -p tcp --dport 3128 -j REJECT".split())
    assert (
        registry.exec(
            "curl -I -4 https://www.google.com".split(),
            check=False,
            capture_output=True,
        ).returncode
        == 7
    )

    restrict_network(instance, allow_ports=[REGISTRY_PORT])
    util.setup_k8s_snap(instance)
    registry.apply_configuration(instance)
    util.bootstrap(instance, datastore_type=datastore_type)
    util.wait_until_k8s_ready(instance, [instance])
