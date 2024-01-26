#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

from e2e_util import harness, util

LOG = logging.getLogger(__name__)


def test_smoke(instances: List[harness.Instance]):
    util.wait_until_k8s_ready(instances[0], instances)
