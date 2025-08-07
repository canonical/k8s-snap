# Canonical Kubernetes documentation

This part of the repository contains the tools and the source for generating
documentation for the Canonical Kubernetes.

The directories are organized like this:

```

├── README.md
├── canonicalk8s
│   ├── _build
│   │   ├── {contains the generated docs}
│   ├──{sphinx build tools for creating the docs for Canonical K8s}
│   ├──{source files for canonicalk8s docs}
```

## Building the docs

This documentation uses the `canonicalk8s/Makefile` to generate HTML docs from
the sources. This can also run specific local tests such as spelling and
link checking.

## Contributing to the docs

Contributions to this documentation are welcome. Generally these follow the
same rules and process as other contributions - modify the docs source and
submit a PR.

## The docs release process

When generating a release for Canonical Kubernetes, it is important to keep the
docs in this folder up to date. Below are a list of steps you must complete
before you can call a release complete from the docs perspective.

### Write the release notes

Create the release notes for the snap or the charm and have them reviewed by the
team and the technical author.

### Update the automated files

We use a single file to update all our install commands across our docs. Update
the `canonicalk8s/_parts/install.md` file with the correct version.

The file contains both snap and charm install commands. Be sure to **only**
update the commands related to the release you are doing.

We also use a file for our substitutions. Update the
`canonicalk8s/reuse/substitutions.yaml` file with the latest version and
channel.

### Update other relevant files

There are certain files that cannot be included in the automated files that
also need to be updated.

- `README.md`
- `docs/canonicalk8s/capi/explanation/in-place-upgrades.md`
- `docs/canonicalk8s/capi/howto/custom-ck8s.md`
- `docs/canonicalk8s/charm/howto/validate.md`
- `docs/canonicalk8s/assets/how-to-epa-maas-cloud-init.md`
- `docs/canonicalk8s/conf.py` (update the `sitemap_url_scheme` to match the
current version)

### Update the release branch

All updates are made to the main branch and then release branch is created from
main. If the release branch has already been created before you are updating
these docs, make sure you back port any changes to the release branch.

### Create a version of the release

Read the Docs versions can be created by the technical author. They are
based on branches. Make sure the branch is set to public and not hidden in order
for it to show up to all users.
