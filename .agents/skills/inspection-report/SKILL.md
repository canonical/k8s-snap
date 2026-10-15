<!--
Copyright 2026 Canonical, Ltd.
-->
# Inspection reports

An inspection report is a tarball of diagnostics pulled from a single
Canonical Kubernetes node: service status and logs, the cluster's own view
of itself, system and network diagnostics, and more.
`k8s/scripts/inspect.sh` is the script that builds it, and it is the
authoritative source for everything below -- when this skill and anything
else disagree, trust the script.

This skill covers the tarball's layout, an order to read it in, and the
traps that make an incomplete read look like a healthy system. It assumes
you already have a report tarball in hand.

## Obtaining a report

There are two equivalent, real invocations, and both need root:

1. `sudo k8s inspect [output-file]` -- the `k8s inspect` CLI subcommand
   (`docs/canonicalk8s/_parts/commands/k8s_inspect.md`).
2. `sudo /snap/k8s/current/k8s/scripts/inspect.sh [output_file]` -- the
   script directly, at its real installed path. This is the exact command
   the project's own bug report template asks reporters to run
   (`.github/ISSUE_TEMPLATE/bug_report.yaml`), and what the integration
   test suite calls on every test instance at teardown
   (`_generate_inspection_report` in
   `tests/integration/tests/conftest.py`).

Without root, the script prints "Elevated permissions are needed for this
command. Please use sudo." and exits.

`output-file` (`output_file` in the script) is an optional positional
argument either way. Omit it and the script writes a default name,
`inspection-report-<timestamp>.tar.gz` (e.g.
`inspection-report-20250109_132806.tar.gz`), into the current directory.

Flags that change what gets collected, with their defaults from
`inspect.sh` (the CLI reference documents the identical set):

- `--all-namespaces` (default off). Without it, `kubectl cluster-info dump`
  only covers the `default` and `kube-system` namespaces; the script
  defaults this off specifically "to avoid logging potentially sensitive
  user data" (its own comment). It only widens `cluster-info/` and
  `cluster-info/upgrades.json` -- nothing else in the report changes.
- `--num-snap-log-entries` (default `100000`). Caps how many lines
  `journalctl` and `snap logs k8s` keep, for every `<service>/journal.log`
  and for `snap-logs-k8s.log`. It does **not** change
  `mirrors/<unit>.log`: the registry-mirror journal collector hardcodes
  100000 regardless of this flag.
- `--timeout` (default `180s`). Caps every command that talks to `k8sd` or
  the API server. Purely local commands (`systemctl`, `journalctl`, `ps`,
  `dmesg`, and friends) are not wrapped in a timeout at all -- see the
  `timeout_warnings.log` trap below.
- `--core-dump-dir` (default `/var/crash`). Where `core_dumps/` is copied
  from.

The integration test suite always runs with `--all-namespaces` (see
`_generate_inspection_report`), so a report pulled from CI covers every
namespace; a report a user attaches to a bug report normally will not.
That same helper also captures `inspect.sh`'s own stdout/stderr into a
sibling `inspection_report_logs.txt` next to the tarball -- that is where
the "N commands timed out" and "service has restarted N times" console
messages end up, but it is not part of the tarball itself.

## The shape of the tarball

It is a gzip-compressed tar. Whatever you name the output file,
`inspect.sh` always tars a directory literally called `inspection-report/`
(`build_report_tarball` runs `tar -C "$(pwd)" -cf <out> inspection-report`
then `gzip`s it), so expanding it always produces that directory name,
regardless of what the `.tar.gz` itself is called:

```
tar xzf inspection-report-20250109_132806.tar.gz
cd inspection-report
```

Below that one directory it is a flat set of top-level files
(`k8s-status.log`, `nrestarts.log`, and so on) plus a handful of
subdirectories: one per collected service (`k8s.kubelet/`,
`k8s.containerd/`, ...), plus `sys/`, `args/`, `k8s.k8sd/`,
`cluster-info/` and `core_dumps/`, and conditionally `mirrors/`. Every
path from here on is relative to this directory.

