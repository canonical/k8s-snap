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
  export LD_LIBRARY_PATH="$SNAP_LIBRARY_PATH:$SNAP_CURRENT/lib:$SNAP_CURRENT/usr/lib:$SNAP_CURRENT/lib/$SNAPCRAFT_ARCH_TRIPLET:$SNAP_CURRENT/usr/lib/$SNAPCRAFT_ARCH_TRIPLET:$SNAP_CURRENT/usr/lib/$SNAPCRAFT_ARCH_TRIPLET/ceph:${REAL_LD_LIBRARY_PATH:-}"

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

# Run a dqlite CLI command against the k8s-dqlite cluster
# Example: 'k8s::cmd::dqlite k8s .help'
k8s::cmd::dqlite() {
  k8s::common::setup_env

  K8S_DQLITE_DIR="${SNAP_COMMON}/var/lib/k8s-dqlite"
  "${SNAP}/bin/dqlite" -s "file://${K8S_DQLITE_DIR}/cluster.yaml" -c "${K8S_DQLITE_DIR}/cluster.crt" -k "${K8S_DQLITE_DIR}/cluster.key" "${@}"
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

# Execute a "$SNAP/bin/$service" with arguments from "$SNAP_DATA/args/$service"
# Example: 'k8s::common::execute_service kubelet'
k8s::common::execute_service() {
  service_name="$1"

  k8s::common::setup_env

  # Source arguments and substitute environment variables. Will fail if we cannot read the file.
  declare -a args="($(cat "${SNAP_COMMON}/args/${service_name}"))"

  set -xe
  ulimit -c unlimited
  export GOTRACEBACK="crash"
  if [[ -f "${SNAP_COMMON}/args/${service_name}-env" ]]; then
    mapfile -t env_vars < "${SNAP_COMMON}/args/${service_name}-env"
    exec env -S "${env_vars[@]}" "${SNAP}/bin/${service_name}" "${args[@]}"
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
