#!/bin/bash
set -euo pipefail

source "${BUILD_SCRIPTS_DIR}/lib.sh"

VERSION="${2}"
INSTALL="${1}/bin"

mkdir -p "${INSTALL}"

export GOTOOLCHAIN=local
export CGO_ENABLED=1
# Only needed for Go < 1.27
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
	export GOEXPERIMENT=opensslcrypto
fi
export GOFLAGS="-tags=linux,cgo,ms_tls13kdf"

semver_lt() {
	python3 -c '
import sys

def parse(v):
    v = v.strip().lstrip("v").split("+", 1)[0]
    core, pre_flag, pre_ids = (v.split("-", 1)[0], 0, tuple((0, int(x)) if x.isdigit() else (1, x) for x in v.split("-", 1)[1].split("."))) if "-" in v else (v, 1, ())
    parts = [int(x) for x in core.split(".")]
    while len(parts) < 3:
        parts.append(0)
    return (parts[0], parts[1], parts[2], pre_flag, pre_ids)

sys.exit(0 if parse(sys.argv[1]) < parse(sys.argv[2]) else 1)
' "${1}" "${2}"
}

# Bump dependencies when below minimum security floors
floors=(
	"go.opentelemetry.io/otel v1.47.0"
	"go.opentelemetry.io/otel/sdk v1.47.0"
	"go.opentelemetry.io/otel/trace v1.47.0"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.70.0"
	"golang.org/x/crypto v0.59.0"
)

mod_versions="$(cd server && retry 5 5 go list -m -f '{{.Path}} {{.Version}}' all)"

bumps=()
for item in "${floors[@]}"; do
	mod="${item%% *}"
	floor="${item##* }"
	ver="$(awk -v m="${mod}" '$1 == m {print $2; exit}' <<<"${mod_versions}")"
	if [[ -n "${ver}" ]] && semver_lt "${ver}" "${floor}"; then
		bumps+=("${mod}@${floor}")
	fi
done

if (( ${#bumps[@]} > 0 )); then
	(
		cd server
		retry 5 5 go get "${bumps[@]}"
		retry 5 5 go mod tidy
	)
	retry 5 5 go mod tidy
	(cd etcdutl && retry 5 5 go mod tidy)
	(cd etcdctl && retry 5 5 go mod tidy)
fi

make build
cp bin/etcd "${INSTALL}/etcd"
