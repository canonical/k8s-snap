# Integrating with COS Lite

It is often advisable to have a monitoring solution which will run whether the
cluster itself is running or not. It may also be useful to integrate monitoring
into existing setups.

To make monitoring your cluster a delightful experience, Canonical provides
first-class integration between {{product}} and COS Lite (Canonical
Observability Stack). This guide will help you integrate a COS Lite
deployment with a {{product}} deployment.

This document assumes you have a controller with an installation of Canonical
Kubernetes. If you have not yet installed {{product}}, please see
["Installing {{product}}"][how-to-install].

## Preparing a platform for COS Lite

If you are unfamiliar with Juju models, the documentation can be found
[here][juju-models]. In this section, we'll be adding a new model to keep
observability separate from the Kubernetes model.

First, create a MicroK8s model to act as a deployment cloud for COS Lite:

```
juju add-model --config logging-config='<root>=DEBUG' microk8s-ubuntu
```

We also set the logging level to DEBUG so that helpful debug information is
shown when you use `juju debug-log` (see [juju debug-log][juju-debug-log]).

Use the Ubuntu charm to deploy an application named “microk8s”:

```
juju deploy ubuntu microk8s --series=focal --constraints="mem=8G cores=4 root-disk=30G"
```

Deploy MicroK8s on Ubuntu by accessing the unit you created at the last step
with `juju ssh microk8s/0` and following the 
[Install Microk8s][how-to-install-microk8s] guide for configuration.

```{note} Make sure to enable the hostpath-storage and MetalLB addons for 
Microk8s.
```

Export the Microk8s kubeconfig file to your current directory after
configuration:

```
juju ssh microk8s/0 -- microk8s config > microk8s-config.yaml
```

Register MicroK8s as a Juju cloud using add-k8s (see ["juju
add-k8s"][add-k8s] for details on the add-k8s
command):

```
KUBECONFIG=microk8s-config.yaml juju add-k8s microk8s-cloud
```

## Deploying COS Lite on the Microk8s cloud

On the Microk8s cloud, create a new model and deploy the `cos-lite` bundle:

```
juju add-model cos-lite microk8s-cloud
juju deploy cos-lite
```

Make COS Lite’s endpoints available for 
[cross-model integration][cross-model-integration]:

```
juju offer grafana:grafana-dashboard
juju offer prometheus:receive-remote-write
```

Use `juju status --relations` to verify that both `grafana` and `prometheus`
offerings are listed.

At this point, you’ve established a MicroK8s model on Ubuntu and incorporated
it into Juju as a Kubernetes cloud. You then used this cloud as a substrate for
the COS Lite deployment. You therefore have 2 models on the same controller.

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

[how-to-install]: ../howto/charm
[add-k8s]: https://juju.is/docs/juju/juju-add-k8s
[cos-lite-docs]: https://charmhub.io/topics/canonical-observability-stack
[juju-models]: https://juju.is/docs/juju/model
[juju-debug-log]: https://juju.is/docs/juju/juju-debug-log
[cross-model-integration]: https://juju.is/docs/juju/relation#heading--cross-model
[how-to-install-microk8s]: https://microk8s.io/docs/getting-started