#
# Copyright 2026 Canonical, Ltd.
#
from pytest_tagging import combine_tags

PULL_REQUEST = "pull_request"
NIGHTLY = "nightly"
WEEKLY = "weekly"
GPU = "gpu"
CONFORMANCE = "conformance_tests"
PERFORMANCE = "performance"

# Each test needs to be tagged with at least one of the following tags.
# This is enforced by conftest.
TEST_TAGS = [PULL_REQUEST, NIGHTLY, WEEKLY, CONFORMANCE, PERFORMANCE, GPU]

# Those tags can be used for a convenient way to run multiple test levels.
combine_tags("up_to_nightly", PULL_REQUEST, NIGHTLY)
combine_tags("up_to_weekly", PULL_REQUEST, NIGHTLY, WEEKLY)
