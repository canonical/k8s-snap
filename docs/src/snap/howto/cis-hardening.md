# CIS compliance 

CIS Hardening refers to the process of implementing security configurations that
align with the benchmarks set by the [Center for Internet Security (CIS)][]. The
open source tool [kube-bench][] is designed to automatically check whether
your Kubernetes clusters are configured according to the 
[CIS Kubernetes Benchmark][]. This guide covers how to setup your {{product}} 
cluster with kube-bench. 

## What you'll need

This guide assumes the following:

- You have a bootstrapped {{product}} cluster (see the [Getting Started]
[getting-started-guide] guide)
- You have root or sudo access to the machine

## Install kube-bench 

Download the latest [kube-bench release][] on your Kubernetes nodes. Make sure 
to select the appropriate binary version.

For example, to download the Linux binary, use the following command. Replace 
KB by the version listed in the releases page.

``` 
KB=8.0
mkdir kube-bench
cd kube-bench
curl -L https://github.com/aquasecurity/kube-bench/releases/download/v0.$KB/kube-bench_0.$KB\_linux_amd64.tar.gz -o kube-bench_0.$KB\_linux_amd64.tar.gz
```

Extract the downloaded tarball and move the binary to a directory in your PATH:

```
tar -xvf kube-bench_0.$KB\_linux_amd64.tar.gz
sudo mv kube-bench /usr/local/bin/
``` 

Verify kube-bench installation.

```
kube-bench version
``` 

The output should list the version installed.

Install `kubectl` and configure it to interact with the cluster.

```{warning}
This will override your ~/.kube/config if you already have kubectl installed in your cluster. 
```

```
sudo snap install kubectl --classic
mkdir ~/.kube/
sudo k8s kubectl config view --raw > ~/.kube/config
export KUBECONFIG=~/.kube/config
```

Get CIS hardening checks applicable for {{product}}:

```
git clone -b ck8s https://github.com/canonical/kube-bench.git kube-bench-ck8s-cfg
```

Test-run kube-bench against {{product}}: 

```
sudo -E kube-bench --version ck8s-dqlite-cis-1.24 --config-dir ./kube-bench-ck8s-cfg/cfg/ --config ./kube-bench-ck8s-cfg/cfg/config.yaml
```

## Harden your deployments

Before running a CIS Kubernetes audit, it is essential to first harden your
Canonical Kubernetes deployment to minimize vulnerabilities and ensure 
compliance with industry best practices, as defined by the CIS Kubernetes
Benchmark.

### Control plane nodes

#### Configure auditing

Create an audit-policy.yaml file under /var/snap/k8s/common/etc/ and specify the
level of auditing you desire based on the [upstream instructions][]. Here is a 
minimal example of such a policy file.

```sh
sudo sh -c 'cat >/var/snap/k8s/common/etc/audit-policy.yaml <<EOL
# Log all requests at the Metadata level.
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: Metadata
EOL'
```

Enable auditing at the API server by adding the following arguments.

```sh
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL
--audit-log-path=/var/log/apiserver/audit.log
--audit-log-maxage=30
--audit-log-maxbackup=10
--audit-log-maxsize=100
--audit-policy-file=/var/snap/k8s/common/etc/audit-policy.yaml
EOL'
```

Restart the API server:

```
sudo systemctl restart snap.k8s.kube-apiserver
```

#### Set event rate limits

Create a configuration file with the [rate limits][] and place it under 
`/var/snap/k8s/common/etc/`.
For example:

```sh
sudo sh -c 'cat >/var/snap/k8s/common/etc/eventconfig.yaml <<EOL
apiVersion: eventratelimit.admission.k8s.io/v1alpha1
kind: Configuration
limits:
  - type: Server
    qps: 5000
    burst: 20000
EOL'
```

Create an admissions control config file under /var/k8s/snap/common/etc/ .

```sh
sudo sh -c 'cat >/var/snap/k8s/common/etc/admission-control-config-file.yaml <<EOL
apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
  - name: EventRateLimit
    path: eventconfig.yaml
EOL'
```

Make sure the EventRateLimit admission plugin is loaded in the 
/var/snap/k8s/common/args/kube-apiserver .

```sh
--enable-admission-plugins=...,EventRateLimit,...
```

Load the admission control config file.

```sh
sudo sh -c 'cat >>/var/snap/k8s/common/args/kube-apiserver <<EOL
--admission-control-config-file=/var/snap/k8s/common/etc/admission-control-config-file.yaml
EOL'
```

Restart the API server.

```
sudo systemctl restart snap.k8s.kube-apiserver
```

#### Enable AlwaysPullImages admission control plugin

Make sure the AlwaysPullImages admission plugin is loaded in the 
/var/snap/k8s/common/args/kube-apiserver

```sh
--enable-admission-plugins=...,AlwaysPullImages,...
```

Restart the API server.

```
sudo systemctl restart snap.k8s.kube-apiserver
```

### Worker nodes

#### Protect kernel defaults

Kubelet will not start if it finds kernel configurations incompatible with its
 defaults.

```sh
sudo sh -c 'cat >>/var/snap/k8s/common/args/kubelet <<EOL
--protect-kernel-defaults=true
EOL'
```

Restart `kubelet`.

```
sudo systemctl restart snap.k8s.kubelet
``` 

Reload the system daemons:

```
sudo systemctl daemon-reload
```

## Audit your deployments

Run kube-bench against {{product}} control-plane nodes:

```
sudo -E kube-bench --version ck8s-dqlite-cis-1.24  --config-dir ./kube-bench-ck8s-cfg/cfg/ --config ./kube-bench-ck8s-cfg/cfg/config.yaml
```

Verify that there are no checks failed for control or worker nodes in any of 
the sets, including the dqlite specific checks in the output.

```sh
[INFO] 1 Control Plane Security Configuration
...
[PASS] 1.1.7 Ensure that the dqlite configuration file permissions are set to 644 or more restrictive (Automated)
[PASS] 1.1.8 Ensure that the dqlite configuration file ownership is set to root:root (Automated)
...
[PASS] 1.1.11 Ensure that the dqlite data directory permissions are set to 700 or more restrictive (Automated)
[PASS] 1.1.12 Ensure that the dqlite data directory ownership is set to root:root (Automated)
...
== Summary master ==
55 checks PASS
0 checks FAIL
4 checks WARN
0 checks INFO

[INFO] 3 Control Plane Configuration
...
== Summary controlplane ==
1 checks PASS
0 checks FAIL
2 checks WARN
0 checks INFO

[INFO] 4 Worker Node Security Configuration
...
== Summary node ==
23 checks PASS
0 checks FAIL
0 checks WARN
0 checks INFO

[INFO] 5 Kubernetes Policies
...
== Summary policies ==
0 checks PASS
0 checks FAIL
30 checks WARN
0 checks INFO

== Summary total ==
79 checks PASS
0 checks FAIL
36 checks WARN
0 checks INFO

```

<!-- Links -->
[Center for Internet Security (CIS)]:https://www.cisecurity.org/
[kube-bench]:https://aquasecurity.github.io/kube-bench/v0.6.15/
[CIS Kubernetes Benchmark]:https://www.cisecurity.org/benchmark/kubernetes
[getting-started-guide]: /snap/tutorial/getting-started
[kube-bench release]: https://github.com/aquasecurity/kube-bench/releases
[upstream instructions]:https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[rate limits]:https://kubernetes.io/docs/reference/config-api/apiserver-eventratelimit.v1alpha1
