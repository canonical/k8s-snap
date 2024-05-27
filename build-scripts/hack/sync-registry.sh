#!/bin/bash

docker run -v "./sync-registry-config.yaml":/config.yaml quay.io/skopeo/stable:v1.15 sync \
  --src yaml \
  --dest docker \
  /config.yaml ghcr.io/canonical \
  --format oci \
  --all \
  --dest-creds "${DEST_CREDS}"
