---
myst:
  html_meta:
    description: An overview of how Canonical Kubernetes documentation is structured, written, versioned and maintained.
---

# {{product}} documentation 

Our aim with this documentation set is to provide easy-to-understand 
documentation on all aspects of {{product}}. The source documentation files for 
snap, charm and CAPI based deployment as well as the system configuration used 
to build the docs are included in the
[main repository for the {{product}} snap](https://github.com/canonical/k8s-snap)
.

## Structure

This documentation has adopted the Diátaxis framework. You can read more about
it on the [Diátaxis website]. In essence though, this guides the way we
categorize and write our documentation. You can see there are four main
categories of documentation:

- **Tutorials** for guided walk-throughs
- **How to** pages for specific tasks and goals
- **Explanation** pages which give background reasons and, well, explanations
- **Reference**, where you will find the commands, the roadmap, etc.

Every page of documentation should fit into one of those categories. 

We have included some tips and outlines on the different types of docs we
create to help you better understand our documentation structure 
or get you started if you want to contribute:

- [Tutorial template]
- [How to template]
- [Explanation template]
- [Reference template]

## MyST, Markdown and Sphinx

We use Canonical's [Sphinx Stack] to build the documentation which
is then hosted on ReadtheDocs. The documentation source files are kept in the
`docs/canonicalk8s` directory.

Although Sphinx is normally associated with the `ReSTructured text` format, we
write all our documentation in Markdown to make it easier for humans to work
with. There are a few extra things that come with this - certain features need
to be specially marked up (e.g. admonitions) to be processed properly. There is
a guide to using `MyST` (which is a Markdown extension for Sphinx) directives
and formatting available at [Canonical Sphinx Stack documentation].

## Versioning

We version our documentation to align with {{product}} releases. Each 
{{product}} release has corresponding documentation that describes the 
features, changes and upgrade information for that version. 

Multiple versions of {{product}} documentation are available in the fly out 
ReadtheDocs menu on the bottom right of each page or via the URL:

- **Latest** - The most recent updates available on main. This should be treated
as edge and not stable. 
- **Stable versions** - Versioned documentation that correspond with {{product}}
releases.

Our documentation is updated whenever a new {{product}} version is released. We 
also backport critical documentation updates to earlier versions when needed. 

## Maintaining the docs 

We do our best to maintain a high quality in our documentation and this 
responsibility is shared across the entire team. However, that doesn't mean 
that our documentation can't be improved. If there are any improvements you 
would like to see, every page has a "Give Feedback" button on the top right 
which takes you to GitHub to file an issue. We also include a "Contribute to 
this page" link (pencil icon) which takes you to the GitHub editor to make 
small changes. For larger contributions, please see the 
[docs contribution guide](/snap/howto/contribute)
.

<!-- LINKS -->

[Diátaxis website]: https://diataxis.fr/
[community page]: /community
[Tutorial template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-tutorial
[How to template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-howto
[Explanation template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-explanation
[Reference template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-reference
[Canonical Sphinx Stack documentation]: https://canonical-sphinx-stack.readthedocs-hosted.com/latest/reference/myst-syntax/
[Sphinx Stack]: https://github.com/canonical/sphinx-stack