#!/bin/bash

docker run -v "./sync-registry-config.yaml":/config.yaml ghcr.io/canonical/stable:1.15.0 sync \
  --src yaml \
  --dest docker \
  /config.yaml ghcr.io/canonical \
  --format oci \
  --all \
  --dest-creds "${DEST_CREDS}"
