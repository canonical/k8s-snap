#!/bin/bash
DIR=$(realpath $(dirname "${0}"))

docker run -v "${DIR}/registry-k8s-io.yaml":/config.yaml quay.io/skopeo/stable:v1.15 sync \
  --src yaml \
  --dest docker \
  /config.yaml ghcr.io/canonical \
  --format oci \
  --dest-creds "${DEST_CREDS}"
