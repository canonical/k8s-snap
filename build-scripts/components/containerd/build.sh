#!/bin/bash
set -euo pipefail

source "${BUILD_SCRIPTS_DIR}/lib.sh"

# Store the current Go snap revision
INITIAL_GO_REVISION=$(snap list go | grep -E '^go\s' | awk '{print $3}')
echo "Current Go snap revision: ${INITIAL_GO_REVISION}"

# Restore the initial Go snap revision on exit, even if the build fails
restore_go() {
  echo "Restoring Go snap to initial revision: ${INITIAL_GO_REVISION}"
  snap revert go --revision="${INITIAL_GO_REVISION}" || true
}
trap restore_go EXIT

maj_min=$(awk '/^go /{print $2}' go.mod | cut -d. -f1,2)
echo "Refreshing to go ${maj_min}-fips/stable channel..."
retry 5 10 snap refresh go --channel="${maj_min}"-fips/stable

INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

VERSION="${2}"
REVISION=$(git rev-parse HEAD)

sed -i "s,^VERSION.*$,VERSION=${VERSION}," Makefile
sed -i "s,^REVISION.*$,REVISION=${REVISION}," Makefile

export GOTOOLCHAIN=local
# Only needed for Go < 1.27
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
	export GOEXPERIMENT=opensslcrypto
fi

semver_lt() {
  python3 -c '
import sys
def parse(s):
    core, _, pre = s.lstrip("v").split("+", 1)[0].partition("-")
    p = (0, tuple((0, int(x)) if x.isdigit() else (1, x) for x in pre.split("."))) if pre else (1, ())
    return (tuple(int(x) for x in core.split(".")), p)
sys.exit(0 if parse(sys.argv[1]) < parse(sys.argv[2]) else 1)
' "${1}" "${2}"
}

# Bump dependencies to patched minimum floors if go.mod has older versions
bumps=()
floors=(
  "github.com/cilium/ebpf:v0.22.0"
  "github.com/go-jose/go-jose/v4:v4.1.6"
  "github.com/klauspost/compress:v1.18.7"
  "github.com/moby/sys/user:v0.4.1"
  "go.opentelemetry.io/otel:v1.47.0"
  "go.opentelemetry.io/otel/sdk:v1.47.0"
  "go.opentelemetry.io/otel/trace:v1.47.0"
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace:v1.47.0"
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc:v1.47.0"
  "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp:v1.47.0"
  "golang.org/x/crypto:v0.59.0"
  # Raised above CVE floors to satisfy golang.org/x/crypto MVS constraints
  "golang.org/x/mod:v0.42.0"
  "golang.org/x/text:v0.44.0"
  "golang.org/x/net:v0.62.0"
  "google.golang.org/grpc:v1.83.2"
)
mod_versions="$(retry 5 5 go list -mod=readonly -m -f '{{.Path}} {{.Version}}' all)"
for entry in "${floors[@]}"; do
  mod="${entry%%:*}"
  floor="${entry##*:}"
  curr="$(awk -v m="${mod}" '$1 == m {print $2; exit}' <<<"${mod_versions}")"
  if [ -n "${curr}" ] && semver_lt "${curr}" "${floor}"; then
    bumps+=("${mod}@${floor}")
  fi
done
if [ ${#bumps[@]} -gt 0 ]; then
  retry 5 5 go get "${bumps[@]}"
  retry 5 5 go mod tidy
  go mod vendor
fi
export CGO_ENABLED=1
export GO_BUILDTAGS="linux cgo ms_tls13kdf"
export SHIM_CGO_ENABLED=1
export SHIM_GO_BUILDTAGS="linux cgo ms_tls13kdf"
for bin in containerd ctr containerd-shim-runc-v2; do
  make "bin/${bin}"
  cp "bin/${bin}" "${INSTALL}/${bin}"
done
