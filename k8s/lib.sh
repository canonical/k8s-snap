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

  "${SNAP}/bin/kube-proxy" --cleanup || true

  k8s::cmd::k8s x-cleanup network || true
}

# [DANGER] Cleanup containers and runtime state. Note that the order of operations below is crucial.
k8s::remove::containers() {
  k8s::common::setup_env

  # kill all container shims and pause processes
  k8s::cmd::k8s x-print-shim-pids | xargs -r -t kill -SIGKILL || true

  # delete cni network namespaces
  ip netns list | cut -f1 -d' ' | grep -- "^cni-" | xargs -n1 -r -t ip netns delete || true

  # The PVC loopback devices use container paths, making them tricky to identify.
  # We'll rely on the volume mount paths (/var/lib/kubelet/*).
  local LOOP_DEVICES=`cat /proc/mounts | grep /var/lib/kubelet/pods | grep /dev/loop | cut -d " " -f 1`

  # unmount Pod NFS volumes forcefully, as unmounting them normally may hang otherwise.
  cat /proc/mounts | grep /run/containerd/io.containerd. | grep "nfs[34]" | cut -f2 -d' ' | xargs -r -t umount -f || true
  cat /proc/mounts | grep /var/lib/kubelet/pods | grep "nfs[34]" | cut -f2 -d' ' | xargs -r -t umount -f || true

  # unmount Pod volumes gracefully.
  cat /proc/mounts | grep /run/containerd/io.containerd. | cut -f2 -d' ' | xargs -r -t umount || true
  cat /proc/mounts | grep /var/lib/kubelet/pods | cut -f2 -d' ' | xargs -r -t umount || true

  # unmount lingering Pod volumes by force, to prevent potential volume leaks.
  cat /proc/mounts | grep /run/containerd/io.containerd. | cut -f2 -d' ' | xargs -r -t umount -f || true
  cat /proc/mounts | grep /var/lib/kubelet/pods | cut -f2 -d' ' | xargs -r -t umount -f || true

  # unmount various volumes exposed by CSI plugin drivers.
  cat /proc/mounts | grep /var/lib/kubelet/plugins | cut -f2 -d' ' | xargs -r -t umount -f || true

  # remove kubelet plugin sockets, as we don't have the containers associated with them anymore,
  # so kubelet won't try to access inexistent plugins on reinstallation.
  find /var/lib/kubelet/plugins/ -name "*.sock" | xargs rm -f || true
  rm /var/lib/kubelet/plugins_registry/*.sock || true

  cat /proc/mounts | grep /var/snap/k8s/common/var/lib/containerd/ | cut -f2 -d' ' | xargs -r -t umount  || true

  # cleanup loopback devices
  for dev in $LOOP_DEVICES; do
    losetup -d $dev
  done
}

k8s::remove::containerd() {
  k8s::common::setup_env

  # only remove containerd if the snap was already bootstrapped.
  # this is to prevent removing containerd when it is not installed by the snap.
  if [ -f "$SNAP_COMMON/lock/containerd-socket-path" ]; then
     rm -f $(cat "$SNAP_COMMON/lock/containerd-socket-path")
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
  exec "${SNAP}/bin/${service_name}" "${args[@]}"
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
