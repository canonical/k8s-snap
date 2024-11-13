#
# Copyright 2024 Canonical, Ltd.
#
import os
from pathlib import Path

DIR = Path(__file__).absolute().parent

# The following defaults are used to define how long to wait for a condition to be met.
DEFAULT_WAIT_RETRIES = int(os.getenv("TEST_DEFAULT_WAIT_RETRIES") or 60)
DEFAULT_WAIT_DELAY_S = int(os.getenv("TEST_DEFAULT_WAIT_DELAY_S") or 5)

MANIFESTS_DIR = DIR / ".." / ".." / "templates"

# ETCD_DIR contains all templates required to setup an etcd database.
ETCD_DIR = MANIFESTS_DIR / "etcd"

# ETCD_URL is the url from which the etcd binaries should be downloaded.
ETCD_URL = os.getenv("ETCD_URL") or "https://github.com/etcd-io/etcd/releases/download"

# ETCD_VERSION is the version of etcd to use.
ETCD_VERSION = os.getenv("ETCD_VERSION") or "v3.4.34"

# FLAVOR is the flavor of the snap to use.
FLAVOR = os.getenv("TEST_FLAVOR") or ""

# SNAP is the absolute path to the snap against which we run the integration tests.
SNAP = os.getenv("TEST_SNAP")

# SNAP_NAME is the name of the snap under test.
SNAP_NAME = os.getenv("TEST_SNAP_NAME") or "k8s"

# SUBSTRATE is the substrate to use for running the integration tests.
# One of 'local' (default), 'lxd', 'juju', or 'multipass'.
SUBSTRATE = os.getenv("TEST_SUBSTRATE") or "local"

# SKIP_CLEANUP can be used to prevent machines to be automatically destroyed
# after the tests complete.
SKIP_CLEANUP = (os.getenv("TEST_SKIP_CLEANUP") or "") == "1"

# INSPECTION_REPORTS_DIR is the directory where inspection reports are stored.
# If empty, no reports are generated.
INSPECTION_REPORTS_DIR = os.getenv("TEST_INSPECTION_REPORTS_DIR")

# LXD_PROFILE_NAME is the profile name to use for LXD containers.
LXD_PROFILE_NAME = os.getenv("TEST_LXD_PROFILE_NAME") or "k8s-integration"

# LXD_PROFILE is the profile to use for LXD containers.
LXD_PROFILE = (
    os.getenv("TEST_LXD_PROFILE")
    or (DIR / ".." / ".." / "lxd-profile.yaml").read_text()
)

# LXD_DUALSTACK_NETWORK is the network to use for LXD containers with dualstack configured.
LXD_DUALSTACK_NETWORK = os.getenv("TEST_LXD_DUALSTACK_NETWORK") or "dualstack-br0"

# LXD_DUALSTACK_PROFILE_NAME is the profile name to use for LXD containers with dualstack configured.
LXD_DUALSTACK_PROFILE_NAME = (
    os.getenv("TEST_LXD_DUALSTACK_PROFILE_NAME") or "k8s-integration-dualstack"
)

# LXD_DUALSTACK_PROFILE is the profile to use for LXD containers with dualstack configured.
LXD_DUALSTACK_PROFILE = (
    os.getenv("TEST_LXD_DUALSTACK_PROFILE")
    or (DIR / ".." / ".." / "lxd-dualstack-profile.yaml").read_text()
)

# LXD_IPV6_NETWORK is the network to use for LXD containers with ipv6-only configured.
LXD_IPV6_NETWORK = os.getenv("TEST_LXD_IPV6_NETWORK") or "ipv6-br0"

# LXD_IPV6_PROFILE_NAME is the profile name to use for LXD containers with ipv6-only configured.
LXD_IPV6_PROFILE_NAME = (
    os.getenv("TEST_LXD_IPV6_PROFILE_NAME") or "k8s-integration-ipv6"
)

# LXD_IPV6_PROFILE is the profile to use for LXD containers with ipv6-only configured.
LXD_IPV6_PROFILE = (
    os.getenv("TEST_LXD_IPV6_PROFILE")
    or (DIR / ".." / ".." / "lxd-ipv6-profile.yaml").read_text()
)

# LXD_IMAGE is the image to use for LXD containers.
LXD_IMAGE = os.getenv("TEST_LXD_IMAGE") or "ubuntu:22.04"

# LXD_SIDELOAD_IMAGES_DIR is an optional directory with OCI images from the host
# that will be mounted at /var/snap/k8s/common/images on the LXD containers.
LXD_SIDELOAD_IMAGES_DIR = os.getenv("TEST_LXD_SIDELOAD_IMAGES_DIR") or ""

# MULTIPASS_IMAGE is the image to use for Multipass VMs.
MULTIPASS_IMAGE = os.getenv("TEST_MULTIPASS_IMAGE") or "22.04"

# MULTIPASS_CPUS is the number of cpus for Multipass VMs.
MULTIPASS_CPUS = os.getenv("TEST_MULTIPASS_CPUS") or "2"

# MULTIPASS_MEMORY is the memory for Multipass VMs.
MULTIPASS_MEMORY = os.getenv("TEST_MULTIPASS_MEMORY") or "2G"

# MULTIPASS_DISK is the disk size for Multipass VMs.
MULTIPASS_DISK = os.getenv("TEST_MULTIPASS_DISK") or "10G"

# JUJU_MODEL is the Juju model to use.
JUJU_MODEL = os.getenv("TEST_JUJU_MODEL")

# JUJU_CONTROLLER is the Juju controller to use.
JUJU_CONTROLLER = os.getenv("TEST_JUJU_CONTROLLER")

# JUJU_CONSTRAINTS is the constraints to use when creating Juju machines.
JUJU_CONSTRAINTS = os.getenv("TEST_JUJU_CONSTRAINTS", "mem=4G cores=2 root-disk=20G")

# JUJU_BASE is the base OS to use when creating Juju machines.
JUJU_BASE = os.getenv("TEST_JUJU_BASE") or "ubuntu@22.04"

# JUJU_MACHINES is a list of existing Juju machines to use.
JUJU_MACHINES = os.getenv("TEST_JUJU_MACHINES") or ""

# A list of space-separated channels for which the upgrade tests should be run in sequential order.
# First entry is the bootstrap channel. Afterwards, upgrades are done in order.
# Alternatively, use 'recent <num> <flavour>' to get the latest <num> channels for <flavour>.
VERSION_UPGRADE_CHANNELS = (
    os.environ.get("TEST_VERSION_UPGRADE_CHANNELS", "").strip().split()
)

# The minimum Kubernetes release to upgrade from (e.g. "1.31")
# Only relevant when using 'recent' in VERSION_UPGRADE_CHANNELS.
VERSION_UPGRADE_MIN_RELEASE = os.environ.get("TEST_VERSION_UPGRADE_MIN_RELEASE")

# A list of space-separated channels for which the strict interface tests should be run in sequential order.
# Alternatively, use 'recent <num> strict' to get the latest <num> channels for strict.
STRICT_INTERFACE_CHANNELS = (
    os.environ.get("TEST_STRICT_INTERFACE_CHANNELS", "").strip().split()
)
