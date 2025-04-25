#!/bin/bash

VERSION="1.39.2"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../k8s/manifests/charts"

cd "$CHARTS_PATH"

helm pull --repo https://coredns.github.io/helm coredns --version $VERSION
