<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: verify

Goal: decide whether the reproduced behaviour is an actual defect or the
documented, intended behaviour of k8s-snap.

You run straight after the reproduction and before any test is written, so this
is the gate that stops the bot investing in a test and a fix for behaviour that
is working as designed. You have the reproduction evidence, not yet a root-cause
diagnosis; judge the behaviour, not the implementation.

## Procedure

1. Compare the observed behaviour against the documentation under `docs/` and
   the command's own help and spec.
2. Check whether the behaviour is a deliberate design choice (a default, a
   guardrail, a documented limitation) rather than a bug.
3. Weigh the evidence both ways and commit to a verdict.

## Return

- `verdict`:
  - `bug`: the behaviour contradicts documented or clearly-intended behaviour.
  - `intended-behavior`: the behaviour matches how k8s-snap is designed to work;
    the issue is not actionable as a code fix.
  - `unclear`: the evidence does not support a confident call.
- `confidence`: `high` | `medium` | `low` for that verdict.

Only `bug` continues to the reproducer step. Both `intended-behavior` and
`unclear` stop the pipeline and park the issue at `triage/needs-human` for a
maintainer, so do not reach for `unclear` as a safe middle ground: it is the
schema default, and treating it as "proceed anyway" would let an empty or
malformed answer spend a cluster run and touch code on an unproven premise.
Commit to `bug` when the evidence supports it, and say why in your reasoning.
