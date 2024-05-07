#!/bin/bash

actor=${ACTOR}
token=${TOKEN}
dir=${DIR}

docker run -v "$dir":/tmp/ quay.io/skopeo/stable:v1.15 sync \
  --src yaml \
  --dest docker \
  /tmp/.github/data/sync-registry-config.yaml ghcr.io/canonical \
  --format oci \
  --dest-creds "$actor":"$token"