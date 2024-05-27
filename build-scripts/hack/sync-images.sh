#!/bin/bash

# Description:
#   Sync images from upstream repositories under ghcr.io/canonical.
#
# Usage:
#   $ USERNAME="$username" PASSWORD="$password" ./sync-images.sh

DIR="$(realpath "$(dirname "${0}")")"

"${DIR}/../../src/k8s/tools/regsync.sh" once "${DIR}/sync-images.yaml"
