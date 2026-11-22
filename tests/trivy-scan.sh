#!/usr/bin/env bash

SCRIPT_DIR=$(realpath "$(dirname "$BASH_SOURCE")")

set -ex
cd "${SCRIPT_DIR}/.."

REPO_ONLY=false
if [[ "${1:-}" == "--repo-only" ]]; then
  REPO_ONLY=true
else
  SNAP_PATH="${1:-}"
  if [[ ! -f "$SNAP_PATH" ]]; then
    echo "Usage: $0 <snap_path> | --repo-only"
    exit 1
  fi
fi

# Setup Trivy vulnerability scanner
TRIVY_VERSION="v0.69.3"
mkdir -p .trivy/sarifs
trap 'rm -f ./.trivy/sarifs/*.tmp' EXIT

pushd .trivy
TRIVY_TARBALL="trivy_${TRIVY_VERSION#v}_Linux-64bit.tar.gz"
if [[ ! -x ./trivy ]] || [[ "$(./trivy --version | head -n1)" != "Version: ${TRIVY_VERSION#v}" ]]; then
  wget --tries=5 --waitretry=2 -O "$TRIVY_TARBALL" "https://github.com/aquasecurity/trivy/releases/download/${TRIVY_VERSION}/${TRIVY_TARBALL}"
  tar -zxvf "./${TRIVY_TARBALL}"
fi
popd

# Resolve and isolate a consistent cache directory across all Trivy invocations
TRIVY_CACHE_DIR="${TRIVY_CACHE_DIR:-${HOME}/.cache/trivy}"
mkdir -p "$TRIVY_CACHE_DIR"

# Isolate download attempts from global credentials and config to prevent credential exposure across registries
unset TRIVY_USERNAME TRIVY_PASSWORD TRIVY_REGISTRY_TOKEN TRIVY_AUTH_TOKEN TRIVY_CONFIG

# Download vulnerability database with mirror fallbacks and retries
DB_REPOSITORIES=(
  "ghcr.io/aquasecurity/trivy-db:2"
  "public.ecr.aws/aquasecurity/trivy-db:2"
  "mirror.gcr.io/aquasec/trivy-db:2"
)
DB_DOWNLOADED=false
for repo in "${DB_REPOSITORIES[@]}"; do
  for attempt in 1 2 3; do
    echo "Attempting to download Trivy vulnerability DB from ${repo} (attempt ${attempt})..."
    if ./.trivy/trivy image --config /dev/null --cache-dir "$TRIVY_CACHE_DIR" --download-db-only --db-repository "$repo"; then
      DB_DOWNLOADED=true
      break 2
    fi
    echo "Attempt ${attempt} to download Trivy vulnerability DB failed. Retrying..."
    sleep 3
  done
done
if [[ "$DB_DOWNLOADED" != "true" ]]; then
  echo "Failed to download Trivy vulnerability DB from all repositories."
  exit 1
fi

# Run Trivy vulnerability scanner in repo mode.
#
# We'll have two runs:
# * one with SARIF output, used by GitHub
# * one with "json" output
#   * SARIF is also a json but not as well structured
#   * the list of vulnerabilities is easier to parse and compare with the CISA list
#   * the second run will not filter the records based on severity
TRIVY_FS_SARIF="./.trivy/sarifs/trivy-k8s-repo-scan--results.sarif"
TRIVY_FS_JSON="./.trivy/sarifs/trivy-k8s-repo-scan--results.json"

# Run Trivy vulnerability scanner in repo mode
./.trivy/trivy fs . \
  --cache-dir "$TRIVY_CACHE_DIR" \
  --format sarif \
  --skip-db-update \
  --severity "MEDIUM,HIGH,CRITICAL" \
  --ignore-unfixed \
  --output "${TRIVY_FS_SARIF}.tmp"
mv "${TRIVY_FS_SARIF}.tmp" "$TRIVY_FS_SARIF"

./.trivy/trivy fs . \
  --cache-dir "$TRIVY_CACHE_DIR" \
  --format json \
  --skip-db-update \
  --ignore-unfixed \
  --output "${TRIVY_FS_JSON}.tmp"
mv "${TRIVY_FS_JSON}.tmp" "$TRIVY_FS_JSON"

if [[ "$REPO_ONLY" == "false" ]]; then
  # Run Trivy vulnerability scanner in rootfs mode, scanning the snap
  for var in $(env | grep -o '^TRIVY_[^=]*'); do
    if [[ "$var" != "TRIVY_CACHE_DIR" ]]; then
      unset "$var"
    fi
  done

  cp "${SNAP_PATH}" ./.trivy/k8s-test.snap
  rm -rf ./.trivy/squashfs-root
  pushd ./.trivy
  unsquashfs ./k8s-test.snap
  popd

  TRIVY_ROOTFS_SARIF="./.trivy/sarifs/snap.sarif"
  TRIVY_ROOTFS_JSON="./.trivy/sarifs/snap.json"

  ./.trivy/trivy rootfs ./.trivy/squashfs-root/ \
    --cache-dir "$TRIVY_CACHE_DIR" \
    --format sarif \
    --skip-db-update \
    --output "${TRIVY_ROOTFS_SARIF}.tmp"
  mv "${TRIVY_ROOTFS_SARIF}.tmp" "$TRIVY_ROOTFS_SARIF"

  ./.trivy/trivy rootfs ./.trivy/squashfs-root/ \
    --cache-dir "$TRIVY_CACHE_DIR" \
    --format json \
    --skip-db-update \
    --output "${TRIVY_ROOTFS_JSON}.tmp"
  mv "${TRIVY_ROOTFS_JSON}.tmp" "$TRIVY_ROOTFS_JSON"
fi

function split_sarif_runs() {
  local input_sarif="$1"
  local prefix="$2"

  local count
  count=$(jq '.runs | length' "$input_sarif")
  if [[ "$count" -le 1 ]]; then
    echo "$input_sarif"
    return 0
  fi

  echo "Splitting SARIF: $input_sarif (contains $count runs)" >&2
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
ROOTFS_SARIF_FILES=()
if [[ "$REPO_ONLY" == "false" ]]; then
  ROOTFS_SARIF_FILES=($(split_sarif_runs "$TRIVY_ROOTFS_SARIF" "snap"))
fi

echo "Final SARIF files:"
for sarif in "${REPO_SARIF_FILES[@]}" "${ROOTFS_SARIF_FILES[@]}"; do
  echo "$sarif"
done

# Obtain CISA Known Exploited Vulnerabilities list.
if ! curl -fsSL --retry 5 --retry-delay 2 -o ./.trivy/kev.json \
  https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json; then
  echo "Error: failed to download CISA KEV catalog after retries" >&2
  exit 1
fi

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
if [[ "$REPO_ONLY" == "false" ]]; then
  get_cisa_kev_cves ./.trivy/kev.json ./.trivy/sarifs/snap.json
fi
trap - EXIT