## Reading order

`inspect.sh` collects a few dozen files. Reading them in the order they
build context on each other beats an alphabetical pass:

1. **Node role.** Exactly one of `is-worker-node`, `is-control-plane-node`
   or `is-not-bootstrapped-node` exists at the report root (derived from
   `k8s local-node-status`; the script writes exactly one, never more than
   one). This determines which services the report can even show you --
   `inspect.sh` defines two distinct lists and only checks the one that
   applies:
   - control-plane: `k8s.containerd`, `k8s.etcd`, `k8s.kube-proxy`,
     `k8s.k8sd`, `k8s.kube-apiserver`, `k8s.kube-controller-manager`,
     `k8s.kube-scheduler`, `k8s.kubelet`
   - worker: `k8s.containerd`, `k8s.k8s-apiserver-proxy`, `k8s.kubelet`,
     `k8s.k8sd`, `k8s.kube-proxy`

   A worker-node report has no `k8s.kube-apiserver/`, `k8s.etcd/`, and so
   on: those directories only exist for services the script was told to
   check. And on an `is-not-bootstrapped-node` report,
   `check_expected_services` still ran against the full control-plane
   list (the script's non-worker branch always checks it, whether or not
   the node turned out to be bootstrapped) -- so a wall of "should be
   running on this node" warnings there is expected noise, not a finding.

2. **Are services up and flapping.** `nrestarts.log` is a flat
   `service -> count` list, one line per checked service (e.g.
   `k8s.kubelet -> 0`), the fastest way to spot a restart loop.
   `<service>/systemctl.log` (`systemctl status snap.<service>`) has that
   service's current state; the script itself treats a service as up only
   when this output contains the literal string `active (running)`.

3. **The cluster's own view.** `k8s-status.log` (`k8s status`: is the
   cluster ready, the datastore, which features are enabled) and
   `k8s-get.log` (`k8s get`: the actual configured values for
   network/dns/gateway/ingress/local-storage/load-balancer) are this
   node's own belief about itself. `k8s-version.log` has the
   client/server Kubernetes versions. The same collection pass also drops
   `uname.log` (`uname -a`) and the snap's own self-report --
   `snap-version.log`, `snap-list-k8s.log`, `snap-services-k8s.log` --
   worth a glance first to confirm which kernel and which snap
   revision/channel this node actually has installed, before trusting
   anything else here.

4. **Service logs.** Once a specific service looks wrong, its journal is
   `<service>/journal.log` (`journalctl -u snap.<service>`, capped at
   `--num-snap-log-entries`). `snap-logs-k8s.log` (`snap logs k8s`) is the
   separate, snapd-multiplexed view of every `k8s` snap service
   interleaved by timestamp -- useful when you don't yet know which
   service is at fault.

5. **Configuration.** `args/` is a recursive copy of
   `/var/snap/k8s/common/args`: one file per service holding its startup
   flags (`args/kube-apiserver`, `args/kubelet`, `args/etcd`,
   `args/kube-scheduler`, `args/kube-controller-manager`, and so on), plus
   `args/conf.d/` for extra files uploaded at bootstrap/join time. A
   service crash-looping from a bad flag shows up here, not in its
   journal.

6. **Datastore and cluster membership.** `k8s.k8sd/k8sd-cluster.yaml` and
   `k8s.k8sd/k8sd-info.yaml` are raw copies of k8sd's own database state
   (`/var/snap/k8s/common/var/lib/k8sd/state/database/cluster.yaml` and
   `info.yaml`): this node's own record of cluster membership and
   configuration. `k8s.k8sd/k8sd-configmap.log` is the `k8sd-config`
   ConfigMap (`k8s kubectl get cm k8sd-config -n kube-system -o yaml`);
   the broader `k8s-configmaps.log` lists every ConfigMap in
   `kube-system` for comparison. `k8s.k8sd/k8sd-files.log` is an
   `ls -la` of `/var/snap/k8s/common/var/lib/k8sd`, useful for spotting a
   missing or zero-length state file.

