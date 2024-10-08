#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path
from typing import Generator, List, Union

import pytest
from test_util import config, harness, util
from test_util.etcd import EtcdCluster

LOG = logging.getLogger(__name__)


def _harness_clean(h: harness.Harness):
    "Clean up created instances within the test harness."

    if config.SKIP_CLEANUP:
        LOG.warning(
            "Skipping harness cleanup. "
            "It is your job now to clean up cloud resources"
        )
    else:
        LOG.debug("Cleanup")
        h.cleanup()


def _generate_inspection_report(h: harness.Harness, instance_id: str):
    LOG.debug("Generating inspection report for %s", instance_id)

    inspection_path = Path(config.INSPECTION_REPORTS_DIR)
    result = h.exec(
        instance_id,
        ["/snap/k8s/current/k8s/scripts/inspect.sh", "/inspection-report.tar.gz"],
        capture_output=True,
        text=True,
        check=False,
    )

    (inspection_path / instance_id).mkdir(parents=True, exist_ok=True)
    (inspection_path / instance_id / "inspection_report_logs.txt").write_text(
        result.stdout
    )

    try:
        h.pull_file(
            instance_id,
            "/inspection-report.tar.gz",
            (inspection_path / instance_id / "inspection_report.tar.gz").as_posix(),
        )
    except harness.HarnessError as e:
        LOG.warning("Failed to pull inspection report: %s", e)


@pytest.fixture(scope="session")
def h() -> harness.Harness:
    LOG.debug("Create harness for %s", config.SUBSTRATE)
    if config.SUBSTRATE == "local":
        h = harness.LocalHarness()
    elif config.SUBSTRATE == "lxd":
        h = harness.LXDHarness()
    elif config.SUBSTRATE == "multipass":
        h = harness.MultipassHarness()
    elif config.SUBSTRATE == "juju":
        h = harness.JujuHarness()
    else:
        raise harness.HarnessError(
            "TEST_SUBSTRATE must be one of: local, lxd, multipass, juju"
        )

    yield h

    if config.INSPECTION_REPORTS_DIR is not None:
        for instance_id in h.instances:
            LOG.debug("Generating inspection reports for session instances")
            _generate_inspection_report(h, instance_id)

    _harness_clean(h)


def pytest_configure(config):
    config.addinivalue_line(
        "markers",
        "bootstrap_config: Provide a custom bootstrap config to the bootstrapping node.\n"
        "disable_k8s_bootstrapping: By default, the first k8s node is bootstrapped. This marker disables that.\n"
        "no_setup: No setup steps (pushing snap, bootstrapping etc.) are performed on any node for this test.\n"
        "dualstack: Support dualstack on the instances.\n"
        "etcd_count: Mark a test to specify how many etcd instance nodes need to be created (None by default)\n"
        "node_count: Mark a test to specify how many instance nodes need to be created\n"
        "node_config: Mark a test to specify detail configuration for each created node\n",
    )


@pytest.fixture(scope="function")
def node_config(request) -> List[dict]:
    """Return a list of node configurations for the test.

    The configuration can be specified using the `node_count` or `node_config` markers.
    if `node_count` is used, the configuration for each node is empty, only the count is used.

    example)
    ```python
    @pytest.mark.node_count(3)
    def my_test(instances):
        len(instances) == 3   # because of the node_count marker
    ```

    if `node_config` is used, the configuration returned is a list of dicts
    where each dict represents the configuration for a node.

    example)
    ```python
    @pytest.mark.node_config("snap", ["latest", "1.31"])
    def my_other_test(instances):
        len(instances) == 2   # because of the markers
        # instances[0] has snap installed from latest track
        # instances[1] has snap installed from 1.31 track
    ```
    or
    ```python
    @pytest.mark.node_config("snap,extra_args", [("latest", "something"), ("1.31", "something-else")])
    def my_other_test(instances):
        len(instances) == 2   # because of the markers
        # instances[0] has snap installed from latest track and has extra_args="something"
        # instances[1] has snap installed from 1.31 track and has extra_args="something-else"
    ```
    """

    def _coerce_properties(properties, args):
        """Coerce properties into a tuple of the correct length."""
        arg_len = len(args)
        if isinstance(properties, tuple) and len(properties) == arg_len:
            coerced = properties
        elif not isinstance(properties, tuple) and arg_len == 1:
            coerced = (properties,)
        assert (
            len(coerced) == arg_len
        ), f"Expected {arg_len} properties, got {len(coerced)}"
        return coerced

    node_config_marker = request.node.get_closest_marker("node_config")
    if node_config_marker:
        node_args, node_properties, *_ = node_config_marker.args
        node_args = node_args.split(",")
        return [
            dict(zip(node_args, coerced))
            for properties in node_properties
            if (coerced := _coerce_properties(properties, node_args))
        ]
    node_count_marker = request.node.get_closest_marker("node_count")
    if node_count_marker:
        node_count_arg, *_ = node_count_marker.args
        return [{} for _ in range(int(node_count_arg))]
    # if no markings, create one default instance
    return [{}]


