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

`intended-behavior` stops the pipeline and parks the issue as not-actionable.
`bug` and `unclear` both continue to the reproducer step: capturing the
behaviour in a failing test is useful either way, and a maintainer reviewing the
draft PR sees the verdict and can judge whether the expectation is right.
