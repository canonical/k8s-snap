#!/bin/bash

VERSION="3.12.0"
DIR=`realpath $(dirname "${0}")`

CHARTS_PATH="$DIR/../../k8s/components/charts"

cd "$CHARTS_PATH"

helm pull --repo https://kubernetes-sigs.github.io/metrics-server/ metrics-server --version $VERSION