@pytest.fixture(scope="function")
def disable_k8s_bootstrapping(request) -> bool:
    return bool(request.node.get_closest_marker("disable_k8s_bootstrapping"))


@pytest.fixture(scope="function")
def no_setup(request) -> bool:
    return bool(request.node.get_closest_marker("no_setup"))


@pytest.fixture(scope="function")
def bootstrap_config(request) -> Union[str, None]:
    bootstrap_config_marker = request.node.get_closest_marker("bootstrap_config")
    if not bootstrap_config_marker:
        return None
    config, *_ = bootstrap_config_marker.args
    return config


@pytest.fixture(scope="function")
def dualstack(request) -> bool:
    return bool(request.node.get_closest_marker("dualstack"))


@pytest.fixture(scope="function")
def instances(
    h: harness.Harness,
    node_config: List[dict],
    tmp_path: Path,
    disable_k8s_bootstrapping: bool,
    no_setup: bool,
    bootstrap_config: Union[str, None],
    dualstack: bool,
) -> Generator[List[harness.Instance], None, None]:
    """Construct instances for a cluster.

    Bootstrap and setup networking on the first instance, if `disable_k8s_bootstrapping` marker is not set.
    """
    node_count = len(node_config)
    if node_count <= 0:
        pytest.xfail("Test requested 0 or fewer instances, skip this test.")

    LOG.info(f"Creating {node_count} instances")
    instances: List[harness.Instance] = []

    for cfg in node_config:
        # Create <node_count> instances and setup the k8s snap in each.
        instance = h.new_instance(dualstack=dualstack)
        instances.append(instance)
        if not no_setup:
            util.setup_k8s_snap(instance, tmp_path, **cfg)

    if not disable_k8s_bootstrapping and not no_setup:
        first_node, *_ = instances

        if bootstrap_config is not None:
            first_node.exec(
                ["k8s", "bootstrap", "--file", "-"],
                input=str.encode(bootstrap_config),
            )
        else:
            first_node.exec(["k8s", "bootstrap"])

    yield instances

    if config.SKIP_CLEANUP:
        LOG.warning("Skipping clean-up of instances, delete them on your own")
        return

    # Cleanup after each test.
    # We cannot execute _harness_clean() here as this would also
    # remove the session_instance. The harness ensures that everything is cleaned up
    # at the end of the test session.
    for instance in instances:
        if config.INSPECTION_REPORTS_DIR is not None:
            LOG.debug("Generating inspection reports for test instances")
            _generate_inspection_report(h, instance.id)

        h.delete_instance(instance.id)


@pytest.fixture(scope="session")
def session_instance(
    h: harness.Harness, tmp_path_factory: pytest.TempPathFactory
) -> Generator[harness.Instance, None, None]:
    """Constructs and bootstraps an instance that persists over a test session.

    Bootstraps the instance with all k8sd features enabled to reduce testing time.
    """
    LOG.info("Setup node and enable all features")

    tmp_path = tmp_path_factory.mktemp("data")
    instance = h.new_instance()
    util.setup_k8s_snap(instance, tmp_path)

    bootstrap_config_path = "/home/ubuntu/bootstrap-session.yaml"
    instance.send_file(
        (config.MANIFESTS_DIR / "bootstrap-session.yaml").as_posix(),
        bootstrap_config_path,
    )

    instance.exec(["k8s", "bootstrap", "--file", bootstrap_config_path])
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    yield instance


@pytest.fixture(scope="function")
def etcd_count(request) -> int:
    etcd_count_marker = request.node.get_closest_marker("etcd_count")
    if not etcd_count_marker:
        return 0
    etcd_count_arg, *_ = etcd_count_marker.args
    return int(etcd_count_arg)


@pytest.fixture(scope="function")
def etcd_cluster(
    h: harness.Harness, etcd_count: int
) -> Generator[EtcdCluster, None, None]:
    """Construct etcd instances for a cluster."""
    LOG.info(f"Creating {etcd_count} etcd instances")

    cluster = EtcdCluster(h, initial_node_count=etcd_count)

    yield cluster
