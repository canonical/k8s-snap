<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: reproduce

Goal: decide whether the reported issue reproduces on a cluster you build
yourself, or whether it should be skipped as out of scope for automated triage.

This is the first step of the pipeline. Nothing downstream touches product code
until you have seen the failure happen here.

## Procedure

1. Read the issue and any prior report context. Extract the exact commands,
   manifests, k8s-snap version/channel, and the **cluster shape** the reporter
   used: how many control-plane nodes, how many workers.
2. If an inspection tarball is attached, expand and read it first (`k8s status`,
   service logs). It tells you what to look for and often names the mechanism,
   but reading it is **not** a reproduction.
3. Build a cluster with the project helper. It installs any missing tooling,
   builds `k8s.snap` when absent, launches LXD nodes and joins them:
   ```bash
   hack/cluster-up.sh --control-plane <N> --workers <M>
   ```
   Match the reporter's shape: "two nodes, 1 control plane and 1 worker" is
   `--control-plane 1 --workers 1`. Re-running is safe and cheap, it skips
   whatever already exists. See `hack/cluster-up.sh --help`.
4. Drive the cluster exactly as the reporter did. Nodes are named
   `k8s-triage-cp<i>` and `k8s-triage-w<i>`:
   ```bash
   lxc exec k8s-triage-cp1 -- k8s status
   lxc exec k8s-triage-cp1 -- k8s kubectl get pods -A -o wide
   ```
5. Observe. Did the reported failure happen, on this cluster, in front of you?
   Some defects need time or repetition (a restart loop, a rebalance cycle);
   watch long enough to be sure, and say what you watched.

## Return

- `reproducible`: true **only** if you observed the reported failure yourself on
  the cluster you just built. Evidence taken from the inspection tarball, the
  issue text, or the source code is not a reproduction: return false.
- `evidence`: the command you ran and the output lines that show the failure.
  Required whenever `reproducible` is true.
- `skipped` + `skipped_reason` when triage should not proceed:
  - `missing-details`: not enough information to attempt a reproduction.
  - `unsupported-version`: the affected version is EOL / out of support.
  - `host-specific`: needs specific hardware/cloud not available here.
  - `unsupported-runtime`: needs a substrate this environment cannot provide.
  - `not-actionable`: a question or support request, not a defect.
  - `maintainer-override`: a maintainer asked to stop automated triage.
- If you ran a reproduction and the failure did not occur, return
  `reproducible: false` with `skipped: false`.

Do not diagnose, write tests, or change code in this step.
