#!/bin/bash

VERSION="1.17.1"
DIR=`realpath $(dirname "${0}")`

CHARTS_PATH="$DIR/../../k8s/manifests/charts"

cd "$CHARTS_PATH"

helm pull --repo https://helm.cilium.io/ cilium --version $VERSION
