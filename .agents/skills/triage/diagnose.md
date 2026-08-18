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
   version this build actually ships, never at its default branch, into your
   scratch directory:
   ```bash
   git clone --depth 1 --branch "$(cat build-scripts/components/k8sd/version)" \
     "$(cat build-scripts/components/k8sd/repository)" .triage/issue-<n>/k8sd
   ```
   Building `k8sd` locally needs `libdqlite-dev`; when that is missing, read the
   source rather than fighting the build, and let the snap rebuild in the fix
   step be your validation.
3. Correlate what you have: log lines, service names, the args files under
   `/var/snap/k8s/common/args`, and the code path you just read. The cluster
   from the reproduction is still up, so you can go back and read live state
   instead of guessing.
4. Form a single, specific hypothesis: the file and function, and the mechanism
   by which it produces the failure the test observes. Note whether the fix
   would land in this repository or in an adjacent one.

## Return

- `confidence`:
  - `high`: you identified a specific code path and the mechanism is clear.
  - `medium`: a plausible area is identified but the exact cause is unconfirmed.
  - `low`: the failure is real but the cause is still unclear.

Record the hypothesis and the evidence in your reasoning so the fix step can
build on it. Do not modify code in this step.
