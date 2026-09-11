#
# Copyright 2026 Canonical, Ltd.
#
"""CI failure classification and metrics toolkit.

See `docs/proposals/003-ci-failure-classification-and-metrics.md` for the design.
"""

# Schema version for records written to the metrics store. Bump on any
# breaking change to the record shapes in `models.py`.
SCHEMA_VERSION = "v1"

# Version of the ingest logic. Bump when the meaning of an ingested field
# changes, so records can be identified as needing re-ingestion.
INGEST_VERSION = 1

# Version of the failure taxonomy (the `class`/`subclass` value space).
# Stored on every classified record so historical comparisons stay honest
# when the taxonomy evolves.
TAXONOMY_VERSION = 2

__all__ = [
    "SCHEMA_VERSION",
    "INGEST_VERSION",
    "TAXONOMY_VERSION",
]
