# End To End Tests

## Overview

End to end tests are written in Python. They are built on top of a [Harness](./tests/conftest.py) fixture so that they can run on multiple environments like LXD, Multipass, Juju, or the local machine.

End to end tests can be configured using environment variables. You can see all available options in [./tests/config.py](./tests/config.py).

## Running end to end tests

Running the end to end tests requires `python3` and `tox`. Install with:

```bash
sudo apt install python3-virtualenv
virtualenv .venv
. .venv/bin/activate
pip install 'tox<5'
```

Further, make sure that you have built `k8s.snap`:

```bash
snapcraft --use-lxd
mv k8s_*.snap k8s.snap
```

In general, all end to end tests will require specifying the local path to the snap package under test, using the `TEST_SNAP` environment variable. Make sure to specify the full path to the file.

End to end tests are typically run with: `cd tests/integration && tox -e integration`

### Running end to end tests on the local machine

```bash
export TEST_SNAP=$PWD/k8s.snap
export TEST_SUBSTRATE=local

cd tests/integration && tox -e integration
```

> *NOTE*: When running locally, end to end tests that create more than one instance will fail.

### Running end to end tests on LXD containers

First, make sure that you have initialized LXD:

```bash
sudo lxd init --auto
```

Then, run the tests with:

```bash
export TEST_SNAP=$PWD/k8s.snap
export TEST_SUBSTRATE=lxd

export TEST_LXD_IMAGE=ubuntu:22.04          # (optionally) specify which image to use for LXD containers
export TEST_LXD_PROFILE=k8s-integration     # (optionally) specify profile name to configure
export TEST_SKIP_CLEANUP=1                  # (optionally) do not destroy machines after tests finish

cd tests/integration && tox -e integration
```

### Running end to end tests on multipass VMs

First, make sure that you have installed Multipass:

```bash
sudo snap install multipass
```

Then, run the tests with:

```bash
export TEST_SNAP=$PWD/k8s.snap
export TEST_SUBSTRATE=multipass

export TEST_MULTIPASS_IMAGE=22.04           # (optionally) specify ubuntu version for VMs
export TEST_MULTIPASS_CPUS=4                # (optionally) specify how many cpus each VM should have
export TEST_MULTIPASS_MEMORY=2G             # (optionally) specify how much RAM each VM should have
export TEST_MULTIPASS_DISK=10G              # (optionally) specify how much disk each VM should have

cd tests/integration && tox -e integration
```

Multipass can also be used to run the tests on Ubuntu Core:

```bash
export TEST_SNAP=$PWD/k8s.snap
export TEST_SUBSTRATE=multipass
export TEST_MULTIPASS_IMAGE=core20

cd tests/integration && tox -e integration
```

### Running end to end tests on Juju

First, make sure you have installed Juju and bootstrapped a Juju controller. You can provision a local controller on LXD and create a `k8s-integration` model using:

```bash
sudo snap install juju
mkdir -p ~/.local/share
juju bootstrap localhost
juju add-model k8s-integration
```

Then, run the tests with:

```bash
export TEST_SNAP=$PWD/k8s.snap
export TEST_SUBSTRATE=juju
export TEST_JUJU_MODEL=k8s-integration

export TEST_JUJU_CONTROLLER=localhost       # (optionally) specify Juju controller to use for running the tests
export TEST_JUJU_BASE=ubuntu@22.04          # (optionally) specify base OS to use for new Juju machines
export TEST_JUJU_CONSTRAINTS='mem=4G'       # (optionally) specify constraints for new Juju machines

cd tests/integration && tox -e integration
```

Alternatively, you can specify a list of existing Juju machines to use for the tests (e.g. machines created using `juju add-machine`):

```bash
export TEST_SNAP=$PWD/k8s.snap
export TEST_SUBSTRATE=juju
export TEST_JUJU_MODEL=k8s-integration

export TEST_JUJU_MACHINES=0,1,2

cd tests/integration && tox -e integration
```

## Writing an End to End test

For a simple way to write end to end tests, have a look at [`test_smoke.py`](./tests/test_smoke.py), which spins up a single instance, installs k8s and ensures that the kubelet node registers in the cluster.

Make sure to use the [Harness](./tests/conftest.py) fixture. That way, there _should not_ be a need for extra logic to handle running the tests in LXD, Multipass, Juju or locally.

A typical end to end test for feature `<feature>` should look like this:

```python
# tests/integration/tests/test_<feature>.py
#
# Copyright 2024 Canonical, Ltd.
#
import logging

from test_util import harness, util

LOG = logging.getLogger(__name__)
FEATURE_NODE_COUNT = 3  # number of machines necessary for the test


@pytest.mark.node_count(FEATURE_NODE_COUNT)
def test_feature(instances: List[harness.Instance]):
    # The cluster is bootstrapped, with only networking setup.
    first_node, *others_nodes = instances
    first_node.exec(["k8s", "something"])
```
