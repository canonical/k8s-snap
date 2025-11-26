#
# Copyright 2025 Canonical, Ltd.
#
from pytest_tagging import combine_tags

PULL_REQUEST = "pull_request"
NIGHTLY = "nightly"
WEEKLY = "weekly"
GPU = "gpu"
CONFORMANCE = "conformance_tests"

TEST_LEVELS = [PULL_REQUEST, NIGHTLY, WEEKLY, CONFORMANCE]

# Those tags can be used for a convenient way to run multiple test levels.
combine_tags("up_to_nightly", PULL_REQUEST, NIGHTLY)
combine_tags("up_to_weekly", PULL_REQUEST, NIGHTLY, WEEKLY)
