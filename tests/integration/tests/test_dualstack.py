#
# Copyright 2024 Canonical, Ltd.
#
import yaml
import logging
from pathlib import Path

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(2)
def test_dualstack(h: harness.Harness, tmp_path: Path):
    util.run(
        [
            "lxc",
            "network",
            "create",
            config.LXD_DUALSTACK_NETWORK,
            "ipv4.address=10.90.60.1/24",
            "ipv6.address=fd42:1e1e:7a2f:326e::1/64",
            "ipv4.nat=true",
            "ipv6.nat=true",
        ],
        check=False
    )

    util.run(
        [
            "lxc",
            "network",
            "ls"
        ]
    )

    out = util.run(
        [
            "lxc",
            "network",
            "show",
            config.LXD_DUALSTACK_NETWORK
        ],
        text=True,
        capture_output=True
    )
    parsed = yaml.safe_load(out.stdout)

    print(parsed)

    snap_path = (tmp_path / "k8s.snap").as_posix()
    main = h.new_instance(dualstack=True)
    util.setup_k8s_snap(main, snap_path)
