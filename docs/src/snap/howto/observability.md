# How to use Prometheus with {{product}}

Observability is an essential component in any system for understanding,
managing, and improving its performance and reliability. The main pillars of
observability are metrics, logs and traces.

One of these pillars is covered by [Prometheus][Prometheus], an open-source
systems monitoring and alerting toolkit designed to collect, process, and query
time-series metrics.

This guide walks you through installing Prometheus in a {{product}} environment.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine.
- You have installed the {{product}} snap.
  (see How-to [Install {{product}} from a snap][snap-install-howto]).
- You have a bootstrapped {{product}} cluster (see the [Getting Started][
  getting-started-guide] guide).
- You have enabled a persistent storage solution in your cluster
  (see How-to [Enable persistent storage][enable-storage]).

## Install Prometheus

Prometheus and its operator can be installed with a Helm chart. Start by
adding the community Helm chart repository to your system:

```
sudo k8s helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
sudo k8s helm repo update
```

Before deploying the Helm chart, you can customize it with a `values.yaml`
file. You can generate it by running:

```
sudo k8s helm show values prometheus-community/kube-prometheus-stack > values.yaml
```

In order to ensure High Availability for the Prometheus services, make sure to
configure the `volumeClaimTemplate` sections appropriately for Alertmanager and
Prometheus (and ThanosRuler, if enabled), and the `persistence` section for
Grafana. If the `storageClassName` field is not set, the cluster's default
`StorageClass` will be used instead. Make sure you have a `StorageClass`
created for your persistent storage solution of choice. You can list them by
running the following:

```
sudo k8s kubectl get storageclass
```

After the Prometheus deployment has been customized with the
`values.yaml` file, run the following command:

```
sudo k8s helm install prometheus prometheus-community/kube-prometheus-stack \
  --create-namespace --namespace observability -f values.yaml
```

Note that this Helm chart installs a few dependent charts:

- `prometheus-community/kube-state-metrics`
- `prometheus-community/prometheus-node-exporter`
- `grafana/grafana`

## Verify that Prometheus is running

It is recommended to ensure that Prometheus initialises properly and is running
without issues. Check that the Prometheus pods are running:

```
sudo k8s kubectl get pods -n observability -l "app.kubernetes.io/name=prometheus"
```

Next, connect to the Prometheus dashboard through its Kubernetes Service:

```
SVC_NAME="prometheus-kube-prometheus-prometheus"
sudo k8s kubectl get -n observability svc/$SVC_NAME
CLUSTER_IP="$(sudo k8s kubectl get -n observability svc/$SVC_NAME -o jsonpath='{.spec.clusterIP}')"
CLUSTER_IP_PORT="$(sudo k8s kubectl get -n observability svc/$SVC_NAME -o jsonpath='{.spec.ports[0].port}')"
echo "Prometheus dashboard URL (ClusterIP): http://${CLUSTER_IP}:${CLUSTER_IP_PORT}/graph"
```

If you do not have access to the cluster network, or if the Prometheus
Kubernetes service is not exposed externally, you can instead create a
temporary local port-forward to the Prometheus dashboard:

```
export POD_NAME=$(sudo k8s kubectl get pods --namespace observability -l "app.kubernetes.io/name=prometheus" -o jsonpath="{.items[0].metadata.name}")
sudo k8s kubectl --namespace observability port-forward $POD_NAME 9090
```

You can check the metrics that have been scraped so far by running:

```
curl -s http://${CLUSTER_IP}:${CLUSTER_IP_PORT}/metrics
```

## Accessing Grafana

[Grafana][Grafana] is an open-source analytics and visualization web
application that enables you query, visualize, alert on, and explore metrics,
logs, and traces.

If you've deployed Prometheus with the Helm chart above, you should already
have Grafana deployed in your cluster:

```
sudo k8s kubectl get pods -n observability -l "app.kubernetes.io/name=grafana"
```

Next, connect to the Grafana dashboard through its Kubernetes service:

```
SVC_NAME="prometheus-grafana"
sudo k8s kubectl get -n observability svc/$SVC_NAME
CLUSTER_IP="$(sudo k8s kubectl get -n observability svc/$SVC_NAME -o jsonpath='{.spec.clusterIP}')"
CLUSTER_IP_PORT="$(sudo k8s kubectl get -n observability svc/$SVC_NAME -o jsonpath='{.spec.ports[0].port}')"
echo "Grafana dashboard URL (ClusterIP): http://${CLUSTER_IP}:${CLUSTER_IP_PORT}/"
```

If you do not have access to the cluster network, or if the Grafana Kubernetes
service is not exposed externally, you can instead create a temporary local
port-forward to the Grafana dashboard:

```
export POD_NAME=$(sudo k8s kubectl get pods --namespace observability -l "app.kubernetes.io/name=grafana" -o jsonpath="{.items[0].metadata.name}")
sudo k8s kubectl --namespace observability port-forward $POD_NAME 3000
```

The default username/password for Grafana are: `admin`/`prom-operator`

## Removing Prometheus

Prometheus and its related components (including Grafana) can be removed by
running:

```
sudo k8s helm delete prometheus -n observability
```

> **_NOTE:_**: The Persistent Volumes created for Prometheus and its related
> services may not deleted when removing Prometheus. You can check them
> by running:

```
sudo k8s kubectl get -n observability pvc
sudo k8s kubectl get pv
```

<!-- LINKS -->

[Prometheus]: https://prometheus.io/
[snap-install-howto]: ./install/snap.md
[getting-started-guide]: ../tutorial/getting-started.md
[enable-storage]: ./storage/index.md
[Grafana]: https://grafana.com/
