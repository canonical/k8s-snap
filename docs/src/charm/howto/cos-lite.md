# Integrating with COS Lite

It is often advisable to have a monitoring solution which will run whether the
cluster itself is running or not. It may also be useful to integrate monitoring
into existing setups.

To make monitoring your cluster a delightful experience, Canonical provides
first-class integration between {{product}} and COS Lite (Canonical
Observability Stack). This guide will help you integrate a COS Lite
deployment with a {{product}} deployment.

This document assumes that you have a controller with an installation of
{{product}}. If you have not yet installed {{product}}, please see
[Installing {{product}}][how-to-install].

## Preparing a platform for COS Lite

If you are unfamiliar with Juju models, the documentation can be found
[here][juju-models]. This section adds a new model to keep observability
separate from the Kubernetes model.

First, create a new model to act as a deployment cloud for COS Lite:

```
juju add-model --config logging-config='<root>=DEBUG' cos-cluster
```

Set the logging level to DEBUG so that helpful debug information is shown when
you use `juju debug-log` (see [juju debug-log][juju-debug-log]).

Next, deploy your observability cluster using the `k8s` charm:

```
juju deploy k8s --constraints="mem=8G cores=4 root-disk=30G"
```

```{note} local-storage and load-balancer are essential features for the COS
Lite to function correctly. You can enable these features using the charm
configuration [options][k8s-config].
```

Once the cluster is in the `active/idle` state, export the kubeconfig file:

```
juju run k8s/leader get-kubeconfig | yq eval '.kubeconfig' > kubeconfig
```

Register this cluster as a Juju cloud using add-k8s (see ["juju
add-k8s"][add-k8s] for details on the add-k8s
command):

```
KUBECONFIG=./kubeconfig juju add-k8s k8s-cloud
```

## Deploying COS Lite on the K8s cloud

On the K8s cloud, create a new model and deploy the `cos-lite` bundle.
Use the --trust flag to grant the applications access to your cloud credentials.

```
juju add-model cos-lite k8s-cloud
juju deploy cos-lite --trust
```

Make the COS Lite endpoints available for
[cross-model integration][cross-model-integration]:

```
juju offer grafana:grafana-dashboard
juju offer prometheus:receive-remote-write
```

Use `juju status --relations` to verify that both `grafana` and `prometheus`
offerings are listed.

At this stage, you’ve set up a Kubernetes cluster, registered it as a Juju
cloud, and deployed COS Lite on it. This creates two models on the same
controller.

## Integrating COS Lite with {{product}}

Switch to your {{product}} model (if you forgot the name of your model,
you can run `juju models` to see a list of available models):

```
juju switch <canonical-kubernetes-model>
```

Consume the COS Lite endpoints:

```
juju consume cos-lite.grafana
juju consume cos-lite.prometheus
```

Deploy the grafana-agent:

```
juju deploy grafana-agent
```

Relate `grafana-agent` to `k8s`:

```
juju integrate grafana-agent:cos-agent k8s:cos-agent
```

Relate `grafana-agent` to the COS Lite offered interfaces:

```
juju integrate grafana-agent grafana
juju integrate grafana-agent prometheus
```

Get the credentials and login URL for Grafana:

```
juju run grafana/0 get-admin-password -m cos-lite
```

The above command will output:

```
admin-password: b9OhxF5ndUDO
url: http://10.246.154.87/cos-lite-grafana
```

The username for this credential is `admin`.

You’ve successfully gained access to a comprehensive observability stack. Visit
the URL and use the credentials to log in.

Once you feel ready to dive deeper into your shiny new observability platform,
you can head over to the [COS Lite documentation][cos-lite-docs].

<!-- LINKS -->

[how-to-install]: ../howto/charm/install
[add-k8s]: https://juju.is/docs/juju/juju-add-k8s
[cos-lite-docs]: https://charmhub.io/topics/canonical-observability-stack
[juju-models]: https://juju.is/docs/juju/model
[juju-debug-log]: https://juju.is/docs/juju/juju-debug-log
[cross-model-integration]: https://juju.is/docs/juju/relation#heading--cross-model
[k8s-config]: https://charmhub.io/k8s/configurations
