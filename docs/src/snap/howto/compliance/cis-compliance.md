# CIS compliance

Canonical Kubernetes is shipped as 'secure by default'. However, to fully
adhere to the complete specification of the [CIS Benchmark][], this requires
some opinionated install and configuration choices and a few manual steps. This
document will present the recommended setup, install and configuration settings
to achieve full compliance.  

## Overview

- Full compliance is not possible without an external **etcd** datastore. This
  guide includes the steps necessary to prepare the datastore.
- To better reflect a real-world install, these instructions include setting up
  one control-plane node and one worker node. These can be scaled as desired.

## What you will need


## Single node etcd setup

This reference requires single node **etcd** cluster to be used as the datastore
for Canonical Kubernetes. The **etcd** cluster must be installed on the same node
as the Kubernetes control plane. 

### Prepare certificates

Create a file `etcd-tls.conf` with the minimal required configuration:

```
cat <<EOF > etcd-tls.conf
[req]
default_bits  = 4096
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
countryName = US
stateOrProvinceName = CA
localityName = San Francisco
organizationName = etcd
commonName = etcd-host

[v3_req]
keyUsage = digitalSignature, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
IP.1 = 127.0.0.1
DNS.1 = localhost
EOF

```

#### Generate CA

```
openssl req -x509 -nodes -newkey rsa:4096 -subj /CN=etcdRootCA \
 -keyout ca-key.pem -out ca-cert.pem
```

3. Generate client key and CSR for the client key

```
openssl req -nodes -newkey rsa:4096 -keyout client-key.pem \
-out client-cert.csr -config etcd-tls.conf
```

4. Generate client certificate

```
openssl x509 -req -in client-cert.csr -CA ca-cert.pem \
-CAkey ca-key.pem -out client-cert.pem -extensions v3_req \
-extfile etcd-tls.conf -CAcreateserial
```

5. Generate server key and CSR for the server key

```
openssl req -nodes -newkey rsa:4096 -keyout server-key.pem \
-out server-cert.csr -config etcd-tls.conf
```

6. Generate server certificate

```
openssl x509 -req -in server-cert.csr -CA ca-cert.pem \
-CAkey ca-key.pem -out server-cert.pem -extensions v3_req \
-extfile etcd-tls.conf -CAcreateserial
```

#### Verify certificates generation

We now have the following certificates:

- CA: 
  - `ca-cert.pem`
  - `ca-key.pem`

- Client certificate: 
  - `client-cert.pem` 
  - `client-key.pem`

- Certificate for etcd peer communication: 
  - `server-cert.pem` 
  - `server-key.pem`

### Bootstrap single-node etcd


Install the `etcd` server: 

```
sudo apt-get update
sudo apt-get install etcd-server -y
```

Add the certificates:

```
sudo mkdir /etc/etcd-certs
sudo cp ./ca-* /etc/etcd-certs/ 
sudo cp ./client-* /etc/etcd-certs/
sudo cp ./server-* /etc/etcd-certs/
sudo chown -R etcd:etcd /etc/etcd-certs/
```

Configure etcd

```
sudo sh -c 'cat >>/etc/default/etcd <<EOL
ETCD_NAME="infra0"
ETCD_DATA_DIR="/var/lib/etcd"
ETCD_LISTEN_CLIENT_URLS="https://127.0.0.1:2379"
ETCD_CLIENT_CERT_AUTH="true"
ETCD_TRUSTED_CA_FILE="/etc/etcd-certs/ca-cert.pem"
ETCD_CERT_FILE="/etc/etcd-certs/client-cert.pem"
ETCD_KEY_FILE="/etc/etcd-certs/client-key.pem"
ETCD_ADVERTISE_CLIENT_URLS="https://127.0.0.1:2379"
ETCD_LISTEN_CLIENT_URLS="https://127.0.0.1:2379"
ETCD_PEER_CLIENT_CERT_AUTH="true"
ETCD_PEER_TRUSTED_CA_FILE="/etc/etcd-certs/ca-cert.pem"
ETCD_PEER_CERT_FILE="/etc/etcd-certs/server-cert.pem"
ETCD_PEER_KEY_FILE="/etc/etcd-certs/server-key.pem"
EOL'
```

