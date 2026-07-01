# build-scripts

Builds each bundled component from upstream source and stages binaries for snapcraft.

## Component Structure

Every component lives under `build-scripts/components/<name>/`:

```
repository       upstream git URL to clone
version          tag or commit to build
build.sh         build script, called as: build.sh <output_dir> <version>
pre-patch.sh     optional hook run before patches are applied
patches/         version-tagged patch directories; optional (see below)
```

Entry point is `build-component.sh <name>`, which: clones the repo at `version`, runs
`pre-patch.sh` (if present), applies the matching patch set, then runs `build.sh`.

## Patching

Patches are organized into version-tagged subdirectories:

```
patches/default/      fallback when version is not semver and no version dir matches (optional)
patches/v1.31.0/      applied for v1.31.x
patches/v1.32.0/      applied for v1.32.x (supersedes v1.31.0 for v1.32+)
```

Resolution algorithm: find the most recent directory whose version is `<=` the build version.
Extra semver labels (e.g. `-beta.0`) are stripped before comparison, but exact-match entries
take priority. A non-semver `version` value (e.g. a branch name) falls back to `default/` if
that directory exists; otherwise no patches are applied.

To verify which patches would apply before building:

```
./build-scripts/print-patches-for.py <component> <version>
```

When a new upstream release breaks an existing patch, create a new versioned directory rather
than editing the existing one — this preserves patch applicability for older versions.

## Adding or Updating a Component

1. Update `components/<name>/version` with the new tag.
2. If the patch no longer applies cleanly, create `components/<name>/patches/<new_version>/`
   with the rebased patch(es).
3. Run `print-patches-for.py <name> <new_version>` to confirm the right set is selected.
4. Build locally: `snapcraft --use-lxd` (or override-build the part directly in LXD).
