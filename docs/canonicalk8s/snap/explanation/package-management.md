# Package management with Helm

There are multiple ways of installing and managing packages in Kubernetes.
This explanation covers the popular package management tool [Helm][] and
provides details on using Helm with {{product}}.

## Installing Helm

Please check out the upstream documentation for [installing Helm CLI][].

## Charts

[Charts][] are the packaging format used by Helm. A chart is a collection of 
Kubernetes resource manifests related to an application or to a group of 
applications that can be packaged into versioned archives. 
Charts utilize templating and template variables called "values" to enable 
customization of the resources.

## Chart repository

The [Chart repository][] is a web server that usually consists of an index and 
packaged charts. Repositories are the preferred way of distributing charts.

A Helm repository can be added to the local client with:

```
helm repo add <repository-name> <repository-url>
```
This command fetches the index from the repository and
stores it in a cache directory.
This index is used when doing an installation or performing an upgrade.

Update the cached index regularly to get the latest list of
available charts and versions.
Run the following command before performing an upgrade:

```
helm repo update
```

## Installing and managing charts

A [Helm installation][] consists of templates with configurable values that are
rendered into standard Kubernetes resource manifests.

User supplied values are defined in the `values.yaml` file that contains
defaults, which can be overwritten by users at install time.

View the configuration of a chart's `values.yaml`:

```
helm show values <repository-name>/<chart-name>
```

Helm charts can be installed on any Kubernetes cluster, including {{product}}
because Helm uses a kubeconfig file just like `kubectl` to apply
the generated manifests to the cluster.

```{note}
Retrieve a kubeconfig for your {{product}} cluster in accordance to
your installation method.
The kubeconfig file can be placed under the default `~/.kube/config` path.
Alternatively, set the `--kubeconfig <path/to/kubeconfig>` flag to
point Helm to the kubeconfig file.
```

A chart installation can be performed with:

```
helm install <release-name> <repository-name>/<chart-name> --version <version> --values <path/to/values.yaml> --namespace <namespace>
```

```{note}
Helm will use the latest available version of a chart if
the `--version` flag is emitted.
```

A single chart can be used to install an application multiple times.
Each Helm installation creates a **Release**,
which is a group of Kubernetes resources that is managed by Helm.
The `<release-name>` is used in templates to generate unique
Kubernetes resources so the same chart can be used for multiple installations.

Existing releases under a namespace can be listed with:

```
helm list --namespace <namespace>
```

Upgrading a release with a newer version of a chart is similar
to the installation process. An [upgrade][] can be performed with:

```
helm upgrade <release-name> <repository-name>/<chart-name> --version <version> --values <path/to/values.yaml> --namespace <namespace>
```

```{note}
The upgrade command can also be used to change the values of an existing release without having to upgrade to a newer version.
```

Each upgrade operation creates a new release revision.
This makes rollbacks possible in case things go wrong and
there is a need to go back to the previous state of a release.

A [rollback][] can be performed with:

```
helm rollback <release-name> <revision> --namespace <namespace>
```

A release can be [uninstalled][] from the cluster with:

```
helm uninstall <release-name> --namespace <namespace>
```

<!-- LINKS -->

[Helm]: https://helm.sh/
[installing Helm CLI]: https://helm.sh/docs/intro/install/
[Charts]: https://helm.sh/docs/topics/charts/
[Chart repository]: https://helm.sh/docs/topics/chart_repository/
[helm installation]: https://helm.sh/docs/helm/helm_install/
[upgrade]: https://helm.sh/docs/helm/helm_upgrade/
[rollback]: https://helm.sh/docs/helm/helm_rollback/
[uninstalled]: https://helm.sh/docs/helm/helm_uninstall/
