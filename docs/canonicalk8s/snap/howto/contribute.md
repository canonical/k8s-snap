---
myst:
  html_meta:
    description: How to contribute to Canonical Kubernetes. Learn to how to fix issues, improve existing pages and build the docs locally. 
---

# How to contribute to {{product}} documentation 

{{product}} is proudly open source, published under the GPLv3 license. Our aim 
is to provide easy-to-understand documentation on all aspects of
{{product}}, so we greatly appreciate your feedback and contributions.
See our [community page][] for ways of getting in touch.

The source of the documentation and the system used to build it are included in
the [main repository for the {{product}} snap][code repo].

## What we welcome

Our documentation is focused on {{product}} itself - the features, components 
and workflows that {{product}} provides directly. With that in mind, we welcome 
the following contributions:

- **Fixes** - Corrections to typos, broken links and outdated information 
are always appreciated no matter how small
- **Improvements to existing pages** - Clarifications, better examples or 
additional detail that helps the user better understand {{product}}
- **New pages about {{product}} features** - Any feature or behavior that 
{{product}} currently provides but is not documented yet 

## Make a small change

If you are simply correcting a typo or updating a link, follow the
'Contribute to this page' link (the pencil icon) on any page. This opens the 
online GitHub editor directly. You will still need to 
raise a pull request and provide a brief explanation of your change.

## Make a larger contribution

For new pages or significant additions to existing pages, please open a GitHub 
issue first and describe what you would like to add. This allows us to provide 
early feedback to ensure the scope is aligned with what is needed for the 
project. 

When you are ready to write, the 
[documentation explanation page](/snap/explanation/documentation.md) has useful 
background on our structure and the tools we use. We also provide templates to 
help you get started:

- [Tutorial template]
- [How to template]
- [Explanation template]
- [Reference template]

## Test your changes locally

To test your changes locally, you can build a local version of the
documentation. Open a terminal and go to the `/docs/canonicalk8s` directory. 
From there you can run the command:

```
make run
```

This will create a local environment, install all the dependencies and build
the docs. The output will then be served locally - check the output for the
URL. Using the `run` option means that the docs will automatically be
regenerated when you change any of the source files too (though remember to
press `F5` in your browser to reload the page without caching)!

## Report an issue 

If you would rather not work on the docs yourself or simply want to suggest 
improvements, please raise an issue on the k8s snap repository. The 
"Give Feedback" button on the top of each documentation page will bring you 
directly to GitHub issues page.

<!-- LINKS -->

[code repo]: https://github.com/canonical/k8s-snap
[Diátaxis website]: https://diataxis.fr/
[community page]: /community
[Tutorial template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-tutorial
[How to template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-howto
[Explanation template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-explanation
[Reference template]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/canonicalk8s/_templates/template-reference
[Canonical Sphinx Stack documentation]: https://canonical-sphinx-stack.readthedocs-hosted.com/latest/reference/myst-syntax/
[Sphinx Stack]: https://github.com/canonical/sphinx-stack