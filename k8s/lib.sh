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

# Cleanup configuration left by the network component
#   - Iptables Rules
#   - Network Interfaces
#   - Traffic Control(tc) rules
# https://github.com/cilium/cilium/blob/7318ce2d0d89a91227e3f313adebce892f3c388e/cilium-dbg/cmd/cleanup.go#L132-L139
k8s::remove::network() {
  k8s::common::setup_env

  local default_interface

  for link in cilium_vxlan cilium_host cilium_net
  do
    ip link delete ${link} || true
  done

  iptables-save | grep -iv cilium | iptables-restore
  ip6tables-save | grep -iv cilium | ip6tables-restore
  iptables-legacy-save | grep -iv cilium | iptables-legacy-restore
  ip6tables-legacy-save | grep -iv cilium | ip6tables-legacy-restore

  default_interface="$(k8s::util::default_interface)"

  for d in ingress egress
  do
    tc filter del dev $default_interface ${d} || true
  done
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

k8s::util::default_interface() {
  k8s::common::setup_env

  ip route show default | awk '{for(i=1; i<NF; i++) if ($i=="dev") print $(i+1)}' | head -1
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
  default_interface="$(k8s::util::default_interface)"
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

# Initialize a single-node k8s-dqlite
k8s::init::k8s_dqlite() {
  k8s::common::setup_env

  k8s::cmd::openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes \
    -keyout /var/lib/k8s-dqlite/cluster.key -out /var/lib/k8s-dqlite/cluster.crt \
    -config "$SNAP/k8s/csr/csr.conf" -extensions v3_ext \
    -subj /CN=k8s \
    -addext "subjectAltName = DNS:k8s-dqlite, IP:127.0.0.1"
  echo 'Address: "127.0.0.1:2380"' > /var/lib/k8s-dqlite/init.yaml

  mkdir -p "$SNAP_DATA/args"
  cp "$SNAP/k8s/args/k8s-dqlite" "$SNAP_DATA/args/k8s-dqlite"
}

# Initialize a single-node k8sd cluster
k8s::init::k8sd() {
  k8s::common::setup_env

  mkdir -p "$SNAP_DATA/args"
  cp "$SNAP/k8s/args/k8sd" "$SNAP_DATA/args/k8sd"
}

# Initialize containerd for the local node
k8s::init::containerd() {
  k8s::common::setup_env

  mkdir -p "$SNAP_DATA/args"
  cp "$SNAP/k8s/args/containerd" "$SNAP_DATA/args/containerd"
  cp "$SNAP/k8s/config/containerd/config.toml" /etc/containerd/config.toml
  cp "$SNAP/opt/cni/bin/"* /opt/cni/bin/
}

# Initialize Kubernetes PKI CA (self-signed)
# Example: 'k8s::init::ca'
k8s::init::ca() {
  k8s::common::setup_env

  mkdir -p /etc/kubernetes/pki

  for key in serviceaccount ca front-proxy-ca; do
    k8s::util::pki::generate_key "/etc/kubernetes/pki/${key}.key"
  done

  # Generate Kubernetes CA
  k8s::cmd::openssl req -x509 -new -sha256 -nodes -days 3650 -key /etc/kubernetes/pki/ca.key -subj "/CN=kubernetes-ca" -out /etc/kubernetes/pki/ca.crt
  # Generate Front Proxy CA
  k8s::cmd::openssl req -x509 -new -sha256 -nodes -days 3650 -key /etc/kubernetes/pki/front-proxy-ca.key -subj "/CN=kubernetes-front-proxy-ca" -out /etc/kubernetes/pki/front-proxy-ca.crt
}

# Initialize Kuberentes server and client certificates, using our own self-signed CA.
# Example: 'k8s::init::pki'
k8s::init::pki() {
  k8s::common::setup_env

  # Generate kube-apiserver certificate
  # TODO(neoaggelos): add IP addresses of machine, add extra SANs from user configuration
  k8s::util::pki::generate_csr "/CN=kube-apiserver" /etc/kubernetes/pki/apiserver.key -addext "$(echo "subjectAltName =
    DNS: localhost,
    DNS: kubernetes,
    DNS: kubernetes.default,
    DNS: kubernetes.default.svc,
    DNS: kubernetes.default.svc.cluster,
    DNS: $(k8s::cmd::hostname),

    IP: 127.0.0.1,
    IP: 10.152.183.1,
    IP: $(k8s::util::default_ip)
  " | tr '\n' ' ')" | k8s::util::pki::sign_cert > /etc/kubernetes/pki/apiserver.crt

  # Generate front-proxy-client certificate (signed by front-proxy-ca)
  k8s::util::pki::generate_csr /CN=front-proxy-client /etc/kubernetes/pki/front-proxy-client.key -config "$SNAP/k8s/csr/csr.conf" |
    k8s::util::pki::sign_cert -extensions v3_ext -extfile "$SNAP/k8s/csr/csr.conf" -CA "/etc/kubernetes/pki/front-proxy-ca.crt" -CAkey "/etc/kubernetes/pki/front-proxy-ca.key" \
      > /etc/kubernetes/pki/front-proxy-client.crt

  # Generate kubelet certificates
  # TODO(neoaggelos): add IP addresses of machine
  k8s::util::pki::generate_csr "/CN=system:node:$(k8s::cmd::hostname)/O=system:nodes" /etc/kubernetes/pki/kubelet.key -addext "$(echo "subjectAltName =
    DNS: $(k8s::cmd::hostname),
    IP: 127.0.0.1,
    IP: $(k8s::util::default_ip)
  " | tr '\n' ' ')" | k8s::util::pki::sign_cert > /etc/kubernetes/pki/kubelet.crt

  # Generate the rest of the client certificates
  k8s::util::pki::generate_csr /CN=kubernetes-admin/O=system:masters /etc/kubernetes/pki/admin.key | k8s::util::pki::sign_cert > /etc/kubernetes/pki/admin.crt
  k8s::util::pki::generate_csr /CN=system:kube-proxy /etc/kubernetes/pki/proxy.key | k8s::util::pki::sign_cert > /etc/kubernetes/pki/proxy.crt
  k8s::util::pki::generate_csr /CN=system:kube-scheduler /etc/kubernetes/pki/scheduler.key | k8s::util::pki::sign_cert > /etc/kubernetes/pki/scheduler.crt
  k8s::util::pki::generate_csr /CN=system:kube-controller-manager /etc/kubernetes/pki/controller-manager.key | k8s::util::pki::sign_cert > /etc/kubernetes/pki/controller-manager.crt
  k8s::util::pki::generate_csr /CN=kube-apiserver-kubelet-client/O=system:masters /etc/kubernetes/pki/apiserver-kubelet-client.key | k8s::util::pki::sign_cert > /etc/kubernetes/pki/apiserver-kubelet-client.crt
}

k8s::init::kubeconfigs() {
  k8s::util::generate_x509_kubeconfig /etc/kubernetes/pki/admin.crt /etc/kubernetes/pki/admin.key /etc/kubernetes/pki/ca.crt > /etc/kubernetes/admin.conf
  k8s::util::generate_x509_kubeconfig /etc/kubernetes/pki/kubelet.crt /etc/kubernetes/pki/kubelet.key /etc/kubernetes/pki/ca.crt > /etc/kubernetes/kubelet.conf
  k8s::util::generate_x509_kubeconfig /etc/kubernetes/pki/proxy.crt /etc/kubernetes/pki/proxy.key /etc/kubernetes/pki/ca.crt > /etc/kubernetes/proxy.conf
  k8s::util::generate_x509_kubeconfig /etc/kubernetes/pki/controller-manager.crt /etc/kubernetes/pki/controller-manager.key /etc/kubernetes/pki/ca.crt > /etc/kubernetes/controller-manager.conf
  k8s::util::generate_x509_kubeconfig /etc/kubernetes/pki/scheduler.crt /etc/kubernetes/pki/scheduler.key /etc/kubernetes/pki/ca.crt > /etc/kubernetes/scheduler.conf
}

# Generate a kubeconfig file that uses x509 certificates for authentication.
# Example: 'k8s::util::generate_x509_kubeconfig /etc/kubernetes/pki/admin.crt /etc/kubernetes/pki/admin.key /etc/kubernetes/pki/ca.crt 127.0.0.1 6443 > /etc/kubernetes/admin.conf'
k8s::util::generate_x509_kubeconfig() {
  k8s::common::setup_env

  cert_data="$(base64 -w 0 < "$1")"
  key_data="$(base64 -w 0 < "$2")"
  ca_data="$(base64 -w 0 < "$3")"

  # optional arguments (apiserver IP and port)
  apiserver="${4:-127.0.0.1}"
  apiserver_port="$(cat "$SNAP_DATA/args/kube-apiserver" | grep -- --secure-port | tr '=' ' ' | cut -f2 -d' ')"
  port="${5:-$apiserver_port}"

  cat "$SNAP/k8s/config/kubeconfig" |
    sed 's/CADATA/'"${ca_data}"'/g' |
    sed 's/CERTDATA/'"${cert_data}"'/g' |
    sed 's/KEYDATA/'"${key_data}"'/g' |
    sed 's/APISERVER/'"${apiserver}"'/g' |
    sed 's/PORT/'"${port}"'/g'
}

# Configure default arguments for Kubernetes services
# Example: 'k8s::init::kubernetes'
k8s::init::kubernetes() {
  k8s::common::setup_env

  mkdir -p "$SNAP_DATA/args"
  cp "$SNAP/k8s/args/kubelet" "$SNAP_DATA/args/kubelet"
  cp "$SNAP/k8s/args/kube-apiserver" "$SNAP_DATA/args/kube-apiserver"
  cp "$SNAP/k8s/args/kube-proxy" "$SNAP_DATA/args/kube-proxy"
  cp "$SNAP/k8s/args/kube-scheduler" "$SNAP_DATA/args/kube-scheduler"
  cp "$SNAP/k8s/args/kube-controller-manager" "$SNAP_DATA/args/kube-controller-manager"
}

# Configure permissions for important cluster config files
# Example: 'k8s::init::permissions'
k8s::init::permissions() {
  k8s::common::setup_env

  chmod go-rxw -R "$SNAP_DATA/args" "$SNAP_COMMON/opt" "$SNAP_COMMON/etc" "$SNAP_COMMON/var/lib" "$SNAP_COMMON/var/log"
}

# Initialize all services to run a single-node cluster
# Example: 'k8s::init'
k8s::init() {
  k8s::init::containerd
  k8s::init::k8s_dqlite
  k8s::init::k8sd
  k8s::init::ca
  k8s::init::pki
  k8s::init::kubernetes
  k8s::init::kubeconfigs
  k8s::init::permissions
}
