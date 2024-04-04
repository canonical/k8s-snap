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
- **Owner**:

<!-- Set this to the name of the team-member that will sign-off the tasks -->
- **Reviewer**:

<!-- Link to PR to initialize the release branch (see below) -->
- **PR**:

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

- [ ] **Owner**: Create `release-1.xx-strict` branch from latest `autoupdate/strict`
  - `git switch autoupdate/strict`
  - `git pull`
  - `git checkout -b release-1.xx-strict`
  - `git push origin release-1.xx-strict`
- [ ] **Owner**: Create `release-1.xx` branch from latest `main`
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
- [ ] **Owner**: Create `1.xx` branch from latest `master` in k8s-dqlite
  - `git clone git@github.com:canonical/k8s-dqlite.git ~/tmp/release-1.xx`
  - `pushd ~/tmp/release-1.xx`
  - `git switch main`
  - `git pull`
  - `git checkout -b 1.xx`
  - `git push origin 1.xx`
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
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
  - `popd`
  - `rm -rf ~/tmp/release-1.xx`
- [ ] **Reviewer**: Ensure `release-1.xx` branch is based on latest changes on `main` at the time of the release cut.
- [ ] **Reviewer**: Ensure `release-1.xx-strict` branch is based on latest changes on `autoupdate/strict` at the time of the release cut.
- [ ] **Owner**: Create PR to initialize `release-1.xx` branch:
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/cla.yaml][]
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/cron-jobs.yaml][]
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/go.yaml][]
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/integration.yaml][]
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/python.yaml][]
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/sbom.yaml][]
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/strict-integration.yaml][]
  - [ ] Update branch from `main` to `release-1.xx` in [.github/workflows/strict.yaml][]
  - [ ] Update branch from `autoupdate/strict` to `release-1.xx-strict` in [.github/workflows/go.yaml][]
  - [ ] Update branch from `autoupdate/strict` to `release-1.xx-strict` in [.github/workflows/strict-integration.yaml][]
  - [ ] Update branch from `autoupdate/strict` to `release-1.xx-strict` in [.github/workflows/python.yaml][]
  - [ ] Update branch from `autoupdate/strict` to `release-1.xx-strict` in [.github/workflows/integration.yaml][]
  - [ ] Update branch from `autoupdate/strict` to `release-1.xx-strict` in [.github/workflows/sbom.yaml][]
  - [ ] Update branch from `autoupdate/strict` to `release-1.xx-strict` in [.github/workflows/strict.yaml][]
  - [ ] Update `KUBE_TRACK` to `1.xx` in [/build-scripts/components/kubernetes/version.sh][]
  - [ ] Update `master` to `1.xx` in [/build-scripts/components/k8s-dqlite/version.sh][]
  - [ ] Update `"main"` to `"release-1.xx"` in [/build-scripts/hack/generate-sbom.py][]
  - [ ] `git commit -m 'Release 1.xx'`
  - [ ] Create PR with the changes and request review from **Reviewer**. Make sure to update the issue `Information` section with a link to the PR.
- [ ] **Reviewer**: Review and merge PR to initialize branch.
- [ ] **Reviewer**: On merge, confirm [Auto-update strict branch] action runs to completion
- [ ] **Owner**: Create launchpad builders for `release-1.xx`
  - [ ] Go to [lp:k8s][] and do **Import now** to pick up all latest changes.
  - [ ] Under **Branches**, select `release-1.xx`, then **Create snap package**
  - [ ] Set **Snap recipe name** to `k8s-snap-1.xx`
  - [ ] Set **Owner** to `Canonical Kubernetes (containers)`
  - [ ] Set **The project that this Snap is associated with** to `k8s`
  - [ ] Set **Series** to Infer from snapcraft.yaml
  - [ ] Set **Processors** to `AMD x86-64 (amd64)` and `ARM ARMv8 (arm64)`
  - [ ] Enable **Automatically upload to store**
  - [ ] Set **Registered store name** to `k8s`
  - [ ] In **Store Channels**, set **Track** to `1.xx-classic` and **Risk** to `edge`. Leave **Branch** empty
  - [ ] Click **Create snap package** at the bottom of the page.
- [ ] **Owner**: Create launchpad builders for `release-1.xx-strict`
  - [ ] Return to [lp:k8s][].
  - [ ] Under **Branches**, select `release-1.xx-strict`, then **Create snap package**
  - [ ] Set **Snap recipe name** to `k8s-snap-1.xx-strict`
  - [ ] Set **Owner** to `Canonical Kubernetes (containers)`
  - [ ] Set **The project that this Snap is associated with** to `k8s`
  - [ ] Set **Series** to Infer from snapcraft.yaml
  - [ ] Set **Processors** to `AMD x86-64 (amd64)` and `ARM ARMv8 (arm64)`
  - [ ] Enable **Automatically upload to store**
  - [ ] Set **Registered store name** to `k8s`
  - [ ] In **Store Channels**, set **Track** to `1.xx` and **Risk** to `edge`. Leave **Branch** empty
  - [ ] Click **Create snap package** at the bottom of the page.
- [ ] **Reviewer**: Ensure snap recipes are created in [lp:k8s/+snaps][]
  - look for `k8s-snap-1.xx`
  - look for `k8s-snap-1.xx-strict`

#### After release

**Owner** follows up with the **Reviewer** and team about things to improve around the process.

<!-- LINKS -->
[Auto-update strict branch]: https://github.com/canonical/k8s-snap/actions/workflows/strict.yaml
[snapstore track-request]: https://forum.snapcraft.io/t/tracks-request-for-k8s-snap/39122/2
[.github/workflows/cla.yaml]: ../workflows/cla.yaml
[.github/workflows/cron-jobs.yaml]: ../workflows/cron-jobs.yaml
[.github/workflows/go.yaml]: ../workflows/go.yaml
[.github/workflows/integration.yaml]: ..workflows/integration.yaml
[.github/workflows/python.yaml]: ../workflows/python.yaml
[.github/workflows/sbom.yaml]: ../workflows/sbom.yaml
[.github/workflows/strict-integration.yaml]: ../workflows/strict-integration.yaml
[.github/workflows/strict.yaml]: ..workflows/strict.yaml
[/build-scripts/components/kubernetes/version.sh]: ../../build-scripts/components/kubernetes/version.sh
[/build-scripts/components/k8s-dqlite/version.sh]: ../../build-scripts/components/k8s-dqlite/version.sh
[/build-scripts/hack/generate-sbom.py]: ../..//build-scripts/hack/generate-sbom.py
[lp:k8s]: https://code.launchpad.net/~cdk8s/k8s/+git/k8s-snap
[lp:k8s/+snaps]: https://launchpad.net/k8s/+snaps