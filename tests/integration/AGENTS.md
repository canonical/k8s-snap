# tests/integration

pytest-based integration test suite. Tests run against a real k8s snap on actual machines
provisioned by a substrate harness (LXD, Juju, or Multipass).

## Running Tests

```
cd tests/integration
tox -e integration -- tests/<test_file.py>          # single file
tox -e integration -- tests/ -k test_name           # by name
tox -e lint                                         # check style
tox -e format                                       # auto-fix style
```

## Configuration

All runtime config is via `TEST_*` environment variables. Key variables:

| Variable | Default | Purpose |
|----------|---------|---------|
| `TEST_SUBSTRATE` | `lxd` | Harness backend: `lxd`, `juju`, `multipass` |
| `TEST_SNAP` | _(none)_ | Absolute path to a local `.snap` file to test |
| `TEST_SNAP_NAME` | `k8s` | Snap name when installing from store |
| `TEST_FLAVOR` | `classic` | Snap flavor: `classic` or `strict` |
| `TEST_SKIP_CLEANUP` | `0` | Set to `1` to leave instances alive after tests |
| `TEST_LXD_IMAGE` | `ubuntu:22.04` | LXD image for containers |
| `TEST_VERSION_UPGRADE_CHANNELS` | _(none)_ | Space-separated channel list for upgrade tests |
| `TEST_INSPECTION_REPORTS_DIR` | _(none)_ | Directory for post-test inspection reports |

Constants (paths, URLs, versions) live in `tests/test_util/config.py`. Read it before
adding new configuration values.

## Test Tagging

Every test function **must** carry at least one `@pytest.mark.tags(...)`. Missing tags
cause a conftest assertion failure.

Import tags from `test_util.tags`:

```python
from test_util import tags

@pytest.mark.tags(tags.PULL_REQUEST)
def test_something(...): ...
```

| Constant | Tag string | When run |
|----------|-----------|----------|
| `tags.PULL_REQUEST` | `pull_request` | every PR |
| `tags.NIGHTLY` | `nightly` | nightly CI |
| `tags.WEEKLY` | `weekly` | weekly CI |
| `tags.CONFORMANCE` | `conformance_tests` | CNCF conformance |
| `tags.PERFORMANCE` | `performance` | benchmark runs |
| `tags.GPU` | `gpu` | GPU operator tests |
| `tags.PROMOTE_CANDIDATE` | `beta_to_candidate` | channel promotion gate |
| `tags.PROMOTE_STABLE` | `candidate_to_stable` | channel promotion gate |

Tag combinations: `up_to_nightly` (PR + nightly), `up_to_weekly` (PR + nightly + weekly).

## Test Structure

Tests declare their cluster shape with markers; the `instances` fixture provisions and
bootstraps the nodes:

```python
@pytest.mark.node_count(3)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-ha.yaml").read_text()
)
@pytest.mark.tags(tags.NIGHTLY)
def test_ha_cluster(instances: List[harness.Instance]):
    cp, worker1, worker2 = instances[0], instances[1], instances[2]
    cp.exec(["k8s", "status", "--wait-ready"])
```

Common markers:

| Marker | Purpose |
|--------|---------|
| `@pytest.mark.node_count(N)` | number of instances to provision |
| `@pytest.mark.bootstrap_config(yaml_str)` | bootstrap config passed to `k8s bootstrap` |
| `@pytest.mark.network_type(type)` | network configuration override |
| `@pytest.mark.disable_k8s_bootstrapping` | skip automatic bootstrap |
| `@pytest.mark.no_setup` | skip all setup (snap install + bootstrap) |

## Harness and Instance API

The `Instance` class (`test_util/harness/base.py`) abstracts over all substrates:

```python
instance.exec(["k8s", "status"])              # run command on instance; never use subprocess.run
instance.exec(cmd, capture_output=True)       # capture stdout/stderr
instance.arch                                 # cached_property: machine architecture string
instance.id                                   # instance identifier
instance.delete()                             # destroy the instance
instance.restart()                            # reboot the instance
instance.open_ports([6443, 16443])            # open firewall ports (Juju/cloud substrates)
```

For host architecture outside an instance context, use `platform.machine()`.

## Test Utilities

Before writing a new helper, check `test_util/snap.py` and `test_util/util.py` — most common
operations are already there.

### `test_util/snap.py`

Snap Store API and channel utilities. Reuse these instead of inline HTTP requests:

- `get_snap_info(snap_name=SNAP_NAME)` — fetch snap metadata from the store
- `get_channels(num, flavor, arch, risk_level)` — list matching channels
- `get_most_stable_channels(...)` — ascending list of channels by stability
- Constants: `SNAPSTORE_INFO_API`, `SNAPSTORE_HEADERS`, `RISK_LEVELS`, `SNAP_NAME`

### `test_util/util.py`

General-purpose helpers:

- `stubbornly(retries, delay_s).on(instance).exec(cmd)` — retry a command until it succeeds
- `stubbornly(...).until(condition_fn).exec(cmd)` — retry until lambda returns truthy
- `setup_k8s_snap(instance, channel?, snap?)` — install and connect interfaces
- `snap_refresh(instance, channel)` — refresh snap with retry logic
- `wait_until_k8s_ready(instance)` — wait for cluster ready state
- `wait_for_dns(instance)`, `wait_for_network(instance)`, `wait_for_load_balancer(instance)`
- `check_snap_services_ready(instance)` — assert all expected services are active
- `check_service_restarts(instance)` — assert no unexpected service restarts
- `check_service_logs_for_panics(instance)` — scan logs for Go panics
- `get_join_token(cp, joining_node)`, `join_cluster(instance, token)`
- `major_minor(version)` — parse `(major, minor)` tuple from version string
- `previous_track(snap_version)` — resolve the preceding snap track

## License Headers

Every Python file must start with:

```python
#
# Copyright 2026 Canonical, Ltd.
#
```

The year must match the current year. `tox -e format` applies headers automatically.

## Linting

Tools: black, isort (profile=black), flake8 (max-line-length=120), codespell, licenseheaders, bandit.

```
tox -e lint      # check (non-destructive)
tox -e format    # apply fixes
```

Run `tox -e lint` before committing. The same toolchain with identical settings is used
in `ci/` (separate `tox.ini`).
