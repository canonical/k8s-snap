#!/bin/bash

VERSION="v1.17.1"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../src/k8s/pkg/k8sd/features/cilium/charts"

cd "$CHARTS_PATH" || exit

helm pull --repo https://helm.cilium.io cilium --version $VERSION
