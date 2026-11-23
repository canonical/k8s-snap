#!/bin/bash
set -euo pipefail

source "${BUILD_SCRIPTS_DIR}/lib.sh"

VERSION="${2}"
INSTALL="${1}/bin"
mkdir -p "${INSTALL}"

# Retain clean upstream version commit before modifying dependencies.
COMMIT="$(git describe --long --always "${VERSION}")"

# Static binary cannot dlopen libcrypto via OpenSSL backend.
export GOTOOLCHAIN=local
export MS_GO_NOSYSTEMCRYPTO=1
if [[ "$(go env GOVERSION)" < "go1.27" ]]; then
  export GOEXPERIMENT=nosystemcrypto
fi

semver_lt() {
  python3 -c '
import sys

def parse(v):
    v = v.lstrip("v").split("+")[0]
    parts = v.split("-", 1)
    core = tuple(int(x) for x in parts[0].split("."))
    if len(parts) == 1:
        return (core, 1, ())
    prerelease = []
    for id_ in parts[1].split("."):
        prerelease.append((0, int(id_)) if id_.isdigit() else (1, id_))
    return (core, 0, tuple(prerelease))

sys.exit(0 if parse(sys.argv[1]) < parse(sys.argv[2]) else 1)
' "$1" "$2"
}

# Bump dependencies that do not meet patched security floors.
bumps=()
mod_versions="$(retry 5 5 go list -mod=readonly -m -f '{{.Path}} {{.Version}}' all)"
for target in \
  "golang.org/x/net v0.62.0" \
  "github.com/moby/sys/user v0.4.1" \
  "github.com/cilium/ebpf v0.22.0"; do
  mod="${target%% *}"
  floor="${target##* }"
  cur="$(awk -v m="${mod}" '$1 == m {print $2; exit}' <<<"${mod_versions}")"
  if [[ -n "${cur}" ]] && semver_lt "${cur}" "${floor}"; then
    bumps+=("${mod}@${floor}")
  fi
done

if [[ ${#bumps[@]} -gt 0 ]]; then
  retry 5 5 go get "${bumps[@]}"
  retry 5 5 go mod tidy
  go mod vendor
fi

make COMMIT="${COMMIT}" BUILDTAGS="seccomp apparmor" EXTRA_LDFLAGS="-s -w" static
cp runc "${INSTALL}/runc"
