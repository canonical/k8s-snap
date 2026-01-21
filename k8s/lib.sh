#!/bin/bash -e

# shellcheck disable=SC2002,SC2030,SC2031

# Configure execution environment, locales and XDG to use paths from SNAP
# Example: 'k8s::common::setup_env'
k8s::common::setup_env() {
  if [ -n "$_K8S_ENV_SETUP_ONCE" ]; then
    return 0
  fi

  local SNAP_CURRENT="${SNAP%"${SNAP_REVISION}"}current"

  # Configure PATH, LD_LIBRARY_PATH
  export PATH="$SNAP_CURRENT/usr/bin:$SNAP_CURRENT/bin:$SNAP_CURRENT/usr/sbin:$SNAP_CURRENT/sbin:$REAL_PATH"
  export LD_LIBRARY_PATH="$SNAP_LIBRARY_PATH:$SNAP_CURRENT/lib:$SNAP_CURRENT/usr/lib:$SNAP_CURRENT/lib/$CRAFT_ARCH_TRIPLET_BUILD_FOR:$SNAP_CURRENT/usr/lib/$CRAFT_ARCH_TRIPLET_BUILD_FOR:$SNAP_CURRENT/usr/lib/$CRAFT_ARCH_TRIPLET_BUILD_FOR/ceph:${REAL_LD_LIBRARY_PATH:-}"

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

# Cleanup systemd overrides
k8s::remove::system_tuning() {
  if ! k8s::common::is_strict; then
    # Find files matching pattern: digits followed by '-k8s.conf'
    files_to_remove=$(find /etc/sysctl.d/ -maxdepth 1 -type f -regextype posix-extended -regex '.*/[0-9]+-k8s\.conf')


    if [ -n "$files_to_remove" ]; then
      echo "$files_to_remove" | xargs sudo rm -f
      sudo sysctl --system
    fi
  fi
}

k8s::remove::resources() {
  if [ -d "$SNAP_COMMON/etc" ]; then
    sudo rm -rf "$SNAP_COMMON/etc"
  fi
}

k8s::remove::kubelet_logs() {
  if ! k8s::common::is_strict; then
    if [ -d "$SNAP_COMMON/var/log/pods" ]; then
        sudo rm -rf "$SNAP_COMMON/var/log/pods"
    fi

    if [ -d "$SNAP_COMMON/var/log/containers" ]; then
      sudo rm -rf "$SNAP_COMMON/var/log/containers"
    fi
  fi
}

# Run a ctr command against the local containerd socket
# Example: 'k8s::cmd::ctr image ls -q'
k8s::cmd::ctr() {
  env \
    CONTAINERD_NAMESPACE="${CONTAINERD_NAMESPACE:-k8s.io}" \
    CONTAINERD_ADDRESS="${CONTAINERD_ADDRESS:-$SNAP_COMMON/run/containerd.sock}" \
    "${SNAP}/bin/ctr" "${@}"
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

# Wait for containerd socket to be ready
# Example: 'k8s::util::wait_containerd_socket'
k8s::util::wait_containerd_socket() {
  while ! k8s::cmd::ctr --connect-timeout 1s > /dev/null; do
    echo Waiting for containerd to start
    sleep 3
  done
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
  ulimit -c unlimited
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

  modprobe $@
}

k8s::containerd::ensure_systemd_defaults() {
  k8s::common::setup_env

  local override_dir="/etc/systemd/system/snap.k8s.containerd.service.d"
  local override_file="$SNAP/k8s/systemd/containerd-defaults.conf"

  if ! [ -f "$override_dir/containerd-defaults.conf" ]; then
    mkdir -p "$override_dir"
    cp "$override_file" "$override_dir/"
  fi
}

# Sanitize feature gates in kube-apiserver arguments
# This removes any feature gates that are not present in the current apiserver version
# from the --feature-gates argument in /var/snap/k8s/common/args/kube-apiserver.
# Usage: k8s::apiserver::sanitize_feature_gates
k8s::apiserver::sanitize_feature_gates() {
  local args_file="/var/snap/k8s/common/args/kube-apiserver"

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
  if [ -x "/snap/k8s/current/bin/kube-apiserver" ]; then
    # Extract feature gate names from help output (format: kube:FeatureName=true|false)
    supported_gates=$(/snap/k8s/current/bin/kube-apiserver --help 2>/dev/null | awk '/^ *kube:/{print $1}' | sed 's/^kube://' | sed 's/=.*//')
  fi

  # If we couldn't get supported gates, return without changes
  if [ -z "$supported_gates" ]; then
    return 0
  fi

  # Convert supported gates to array
  declare -A supported_gates_map
  while IFS= read -r gate; do
    [[ -n "$gate" ]] && supported_gates_map["$gate"]=1
  done <<< "$supported_gates"

  # Get the current feature gates line
  local current_line=$(grep "^--feature-gates=" "$args_file")
  local feature_gates_value="${current_line#--feature-gates=}"

  # Remove surrounding quotes if present
  feature_gates_value="${feature_gates_value%\"}"
  feature_gates_value="${feature_gates_value#\"}"

  local updated_gates=""

  # Split by comma and filter out unsupported gates
  IFS=',' read -ra gates <<< "$feature_gates_value"
  for gate in "${gates[@]}"; do
    local gate_name="${gate%%=*}"

    # Skip unsupported gates
    if [[ -z "${supported_gates_map[$gate_name]}" ]]; then
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
    # Replace the line with updated gates (add quotes back)
    sed -i "s/^--feature-gates=.*$/--feature-gates=\"${updated_gates}\"/" "$args_file"
  else
    # Remove the line if all gates were removed
    sed -i '/^--feature-gates=/d' "$args_file"
  fi
}
