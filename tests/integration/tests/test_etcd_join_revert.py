#
# Copyright 2026 Canonical, Ltd.
#
import base64
import concurrent.futures
import json
import logging
import time
from typing import Callable, List, Optional

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

ETCD_PEER_PORT = 2380

# kube-apiserver refuses to start with an unknown flag. The joining node waits
# for its local kube-apiserver, so the join times out, but only after its etcd
# learner has been promoted to a voter. That leaves a wide, deterministic window
# between the promotion and the join revert.
FAILING_JOIN_CONFIG = """
extra-node-kube-apiserver-args:
  --integration-test-invalid-flag: "true"
"""


def _wait_for(
    condition: Callable[[], bool], timeout_s: int, interval_s: int, what: str
):
    deadline = time.monotonic() + timeout_s
    last_error = None
    while time.monotonic() < deadline:
        try:
            if condition():
                return
            last_error = None
        except Exception as e:  # noqa: BLE001 - retried until the deadline
            last_error = e
        time.sleep(interval_s)
    raise TimeoutError(f"timed out after {timeout_s}s waiting for {what}: {last_error}")


def _etcd_api(
    instance: harness.Instance, path: str, body: Optional[dict] = None
) -> dict:
    """POST to the node's local etcd gRPC gateway with the apiserver client certificate."""
    proc = instance.exec(
        [
            "curl",
            "--silent",
            "--show-error",
            "--fail",
            "--max-time",
            "10",
            "--cacert",
            "/etc/kubernetes/pki/etcd/ca.crt",
            "--cert",
            "/etc/kubernetes/pki/apiserver-etcd-client.crt",
            "--key",
            "/etc/kubernetes/pki/apiserver-etcd-client.key",
            f"https://127.0.0.1:2379{path}",
            "-X",
            "POST",
            "-d",
            json.dumps(body or {}),
        ],
        capture_output=True,
        text=True,
    )
    return json.loads(proc.stdout)


def _etcd_members(instance: harness.Instance) -> List[dict]:
    return _etcd_api(instance, "/v3/cluster/member/list").get("members", [])


def _etcd_member_by_ip(instance: harness.Instance, ip: str) -> Optional[dict]:
    peer_url = f"https://{ip}:{ETCD_PEER_PORT}"
    for member in _etcd_members(instance):
        if peer_url in member.get("peerURLs", []):
            return member
    return None


def _etcd_put(instance: harness.Instance, key: str) -> bool:
    """Write a key through raft. Fails (raises) when etcd has no quorum."""
    _etcd_api(
        instance,
        "/v3/kv/put",
        {
            "key": base64.b64encode(key.encode()).decode(),
            "value": base64.b64encode(str(time.time()).encode()).decode(),
        },
    )
    return True


def _diagnostics(cp2: harness.Instance, cp3: harness.Instance) -> str:
    try:
        members = [
            (m.get("name", ""), m.get("isLearner", False)) for m in _etcd_members(cp2)
        ]
    except Exception as e:  # noqa: BLE001 - diagnostics only
        members = f"<unavailable: {e}>"
    cp3_etcd = cp3.exec(
        ["systemctl", "is-active", "snap.k8s.etcd"],
        capture_output=True,
        text=True,
        check=False,
    ).stdout.strip()
    return f"etcd members seen from cp2: {members}, etcd service on cp3: {cp3_etcd}"


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_failed_join_revert_keeps_etcd_quorum(instances: List[harness.Instance]):
    """A failed control-plane join must not take etcd quorum down with it.

    Regression test for canonical/k8sd#159. The join revert ignored etcd
    rejecting the removal of the promoted joiner, then stopped the joiner's
    etcd and deleted its data anyway.

    cp1 and cp2 are etcd voters. cp3's join fails after its learner has been
    promoted, and cp1's etcd goes down before the revert. etcd rejects removing
    cp3, because that would leave only 1 of 2 voters active. cp3 must keep
    running etcd: with cp1 down, cp2 and cp3 hold the quorum.
    """
    cp1, cp2, cp3 = instances
    cp3_ip = util.get_default_ip(cp3)

    util.wait_until_k8s_ready(cp1, [cp1])
    util.join_cluster(cp2, util.get_join_token(cp1, cp2))
    util.wait_until_k8s_ready(cp1, [cp1, cp2])
    join_token = util.get_join_token(cp1, cp3)

    with concurrent.futures.ThreadPoolExecutor(max_workers=1) as executor:
        join = executor.submit(
            cp3.exec,
            ["k8s", "join-cluster", join_token, "--file", "-", "--timeout", "120s"],
            input=FAILING_JOIN_CONFIG.encode(),
            capture_output=True,
            check=False,
        )

        def _cp3_is_voter() -> bool:
            if join.done():
                raise AssertionError(
                    "join of cp3 ended before its etcd member was promoted"
                )
            member = _etcd_member_by_ip(cp2, cp3_ip)
            return member is not None and not member.get("isLearner", False)

        _wait_for(
            _cp3_is_voter,
            timeout_s=110,
            interval_s=1,
            what="cp3 to become an etcd voter",
        )
        assert not join.done(), "join of cp3 ended before cp1's etcd could be stopped"

        LOG.info(
            "cp3 is an etcd voter, stopping etcd on cp1 before the join is reverted"
        )
        cp1.exec(["snap", "stop", "k8s.etcd"])

        result = join.result(timeout=300)

    assert result.returncode != 0, "join of cp3 was expected to fail"
    LOG.info("join of cp3 failed as expected: %s", result.stderr.decode().strip())

    # The CLI returns at its timeout while the server-side revert (which
    # retries a rejected removal for up to 30s) and microcluster's forced
    # deletion of cp3 (which runs cp3's PreRemove hook) may still be running.
    # k8s status can't be used to wait here: it needs kube-apiserver, which
    # needs the etcd quorum this test is about.
    time.sleep(90)

    # Core assertion. Before the fix, cp3's etcd was stopped after the
    # rejected removal, leaving cp2 as the only live voter out of three.
    try:
        _wait_for(
            lambda: _etcd_put(cp2, "/integration-test/join-revert"),
            timeout_s=60,
            interval_s=3,
            what="an etcd write on cp2",
        )
    except TimeoutError as e:
        pytest.fail(
            f"etcd lost quorum after the failed join of cp3 was reverted: {e}. "
            + _diagnostics(cp2, cp3)
        )

    # Bring cp1 back and remove the leftover cp3 voter, as the revert's log
    # instructs. The cluster must end up healthy with only cp1 and cp2.
    cp1.exec(["snap", "start", "k8s.etcd"])

    def _cp3_removed_from_etcd() -> bool:
        member = _etcd_member_by_ip(cp1, cp3_ip)
        if member is not None:
            _etcd_api(cp1, "/v3/cluster/member/remove", {"ID": member["ID"]})
        return _etcd_member_by_ip(cp1, cp3_ip) is None

    _wait_for(
        _cp3_removed_from_etcd,
        timeout_s=120,
        interval_s=3,
        what="removal of cp3 from etcd",
    )
    _wait_for(
        lambda: not util.diverged_cluster_memberships(cp1, [cp1, cp2]),
        timeout_s=120,
        interval_s=5,
        what="cluster memberships to converge on cp1 and cp2",
    )
    assert _etcd_put(cp1, "/integration-test/join-revert")
