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

<!-- Link to PRs to initialize the release branch (see below) -->
- **PR (release-1.xx)**:
- **PR (moonray/release-1.xx)**:
- **PR (strict/release-1.xx)**:

<!-- Link to PR to initialize auto-update job for the release branch (see below) -->
- **PR**:

#### Actions

The steps are to be followed in-order, each task must be completed by the person specified in **bold**. Do not perform any steps unless all previous ones have been signed-off. The **Reviewer** closes the issue once all steps are complete.

- [ ] **Owner**: Add the assignee and reviewer as assignees to the GitHub issue
- [ ] **Owner**: Ensure that you are part of the ["containers" team](https://launchpad.net/~containers)
- [ ] **Owner**: Ensure that are no [fast-forward PRs](https://github.com/canonical/k8s-snap/pulls) open against the `moonray/main` and `strict/main` branches.
- [ ] **Owner**: Request a new `1.xx` Snapstore track for the snaps similar to the  [snapstore track-request][].
  - #### Post template on https://discourse.charmhub.io/

    **Title:** Request for 1.xx tracks for the k8s snap

    **Category:** store-requests

    **Body:**

      Hi,

      Could we please have the following tracks for k8s-snap?

      - "1.xx"
      - "1.xx-classic"
      - "1.xx-moonray"

      Thank you, $name

- [ ] **Owner**: Create `release-1.xx` branch from latest `master` in k8s-dqlite
  - `git clone git@github.com:canonical/k8s-dqlite.git ~/tmp/release-1.xx`
  - `pushd ~/tmp/release-1.xx`
  - `git switch main`
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
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
  - `popd`
  - `rm -rf ~/tmp/release-1.xx`
- [ ] **Owner**: Create `release-1.xx` branch from latest `main`
  - `git switch main`
  - `git pull`
  - `git checkout -b release-1.xx`
  - `git push origin release-1.xx`
- [ ] **Reviewer**: Ensure `release-1.xx` branch is based on latest changes on `main` at the time of the release cut.
- [ ] **Owner**: Create `moonray/release-1.xx` branch from latest `moonray/main`
  - `git switch moonray/main`
  - `git pull`
  - `git checkout -b moonray/release-1.xx`
  - `git push origin moonray/release-1.xx`
- [ ] **Reviewer**: Ensure `moonray/release-1.xx` branch is based on latest changes on `moonray/main` at the time of the release cut.
- [ ] **Owner**: Create `strict/release-1.xx` branch from latest `strict/main`
  - `git switch strict/main`
  - `git pull`
  - `git checkout -b strict/release-1.xx`
  - `git push origin strict/release-1.xx`
- [ ] **Reviewer**: Ensure `strict/release-1.xx` branch is based on latest changes on `strict/main` at the time of the release cut.
- [ ] **Owner**: Create PR to initialize `release-1.xx` branch:
  - [ ] Update `KUBERNETES_RELEASE_MARKER` to `stable-1.xx` in [/build-scripts/hack/update-component-versions.py][]
  - [ ] Update `master` to `release-1.xx` in [/build-scripts/components/k8s-dqlite/version][]
  - [ ] Update `"main"` to `"release-1.xx"` in [/build-scripts/hack/generate-sbom.py][]
  - [ ] `git commit -m 'Release 1.xx'`
  - [ ] Create PRs against `release-1.xx` with the changes and request review from **Reviewer**. Make sure to update the issue `Information` section with link to the PR.
- [ ] **Reviewer**: Ensure `release-1.xx` PR is merged and builds Kubernetes 1.xx.
- [ ] **Owner**: Create PRs to initialize `moonray/release-1.xx` branch.
  - [ ] `git checkout moonray/release-1.xx`
  - [ ] `git merge release-1.xx`
  - [ ] Create PR against `moonray/release-1.xx` with the changes and request review from **Reviewer**. Make sure to update the issue `Information` section with link to the PR.
- [ ] **Owner**: Create PRs to initialize `strict/release-1.xx` branch.
  - [ ] `git checkout strict/release-1.xx`
  - [ ] `git merge release-1.xx`
  - [ ] Create PR against `strict/release-1.xx` with the changes and request review from **Reviewer**. Make sure to update the issue `Information` section with link to the PR.
- [ ] **Reviewer**: Review and merge PRs to initialize the release branches for `moonray/release-1.xx` and `strict/release-1.xx`.
- [ ] **Owner**: Create PR to initialize `update-components.yaml` job for `release-1.xx` branch:
  - [ ] Add `release-1.xx` in [.github/workflows/update-components.yaml][]
  - [ ] Remove unsupported releases from the list (if applicable, consult with **Reviewer**)
  - [ ] Create PR against `main` with the changes and request review from **Reviewer**. Make sure to update the issue information with a link to the PR.
- [ ] **Owner**: Create launchpad builders for `release-1.xx`
  - [ ] Go to [lp:k8s][] and do **Import now** to pick up all latest changes.
  - [ ] Under **Branches**, select `release-1.xx`, then **Create snap package**
  - [ ] Set **Snap recipe name** to `k8s-snap-1.xx`
  - [ ] Set **Owner** to `Canonical Kubernetes (containers)`
  - [ ] Set **The project that this Snap is associated with** to `k8s`
  - [ ] Set **Series** to Infer from snapcraft.yaml
  - [ ] Set **Processors** to `AMD x86-64 (amd64)` and `ARM ARMv8 (arm64)`
  - [ ] Enable **Automatically build when branch changes**
  - [ ] Enable **Automatically upload to store**
  - [ ] Set **Registered store name** to `k8s`
  - [ ] In **Store Channels**, set **Track** to `1.xx-classic` and **Risk** to `edge`. Leave **Branch** empty
  - [ ] Click **Create snap package** at the bottom of the page.
- [ ] **Owner**: Create launchpad builders for `strict/release-1.xx`
  - [ ] Return to [lp:k8s][].
  - [ ] Under **Branches**, select `strict/release-1.xx`, then **Create snap package**
  - [ ] Set **Snap recipe name** to `k8s-snap-1.xx-strict`
  - [ ] Set **Owner** to `Canonical Kubernetes (containers)`
  - [ ] Set **The project that this Snap is associated with** to `k8s`
  - [ ] Set **Series** to Infer from snapcraft.yaml
  - [ ] Set **Processors** to `AMD x86-64 (amd64)` and `ARM ARMv8 (arm64)`
  - [ ] Enable **Automatically build when branch changes**
  - [ ] Enable **Automatically upload to store**
  - [ ] Set **Registered store name** to `k8s`
  - [ ] In **Store Channels**, set **Track** to `1.xx` and **Risk** to `edge`. Leave **Branch** empty
  - [ ] Click **Create snap package** at the bottom of the page.
- [ ] **Owner**: Create launchpad builders for `moonray/release-1.xx`
  - [ ] Return to [lp:k8s][].
  - [ ] Under **Branches**, select `moonray/release-1.xx`, then **Create snap package**
  - [ ] Set **Snap recipe name** to `k8s-snap-1.xx-moonray`
  - [ ] Set **Owner** to `Canonical Kubernetes (containers)`
  - [ ] Set **The project that this Snap is associated with** to `k8s`
  - [ ] Set **Series** to Infer from snapcraft.yaml
  - [ ] Set **Processors** to `AMD x86-64 (amd64)` and `ARM ARMv8 (arm64)`
  - [ ] Enable **Automatically build when branch changes**
  - [ ] Enable **Automatically upload to store**
  - [ ] Set **Registered store name** to `k8s`
  - [ ] In **Store Channels**, set **Track** to `1.xx-moonray` and **Risk** to `edge`. Leave **Branch** empty
  - [ ] Click **Create snap package** at the bottom of the page.
- [ ] **Reviewer**: Ensure snap recipes are created in [lp:k8s/+snaps][]
  - look for `k8s-snap-1.xx`
  - look for `k8s-snap-1.xx-moonray`
  - look for `k8s-snap-1.xx-strict`

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
[/build-scripts/hack/generate-sbom.py]: ../..//build-scripts/hack/generate-sbom.py
[lp:k8s]: https://code.launchpad.net/~cdk8s/k8s/+git/k8s-snap
[lp:k8s/+snaps]: https://launchpad.net/k8s/+snaps