Restart etcd

```
sudo systemctl restart etcd.service
```

#### Verify etcd cluster is setup

If you want to verify your configuration is valid you can send a verbose https
request to the etcd server.

```
curl --cacert ca-cert.pem --cert ./client-cert.pem \
--key ./client-key.pem -L https://127.0.0.1:2379/version -v
```

## Canonical Kubernetes setup

Canonical Kubernetes is delivered as a snap (https://snapcraft.io/k8s). 

### Control plane and worker node

Install the snap from the desired track.

```
sudo snap install k8s --classic --channel=1.30/edge
```

Create a file called *configuration.yaml*. In this configuration file we let
the snap start with its default CNI and DNS deployed and we also
point k8s to the external etcd:

<!-- TO DO:
 Is this config relevant for generic k8s? (e.g. cpu) -->

```
cluster-config:
  network:
    enabled: true
  dns:
    enabled: true
  local-storage:
    enabled: true
extra-node-kubelet-args:
  --reserved-cpus: "0-31"
  --cpu-manager-policy: "static"
  --topology-manager-policy: "best-effort"
datastore-type: external
datastore-servers:
  - https://127.0.0.1:2379
datastore-ca-crt: |
  ### insert the contents of ca-cert.pem
datastore-client-crt: |
  ### insert the contents of client-cert.pem
datastore-client-key: |
  ### insert the contents of client-key.pem

```

Bootstrap Canonical Kubernetes using the above configuration file.

```
sudo k8s bootstrap --file configuration.yaml
```

#### Verify single node k8s cluster is up

After a few seconds you can query the API server with:

```
sudo k8s kubectl get all -A
```

### Second k8s node as worker

1. Install the k8s snap on the second node

```
sudo snap install k8s --classic --channel=1.30/edge
```

2. On the control plane node generate a join token to be used for joining the second node

 ```
controlplane$ sudo k8s get-join-token --worker
```

3. On the worker node create  the followingconfiguration.yaml

```
extra-node-kubelet-args:  
  --reserved-cpus: "0-31"  
  --cpu-manager-policy: "static"  
  --topology-manager-policy: "best-effort"
```

4. On the worker node use the token to join the cluster

```
sudo k8s join-cluster --file configuration.yaml <token-generated-on-the-control-plane-node>
```

#### Verify the two node cluster is ready

After a few seconds the second worker node will register with the control plane. You can query the available workers from the first node.

```
sudo k8s kubectl get no
```


## Hardening the control plane node

### Configure auditing

1. Create an *audit-policy.yaml *file under /var/snap/k8s/common/etc/ and specify the level of auditing you desire based on the [upstream instructions](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/). Here is a minimal example of such a policy file.

```
sudo sh -c 'cat >/var/snap/k8s/common/etc/audit-policy.yaml <<EOL
# Log all requests at the Metadata level.
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: Metadata
EOL'
```

2. Enable auditing at the API server by adding the following arguments.

```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL--audit-log-path=/var/log/apiserver/audit.log--audit-log-maxage=30--audit-log-maxbackup=10--audit-log-maxsize=100--audit-policy-file=/var/snap/k8s/common/etc/audit-policy.yamlEOL'
```

3. Restart the API server.

3. ```
sudo systemctl restart snap.k8s.kube-apiserver
```

### Set event rate limits

1. Create a configuration file with the [rate limits](https://kubernetes.io/docs/reference/config-api/apiserver-eventratelimit.v1alpha1) and place it under /var/snap/k8s/common/etc/ . For example:

```
sudo sh -c 'cat >/var/snap/k8s/common/etc/eventconfig.yaml <<EOLapiVersion: eventratelimit.admission.k8s.io/v1alpha1kind: Configurationlimits:  - type: Server    qps: 5000    burst: 20000EOL'
```

2. Create an admissions control config file under /var/k8s/snap/common/etc/

```
sudo sh -c 'cat >/var/snap/k8s/common/etc/admission-control-config-file.yaml <<EOL`apiVersion:`` ``apiserver.config.k8s.io/v1`
`kind:`` ``AdmissionConfiguration`
plugins:`  ``-`` ``name:`` ``EventRateLimit`
`    ``path:`` ``eventconfig.yaml`
EOL'
```

3. Make sure the EventRateLimit admission plugin is loaded in the /var/snap/k8s/common/args/kube-apiserver.

3. ```
--enable-admission-plugins=...,EventRateLimit,...
```

4. Load the admission control config file.

4. ```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL`--admission-control-config-file=``/var/snap/k8s/common/etc/admission-control-config-file.yaml`
EOL'
```

Restart the API server.

```
sudo systemctl restart snap.k8s.kube-apiserver
```

### AlwaysPullImages admission control plugin

1. Make sure the AlwaysPullImages admission plugin is loaded in the /var/snap/k8s/common/args/kube-apiserver.

```
--enable-admission-plugins=...,AlwaysPullImages,...
```

2. Restart the API server.

2. ```
sudo systemctl restart snap.k8s.kube-apiserver
```

## Hardening the workers

### Protect kernel defaults

Kubelet will not start if it finds kernel configurations incompatible with its defaults.

1. Configure kubelet.

1. ```
sudo sh -c 'cat >>/var/snap/k8s/common/args/kubelet <<EOL--protect-kernel-defaults=trueEOL'
```

2. Restart kubelet.

```
sudo systemctl restart snap.k8s.kubelet
```

## Run kube-bench on Canonical Kubernetes

1. Download the KubeBench.

```
$ mkdir kube-bench$ cd kube-bench$ curl -L https://github.com/aquasecurity/kube-bench/releases/download/v0.7.3/kube-bench_0.7.3_linux_amd64.tar.gz -o kube-bench_0.7.3_linux_amd64.tar.gz$ tar -xvf kube-bench_0.7.3_linux_amd64.tar.gz
```

2. Fetch the CIS hardening checks applicable for Canonical Kubernetes.

2. ```
git clone -b ck8s https://github.com/canonical/kube-bench.git kube-bench-ck8s-cfg
```

3. Install kubectl and configure it to interact with the cluster

```
sudo snap install kubectl --classic$ mkdir ~/.kube/$ sudo k8s kubectl config view --raw > ~/.kube/config$ export KUBECONFIG=~/.kube/config
```

Run kube-bench against Canonical Kubernetes. Make sure you do not have any files prefixed with “etcd” in the current working directory.

```
sudo -E ./kube-bench --version cis-1.24-ck8s --config-dir ./kube-bench-ck8s-cfg/cfg/ --config ./kube-bench-ck8s-cfg/cfg/config.yaml
```


<!--
##### References

etcd project: [https://github.com/etcd-io/etcd](https://github.com/etcd-io/etcd/releases/download/v3.5.14/etcd-v3.5.14-linux-amd64.tar.gz)

k8s snap on snapstore: [https://snapcraft.io/k8s](https://snapcraft.io/k8s)

CNCF conformance instructions [https://github.com/cncf/k8s-conformance/blob/master/instructions.md](https://github.com/cncf/k8s-conformance/blob/master/instructions.md)

sonobuoy project [https://github.com/vmware-tanzu/sonobuoy](https://github.com/vmware-tanzu/sonobuoy)

auditing kubernetes [https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/)

Multus: [https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md)

Sriov network device plugin: [https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin/tree/master?tab=readme-ov-file#quick-start](https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin/tree/master?tab=readme-ov-file#quick-start)

Sriov cni: [https://github.com/k8snetworkplumbingwg/sriov-cni?tab=readme-ov-file#kubernetes-quick-start](https://github.com/k8snetworkplumbingwg/sriov-cni?tab=readme-ov-file#kubernetes-quick-start)
-->
