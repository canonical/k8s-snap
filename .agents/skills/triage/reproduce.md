<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: reproduce

Goal: decide whether the reported issue reproduces on a cluster you build
yourself, or whether it should be skipped as out of scope for automated triage.

This is the first step of the pipeline. Nothing downstream touches product code
until you have seen the failure happen here.

> Uses: `local-cluster`, `inspection-report`

## Procedure

1. Read the issue and any prior report context. Extract the exact commands,
   manifests, k8s-snap version/channel, and the **cluster shape** the reporter
   used: how many control-plane nodes, how many workers.
2. If an inspection tarball is attached, expand and read it before you build
   anything: it usually names the mechanism and saves you a search. Reading it
   is **not** a reproduction.
3. Build a cluster that matches the reporter's shape. "Two nodes, 1 control
   plane and 1 worker" means exactly that, not two of either: a shape you
   invent exercises a different system than the one that failed.
4. Drive the cluster exactly as the reporter did, with their commands and their
   manifests. Deviating from them gives you a different experiment and a
   worthless answer.
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