7. **Kubernetes object state.** `cluster-info/` is the output directory
   of `kubectl cluster-info dump` (`default`/`kube-system` unless the
   report was taken with `--all-namespaces`): the Kubernetes-level object
   state, as opposed to the service-level view above it.
   `cluster-info/upgrades.json` is the in-progress-upgrade object, dumped
   separately (`k8s kubectl get upgrades -ojson`).

8. **Host and network context.** `sys/` holds `ps -ef` (`sys/ps`), disk
   usage (`sys/disk_usage`, `df -h`), `/proc/mounts`
   (`sys/proc-mounts`), memory (`sys/memory_usage`, `free -m`), swap
   (`sys/swap`, `swapon`), `uptime` (`sys/uptime`), `/etc/os-release`
   (`sys/etc-os-release`), loaded kernel modules
   (`sys/loaded_kernel_modules`, `lsmod`) and `dmesg -H` (`sys/dmesg`,
   worth a grep for OOM kills). Alongside it: `ip-a.log` (`ip a`),
   `ip-r.log` (`ip r`), the firewall dumps `iptables.log`,
   `iptables-legacy.log`, `iptables6.log`, `iptables6-legacy.log`, the
   listening-socket dumps `ss-plnt.log` (`ss -plnt`, TCP) and
   `ss-plntu.log` (`ss -plntu`, TCP+UDP), and `proxy_in_etc_environment`
   (any `HTTP_PROXY`/`HTTPS_PROXY`/`NO_PROXY` lines from
   `/etc/environment`).

9. **The remaining specials.** `sbom.json` is a copy of the snap's own
   `bom.json` (software bill of materials): exact component versions
   shipped, useful for confirming this is even the version the reporter
   thinks it is. `mirrors/<unit>.log` exists only when the node has
   enabled systemd units matching `registry-*` (containerd registry
   mirrors) -- one journal per enabled unit, and the directory itself is
   only created if at least one exists. `core_dumps/` is always created
   (even empty) and holds anything copied from `--core-dump-dir` (default
   `/var/crash`); empty means either nothing crashed or core dumps were
   never enabled on that host, and the report cannot tell you which.

## Traps

- **Check `timeout_warnings.log` before you trust an absence.** It lists
  every command that hit `--timeout` (default 180s): `run_with_timeout`
  appends the command line whenever it exits 124. Only commands that talk
  to `k8sd` or the API server go through it: `k8s kubectl version`,
  `k8s status`, `k8s get`, both `k8s kubectl get cm ...` calls, and both
  `cluster-info` collectors. If `k8s-status.log`, `k8s-get.log`,
  `k8s-version.log`, `k8s-configmaps.log`, `k8s.k8sd/k8sd-configmap.log`
  or `cluster-info/` is missing, empty or cut off, check
  `timeout_warnings.log` first: it can mean the command hung for 180
  seconds, not that whatever it was asking about is healthy. A slow or
  wedged API server produces exactly this pattern, and that pattern is
  itself a finding.

- **Service names in the report are snap unit names.** Directories,
  `nrestarts.log` entries and `args/` files all use the bare service name
  as `inspect.sh` sees it -- `k8s.kubelet`, `k8s.k8sd`, `k8s.containerd`,
  `k8s.kube-apiserver`, and so on (full lists under "Node role" above).
  The journals underneath were collected with one more `snap.` prefix,
  `journalctl -u snap.<name>`, matching the unit name systemd itself uses
  (`sudo journalctl -xe -u snap.k8s.<service>` is the same convention
  live, per the troubleshooting guide).

