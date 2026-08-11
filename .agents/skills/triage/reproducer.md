<!--
Copyright 2026 Canonical, Ltd.
-->
# Step: reproducer

Goal: turn the reproduction you just observed into an end-to-end test that fails
because of this bug, and prove that it fails.

That failing test is the specification the fix must satisfy. A test that already
passes proves nothing about a later change, so this step is only finished once
you have watched it fail for the reported reason.

## The cluster, and the test's own cluster

These are two different things, and mixing them up wastes a lot of time:

- `hack/cluster-up.sh` gave you a cluster (`k8s-triage-*`) to poke at by hand.
- `tests/integration` builds and tears down **its own** instances from the
  `harness.Instance` fixtures. Your test must not assume the manual cluster.

They use separate names and LXD profiles, so they coexist. On a small machine
they compete for CPU and disk: `hack/cluster-up.sh --destroy` first if the test
run is starved.

## Procedure

1. You are already in a dedicated worktree with `triage/fix-<issue>` checked
   out. Never create, switch or reset branches; just commit here.
2. Find the closest existing test under `tests/integration/tests` and follow its
   conventions rather than inventing your own (`test_dns.py` for DNS, and so
   on). Reuse the helpers in `tests/integration/tests/test_util/util.py`.
3. Declare the cluster shape the reporter used. `node_count` **defaults to 1**,
   so a missing marker silently tests a single-node cluster and quietly proves
   nothing about a multi-node bug:
   ```python
   @pytest.mark.node_count(2)   # the reporter's 1 control-plane + 1 worker
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
5. Run only your test:
   ```bash
   export TEST_SNAP=$PWD/k8s.snap TEST_SUBSTRATE=lxd
   cd tests/integration && tox -e integration -- -k <test_name>
   ```
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
