#!/bin/bash

VERSION="v0.7.1"
DIR=`realpath $(dirname "${0}")`

CHARTS_PATH="$DIR/../k8s/components/charts"

cd "$CHARTS_PATH"

git clone https://github.com/kubernetes-sigs/gateway-api --depth 1 -b "${VERSION}" gateway-api-src

rm -rf "gateway-api-${VERSION:1}.tgz"

helm create gateway-api
rm -rf gateway-api/templates/*
rm -rf gateway-api/charts
cp gateway-api-src/config/crd/standard/* gateway-api/templates/
cp gateway-api-src/config/crd/experimental/gateway.networking.k8s.io_tlsroutes.yaml gateway-api/templates/
sed -i 's/^\(version: \).*$/\1'"${VERSION:1}"'/' gateway-api/Chart.yaml
sed -i 's/^\(appVersion: \).*$/\1'"${VERSION:1}"'/' gateway-api/Chart.yaml
sed -i 's/^\(description: \).*$/\1'"A Helm Chart containing Gateway API CRDs"'/' gateway-api/Chart.yaml
helm package gateway-api

rm -rf gateway-api-src
rm -rf gateway-api