- **One report is one node, at one moment.** A fault specific to one
  cluster member will not appear in another member's report, and a
  report you don't have is not evidence that the node it would describe
  is healthy -- it is just a report you don't have. A control-plane-only
  symptom needs that control-plane node's own report; a worker's report
  has no `k8s.kube-apiserver/` directory to even look in.

- **Reading a report is evidence, never a reproduction.** It shows what
  this node looked like at collection time. It does not show that the
  fault still reproduces, or reproduces the same way, right now.

- **Treat every report as sensitive by default.** `inspect.sh` greps the
  fully-collected report for certificates and keys
  (`grep -rlEi "BEGIN CERTIFICATE|PRIVATE KEY" inspection-report`) and
  prints a red warning naming the offending files -- but it does not
  strip or redact anything; the tarball is built exactly as found. The
  `k8s inspect` CLI reference claims the opposite ("This command removes
  sensitive information, such as secrets and certificates, from the
  generated report"); that sentence is inconsistent with what the script
  actually does, so trust the script. The report format reference is
  more careful: it only says users are "free to inspect the report and
  remove any information deemed sensitive before sharing it". Assume any
  report you are handed still contains certificates, keys, and whatever
  workload data was live in `default`/`kube-system` (or every namespace,
  if it was collected with `--all-namespaces`). Never paste its contents
  into a public comment or issue, and keep it inside a private scratch
  directory.

## Worked examples

**"A pod keeps restarting."**

1. Confirm the node role (step 1 above); a worker has `k8s.kubelet` and
   `k8s.containerd` but no `k8s.kube-apiserver`.
2. `cat nrestarts.log` -- if `k8s.kubelet` or `k8s.containerd` itself is
   restarting, chase that first; it is a more likely cause than anything
   pod-specific.
3. `grep -i <pod-name> k8s.kubelet/journal.log` -- kubelet's journal
   records the pod's own lifecycle: image pulls, probe failures, backoff.
4. `grep -iE "oom|error" k8s.containerd/journal.log` -- runtime-level
   failures that kubelet only reports secondhand.
5. `grep -i "out of memory" sys/dmesg` -- a kernel OOM kill of the
   container shows here even when neither service log mentions it.
6. `grep -rl <pod-name> cluster-info/` -- find the pod's own dumped
   object and logs in the `kubectl cluster-info dump` tree (only present
   if its namespace was in scope: `default`/`kube-system` unless the
   report was taken with `--all-namespaces`).

**"The cluster is unreachable."**

1. `wc -l timeout_warnings.log` -- if `k8s kubectl version`, `k8s status`
   or `k8s get` show up here, the API server or k8sd was already not
   answering when the report was taken. That is the answer, and every
   file below may be empty or missing as a direct consequence, not a
   separate problem.
2. `cat k8s-status.log` -- the node's own last-known status line. Empty
   or truncated, combined with step 1, confirms the API/k8sd path was
   hanging rather than merely reporting bad news.
3. Confirm this is a control-plane report (step 1 above):
   `k8s.kube-apiserver/` only exists there. `grep k8s.kube-apiserver
   nrestarts.log` and `cat k8s.kube-apiserver/systemctl.log` for whether
   the process is even up.
4. `cat k8s.kube-apiserver/journal.log` for the API server's own error
   output: etcd connection refused, listener bind failure, certificate
   errors.
5. If the API server looks healthy, check membership instead:
   `k8s.k8sd/journal.log` and `k8s.k8sd/k8sd-cluster.yaml`, for whether
   this node still agrees with the rest of the cluster about who its
   members are.
6. `grep ':6443' ss-plnt.log` (API server) or `grep ':6400' ss-plnt.log`
   (k8sd) to confirm the relevant port is actually `LISTEN`ing on this
   node.
7. This is still one node's view: if it looks fine, the unreachable
   symptom may belong to a different member. Get that node's own report
   before concluding anything.
