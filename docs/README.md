# K8s snap documentation

This part of the repository contains the tools and the source for generating
documentation for the Canonical Kubernetes snap.

The directories are organised like this:

```

├── README.md
├── canonicalk8s
│   ├── _build
│   │   ├── {contains the generated docs}
│   ├──{sphinx build tools for creating the docs for Canonical K8s}
│   ├──{source files for canonicalk8s docs}
```

## Building the docs

This documentation uses the /canonicalk8s/Makefile to generate HTML docs from
the sources. This can also run specific local tests such as spelling and
linkchecking.

## Contributing to the docs

Contributions to this documentation are welcome. Generally these follow the
same rules and process as other contributions - modify the docs source and
submit a PR.
