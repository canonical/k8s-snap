# Inspection reports

{{product}} ships with a command to compile a complete report on {{product}} and
its underlying system. This is an essential tool for bug reports and for
investigating why a system is not working.

The resulting report is a tarball containing service arguments and logs, SBOM,
system diagnostics, network diagnostics and more.

```{important}
The collected data will not be submitted automatically. The users are free to
inspect the report and remove any information deemed sensitive before sharing
it.
```

The command tries to limit the report size and avoid private user data. It
also accepts a few arguments that control how and what will be collected.
See the following sections for more details.

Check the following script to see how the inspection report gets generated:
https://github.com/canonical/k8s-snap/blob/release-1.32/k8s/scripts/inspect.sh

## Using the built-in inspection command

Use the following command to generate an inspection report. Note that admin
privileges are required to collect the data.

```
sudo k8s inspect
```

The command output is similar to the following:

```
Collecting service information
Running inspection on a control-plane node
 INFO:  Service k8s.containerd is running
 INFO:  Service k8s.etcd is not-running
 INFO:  Service k8s.kube-proxy is running
 INFO:  Service k8s.etcd is running
 INFO:  Service k8s.k8sd is running
 INFO:  Service k8s.kube-apiserver is running
 INFO:  Service k8s.kube-controller-manager is running
 INFO:  Service k8s.kube-scheduler is running
 INFO:  Service k8s.kubelet is running
Collecting registry mirror logs
Collecting service arguments
 INFO:  Copy service args to the final report tarball
Collecting k8s cluster-info
 INFO:  Copy k8s cluster-info dump to the final report tarball
Collecting SBOM
 INFO:  Copy SBOM to the final report tarball
Collecting system information
 INFO:  Copy uname to the final report tarball
 INFO:  Copy snap diagnostics to the final report tarball
 INFO:  Copy k8s diagnostics to the final report tarball
Collecting networking information
 INFO:  Copy network diagnostics to the final report tarball
Building the report tarball
 SUCCESS:  Report tarball is at /root/inspection-report-20250109_132806.tar.gz
```

Use the report to ensure that all necessary services are running and dive into
every aspect of the system.

## Command arguments

### ``--all-namespaces``

The ``inspect`` command aims to avoid sensitive data, so by default it only
retrieves information from the ``default`` and  ``kube-system`` namespaces.

To collect logs from all namespaces, use the ``--all-namespaces`` argument.

### ``--num-snap-log-entries``

To keep the inspection report size reasonable, it will collect at most
100,000 entries from the snap logs.

If necessary, use the ``--num-snap-log-entries`` argument to increase the limit.

### ``--core-dump-dir``

Core dumps can help determine the cause of process crashes, especially when
the logs do not contain stack traces.

The ``inspect`` command will collect all core dump files located in the
``/var/crash`` folder.

To use a different core dump location, specify the ``--core-dump-dir``
argument.

Core dumps can be enabled like so:

```
sudo su
echo "/var/crash/core-%e.%p.%h" > /proc/sys/kernel/core_pattern
echo 1 > /proc/sys/fs/suid_dumpable
snap set system system.coredump.enable=true
```

### ``--timeout``

This argument adjusts the timeout used when executing various commands that
collect report data.

By default, it will wait up to 180 seconds for each of these commands.
