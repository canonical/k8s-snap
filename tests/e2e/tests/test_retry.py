#
# Copyright 2024 Canonical, Ltd.
#
from e2e_util import util


def test_retry():
    util.stubbornly().exec("boo")
