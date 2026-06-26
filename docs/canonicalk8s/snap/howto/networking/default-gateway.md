---
myst:
  html_meta:
    description: "How to enable, disable and configure the Gateway API controller in Canonical Kubernetes for advanced Kubernetes network traffic management."
---

# How to use the default Gateway

<!-- SPREAD SUITE: snap_bootstrapped -->

{{product}} enables you to configure advanced networking of your cluster using
[gateway API](https://gateway-api.sigs.k8s.io/). When enabled, the necessary CRDs and GatewayClass are generated
to enable the CNI controllers configure traffic and provision infrastructure to
the cluster.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstrapped {{product}} cluster (see the
   [Getting Started](/snap/tutorial/getting-started) guide).

## Check Gateway status

Gateway should be enabled by default. Find out whether Gateway is enabled or
disabled with the following command:

```
sudo k8s status
```

<!-- SPREAD
source ${SPREAD_PATH}/docs/tools/repeat_checks.sh
sudo k8s status | grep "gateway                   enabled"
-->

Please ensure that Gateway is enabled on your cluster.

## Enable Gateway

To enable Gateway, run:

```
sudo k8s enable gateway
```

<!-- SPREAD
sudo k8s get gateway | grep "enabled: true"
# Ensure all pods are up before continuing 
sudo k8s kubectl rollout status daemonset/cilium -n kube-system --timeout=10m
sudo k8s kubectl rollout status deployment/cilium-operator -n kube-system --timeout=10m
sudo k8s kubectl wait --for=condition=Ready pods --all -n kube-system --timeout=10m
sudo k8s kubectl wait --for=jsonpath='{.status.conditions[?(@.type=="Accepted")].status}'=True gatewayclass/ck-gateway --timeout=10m
--> 

## Deploy sample workload

As Gateway is enabled, the GatewayClass called `ck-gateway` is already
deployed. View the default GatewayClass:

```
sudo k8s kubectl get GatewayClass
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get GatewayClass" "ck-gateway"
--> 

A [sample workload](https://raw.githubusercontent.com/canonical/k8s-snap/refs/heads/main/tests/integration/templates/gateway-test.yaml) is available as part of our integration test
suite. This deploys a standard Nginx server with a Service to expose the
ClusterIP. A Gateway that points to our GatewayClass and a HTTPRoute that
specifies routing of HTTP requests from our Gateway to the Nginx Service
are also deployed.

Deploy the sample workload:

```
sudo k8s kubectl apply -f https://raw.githubusercontent.com/canonical/k8s-snap/refs/heads/main/tests/integration/templates/gateway-test.yaml
```

<!-- SPREAD
repeat_checks "sudo k8s kubectl get pods" "Running"
--> 

View the workload and service deployed:

```
sudo k8s kubectl get all -owide
```

<!-- SPREAD 
sudo k8s kubectl get service -owide | grep "my-nginx"
-->

The output should look similar to below:

<!-- SPREAD SKIP -->

```
NAME                            READY   STATUS    RESTARTS   AGE     IP
pod/my-nginx-6d596599f5-cddp2   1/1     Running   0          4m19s   10.1.0.141
...
NAME                                TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE     SELECTOR
service/cilium-gateway-my-gateway   LoadBalancer   10.152.183.189   <pending>     80:30230/TCP   4m19s   <none>
service/my-nginx                    ClusterIP      10.152.183.37    <none>        80/TCP         4m19s   run=my-nginx
```

<!-- SPREAD SKIP END -->

Curling the ClusterIP of `cilium-gateway-my-gateway` or `my-nginx`
should return the welcome to Nginx message. This means the Nginx
server is accessible from within the cluster. In this example
the IP address is 10.152.183.189:80:

<!-- SPREAD SKIP -->

```
curl 10.152.183.189:80
```

<!-- SPREAD SKIP END -->

<!-- SPREAD
GATEWAY_IP=$(sudo k8s kubectl get service cilium-gateway-my-gateway -o jsonpath='{.spec.clusterIP}')
repeat_checks "curl --connect-timeout 2 --max-time 4 $GATEWAY_IP:80" "Welcome to nginx"
--> 

To gain access from outside of the cluster, the Gateway needs an
external IP address which will be provided with the load balancer.

```
sudo k8s enable load-balancer
```

<!-- SPREAD
sudo k8s get load-balancer | grep "enabled: true"
--> 

Configure the load balancer CIDR. Choose an appropriate value
depending on your cluster.
This will assign an external IP to `cilium-gateway-my-gateway`.

```
sudo k8s set load-balancer.cidrs=10.0.1.0/28 load-balancer.l2-mode=true
sudo k8s kubectl get service cilium-gateway-my-gateway
```

<!-- SPREAD
sudo k8s get load-balancer | grep "10.0.1.0/28"
sudo k8s get load-balancer | grep "l2-mode: true"
--> 

Get the external IP of the Gateway from the output.

<!-- SPREAD SKIP -->

```
NAME                        TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
cilium-gateway-my-gateway   LoadBalancer   10.152.183.189   10.0.1.0      80:30230/TCP   6m
```

Verify access from the external IP with the target port.

```
curl 10.0.1.0:80
```

<!-- SPREAD SKIP END -->

The output should display a welcome to Nginx message.

<!-- SPREAD
# Ensure my-gateway is ready
sudo k8s kubectl wait --for=condition=Programmed gateway/my-gateway --timeout=5m
repeat_checks "sudo k8s kubectl get service cilium-gateway-my-gateway" "10.0.1." 30
LOADBALANCER_IP=$(sudo k8s kubectl get service cilium-gateway-my-gateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
repeat_checks "curl --connect-timeout 2 --max-time 4 $LOADBALANCER_IP:80" "Welcome to nginx"
--> 

## Disable gateway

You can `disable` the built-in Gateway:

```{warning}
If you have an active cluster, disabling Gateway may impact external access to services within your cluster. Ensure that you have alternative configurations in place before disabling Gateway.
```

```
sudo k8s disable gateway
```

<!-- SPREAD
sudo k8s get gateway | grep "enabled: false"
--> 
