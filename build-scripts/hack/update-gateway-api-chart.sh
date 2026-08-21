#!/bin/bash

VERSION="v1.4.1"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../k8s/manifests/charts"

cd "$CHARTS_PATH"

git clone https://github.com/kubernetes-sigs/gateway-api --depth 1 -b "${VERSION}" gateway-api-src

rm -rf "gateway-api-${VERSION:1}.tgz"

helm create gateway-api
rm -rf gateway-api/templates/*
rm -rf gateway-api/charts
cp gateway-api-src/config/crd/standard/* gateway-api/templates/
# The standard channel also ships a ValidatingAdmissionPolicy that restricts how
# the CRDs may be upgraded. This chart only carries the CRDs themselves.
rm -f gateway-api/templates/gateway.networking.k8s.io_vap_safeupgrades.yaml
sed -i 's/^\(version: \).*$/\1'"${VERSION:1}"'/' gateway-api/Chart.yaml
sed -i 's/^\(appVersion: \).*$/\1'"${VERSION:1}"'/' gateway-api/Chart.yaml
sed -i 's/^\(description: \).*$/\1'"A Helm Chart containing Gateway API CRDs"'/' gateway-api/Chart.yaml
helm package gateway-api

rm -rf gateway-api-src
rm -rf gateway-api
