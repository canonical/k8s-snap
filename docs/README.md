# K8s snap documentation

This part of the repository contains the tools and the source for generating documentation for the Canonical Kubernetes snap.

The directories are organised like this:

```

├── _build
│   ├── {contains the generated docs}
├── README.md
├── src
│   ├──{source files for the docs}
└── tools
    ├──{sphinx build tools for creating the docs}
```

## Building the docs

This documentation uses the /tools/Makefile to generate HTML docs from the sources
This can also run specific local tests such as spelling and linkchecking

## Contributing to the docs