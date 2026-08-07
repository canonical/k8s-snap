<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: fix

Goal: make the failing end-to-end test pass with the smallest correct change.

The test from the reproducer step is already committed on `triage/fix-<issue>`
and is known to fail. It is the specification. **Do not weaken, skip, delete or
rewrite it to get a green run.** If the test is wrong, say so and stop; do not
quietly change what is being asserted.

## Procedure

1. You are in your own worktree, on `triage/fix-<issue>`, with the reproducer
   test already committed. Confirm before changing anything:
   ```bash
   git status --short && git log --oneline -1
   ```
2. Make the **smallest** change that addresses the root cause from the
   diagnosis. Match surrounding style; no unrelated refactors, no new
   dependencies. Self-explanatory code over comments.
3. If the root cause is in an adjacent repository (`k8sd`, `k8s-snap-api`),
   follow `CONTRIBUTING.md`: clone it, change it there, and add a `replace`
   directive so this build picks up your checkout. A separate PR in that
   repository is required; note that in your reasoning so the orchestrator can
   say so on the issue.
4. Rebuild the snap so the artefact under test actually contains your change,
   then re-run the one test. The build **must run from the primary checkout**
   (the real directory, not this worktree): snapcraft copies every project file
   into the LXD build instance, and the sshfs mount that backs this worktree
   refuses those copies with a permission error. Use the `REPO_ROOT` variable
   that `hack/cluster-up.sh` derives — or find the primary tree from the git
   worktree list:
   ```bash
   # Find the primary checkout (first entry in worktree list)
   PRIMARY="$(git worktree list --porcelain | sed -n '1s/^worktree //p')"
   # Reuse an existing snap rather than spending tens of minutes rebuilding
   SNAP="$PRIMARY/k8s.snap"
   if [ ! -f "$SNAP" ]; then
     (cd "$PRIMARY" && snapcraft --use-lxd && mv k8s_*.snap k8s.snap)
   fi
   export TEST_SNAP="$SNAP" TEST_SUBSTRATE=lxd
   cd tests/integration && tox -e integration -- -k <test_name>
   ```
   A change to `k8sd` only reaches the cluster through a rebuild. Skipping it
   means your green run proves nothing. The build takes tens of minutes.
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
