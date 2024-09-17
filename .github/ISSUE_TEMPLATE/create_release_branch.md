---
name: "[Runbook] Create release branch"
about: Create a new branch for a new stable Kubernetes release
---

#### Summary

Make sure to follow the steps below and ensure all actions are completed and signed-off by one team member.

#### Information

<!-- Replace with the version to create the branch for, e.g. 1.28 -->
- **K8s version**: 1.xx

<!-- Set this to the name of the person responsible for running the release tasks, e.g. @neoaggelos -->
- **Owner**: `who plans to do the work`

<!-- Set this to the name of the team-member that will sign-off the tasks -->
- **Reviewer**:  `who plans to review the work`

<!-- Link to PR to initialize the release branch (see below) -->
- **PR**: https://github.com/canonical/k8s-snap/pull/`<int>`

<!-- Link to PR to initialize auto-update job for the release branch (see below) -->
- **PR**: https://github.com/canonical/k8s-snap/pull/`<int>`

#### Actions

The steps are to be followed in-order, each task must be completed by the person specified in **bold**. Do not perform any steps unless all previous ones have been signed-off. The **Reviewer** closes the issue once all steps are complete.

- [ ] **Owner**: Add the assignee and reviewer as assignees to the GitHub issue
- [ ] **Owner**: Ensure that you are part of the ["containers" team](https://launchpad.net/~containers)
- [ ] **Owner**: Request a new `1.xx` Snapstore track for the snaps similar to the  [snapstore track-request][].
  - #### Post template on https://discourse.charmhub.io/

    **Title:** Request for 1.xx tracks for the k8s snap

    **Category:** store-requests

    **Body:**

      Hi,

      Could we please have a track "1.xx-classic" and "1.xx" for the respective K8s snap release?

      Thank you, $name

- [ ] **Owner**: Create `release-1.xx` branch from latest `main`
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
- [ ] **Owner**: Create `release-1.xx` branch from latest `master` in k8s-dqlite
  - `git clone git@github.com:canonical/k8s-dqlite.git ~/tmp/release-1.xx`
  - `pushd ~/tmp/release-1.xx`
  - `git switch master`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
  - `popd`
  - `rm -rf ~/tmp/release-1.xx`
- [ ] **Owner**: Create `release-1.xx` branch from latest `main` in cilium-rocks
  - `git clone git@github.com:canonical/cilium-rocks.git ~/tmp/release-1.xx`
  - `pushd ~/tmp/release-1.xx`
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
  - `popd`
  - `rm -rf ~/tmp/release-1.xx`
- [ ] **Owner**: Create `release-1.xx` branch from latest `main` in coredns-rock
  - `git clone git@github.com:canonical/coredns-rock.git ~/tmp/release-1.xx`
  - `pushd ~/tmp/release-1.xx`
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
  - `popd`
  - `rm -rf ~/tmp/release-1.xx`
- [ ] **Owner**: Create `release-1.xx` branch from latest `main` in metrics-server-rock
  - `git clone git@github.com:canonical/metrics-server-rock.git ~/tmp/release-1.xx`
  - `pushd ~/tmp/release-1.xx`
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
  - `popd`
  - `rm -rf ~/tmp/release-1.xx`
- [ ] **Owner**: Create `release-1.xx` branch from latest `main` in rawfile-localpv
  - `git clone git@github.com:canonical/rawfile-localpv.git ~/tmp/release-1.xx`
  - `pushd ~/tmp/release-1.xx`
  - `git switch rockcraft`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
  - `popd`
  - `rm -rf ~/tmp/release-1.xx`
- [ ] **Reviewer**: Ensure `release-1.xx` branch is based on latest changes on `main` at the time of the release cut.
- [ ] **Owner**: Create PR to initialize `release-1.xx` branch:
  - [ ] Update `KUBERNETES_RELEASE_MARKER` to `stable-1.xx` in [/build-scripts/hack/update-component-versions.py][]
  - [ ] Update `"main"` to `"release-1.xx"` in [/build-scripts/hack/generate-sbom.py][]
  - [ ] `git commit -m 'Release 1.xx'`
  - [ ] Create PR against `release-1.xx` with the changes and request review from **Reviewer**. Make sure to update the issue `Information` section with a link to the PR.
- [ ] **Reviewer**: Review and merge PR to initialize branch.
- [ ] **Owner**: Create PR to initialize `update-components.yaml` job for `release-1.xx` branch:
  - [ ] Add `release-1.xx` in [.github/workflows/update-components.yaml][]
  - [ ] Remove unsupported releases from the list (if applicable, consult with **Reviewer**)
  - [ ] Create PR against `main` with the changes and request review from **Reviewer**. Make sure to update the issue information with a link to the PR.
- [ ] **Reviewer**: On merge, confirm [Auto-update strict branch] action runs to completion and that the `autoupdate/release-1.xx-*` flavor branches are created.
   - [ ] autoupdate/release-1.xx-strict
   - [ ] autoupdate/release-1.xx-moonray
- [ ] **Owner**: Create launchpad builders for `release-1.xx` and flavors
  - [ ] Run the [Confirm Snap Builds][] Action
- [ ] **Reviewer**: Ensure snap recipes are created in [lp:k8s/+snaps][]
  - [ ] look for `k8s-snap-1.xx-classic`
  - [ ] look for `k8s-snap-1.xx-strict`
  - [ ] look for `k8s-snap-1.xx-moonray`
  - [ ] make sure each is "Authorized for Store Upload"

#### After release

- [ ] **Owner** follows up with the **Reviewer** and team about things to improve around the process.
- [ ] **Owner**: After a few weeks of stable CI, update default track to `1.xx/stable` via
  - On the snap [releases page][], select `Track` > `1.xx`


<!-- LINKS -->
[Auto-update strict branch]: https://github.com/canonical/k8s-snap/actions/workflows/strict.yaml
[snapstore track-request]: https://forum.snapcraft.io/t/tracks-request-for-k8s-snap/39122/2
[releases-page]: https://snapcraft.io/k8s/releases
[.github/workflows/cla.yaml]: ../workflows/cla.yaml
[.github/workflows/cron-jobs.yaml]: ../workflows/cron-jobs.yaml
[.github/workflows/go.yaml]: ../workflows/go.yaml
[.github/workflows/integration.yaml]: ..workflows/integration.yaml
[.github/workflows/python.yaml]: ../workflows/python.yaml
[.github/workflows/sbom.yaml]: ../workflows/sbom.yaml
[.github/workflows/strict-integration.yaml]: ../workflows/strict-integration.yaml
[.github/workflows/strict.yaml]: ../workflows/strict.yaml
[.github/workflows/update-components.yaml]: ../workflows/update-components.yaml
[/build-scripts/components/hack/update-component-versions.py]: ../../build-scripts/components/hack/update-component-versions.py
[/build-scripts/components/k8s-dqlite/version]: ../../build-scripts/components/k8s-dqlite/version
[/build-scripts/hack/generate-sbom.py]: ../../build-scripts/hack/generate-sbom.py
[lp:k8s]: https://code.launchpad.net/~cdk8s/k8s/+git/k8s-snap
[lp:k8s/+snaps]: https://launchpad.net/k8s/+snaps
[Confirm Snap Builds]: https://github.com/canonical/canonical-kubernetes-release-ci/actions/workflows/create-release-branch.yaml
