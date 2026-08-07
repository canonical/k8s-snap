<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: diagnose

Goal: given an end-to-end test that fails because of this bug, locate the most
likely root cause and state how confident you are.

You run after the reproducer step, so you are not chasing a description: there
is a command that fails, and its output is in the report. Start from that.

## Procedure

1. Re-read the reproduction evidence and the test failure in the report.
2. Trace from the observed failure to the responsible code. Remember that this
   repository holds packaging, hooks and tests, not Go source (see `SKILL.md`):
   - daemon and CLI behaviour lives in `k8sd`
     (https://github.com/canonical/k8sd, pinned by
     `build-scripts/components/k8sd/{repository,version}`). Clone it into your
     per-issue scratch, never into the checkout root:
     ```bash
     git clone --depth 1 --branch "$(cat build-scripts/components/k8sd/version)" \
       "$(cat build-scripts/components/k8sd/repository)" .triage/issue-<n>/k8sd
     ```
     Building it locally needs `libdqlite-dev`; if that is missing, read the
     source rather than fighting the build, and validate through the snap
     rebuild in the fix step.
   - the API contract lives in `k8s-snap-api`
     (https://github.com/canonical/k8s-snap-api)
   - packaging, service definitions and hooks live under `snap/`, `hooks/`,
     `k8s/` and `build-scripts/` in this repository
3. Correlate log lines, service names, and args files
   (`/var/snap/k8s/common/args/*`) with the code path. The node names from the
   reproduction (`k8s-triage-*`) still exist, so you can read live state:
   ```bash
   lxc exec k8s-triage-cp1 -- k8s kubectl logs -n kube-system <pod>
   ```
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
