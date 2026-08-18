<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: fix

Goal: make the failing end-to-end test pass with the smallest correct change.

The test from the reproducer step is already committed on `triage/fix-<issue>`
and is known to fail. It is the specification. **Do not weaken, skip, delete or
rewrite it to get a green run.** If the test is wrong, say so and stop; do not
quietly change what is being asserted.

> Uses: `local-cluster`

## Procedure

1. You are in your own worktree, on `triage/fix-<issue>`, with the reproducer
   test already committed. Confirm before changing anything:
   ```bash
   git status --short && git log --oneline -1
   ```
2. Make the **smallest** change that addresses the root cause from the
   diagnosis. Match surrounding style; no unrelated refactors, no new
   dependencies. Self-explanatory code over comments.
3. If the root cause is in an adjacent repository, change it there, in the clone
   in your scratch directory, and point this build at that clone. This
   repository is not a Go module, so there is no `replace` directive to add
   here: the build resolves a component purely from its two pin files, so
   temporarily aim them at your clone and rebuild.
   ```bash
   echo "$PWD/.triage/issue-<n>/k8sd" > build-scripts/components/k8sd/repository
   git -C .triage/issue-<n>/k8sd rev-parse --abbrev-ref HEAD \
     > build-scripts/components/k8sd/version
   ```
   Those two files are product code: never commit them pointing at a local path.
   Restore them before you commit, and put the real change in a PR against that
   repository, saying so in your reasoning so the orchestrator says so on the
   issue. (`CONTRIBUTING.md`'s `replace` workflow is a different case: a Go
   module edit *inside* `k8sd` to build it against a local `k8s-snap-api`.)
4. Rebuild the snap so the artefact under test actually contains your change,
   then re-run the one test. A change to a Go component reaches the cluster only
   through a rebuild: a green run without one proves nothing. The build costs
   tens of minutes, and it must run from the primary checkout rather than this
   worktree.
5. Commit the fix as its own commit, separate from the test. Stage only the
   files you changed: the checkout may carry unrelated work in progress, and
   `commit -a` would sweep it into the bot's PR.
   ```bash
   git status --short          # confirm what is yours
   git add <the files you edited>
   git commit -m "fix: <concise description>"
   ```
   Do not open the PR yourself; the orchestrator pushes the branch and opens a
   draft PR.

## Return

- `fixed`: true only if the change is complete **and** you watched the
  previously failing test pass against the rebuilt snap.
- `commit_message`: the concise commit subject you used.

If you cannot produce a confident, test-backed fix, return `fixed: false` with a
short reason. This is a normal outcome, not a failure to hide: the branch still
carries the failing test, the orchestrator pushes it and opens a draft PR, and a
maintainer picks it up from a reproducible starting point. Leaving a correct
red test behind is worth more than a speculative fix.

### When you have a fix but cannot verify it

Sometimes you reach step 4 with a change you are confident addresses the
diagnosed root cause, but the **rebuild or test tooling itself** fails, not
your fix: `snapcraft --use-lxd` refuses to run (permission/confinement
errors), the LXD cluster from step 4 becomes unreachable, a required tool is
missing, or similar. This is different from being unsure whether the change
is correct -- do not use this for that. If you cannot tell whether the fix
is right, treat it as no confident fix (above), not this.

When it is specifically the *infrastructure* that failed:

1. Still commit the fix as its own commit (step 5) -- do not discard a
   diagnosed, plausible change just because you could not watch it run.
2. Return `fixed: false`, `verification_blocked: true`, and
   `blocked_reason`: one short phrase naming what failed (e.g. `"snapcraft
   --use-lxd: permission denied"`, `"cluster unreachable after rebuild"`).

The orchestrator opens a draft PR that says plainly this fix is unverified
and needs manual confirmation -- it is never presented as a working fix.
Getting a real diagnosis and candidate in front of a maintainer beats
discarding both because the last step could not run.
