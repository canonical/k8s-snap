#!/bin/bash

VERSION="3.12.2"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../src/k8s/pkg/k8sd/features/metrics-server/charts"

cd "$CHARTS_PATH"

helm pull --repo https://kubernetes-sigs.github.io/metrics-server/ metrics-server --version $VERSION
