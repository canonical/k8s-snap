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
```