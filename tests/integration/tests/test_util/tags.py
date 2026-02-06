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
PROMOTE_CANDIDATE = "beta_to_candidate"
PROMOTE_STABLE = "candidate_to_stable"

# Each test needs to be tagged with at least one of the following tags.
# This is enforced by conftest.
TEST_TAGS = [
    PULL_REQUEST,
    NIGHTLY,
    WEEKLY,
    CONFORMANCE,
    PERFORMANCE,
    GPU,
    PROMOTE_CANDIDATE,
    PROMOTE_STABLE,
]

# Those tags can be used for a convenient way to run multiple test levels.
combine_tags("up_to_nightly", PULL_REQUEST, NIGHTLY)
combine_tags("up_to_weekly", PULL_REQUEST, NIGHTLY, WEEKLY)
combine_tags("edge_to_beta", PULL_REQUEST)
