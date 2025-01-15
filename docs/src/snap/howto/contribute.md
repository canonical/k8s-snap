# How to contribute to {{product}}

{{product}} is proudly open source, published under the GPLv3 license.
We welcome and encourage contributions to the code and the documentation. See
the [community page][] for ways to get in touch and provide feedback.

## Contribute to the code

{{product}} is shipped as a snap package. To contribute to the code,
you should first make sure you can build and test the snap locally.

### Build the snap

To build the snap locally, you will need the following:

- The latest, snap-based version of LXD (see the [install guide here][install
   lxd])
- The Snapcraft build tool, for building the snap (see the [Snapcraft
   documentation][]).

Clone the [GitHub repository for the k8s snap][code repo] and then open a
terminal in that directory. Run the command:

```
snapcraft --use-lxd
```

This will launch an LXD container and use it to build a version of the snap.
This will take some time as the build process fetches dependencies, stages the
‘parts’ of the snap and creates the snap package itself. The snap itself will
be fetched from the build environment and placed in the local project
directory. Note that the LXD container used for building will be stopped, but
not deleted. This is in case there were any errors or artifacts you may wish to
inspect.

### Install the snap

The snap can then be installed locally by using the ‘--dangerous’ option. This
is a safeguard to make sure the user is aware that the snap is not signed by
the snap store, and is not confined:

```
sudo snap install k8s_v1.32.1_multi.snap --dangerous --classic
```

```{note} You will not be able to install this snap if there is already a
   k8s snap installed on your system.
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
lxc delete snapcraft-k8s
```

### Making a change to the API

The Canonical Kubernetes codebase references the `k8s-snap-api` package
extensively. When contributing changes that require API modifications, follow
these steps:

1. Clone the `k8s-snap-api` repository from
   https://github.com/canonical/k8s-snap-api

2. Add a module replace directive in your src/k8s/go.mod file to point to your
   local API copy. For example:

```
module github.com/canonical/k8s

go 1.23.0

replace github.com/canonical/k8s-snap-api => /home/user/ubuntu/k8s-snap/src/k8s/k8s-snap-api

require (
   ...
)
```

3. Make your API changes in the local copy

4. Create a separate PR in the k8s-snap-api repository with your API changes

5. Reference your k8s-snap-api PR in your main k8s-snap PR

### Contribute changes

We welcome any improvements and bug-fixes to the {{product}} code.
Once you have tested your changes, please make a pull request on the [code
repository][code repo] and we will review it as soon as possible.

## PR review process

When you create your PR, a member of the team will review it. Your PR must
receive at least one approval from a Canonical Kubernetes team member before
it's eligible to be merged.

For faster reviews, ensure your PR:

* Passes all automated tests
* Has a clear title and description of the changes
* Links to related issues
* Includes test cases if relevant
* Contains only changes that are relevant to the PRs stated purpose
* Updates relevant documentation

Draft PRs are welcome for early feedback, please mark them as such.

## Contribute to the documentation

Our aim is to provide easy-to-understand documentation on all aspects of
{{product}}, so we greatly appreciate your feedback and contributions.
See our [community page][] for ways of getting in touch.

The source of the documentation and the system used to build it are included in
the [main repository for the {{product}} snap][code repo].

### Documentation framework

This documentation has adopted the Diátaxis framework. You can read more about
it on the [Diátaxis website]. In essence though, this guides the way we
categorise and write our documentation. You can see there are four main
categories of documentation:

- **Tutorials** for guided walk-throughs
- **How to** pages for specific tasks and goals
- **Explanation** pages which give background reasons and, well, explanations
- **Reference**, where you will find the commands, the roadmap, etc.

Every page of documentation should fit into one of those categories. If it
doesn't you may consider if it is actually two pages (e.g. a How to *and* an
explanation).

We have included some tips and outlines of the different types of docs we
create to help you get started:

- [Tutorial template][]
- [How to template][]
- [Explanation template][]
- [Reference template][]

### Small changes

If you are simply correcting a typo or updating a link, you can follow the
'Edit this page on GitHub' link on any page and it will take you to the online
editor to make your change. You will still need to raise a pull request and
explain your change to get it reviewed.

### Myst, Markdown and Sphinx

We use the Sphinx documentation tools to actually build the documentation. You
will find all the Sphinx tooling in the `docs/tools` directory.

Although Sphinx is normally associated with the `ReSTructured text` format, we
write all our documentation in Markdown to make it easier for humans to work
with. There are a few extra things that come with this - certain features need
to be specially marked up (e.g. admonitions) to be processed properly. There is
a guide to using `Myst` (which is a Markdown extension for Sphinx) directives
and formatting contained in the [_parts][] directory of the docs.

### Local testing

To test your changes locally, you can build a local version of the
documentation. Open a terminal and go to the `/docs/tools` directory. From
there you can run the command:

```
make run
```

This will create a local environment, install all the dependencies and build
the docs. The output will then be served locally - check the output for the
URL. Using the `run` option means that the docs will automatically be
regenerated when you change any of the source files too (though remember to
press `F5` in your browser to reload the page without caching)!

<!-- LINKS -->

[install lxd]: https://documentation.ubuntu.com/lxd/en/latest/tutorial/first_steps/
[Snapcraft documentation]: https://snapcraft.io/docs/snapcraft-setup
[code repo]: https://github.com/canonical/k8s-snap
[Diátaxis website]: https://diataxis.fr/
[_parts]: https://github.com/canonical/k8s-snap/blob/main/docs/src/_parts/doc-cheat-sheet-myst.md
[community page]: ../reference/community
[Tutorial template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/_parts/template-tutorial
[How to template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/_parts/template-howto
[Explanation template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/_parts/template-explanation
[Reference template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/_parts/template-reference
