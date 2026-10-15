<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: diagnose

Goal: given an end-to-end test that fails because of this bug, locate the most
likely root cause and state how confident you are.

You run after the reproducer step, so you are not chasing a description: there
is a command that fails, and its output is in the report. Start from that.

> Uses: `inspection-report`, `local-cluster`

## Procedure

1. Re-read the reproduction evidence and the test failure in `report.md`. If the
   issue carries an inspection tarball, read it now as well: it is a snapshot of
   the failure on the reporter's own machine, and it frequently names the
   mechanism your reproduction only shows the symptom of.
2. Trace from the observed failure to the code responsible for it. `SKILL.md`
   says where that code lives; the practical consequence here is that the answer
   is usually **not** in this repository. Clone the adjacent repository at the
   version this build actually ships into your scratch directory:
   ```bash
   git clone --depth 1 --branch "$(cat build-scripts/components/k8sd/version)" \
     "$(cat build-scripts/components/k8sd/repository)" .triage/issue-<n>/k8sd
   ```
   Building `k8sd` locally needs `libdqlite-dev`; when that is missing, read the
   source rather than fighting the build. Cloning it is for *reading*: if the
   root cause turns out to live there, the fix step will report it for a human
   rather than attempt it, because a commit in an adjacent clone cannot be
   built, verified or published by this pipeline (see `fix.md` step 3).
3. Correlate what you have: log lines, service names, the args files under
   `/var/snap/k8s/common/args`, and the code path you just read. If the
   reproducer step did not destroy it, the cluster from the reproduction is
   still up, so you can go back and read live state instead of guessing. If it
   is gone, read the state from the inspection tarball.
4. Form a single, specific hypothesis: the file and function, and the mechanism
   by which it produces the failure the test observes. Note whether the fix
   would land in this repository or in an adjacent one.

## Return

- `confidence`:
  - `high`: you identified a specific code path and the mechanism is clear.
  - `medium`: a plausible area is identified but the exact cause is unconfirmed.
  - `low`: the failure is real but the cause is still unclear.
- `hypothesis`: where the root cause lives -- the file, and the
  function/symbol when you know it.
- `evidence`: what supports it: the log lines, code path or cluster state you
  actually read.

Return both fields rather than leaving them in your reasoning: only the
structured result is written to `report.md`, so anything you work out but do
not return here never reaches the fix step. Do not modify code in this step.
