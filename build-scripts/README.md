# Parts directory

This directory contains the build scripts for Go components built into k8s.

The directory structure looks like this:

```
build-scripts/
    build-component.sh              <-- runs as `build-component.sh $component_name`
                                        - checks out the git repository
                                        - runs the `pre-patch.sh` script (if any)
                                        - applies the patches (if any)
                                        - runs the `build.sh` script to build the component
    components/
        $component_name/
            repository              <-- git repository to clone
            version                 <-- repository tag or commit to checkout
            build.sh                <-- runs as `build.sh $output $version`
                                        first argument is the output directory where
                                        binaries should be placed, second is the component version
            pre-patch.sh            <-- runs as `pre-patch.sh`. takes any action needed before applying
                                        the component patches
            patches/                <-- list of patches to apply after checkout (see section below)
                ...
```

## Applying patches

Most K8s components are retrieved from an upstream source (specified in the `repository`), with a specific tag (specified in `version`), have some patches applied to them (from the `patches/` directory) and are then built (using `build.sh`).

This section explains the directory format for the `patches` directory.

Our patches do not frequently change between versions, but they do have to be rebased from time to time, which breaks compatibility with older versions. For that reason, we maintain a set of patches for each version that introduces a breaking change. Consider the following directory structure for the Kubernetes component.

```
patches/default/0.patch
patches/v1.27.0/a.patch
patches/v1.27.0/b.patch
patches/v1.27.4/c.patch
patches/v1.28.0/d.patch
patches/v1.28.0-beta.0/e.patch
```

The Kubernetes version to build may be decided dynamically while building the snap, or be pinned to a specified version. The following table shows which patches we would apply depending on the Kubernetes version that we build:

| Kubernetes version | Applied patches         | Explanation                                                                                |
| ------------------ | ----------------------- | ------------------------------------------------------------------------------------------ |
| `v1.27.0`          | `a.patch` and `b.patch` |                                                                                            |
| `v1.27.1`          | `a.patch` and `b.patch` | In case there is no exact match, find the most recent older version                        |
| `v1.27.4`          | `c.patch`               | Older patches are not applied                                                              |
| `v1.27.12`         | `c.patch`               | In semver, `v1.27.12 > v1.27.4` so we again must get the most recent patches               |
| `v1.28.0-rc.0`     | `d.patch`               | Extra items from semver are ignored, so we can define the `v1.28.0` patch and be done      |
| `v1.28.0-beta.0`   | `e.patch`               | Extra items from semver are ignored, but due to exact match this patch is used instead     |
| `v1.28.0`          | `d.patch`               | Extra items from semver are ignored, so we can define the `v1.28.0` patch and be done      |
| `v1.28.4`          | `d.patch`               | Picks the patches from the stable versions only, not from beta                             |
| `v1.29.1`          | `d.patch`               | Uses patches from most recent version, even if on a different minor                        |
| `hack/branch`      | `0.patch`               | If not semver and no match, any patches from the `default/` directory are applied (if any) |

Same logic applies for all other components as well.

### Testing which patches would be applied

You can verify which set of patches would be applied in any case using the `print-patches-for.py` script directly:

```bash
$ ./build-scripts/print-patches-for.py kubernetes v1.27.4
/home/ubuntu/k8s/build-scripts/components/kubernetes/patches/v1.27.4/0000-Kubelite-integration.patch
$ ./build-scripts/print-patches-for.py kubernetes v1.27.3
/home/ubuntu/k8s/build-scripts/components/kubernetes/patches/v1.27.0/0000-Kubelite-integration.patch
/home/ubuntu/k8s/build-scripts/components/kubernetes/patches/v1.27.0/0001-Unix-socket-skip-validation-in-component-status.patch
$ ./build-scripts/print-patches-for.py kubernetes v1.28.1
/home/ubuntu/k8s/build-scripts/components/kubernetes/patches/v1.28.0/0001-Set-log-reapply-handling-to-ignore-unchanged.patch
/home/ubuntu/k8s/build-scripts/components/kubernetes/patches/v1.28.0/0000-Kubelite-integration.patch
```

### How to add support for newer versions

When a new release comes out which is no longer compatible with the existing latest patches, simply create a new directory under `patches/` with the new version number. This ensures that previous versions will still work, and newer ones will pick up the fixed patches.

## Updating Component Versions

The `hack/update-component-versions.py` script automates the process of checking for and updating component versions. It supports two modes of operation:

### Regular Update Mode (Default)

```bash
./build-scripts/hack/update-component-versions.py
```

This mode updates component version files directly by:
- Fetching the latest Kubernetes version from upstream
- Updating CNI to match Kubernetes dependencies
- Updating containerd to the latest release in the configured branch
- Updating runc to match containerd's requirements (with upstream patch detection)
- Updating Helm to the latest version
- Updating Go version to match Kubernetes requirements

### JSON Output Mode (for CI/CD)

```bash
./build-scripts/hack/update-component-versions.py --json-output
```

This mode is used by the GitHub Actions workflow. It:
1. Checks all components for available updates
2. Detects **independent upstream patches** - when dependencies have newer patch versions than what parent components require
3. Applies all updates to version files
4. Returns structured JSON with PR title and description

**Example JSON output:**

```json
{
  "title": "Update kubernetes, containerd, and runc",
  "description": "## Component Version Updates\n\n- **kubernetes**: v1.31.0 → v1.31.1\n- **containerd**: v1.7.28 → v1.7.29\n- **runc**: v1.3.0 → v1.3.3\n\n## ⚠️ Independent Patch Updates\n\nThe following updates include patches newer than what parent components require. Please verify compatibility before merging:\n\n- **runc**: v1.3.0 → v1.3.3 (upstream has newer patches than parent containerd v1.7.29 requires)"
}
```

The workflow uses this to create a single PR with all component updates, including warnings for independent patches that need manual review.

### Key Features

- **Dynamic Detection**: No hardcoded EoL lists. The script dynamically compares upstream versions with parent requirements.
- **Semantic Versioning**: Uses proper version comparison to detect patch-level updates within the same major.minor version.
- **Warning Annotations**: Independent updates that diverge from parent requirements are clearly marked with warnings for manual review.
- **Batch Updates**: All component updates are included in a single PR for easier review and testing.
- **GitHub Actions Integration**: The `.github/workflows/update-components.yaml` workflow uses this feature to automatically create PRs.

### Safeguards

The script includes several safeguards:
- Skips updates if the upstream version is older or identical to the current version
- Skips updates if no upstream version information is available
- Logs clear messages indicating when a dependency update diverges from its parent's requirement
- Gracefully handles network failures and missing repositories
