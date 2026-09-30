#!/bin/bash

# Store the current Go snap revision
INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
echo "Current Go snap revision: ${INITIAL_GO_REVISION}"

maj_min=$(awk '/^go /{print $2}' go.mod | cut -d. -f1,2)
echo "Refreshing to go ${maj_min}-fips/stable channel..."
# Retry on transient snap-store failures (e.g. 429 rate limiting) instead of
# failing the whole build outright.
for i in $(seq 1 5); do
  if snap refresh go --channel="${maj_min}"-fips/stable; then
    break
  elif [ "${i}" -eq 5 ]; then
    echo "Failed to refresh go snap after 5 attempts"
    exit 1
  else
    echo "Retrying go snap refresh (${i}/5) in 10s..."
    sleep 10
  fi
done

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

VERSION="${2}"
REVISION=$(git rev-parse HEAD)

sed -i "s,^VERSION.*$,VERSION=${VERSION}," Makefile
sed -i "s,^REVISION.*$,REVISION=${REVISION}," Makefile

export GOTOOLCHAIN=local
export GOEXPERIMENT=opensslcrypto
export CGO_ENABLED=1
export GO_BUILDTAGS="linux cgo ms_tls13kdf"
export SHIM_CGO_ENABLED=1
export SHIM_GO_BUILDTAGS="linux cgo ms_tls13kdf"
for bin in containerd ctr containerd-shim-runc-v2; do
  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done

# Restore the initial Go snap revision
echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"

# Snap revert fails if the revision is already the current one, so we ignore errors
snap revert go --revision="${INITIAL_GO_REVISION}" || true
