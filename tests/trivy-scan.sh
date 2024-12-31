#!/usr/bin/env bash

SCRIPT_DIR=$(dirname "$BASH_SOURCE")

set -ex
cd "${SCRIPT_DIR}/.."

SNAP_PATH="$1"
if [[ ! -f $SNAP_PATH ]]; then
  echo "Usage: $0 <snap_path>"
  exit 1
fi

# Setup Trivy vulnerability scanner
mkdir -p manual-trivy/sarifs
pushd manual-trivy
VER=$(curl --silent -qI https://github.com/aquasecurity/trivy/releases/latest | awk -F '/' '/^location/ {print  substr($NF, 1, length($NF)-1)}');
wget https://github.com/aquasecurity/trivy/releases/download/${VER}/trivy_${VER#v}_Linux-64bit.tar.gz
tar -zxvf ./trivy_${VER#v}_Linux-64bit.tar.gz
popd

# Run Trivy vulnerability scanner in repo mode
./manual-trivy/trivy fs . \
  --format sarif \
  --db-repository public.ecr.aws/aquasecurity/trivy-db \
  --severity "MEDIUM,HIGH,CRITICAL" \
  --ignore-unfixed \
  > ./manual-trivy/sarifs/trivy-k8s-repo-scan--results.sarif

for var in $(env | grep -o '^TRIVY_[^=]*'); do
  unset "$var"
done
cp "${SNAP_PATH}" ./k8s-test.snap
rm -rf ./squashfs-root
unsquashfs k8s-test.snap
./manual-trivy/trivy rootfs ./squashfs-root/ \
  --format sarif \
  --db-repository public.ecr.aws/aquasecurity/trivy-db \
  > ./manual-trivy/sarifs/snap.sarif
