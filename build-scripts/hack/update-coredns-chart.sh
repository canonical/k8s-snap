#!/bin/bash

VERSION="1.36.2"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../src/k8s/pkg/k8sd/features/coredns/charts"

cd "$CHARTS_PATH"

helm pull --repo https://coredns.github.io/helm coredns --version $VERSION
