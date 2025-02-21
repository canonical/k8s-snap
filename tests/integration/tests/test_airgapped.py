from pathlib import Path
import time
from typing import List

import pytest
from test_util import harness, tags, util, registry

def setup_proxy(proxy: harness.Instance):
    """Installs and configures Squid proxy on the given instance."""
    proxy.exec("sudo apt install squid --yes".split())
    proxy.exec("echo 'http_access allow all' >> /etc/squid/conf.d/allow.conf".split())
    time.sleep(1)
    proxy.exec("systemctl restart squid.service".split())

def configure_proxy_env(instance: harness.Instance, proxy_ip: str, extra_no_proxy: str = ""):
    """Sets proxy environment variables on the instance."""
    no_proxy = f"localhost,127.0.0.1,{extra_no_proxy}" if extra_no_proxy else "localhost,127.0.0.1"
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
    instance.exec("iptables -A OUTPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT".split())

def setup_containerd_proxy(instance: harness.Instance, proxy_ip: str):
    """Configures containerd to use the proxy."""
    config = f"""
[Service]
Environment="HTTPS_PROXY=http://{proxy_ip}:3128"
Environment="HTTP_PROXY=http://{proxy_ip}:3128"
Environment="NO_PROXY=10.1.0.0/16,10.152.183.0/24,192.168.0.0/16,127.0.0.1,172.16.0.0/12"
"""
    instance.exec("mkdir -p /etc/systemd/system/snap.k8s.containerd.service.d/".split())
    instance.exec("tee /etc/systemd/system/snap.k8s.containerd.service.d/http-proxy.conf".split(), input=config.encode())

@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.skip()
def test_airgapped_with_proxy(instances: List[harness.Instance]):
    proxy, instance = instances
    proxy_ip = util.get_default_ip(proxy)
    instance_ip = util.get_default_ip(instance)

    setup_proxy(proxy)
    configure_proxy_env(instance, proxy_ip, instance_ip)
    restrict_network(instance, allow_ports=[3128])

    # Verify connectivity
    assert instance.exec("curl -I -4 --noproxy '*' https://www.google.com".split(), check=False, capture_output=True).returncode == 7
    instance.exec("export $(grep -v '^#' /etc/environment | xargs) && curl -I -4 https://www.google.com".split())

    # Install and configure Kubernetes snap
    util.setup_k8s_snap(instance, Path("/"))
    setup_containerd_proxy(instance, proxy_ip)
    instance.exec("sudo k8s bootstrap".split())
    util.wait_until_k8s_ready(instance, [instance])

@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_airgapped_with_proxy_setup_and_image_mirror(instances: List[harness.Instance], registry: registry.Registry):
    proxy, instance = instances
    proxy_ip = util.get_default_ip(proxy)
    registry_ip = util.get_default_ip(registry.instance)

    setup_proxy(proxy)
    configure_proxy_env(registry.instance, proxy_ip, registry_ip)
    restrict_network(registry.instance, allow_ports=[3128])

    # Verify connectivity
    assert registry.instance.exec("curl -I -4 --noproxy '*' https://www.google.com".split(), check=False, capture_output=True).returncode == 7
    registry.instance.exec("export $(grep -v '^#' /etc/environment | xargs) && curl -I -4 https://www.google.com".split())

    setup_containerd_proxy(registry.instance, proxy_ip)
    registry.instance.exec("sudo k8s bootstrap".split())

    # Mirror images
    out = registry.instance.exec(["k8s", "list-images"], capture_output=True, text=True)
    images = out.stdout.splitlines()
    for image in images:
        tag = f"{registry_ip}:5000/{"/".join(image.split("/")[1:])}"
        registry.instance.exec(f"export $(grep -v '^#' /etc/environment | xargs) && /snap/k8s/current/bin/ctr images pull --all-platforms {image}".split())
        registry.instance.exec(f"export $(grep -v '^#' /etc/environment | xargs) && /snap/k8s/current/bin/ctr images tag {image} {tag}".split())

        # The 443 port is required to upload to the local registry. So, we need to temporarily allow it.
        registry.instance.exec("iptables -D OUTPUT -p tcp --dport 443 -j REJECT".split())
        registry.instance.exec("iptables -A OUTPUT -p tcp --dport 443 -j ACCEPT".split())

        registry.instance.exec(f"export $(grep -v '^#' /etc/environment | xargs) && /snap/k8s/current/bin/ctr images push --plain-http {tag}".split())

        registry.instance.exec("iptables -D OUTPUT -p tcp --dport 443 -j ACCEPT".split())
        registry.instance.exec("iptables -A OUTPUT -p tcp --dport 443 -j REJECT".split())

    # Simulate airgap by cutting off proxy
    registry.instance.exec("iptables -D OUTPUT -p tcp --dport 3128 -j ACCEPT".split())
    registry.instance.exec("iptables -A OUTPUT -p tcp --dport 3128 -j REJECT".split())
    assert registry.instance.exec("curl -I -4 https://www.google.com".split(), check=False, capture_output=True).returncode == 7

    restrict_network(instance, allow_ports=[5000])
    util.setup_k8s_snap(instance, Path("/"))
    registry.apply_configuration(instance)
    instance.exec("sudo k8s bootstrap".split())
    util.wait_until_k8s_ready(instance, [instance])
