#!/bin/bash

VERSION="0.15.3"
DIR=$(realpath $(dirname "${0}"))

CHARTS_PATH="$DIR/../../k8s/manifests/charts"

cd "$CHARTS_PATH"

helm pull --repo https://metallb.github.io/metallb metallb --version $VERSION
