#!/bin/bash

CONTOUR_VERSION="v1.28.2"
DIR=$(realpath $(dirname "${0}"))
CHARTS_PATH="$DIR/../../k8s/components/charts"

cd "$CHARTS_PATH"
# Download the common CRDs
echo "Downloading common CRDs from Contour ${CONTOUR_VERSION}"

git clone https://github.com/projectcontour/contour --depth 1 -b "${CONTOUR_VERSION}" contour-src
# curl -s -o "${CHARTS_PATH}/contour/templates/common-crds.yaml" "${COMMON_URL}"

# Common CRDS for contour gateway and ingress
rm -rf "ck-contour-common-${CONTOUR_VERSION:1}.tgz"
helm create ck-contour-common

rm -rf ck-contour-common/templates
rm -rf ck-contour-common/charts
rm -rf ck-contour-common/values.yaml
mkdir -p ck-contour-common/crds

cp contour-src/examples/contour/01-crds.yaml ck-contour-common/crds/
sed -i 's/^\(version: \).*$/\1'"${CONTOUR_VERSION:1}"'/' ck-contour-common/Chart.yaml
sed -i 's/^\(appVersion: \).*$/\1'"${CONTOUR_VERSION:1}"'/' ck-contour-common/Chart.yaml
sed -i 's/^\(description: \).*$/\1'"A Helm Chart containing Contour common CRDs"'/' ck-contour-common/Chart.yaml

helm package ck-contour-common
rm -rf ck-contour-common

# Contour Gateway Provisioner
helm create ck-gateway-contour
rm -rf ck-gateway-contour/templates/*
rm -rf ck-gateway-contour/charts
rm -rf ck-gateway-contour/values.yaml
mkdir -p ck-gateway-contour/crds

cp contour-src/examples/gateway/00-crds.yaml ck-gateway-contour/crds/
cp contour-src/examples/gateway-provisioner/00-common.yaml ck-gateway-contour/templates/
cp contour-src/examples/gateway-provisioner/01-roles.yaml ck-gateway-contour/templates/
cp contour-src/examples/gateway-provisioner/02-rolebindings.yaml ck-gateway-contour/templates/
cp contour-src/examples/gateway-provisioner/03-gateway-provisioner.yaml ck-gateway-contour/templates/

# change gateway provisioner image to use the values from values.yaml
sed -i "s|image: ghcr.io/projectcontour/contour:${CONTOUR_VERSION}|image: \"{{ .Values.projectcontour.image.registry }}/{{ .Values.projectcontour.image.repository }}:{{ .Values.projectcontour.image.tag }}\"|" ck-gateway-contour/templates/03-gateway-provisioner.yaml

# Add image args to the gateway provisioner 
sed -i '/^        - --enable-leader-election$/a\ \ \ \ \ \ \ \ - --envoy-image={{ .Values.envoyproxy.image.registry }}/{{ .Values.envoyproxy.image.repository }}:{{ .Values.envoyproxy.image.tag }}\n\ \ \ \ \ \ \ \ - --contour-image={{ .Values.projectcontour.image.registry }}/{{ .Values.projectcontour.image.repository }}:{{ .Values.projectcontour.image.tag }}' ck-gateway-contour/templates/03-gateway-provisioner.yaml

# Add values.yaml
cat <<EOF >ck-gateway-contour/values.yaml
projectcontour:
  image:
    registry: ghcr.io
    repository: projectcontour/contour
    tag: ${CONTOUR_VERSION}

envoyproxy:
  image:
    registry: docker.io
    repository: envoyproxy/envoy
    tag: ${CONTOUR_VERSION}
EOF

# Add Gateway Class
cat <<EOF >ck-gateway-contour/templates/ck-gateway-class.yaml
---
kind: GatewayClass
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: ck-gateway
spec:
  controllerName: projectcontour.io/gateway-controller
EOF

# Remove the Namespace resource from 00-common.yaml
sed -i '1,5d' ck-gateway-contour/templates/00-common.yaml

sed -i 's/^\(version: \).*$/\1'"${CONTOUR_VERSION:1}"'/' ck-gateway-contour/Chart.yaml
sed -i 's/^\(appVersion: \).*$/\1'"${CONTOUR_VERSION:1}"'/' ck-gateway-contour/Chart.yaml
sed -i 's/^\(description: \).*$/\1'"A Helm Chart containing Contour Gateway Provisioner"'/' ck-gateway-contour/Chart.yaml

helm package ck-gateway-contour
rm -rf ck-gateway-contour

# Remove the github source code
rm -rf contour-src
