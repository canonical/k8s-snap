<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: reproducer

Goal: turn the reproduction you just observed into an end-to-end test that fails
because of this bug, and prove that it fails.

That failing test is the specification the fix must satisfy. A test that already
passes proves nothing about a later change, so this step is only finished once
you have watched it fail for the reported reason.

> Uses: `local-cluster`

## Procedure

1. You are already in a dedicated worktree with `triage/fix-<issue>` checked
   out. Never create, switch or reset branches; just commit here.
2. `tests/integration/AGENTS.md` documents the suite: the markers, the mandatory
   tags, the `Instance` API and the helpers that already exist. Read it before
   writing anything, and follow the closest existing test (`test_dns.py` for
   DNS, and so on) rather than inventing conventions. Reuse the helpers in
   `tests/integration/tests/test_util/`; most of what you need is there.
3. Declare the cluster shape and tag the test. `node_count` **defaults to 1**, so
   a missing marker silently tests a single-node cluster and quietly proves
   nothing about a multi-node bug. A test with no `tags` marker is worse: the
   suite's `conftest.py` fails it at collection, which looks like a red test but
   proves nothing at all.
   ```python
   from test_util import harness, tags

   @pytest.mark.node_count(2)   # the reporter's 1 control-plane + 1 worker
   @pytest.mark.tags(tags.PULL_REQUEST)
   def test_coredns_replicas_spread_across_nodes(instances: List[harness.Instance]):
       ...
   ```
4. Assert **the symptom the reporter saw**, not the mechanism you expect the fix
   to use. Test the observable behaviour end to end: which node each pod landed
   on, that a deployment stops being restarted, that a command now succeeds.
   Asserting an implementation detail (a specific field in a manifest, a
   particular flag) bakes one solution into the test: a different, correct fix
   would leave it red, and the test would pass while the reported symptom
   persists. Assert what should happen, not what is happening today, so the
   test fails now and passes once the bug is gone.
5. Run only your test, as `tests/integration/AGENTS.md` describes, against the
   snap under test.
6. Confirm it fails **for the reported reason**. An `ImportError`, a fixture
   error, a typo or a timeout during bootstrap is not a reproduction: fix your
   test and run it again until the failure is the bug itself.
7. Commit the test on its own, with no product-code changes:
   ```bash
   git add tests/integration/tests/<file>
   git commit -m "test: reproduce <short description> (#<issue>)"
   ```
   This commit matters even if the later fix stage fails: the orchestrator
   pushes the branch either way, so a maintainer always gets the reproducer.

## Return

- `test_path`: the file you added or extended, e.g.
  `tests/integration/tests/test_dns.py`.
- `test_selector`: the `-k` expression that runs just this test.
- `fails_before_fix`: true only if you ran it, watched it fail for the
  reported reason, **and** completed step 7's commit. The orchestrator
  trusts this field alone to decide whether to push the branch and open a
  PR -- a true here with no commit behind it silently loses the reproducer
  for good.
- `failure_output`: the assertion or error lines proving that.

If you cannot get a test to fail for the reported reason, return
`fails_before_fix: false` and say why. Triage stops there and no product code is
touched, which is the correct outcome: without a red test there is nothing to
show a fix works.
