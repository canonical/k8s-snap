#
# Copyright 2023 Canonical, Ltd.
#
import os
from pathlib import Path

DIR = Path(__file__).absolute().parent

# SNAP is the absolute path to the snap against which we run the e2e tests.
SNAP = os.getenv("TEST_SNAP")

# SUBSTRATE is the substrate to use for running the e2e tests.
# One of 'local' (default), 'lxd' or 'multipass'.
SUBSTRATE = os.getenv("TEST_SUBSTRATE") or "local"

# SKIP_CLEANUP can be used to prevent machines to be automatically destroyed
# after the tests complete.
SKIP_CLEANUP = (os.getenv("TEST_SKIP_CLEANUP") or "") == "1"

# LXD_PROFILE_NAME is the profile name to use for LXD containers.
LXD_PROFILE_NAME = os.getenv("TEST_LXD_PROFILE_NAME") or "k8s-e2e"

# LXD_PROFILE is the profile to use for LXD containers.
LXD_PROFILE = (
    os.getenv("TEST_LXD_PROFILE") or (DIR / ".." / "lxd-profile.yaml").read_text()
)

# LXD_IMAGE is the image to use for LXD containers.
LXD_IMAGE = os.getenv("TEST_LXD_IMAGE") or "ubuntu:22.04"

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
