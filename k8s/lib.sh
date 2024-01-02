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

# Run an openssl command
# Example: 'k8s::cmd::openssl genrsa 2048'
k8s::cmd::openssl() {
  k8s::common::setup_env

  env \
    OPENSSL_CONF="${SNAP}/etc/ssl/openssl.cnf" \
    "${SNAP}/usr/bin/openssl" "${@}"
}

# Run a dqlite command against the local dqlite instance
# Example: 'k8s::cmd::dqlite k8s .help'
k8s::cmd::dqlite() {
  k8s::common::setup_env

  "${SNAP}/bin/dqlite" \
    --cert /var/lib/k8s-dqlite/cluster.crt \
    --key /var/lib/k8s-dqlite/cluster.key \
    --servers file:///var/lib/k8s-dqlite/cluster.yaml \
    "${@}"
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

# Run snapctl
# Example: 'k8s::cmd::snapctl start kube-apiserver'
k8s::cmd::snapctl() {
  snapctl "${@}"
}

# Get the local node hostname, in lowercase
# Example: 'hostname="$(k8s::cmd::hostname)"'
k8s::cmd::hostname() {
  k8s::common::setup_env

  hostname | tr '[:upper:]' '[:lower:]'
}

# Get the default host IP
# Example: 'default_ip="$(k8s::util::default_ip)"'
k8s::util::default_ip() {
  k8s::common::setup_env

  local default_interface
  local ip_addr_cidr
  local ip_addr

  # default_interface="eth0"
  # ip_addr_cidr="10.0.1.83/24"
  # ip_addr="10.0.1.83"
  default_interface="$(ip route show default | awk '{for(i=1; i<NF; i++) if ($i=="dev") print $(i+1)}' | head -1)"
  ip_addr_cidr="$(ip -o -4 addr list "${default_interface}" | awk '{print $4}')"
  ip_addr="${ip_addr_cidr%/*}"

  if [ -z "$ip_addr" ]; then
    ip_addr="$(ip route get 255.255.255.255 | awk '{for(i=1; i<NF; i++) if ($i=="src") print $(i+1)})' | head -1)"
  fi
  if [ -z "$ip_addr" ]; then
    ip_addr="127.0.0.1"
  fi

  echo "$ip_addr"
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

# Generate a new RSA key
# Example: 'k8s::util::pki::generate_key /etc/kubernetes/pki/ca.key'
k8s::util::pki::generate_key() {
  if [ ! -f "$1" ]; then
    k8s::cmd::openssl genrsa -out "$1" 2048
    chown 0:0 "$1" || true
    chmod 0600 "$1" || true
  fi
}

# Generate a CSR and private key given the subject
# Example: 'k8s::util::pki::generate_csr "/CN=system:node:$hostname/O=system:nodes" /etc/kubernetes/pki/kubelet.key -addext "subjectAltName = DNS:dns1, IP:ip1, IP:ip2"'
k8s::util::pki::generate_csr() {
  subject="$1"
  key_file="$2"
  shift
  shift

  k8s::util::pki::generate_key "$key_file"
  k8s::cmd::openssl req -new -sha256 -subj "$subject" -key "$key_file" "${@}"
}

# Sign a CSR using the CA of the local node
# Example: 'cat component.csr | k8s::util::pki::sign_cert > component.crt'
k8s::util::pki::sign_cert() {
  k8s::common::setup_env

  csr="$(cat)"

  # Parse SANs from the CSR and add them to the certificate extensions (if any)
  extensions=""
  alt_names="$(echo "$csr" | k8s::cmd::openssl req -text | grep "X509v3 Subject Alternative Name:" -A1 | tail -n 1 | sed 's,IP Address:,IP:,g')"
  if test "x$alt_names" != "x"; then
    extensions="subjectAltName = $alt_names"
  fi

  # Sign certificate and print to stdout
  echo "$csr" | k8s::cmd::openssl x509 -req -sha256 -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -days 3650 -extfile <(echo "${extensions}") "${@}"
}

# Execute a "$SNAP/bin/$service" with arguments from "$SNAP_DATA/args/$service"
# Example: 'k8s::common::execute_service kubelet'
k8s::common::execute_service() {
  service_name="$1"

  k8s::common::setup_env

  # Source arguments and substitute environment variables. Will fail if we cannot read the file.
  declare -a args="($(cat "${SNAP_DATA}/args/${service_name}"))"

  set -xe
  exec "${SNAP}/bin/${service_name}" "${args[@]}"
}
