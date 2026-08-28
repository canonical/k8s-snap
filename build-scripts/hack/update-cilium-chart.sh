#!/bin/bash

VERSION="1.20.0"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../k8s/manifests/charts"

cd "$CHARTS_PATH" || exit

helm pull --repo https://helm.cilium.io cilium --version $VERSION
