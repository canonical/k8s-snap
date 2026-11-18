#!/bin/bash -e

# shellcheck disable=SC2002,SC2030,SC2031

# Configure execution environment, locales and XDG to use paths from SNAP
# Example: 'k8s::common::setup_env'
k8s::common::setup_env() {
  if [ -z "${BASH_VERSION:-}" ]; then
    echo "k8s::common::setup_env requires bash" >&2
    return 1
  fi

  if [ -n "$_K8S_ENV_SETUP_ONCE" ]; then
    return 0
  fi

  local SNAP_CURRENT=""
  if [ -n "${SNAP:-}" ]; then
    SNAP_CURRENT="${SNAP%"${SNAP_REVISION}"}current"
  fi

  # Configure PATH
  local -a path_dirs=()
  local -A seen_path=()

  if [ -n "$SNAP_CURRENT" ]; then
    local snap_bins=(
      "$SNAP_CURRENT/usr/bin"
      "$SNAP_CURRENT/bin"
      "$SNAP_CURRENT/usr/sbin"
      "$SNAP_CURRENT/sbin"
    )
    for d in "${snap_bins[@]}"; do
      local norm="${d%/}"
      seen_path["$norm"]=1
      path_dirs+=("$norm")
    done

    seen_path["${SNAP%/}/usr/bin"]=1
    seen_path["${SNAP%/}/bin"]=1
    seen_path["${SNAP%/}/usr/sbin"]=1
    seen_path["${SNAP%/}/sbin"]=1
  fi

  local host_path="${REAL_PATH:-${PATH:-}}"
  if [ -n "$host_path" ]; then
    local -a raw_path_dirs=()
    IFS=":" read -r -a raw_path_dirs <<< "$host_path"
    for d in "${raw_path_dirs[@]}"; do
      local norm="${d%/}"
      if [[ "$norm" != /* ]] || [ -n "${seen_path[$norm]:-}" ]; then
        continue
      fi
      seen_path["$norm"]=1
      path_dirs+=("$norm")
    done
  fi

  local formatted_path=""
  for d in "${path_dirs[@]}"; do
    formatted_path="${formatted_path:+${formatted_path}:}$d"
  done
  export PATH="$formatted_path"

  # Configure LD_LIBRARY_PATH
  local -a ld_dirs=()
  local -A seen_ld=()

  if [ -n "${SNAP_LIBRARY_PATH:-}" ]; then
    local -a raw_snap_ld=()
    IFS=":" read -r -a raw_snap_ld <<< "$SNAP_LIBRARY_PATH"
    for d in "${raw_snap_ld[@]}"; do
      local norm="${d%/}"
      if [[ "$norm" != /* ]] || [ -n "${seen_ld[$norm]:-}" ]; then
        continue
      fi
      seen_ld["$norm"]=1
      ld_dirs+=("$norm")
    done
  fi

  if [ -n "$SNAP_CURRENT" ]; then
    local arch_triplet="${CRAFT_ARCH_TRIPLET_BUILD_FOR:-${SNAPCRAFT_ARCH_TRIPLET:-}}"
    local snap_libs=(
      "$SNAP_CURRENT/lib"
      "$SNAP_CURRENT/usr/lib"
    )
    if [ -n "$arch_triplet" ]; then
      snap_libs+=(
        "$SNAP_CURRENT/lib/$arch_triplet"
        "$SNAP_CURRENT/usr/lib/$arch_triplet"
        "$SNAP_CURRENT/usr/lib/$arch_triplet/ceph"
      )
    fi
    for d in "${snap_libs[@]}"; do
      local norm="${d%/}"
      if [ -z "${seen_ld[$norm]:-}" ]; then
        seen_ld["$norm"]=1
        ld_dirs+=("$norm")
      fi
    done

    seen_ld["${SNAP%/}/lib"]=1
    seen_ld["${SNAP%/}/usr/lib"]=1
    if [ -n "$arch_triplet" ]; then
      seen_ld["${SNAP%/}/lib/$arch_triplet"]=1
      seen_ld["${SNAP%/}/usr/lib/$arch_triplet"]=1
      seen_ld["${SNAP%/}/usr/lib/$arch_triplet/ceph"]=1
    fi
  fi

  if [ -n "${REAL_LD_LIBRARY_PATH:-}" ]; then
    local -a raw_real_ld=()
    IFS=":" read -r -a raw_real_ld <<< "$REAL_LD_LIBRARY_PATH"
    for d in "${raw_real_ld[@]}"; do
      local norm="${d%/}"
      if [[ "$norm" != /* ]] || [ -n "${seen_ld[$norm]:-}" ]; then
        continue
      fi
      seen_ld["$norm"]=1
      ld_dirs+=("$norm")
    done
  fi

  local formatted_ld=""
  for d in "${ld_dirs[@]}"; do
    formatted_ld="${formatted_ld:+${formatted_ld}:}$d"
  done
  export LD_LIBRARY_PATH="$formatted_ld"

  # NOTE(neoaggelos/2023-08-14):
  # we cannot list system locales from snap. instead, we attempt
  # well-known locales for Ubuntu/Debian/CentOS and check whether
  # they are available on the system.
  # if they are, set them for the current shell.
  for locale in C.UTF-8 en_US.UTF-8 en_US.utf8; do
    if [ -z "$(export LC_ALL=$locale 2>&1)" ]; then
      export LC_ALL="${LC_ALL:-$locale}"
      export LANG="${LC_ALL:-$locale}"
      break
    fi
  done

  _K8S_ENV_SETUP_ONCE="1"
}

# Check if k8s is installed as a strictly confined snap
# Example: 'k8s::common::is_strict && echo running under strict confinement'
k8s::common::is_strict() {
  k8s::common::setup_env

  if cat "${SNAP}/meta/snap.yaml" | grep -q 'confinement: strict'; then
    return 0
  else
    return 1
  fi
}

k8s::common::resources() {
  mkdir -p "$SNAP_COMMON/etc/"
  cp -r "$SNAP/etc/configurations" "$SNAP_COMMON/etc/"
}

# For backwards compatibility, copy DISA-STIG resources from templates to configurations and create symlink
k8s::common::move_resources() {
  local old_dir="$SNAP_COMMON/etc/templates/disa-stig"
  local new_dir="$SNAP_COMMON/etc/configurations/disa-stig"

  # Only migrate if new directory does not exist and old directory exists
  if [ ! -e "$new_dir" ] && [ -e "$old_dir" ]; then
    cp -r "$old_dir" "$new_dir"
    rm -rf "$old_dir"
    mkdir -p "$(dirname "$old_dir")"
    ln -s "$new_dir" "$old_dir"
  fi
}

# Check if FIPS is enabled on the system
# Returns 0 (success) if FIPS is enabled, 1 (failure) otherwise
# Example: 'k8s::common::on_fips_host && echo "FIPS is enabled"'
k8s::common::on_fips_host() {
  if [ -f "/proc/sys/crypto/fips_enabled" ] &&
     [ "$(cat /proc/sys/crypto/fips_enabled 2>/dev/null)" = "1" ]; then
    return 0
  fi
  return 1
}

# Cleanup configuration left by the network feature
k8s::remove::network() {
  k8s::common::setup_env

  k8s::cmd::k8s x-cleanup network || true
}

# [DANGER] Cleanup containers and runtime state. Note that the order of operations below is crucial.
k8s::remove::containers() {
  k8s::common::setup_env

  k8s::cmd::k8s x-cleanup containers || true
}

k8s::remove::containerd() {
  k8s::common::setup_env

  k8s::cmd::k8s x-cleanup containerd || true
}

# Cleanup system tuning sysctl overrides
# Example: 'k8s::remove::system_tuning'
k8s::remove::system_tuning() {
  if ! k8s::common::is_strict; then
    local files_to_remove=()
    if [ -d /etc/sysctl.d ]; then
      mapfile -d '' files_to_remove < <(find /etc/sysctl.d/ -maxdepth 1 -type f -regextype posix-extended -regex '.*/[0-9]+-k8s\.conf' -print0 2>/dev/null)
    fi

    if [ "${#files_to_remove[@]}" -gt 0 ]; then
      k8s::util::run_as_root rm -f "${files_to_remove[@]}"
      k8s::util::run_as_root sysctl --system
    fi
  fi
}

# Cleanup resources in SNAP_COMMON
# Example: 'k8s::remove::resources'
k8s::remove::resources() {
  local common_dir="${SNAP_COMMON:-/var/snap/k8s/common}"
  if [ -d "$common_dir/etc" ]; then
    k8s::util::run_as_root rm -rf "$common_dir/etc"
  fi
}

# Cleanup kubelet pod and container logs
# Example: 'k8s::remove::kubelet_logs'
k8s::remove::kubelet_logs() {
  if ! k8s::common::is_strict; then
    local common_dir="${SNAP_COMMON:-/var/snap/k8s/common}"
    if [ -d "$common_dir/var/log/pods" ]; then
      k8s::util::run_as_root rm -rf "$common_dir/var/log/pods"
    fi

    if [ -d "$common_dir/var/log/containers" ]; then
      k8s::util::run_as_root rm -rf "$common_dir/var/log/containers"
    fi
  fi
}

# Run kubectl as admin
# Example: 'k8s::cmd::kubectl get pod,node -A'
k8s::cmd::kubectl() {
  env KUBECONFIG="${KUBECONFIG:-/etc/kubernetes/admin.conf}" "${SNAP}/bin/kubectl" "${@}"
}

# Run k8s CLI
# Example: 'k8s::cmd::k8s status'
k8s::cmd::k8s() {
  k8s::common::setup_env

  "${SNAP}/bin/k8s" "${@}"
}

# Get the local node hostname, in lowercase
# Example: 'hostname="$(k8s::cmd::hostname)"'
k8s::cmd::hostname() {
  k8s::common::setup_env

  hostname | tr '[:upper:]' '[:lower:]'
}

k8s::util::default_interface() {
  k8s::common::setup_env

  ip route show default | awk '{for(i=1; i<NF; i++) if ($i=="dev") print $(i+1)}' | head -1
}

# Wait for API server to be ready
# Example: 'k8s::util::wait_kube_apiserver'
k8s::util::wait_kube_apiserver() {
  while ! k8s::cmd::kubectl --kubeconfig /etc/kubernetes/kubelet.conf get --raw /readyz >/dev/null; do
    echo Waiting for kube-apiserver to start
    sleep 3
  done
}

# k8s::common::validate_env
#
# Validates a list of environment variables by temporarily exporting them
# with a `K8SD_` prefix and invoking environment setup and validation logic.
# This ensures the user's environment remains untouched after validation.
#
# Exits with status 1 if FIPS mode is requested but the host is not FIPS-enabled.
k8s::common::validate_env() {
  local env_vars=("$@")

  # Export the variables with a K8SD_ prefix
  local env_var key value
  for env_var in "${env_vars[@]}"; do
    key="${env_var%%=*}"
    value="${env_var#*=}"
    export "K8SD_${key}=${value}"
  done

  # Check if FIPS mode is requested (from prefixed vars or fallback to env)
  if { [ "${K8SD_GOFIPS:-${GOFIPS:-}}" = "1" ] || [[ "${K8SD_GODEBUG:-${GODEBUG:-}}" == *"fips140=on"* ]]; } && \
      ! k8s::common::on_fips_host; then
    echo "FIPS mode is requested (GOFIPS=1 or GODEBUG contains fips140=on) but the host system is not FIPS-enabled."
    echo "Please run this service on a FIPS-enabled host to use FIPS mode."
    exit 1
  fi

  # Clean up the K8SD_ variables
  for env_var in "${env_vars[@]}"; do
    key="${env_var%%=*}"
    unset "K8SD_${key}" 2>/dev/null || true
  done
}

# Execute a "$SNAP/bin/$service" with arguments from "$SNAP_DATA/args/$service" and optional additional args
# Environment variables are loaded from "$SNAP_DATA/args/$service-env" and "$SNAP_DATA/args/snap-env"
# Example: 'k8s::common::execute kubelet' or 'k8s::common::execute k8s status'
k8s::common::execute() {
  local service_name="$1"
  shift  # Remove the first argument (service_name), leaving any additional args

  k8s::common::setup_env

  declare -a args=()

  if [[ -f "${SNAP_COMMON}/args/${service_name}" ]]; then
    declare -a file_args="($(cat "${SNAP_COMMON}/args/${service_name}"))"
    args+=("${file_args[@]}")
  fi

  args+=("$@")

  declare -a all_env_vars=()

  # Load global environment variables if the file exists
  if [[ -f "${SNAP_COMMON}/args/snap-env" ]]; then
    mapfile -t global_vars < <(grep -vE '^[[:space:]]*($|#)' "${SNAP_COMMON}/args/snap-env")
    all_env_vars+=("${global_vars[@]}")
  fi

  # Load service-specific environment variables if the file exists
  # Service-specific environment variables take precedence over global environment variables.
  if [[ -f "${SNAP_COMMON}/args/${service_name}-env" ]]; then
    mapfile -t service_vars < <(grep -vE '^[[:space:]]*($|#)' "${SNAP_COMMON}/args/${service_name}-env")
    all_env_vars+=("${service_vars[@]}")
  fi

  # Validate environment with all variables
  k8s::common::validate_env "${all_env_vars[@]}"

  set -e
  ulimit -c unlimited 2>/dev/null || true
  export GOTRACEBACK="crash"

  if (( ${#all_env_vars[@]} > 0 )); then
    exec env -S "${all_env_vars[@]}" "${SNAP}/bin/${service_name}" "${args[@]}"
  else
    exec "${SNAP}/bin/${service_name}" "${args[@]}"
  fi
}

# Initialize a single-node k8sd cluster
k8s::init::k8sd() {
  k8s::common::setup_env

  mkdir -m 0700 -p "$SNAP_COMMON/args"
  cp "$SNAP/k8s/args/k8sd" "$SNAP_COMMON/args/k8sd"
}

# Ensure /var/lib/kubelet is a shared mount
# Example: 'k8s::common::is_strict && k8s::kubelet::ensure_shared_root_dir'
k8s::kubelet::ensure_shared_root_dir() {
  k8s::common::setup_env

  if ! findmnt -o PROPAGATION /var/lib/kubelet -n | grep -q shared; then
    echo "Ensure /var/lib/kubelet mount propagation is rshared"
    mount -o remount --make-rshared "$SNAP_COMMON/var/lib/kubelet" /var/lib/kubelet
  fi
}

# Loads the kernel module names given as arguments
# Example: 'k8s::util::load_kernel_modules mod1 mod2 mod3'
k8s::util::load_kernel_modules() {
  k8s::common::setup_env

  modprobe -- "$@"
}

# Execute command directly when running as root, otherwise with sudo
# Example: 'k8s::util::run_as_root rm -f /path/to/file'
k8s::util::run_as_root() {
  if [ "$(id -u)" -eq 0 ]; then
    "$@"
  else
    sudo "$@"
  fi
}

# Ensure containerd systemd drop-in configuration defaults are applied
# Example: 'k8s::containerd::ensure_systemd_defaults'
k8s::containerd::ensure_systemd_defaults() {
  k8s::common::setup_env

  local override_dir="/etc/systemd/system/snap.${SNAP_INSTANCE_NAME:-k8s}.containerd.service.d"
  local override_file="${SNAP:-/snap/k8s/current}/k8s/systemd/containerd-defaults.conf"

  if [ ! -f "$override_file" ]; then
    return 0
  fi

  if [ ! -f "$override_dir/containerd-defaults.conf" ] || ! cmp -s "$override_file" "$override_dir/containerd-defaults.conf"; then
    mkdir -p "$override_dir"
    cp "$override_file" "$override_dir/"
    systemctl daemon-reload || true
  fi
}

# Sanitize feature gates in kube-apiserver arguments
# Example: 'k8s::apiserver::sanitize_feature_gates'
k8s::apiserver::sanitize_feature_gates() {
  local args_file="${SNAP_COMMON:-/var/snap/k8s/common}/args/kube-apiserver"
  local apiserver_bin="${SNAP:-/snap/k8s/current}/bin/kube-apiserver"

  # Check if the args file exists
  if [ ! -f "$args_file" ]; then
    return 0
  fi

  # Return early if no feature gates are configured
  if ! grep -q "^--feature-gates=" "$args_file"; then
    return 0
  fi

  # Get the list of supported feature gates from kube-apiserver
  local supported_gates=""
  if [ -x "$apiserver_bin" ]; then
    # Extract feature gate names from help output (format: kube:FeatureName=true|false)
    supported_gates=$("$apiserver_bin" --help 2>/dev/null | awk '/^ *kube:/{print $1}' | sed 's/^kube://' | sed 's/=.*//')
  fi

  # If we couldn't get supported gates, return without changes
  if [ -z "$supported_gates" ]; then
    return 0
  fi

  # Convert supported gates to map
  declare -A supported_gates_map
  while IFS= read -r gate; do
    [[ -n "$gate" ]] && supported_gates_map["$gate"]=1
  done <<< "$supported_gates"

  # Get the current feature gates line
  local current_line
  current_line="$(grep -m 1 "^--feature-gates=" "$args_file" || true)"
  local feature_gates_value="${current_line#--feature-gates=}"

  # Remove surrounding quotes if present
  if [[ "$feature_gates_value" == \"*\" ]]; then
    feature_gates_value="${feature_gates_value#\"}"
    feature_gates_value="${feature_gates_value%\"}"
  elif [[ "$feature_gates_value" == \'*\' ]]; then
    feature_gates_value="${feature_gates_value#\'}"
    feature_gates_value="${feature_gates_value%\'}"
  fi

  local updated_gates=""

  # Split by comma and filter out unsupported gates
  IFS=',' read -ra gates <<< "$feature_gates_value"
  for gate in "${gates[@]}"; do
    # Strip leading and trailing whitespace
    gate="${gate#"${gate%%[![:space:]]*}"}"
    gate="${gate%"${gate##*[![:space:]]}"}"
    if [ -z "$gate" ]; then
      continue
    fi

    local gate_name="${gate%%=*}"

    # Skip unsupported gates
    if [[ -z "${supported_gates_map[$gate_name]:-}" ]]; then
      continue
    fi

    # Add the gate to the updated list
    if [ -n "$updated_gates" ]; then
      updated_gates="${updated_gates},${gate}"
    else
      updated_gates="${gate}"
    fi
  done

  # Update the file in-place
  if [ -n "$updated_gates" ]; then
    # Escape special characters for sed replacement (&, \, delimiter |)
    local escaped_gates="${updated_gates//\\/\\\\}"
    escaped_gates="${escaped_gates//|/\\|}"
    escaped_gates="${escaped_gates//&/\\&}"
    sed -i "s|^--feature-gates=.*$|--feature-gates=\"${escaped_gates}\"|" "$args_file"
  else
    # Remove the line if all gates were removed
    sed -i '/^--feature-gates=/d' "$args_file"
  fi
}
