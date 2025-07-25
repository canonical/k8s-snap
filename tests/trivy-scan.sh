#!/usr/bin/env bash

SCRIPT_DIR=$(realpath "$(dirname "$BASH_SOURCE")")

set -ex
cd "${SCRIPT_DIR}/.."

SNAP_PATH="$1"
if [[ ! -f $SNAP_PATH ]]; then
  echo "Usage: $0 <snap_path>"
  exit 1
fi

# Setup Trivy vulnerability scanner
mkdir -p .trivy/sarifs
pushd .trivy
VER=$(curl --silent -qI https://github.com/aquasecurity/trivy/releases/latest | awk -F '/' '/^location/ {print  substr($NF, 1, length($NF)-1)}');
wget https://github.com/aquasecurity/trivy/releases/download/${VER}/trivy_${VER#v}_Linux-64bit.tar.gz
tar -zxvf ./trivy_${VER#v}_Linux-64bit.tar.gz
popd

# Run Trivy vulnerability scanner in repo mode.
#
# We'll have two runs:
# * one with SARIF output, used by GitHub
# * one with "json" output
#   * SARIF is also a json but not as well structured
#   * the list of vulnerabilities is easier to parse and compare with the CISA list
#   * the second run will not filter the records based on severity
TRIVY_FS_SARIF="./.trivy/sarifs/trivy-k8s-repo-scan--results.sarif"

# Run Trivy vulnerability scanner in repo mode
./.trivy/trivy fs . \
  --format sarif \
  --db-repository public.ecr.aws/aquasecurity/trivy-db \
  --severity "MEDIUM,HIGH,CRITICAL" \
  --ignore-unfixed \
  > "$TRIVY_FS_SARIF"

./.trivy/trivy fs . \
  --format json \
  --db-repository public.ecr.aws/aquasecurity/trivy-db \
  --ignore-unfixed \
  > ./.trivy/sarifs/trivy-k8s-repo-scan--results.json

# Run Trivy vulnerability scanner in rootfs mode, scanning the snap
for var in $(env | grep -o '^TRIVY_[^=]*'); do
  unset "$var"
done

cp "${SNAP_PATH}" ./.trivy/k8s-test.snap
rm -rf ./.trivy/squashfs-root
pushd ./.trivy
unsquashfs ./k8s-test.snap
popd

TRIVY_ROOTFS_SARIF="./.trivy/sarifs/snap.sarif"

./.trivy/trivy rootfs ./.trivy/squashfs-root/ \
  --format sarif \
  --db-repository public.ecr.aws/aquasecurity/trivy-db \
  > "$TRIVY_ROOTFS_SARIF"

./.trivy/trivy rootfs ./.trivy/squashfs-root/ \
  --format json \
  --db-repository public.ecr.aws/aquasecurity/trivy-db \
  > ./.trivy/sarifs/snap.json

function split_sarif_runs() {
  local input_sarif="$1"
  local prefix="$2"

  local count
  count=$(jq '.runs | length' "$input_sarif")
  if [[ "$count" -le 1 ]]; then
    echo "$input_sarif"
    return 0
  fi

  echo "Splitting SARIF: $input_sarif (contains $count runs)"
  local output_files=()
  local base="${input_sarif%.sarif}"
  for i in $(seq 0 $((count - 1))); do
    local output="${base}--run${i}.sarif"
    jq --argjson idx "$i" '{version: "2.1.0", runs: [ .runs[$idx] ]}' "$input_sarif" > "$output"
    output_files+=("$output")
  done
  rm "$input_sarif"
  echo "${output_files[@]}"
}


# Split and collect final SARIF files
REPO_SARIF_FILES=($(split_sarif_runs "$TRIVY_FS_SARIF" "trivy-k8s-repo-scan"))
ROOTFS_SARIF_FILES=($(split_sarif_runs "$TRIVY_ROOTFS_SARIF" "snap"))

# Obtain CISA Known Exploited Vulnerabilities list.
curl -s -o ./.trivy/kev.json \
  https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json

function get_cisa_kev_cves() {
  local kevJson=$1
  local trivyJsonReport=$2

  set +x
  local kev_cves=$(jq -r '.vulnerabilities[].cveID' "$kevJson" | sort -u)
  local found_cves=$(jq -r '.Results[] | select(.Vulnerabilities != null) | .Vulnerabilities[].VulnerabilityID' "$trivyJsonReport" | sort -u)
  local matches="$(echo "$found_cves" | grep -F -f <(echo "$kev_cves") || true)"
  set -x

  if [ -n "$matches" ]; then
    echo "KEV listed vulnerabilities found in $trivyJsonReport:"
    echo "$matches"
    exit 1
  fi
}

# Compare the trivy reports with the CISA KEV list
get_cisa_kev_cves ./.trivy/kev.json ./.trivy/sarifs/trivy-k8s-repo-scan--results.json
get_cisa_kev_cves ./.trivy/kev.json ./.trivy/sarifs/snap.json

echo "Final SARIF files:"
for sarif in "${REPO_SARIF_FILES[@]}" "${ROOTFS_SARIF_FILES[@]}"; do
  echo "$sarif"
done
