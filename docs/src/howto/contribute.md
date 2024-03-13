# How to contribute to Canonical Kubernetes

% Include content from [../CONTRIBUTING.md](../CONTRIBUTING.md)
```{include} ../CONTRIBUTING.md
    :start-after: <!-- Include start contributing -->
    :end-before: <!-- Include end contributing -->
```

## Contribute to the code

Canonical Kubernetes is shipped as a snap package. To contribute to the code,
you should first make sure you can build and test the snap locally.

### Build the snap

To build the snap locally, you will need the following:

 - The latest, snap-based version of LXD (see the [install guide here][install
   lxd])
 - The Snapcraft build tool, for building the snap (see the [Snapcraft
   documentation][]).


Clone the [github repository for the k8s snap][code repo] and then open a terminal in that
directory. Run the command:

```
snapcraft --use-lxd
```

This will launch an LXD container and use it to build a version of the snap.
This will take some time as the build process fetches dependencies, stages the
‘parts’ of the snap and creates the snap package itself. The snap itself will
be fetched from the build environment and placed in the local project
directory. Note that the LXD container used for building will be stopped, but
not deleted. This is in case there were any errors or artefacts you may wish to
inspect. 

### Install the snap

The snap can then be installed locally by using the ‘–dangerous’ option. This
is a safeguard to make sure the user is aware that the snap is not signed by
the snap store, and is not confined:

```
sudo snap install k8s_v1.29.2_multi.snap --dangerous --classic
```

```{note} You will not be able to install this snap if there is already a k8s snap installed on your system.
```

Once you have verified the current snap build works, it can be removed with:

```
sudo snap remove k8s --purge
```

The `purge` option is recommended when iterating over code changes, as it also
removes all the installed artefacts which may be associated with the snap.

Now you can iterate over changes to the snap, rebuild and test.

As noted previously, the LXD container used for building is not removed and
will be reused by subsequent build instructions. When you are satisfied it is
no longer needed, this container can be removed:

```
lxc delete snapcraft-microk8s
```

### Contribute changes

We welcome any improvements and bug-fixes to the Canonical Kubernetes code.
Once you have tested your changes, please make a pull request on the [code
repository][code repo] and we will review it as soon as possible.


## Contribute to the documentation

Our aim is to provide easy-to-understand documentation on all aspects of Canonical Kubernetes, so we greatly appreciate your feedback and contributions.

The source of the documentation and the system used to build it are included in the [main repository for the Canonical Kubernetes snap][code repo]. The method for contributing changes to the docs is similar to 

We want LXD to be as easy and straight-forward to use as possible.
Therefore, we aim to provide documentation that contains the information that users need to work with LXD, that covers all common use cases, and that answers typical questions.

You can contribute to the documentation in various different ways.
We appreciate your contributions!

Typical ways to contribute are:

- Add or update documentation for new features or feature improvements that you contribute to the code.
  We'll review the documentation update and merge it together with your code.
- Add or update documentation that clarifies any doubts you had when working with the product.
  Such contributions can be done through a pull request or through a post in the [Tutorials](https://discourse.ubuntu.com/c/lxd/tutorials/146) section on the forum.
  New tutorials will be considered for inclusion in the docs (through a link or by including the actual content).
- To request a fix to the documentation, open a documentation issue on [GitHub](https://github.com/canonical/lxd/issues).
  We'll evaluate the issue and update the documentation accordingly.
- Post a question or a suggestion on the [forum](https://discourse.ubuntu.com/c/lxd/126).
  We'll monitor the posts and, if needed, update the documentation accordingly.
- Ask questions or provide suggestions in the `#lxd` channel on [IRC](https://web.libera.chat/#lxd).
  Given the dynamic nature of IRC, we cannot guarantee answers or reactions to IRC posts, but we monitor the channel and try to improve our documentation based on the received feedback.

If images are added (`doc/images`), prioritize either SVG or PNG format and make sure to optimize PNG images for smaller size using a service like [TinyPNG](https://tinypng.com/) or similar.

% Include content from [README.md](README.md)
```{include} README.md
    :start-after: <!-- Include start docs -->
```

When you open a pull request, a preview of the documentation output is built automatically.
To see the output, view the details for the `docs/readthedocs.com:canonical-lxd` check on the pull request.

### Automatic documentation checks

GitHub runs automatic checks on the documentation to verify the spelling, the validity of links, correct formatting of the Markdown files, and the use of inclusive language.

You can (and should!) run these tests locally as well with the following commands:

- Check the spelling: `make doc-spellcheck`
- Check the validity of links: `make doc-linkcheck`
- Check the Markdown formatting: `make doc-lint`
- Check for inclusive language: `make doc-woke`

### Document configuration options

```{note}
We are currently in the process of moving the documentation of configuration options to code comments.
At the moment, not all configuration options follow this approach.
```

The documentation of configuration options is extracted from comments in the Go code.
Look for comments that start with `lxdmeta:generate` in the code.

When you add or change a configuration option, make sure to include the required documentation comment for it.
See the [`lxd-metadata` README file](https://github.com/canonical/lxd/blob/main/lxd/lxd-metadata/README.md) for information about the format.

Then run `make generate-config` to re-generate the `doc/config_options.txt` file.
The updated file should be checked in.

The documentation includes sections from the `doc/config_options.txt` to display a group of configuration options.
For example, to include the core server options:


<!-- LINKS -->

[install lxd]: https://documentation.ubuntu.com/lxd/en/latest/tutorial/first_steps/
[Snapcraft documentation]: https://snapcraft.io/docs/snapcraft-setup
[code repo]: 