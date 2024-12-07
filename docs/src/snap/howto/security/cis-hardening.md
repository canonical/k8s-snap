# CIS compliance

CIS Hardening refers to the process of implementing security configurations that
align with the benchmarks set by the [Center for Internet Security (CIS)][].
Out of the box {{product}} complies with the majority of the recommended
CIS security configurations. Since implementing all security recommendations
would comes at the expense of compatibility and/or performance we expect
cluster administrators to follow post deployment hardening steps based on their
needs. This guide covers:

  * Post-deployment hardening steps you could consider for your {{product}}
  * Using [kube-bench][] to automatically check whether your Kubernetes
    clusters are configured according to the [CIS Kubernetes Benchmark][]
  * Manually configuring and auditing each CIS hardening recommendation


## What you'll need

This guide assumes the following:

- You have a bootstrapped {{product}} cluster (see the [getting started] guide)
- You have root or sudo access to the machine


## Post-deployment configuration steps

By completing these steps, you can ensure your cluster achieves full compliance
with CIS hardening guidelines.

```{include} ../../../_parts/common_hardening.md
```

## Assess CIS hardening with kube-bench

Download the latest [kube-bench release][] on your Kubernetes nodes. Make sure
to select the appropriate binary version.

For example, to download the Linux binary, use the following command. Replace
`KB` by the version listed in the releases page.

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
git clone -b ck8s-dqlite https://github.com/canonical/kube-bench.git kube-bench-ck8s-cfg
```

Test-run kube-bench against {{product}}:

```
sudo -E kube-bench --version ck8s-cis-1.24 --config-dir ./kube-bench-ck8s-cfg/cfg/ --config ./kube-bench-ck8s-cfg/cfg/config.yaml
```

Review the warnings detected and address any failing checks you see fit.

```
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


## Manually audit CIS hardening recommendations

For manual audits of CIS hardening recommendations, please visit the
[Comprehensive Hardening Checklist][].


<!-- Links -->
[Hardening]:security/hardening.md
[Center for Internet Security (CIS)]:https://www.cisecurity.org/
[kube-bench]:https://aquasecurity.github.io/kube-bench/v0.6.15/
[CIS Kubernetes Benchmark]:https://www.cisecurity.org/benchmark/kubernetes
[getting started]: ../tutorial/getting-started
[kube-bench release]: https://github.com/aquasecurity/kube-bench/releases
[Post-Deployment Configuration Steps]: security/hardening.md#post-deployment-configuration-steps
[Comprehensive Hardening Checklist]: security/hardening.md#comprehensive-hardening-checklist
