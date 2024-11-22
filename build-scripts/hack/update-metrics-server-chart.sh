#!/bin/bash

VERSION="3.12.2"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../k8s/manifests/charts"

cd "$CHARTS_PATH"

helm pull --repo https://kubernetes-sigs.github.io/metrics-server/ metrics-server --version $VERSION
