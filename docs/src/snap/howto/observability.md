# How to use Prometheus with Canonical K8s

Observability is an essential component in any system for understanding,
managing, and improving its performance and reliability. The main pillars
of observability are: metrics, logs, traces.

One of these pillars is covered by [Prometheus][Prometheus], an open-source
systems monitoring and alerting toolkit designed to collect, process, and query
time-series metrics.

This guige walks you through installing Prometheus in a {{product}}
environment.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine.
- You have installed the {{product}} snap.
  (see How-to [Install {{product}} from a snap][snap-install-howto]).
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).
- You have enabled a persistent storage solution in your cluster
  (see How-to [Enable persistent storage][enable-storage]).

## Install Prometheus

First, create the `monitoring` namespace in which Prometheus will be deployed:

```bash
sudo k8s kubectl create namespace monitoring
```

Prometheus can be installed through a Helm chart. Start by adding the community
Helm chart repository to your system:

```bash
sudo k8s helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
sudo k8s helm repo update
```

Before deploying the Helm chart, you can customize it through a `values.yaml`
file. You can generate it by running:

```bash
sudo k8s helm show values prometheus-community/prometheus > values.yaml
```

> **_NOTE:_**: In the `values.yaml` file, you can add more metrics endpoints to
> be scraped in `scrape_configs` (in `prometheus.yml`).

By default, the Prometheus Helm chart requires Persistent Volumes (PVs). It can
be configured to use an existing Persistent Volume Claims (PVC), or it can
create PVCs automatically using the configured `storageClass` (if not set, it
will use the default `StorageClass`). Make sure you have a `StorageClass`
created for your persistent storage solution of choice. You can list them by
running the following:

```baash
sudo k8s kubectl get storageclass
```

After the Prometheus deployment has been customized appropriately through the
`values.yaml` file, run the following command:

```bash
sudo k8s helm install prometheus prometheus-community/prometheus \
  --namespace monitoring -f values.yaml
```

## Verify that Prometheus is running

It is recommended to ensure that Prometheus initialises properly and is
running without issues. Check that the Prometheus Pods are running:

```bash
sudo k8s kubectl get pods -n monitoring
```

Next, connect to the Prometheus dashboard through its Kubernetes Service:

```bash
sudo k8s kubectl get -n monitoring svc/prometheus-server
CLUSTER_IP="$(sudo k8s kubectl get -n monitoring svc/prometheus-server -o jsonpath='{.spec.clusterIP}')"
CLUSTER_IP_PORT="$(sudo k8s kubectl get -n monitoring svc/prometheus-server -o jsonpath='{.spec.ports[0].port}')"
echo "Prometheus dashboard URL (ClusterIP): http://${CLUSTER_IP}:${CLUSTER_IP_PORT}/graph"
```

If you do not have access to the cluster network, and If the `prometheus-server`
service is not exposed externally, you can instead create a temporary local
port-forward to the Prometheus dashboard:

```bash
export POD_NAME=$(sudo k8s kubectl get pods --namespace monitoring -l "app.kubernetes.io/name=prometheus,app.kubernetes.io/instance=prometheus" -o jsonpath="{.items[0].metadata.name}")
sudo k8s kubectl --namespace monitoring port-forward $POD_NAME 9090
```

You can check the metrics scraped by Prometheus by running:

```bash
curl -s http://${CLUSTER_IP}:${CLUSTER_IP_PORT}/metrics
```

# Removing Prometheus

Prometheus can be removed by running:

```bash
sudo k8s helm delete prometheus -n monitoring
```

> **_NOTE:_**: Prometheus' Persistent Volumes may not deleted when when removing
> Prometheus. You can check them by running:

``` bash
sudo k8s kubectl get -n monitoring pvc -l "app.kubernetes.io/instance=prometheus"
sudo k8s kubectl get pv
```

<!-- LINKS -->

[Prometheus]: https://prometheus.io/
[snap-install-howto]: ./install/snap.md
[getting-started-guide]: ../../tutorial/getting-started.md
[enable-storage]: ./storage/index.md
