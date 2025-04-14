#
# Copyright 2025 Canonical, Ltd.
#

import logging

import pytest
from test_util import tags

LOG = logging.getLogger(__name__)


@pytest.mark.tags(tags.PULL_REQUEST)
def test_pytest_err():
    LOG.error("exp err 0")
    raise Exception("expected exc")


@pytest.mark.tags(tags.PULL_REQUEST)
def test_pytest_err2():
    LOG.error("exp err 2")
    raise Exception("expected exc2")


@pytest.mark.tags(tags.PULL_REQUEST)
def test_pytest_pass():
    LOG.info("exp pass")
    pass
