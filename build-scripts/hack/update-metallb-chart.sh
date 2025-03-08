#!/bin/bash

VERSION="0.14.9"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../src/k8s/pkg/k8sd/features/metallb/charts"

cd "$CHARTS_PATH"

helm pull --repo https://metallb.github.io/metallb metallb --version $VERSION
