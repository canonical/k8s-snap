# How to use the default Gateway

{{product}} enables you to configure advanced networking of your cluster using
[gateway API]. When enabled, the necessary CRDs and GatewayClass are generated
to enable the CNI controllers configure traffic and provision infrastructure to
the cluster.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstrapped {{product}} cluster (see the
[Getting Started][getting-started-guide] guide).

## Check Gateway status

Gateway should be enabled by default. Find out whether Gateway is enabled or
disabled with the following command:

```
sudo k8s status
```

Please ensure that Gateway is enabled on your cluster.

## Enable Gateway

To enable Gateway, run:

```
sudo k8s enable gateway
```

## Deploy sample workload

As Gateway is enabled, the GatewayClass called `ck-gateway` is already
deployed. View the default GatewayClass:

```
sudo k8s kubectl get GatewayClass
```

A [sample workload] is available as part of our integration test
suite. This deploys a standard Nginx server with a Service to expose the
ClusterIP. A Gateway that points to our GatewayClass and a HTTPRoute that
specifies routing of HTTP requests from our Gateway to the Nginx Service
are also deployed.

Deploy the sample workload:

```
sudo k8s kubectl apply -f https://raw.githubusercontent.com/canonical/k8s-snap/refs/heads/main/tests/integration/templates/gateway-test.yaml
```

View the workload and service deployed:

```
sudo k8s kubectl get all -owide
```

The output should look similar to below:

```
NAME                            READY   STATUS    RESTARTS   AGE     IP
pod/my-nginx-6d596599f5-cddp2   1/1     Running   0          4m19s   10.1.0.141
...
NAME                                TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE     SELECTOR
service/cilium-gateway-my-gateway   LoadBalancer   10.152.183.189   <pending>     80:30230/TCP   4m19s   <none>
service/my-nginx                    ClusterIP      10.152.183.37    <none>        80/TCP         4m19s   run=my-nginx
```

Curling the ClusterIP of `cilium-gateway-my-gateway` or `my-nginx` 
should return the welcome to Nginx message. This means the Nginx 
server is accessible from within the cluster. In this example
the IP address is 10.152.183.189:80:

```
curl 10.152.183.189:80
```

To gain access from outside of the cluster, the Gateway needs an 
external IP address which will be provided with the load balancer.

```
sudo k8s enable load-balancer
```

Configure the load balancer CIDR. Choose an appropriate value 
depending on your cluster.
This will assign an external IP to `cilium-gateway-my-gateway`.

```
sudo k8s set load-balancer.cidrs=10.0.1.0/28 load-balancer.l2-mode=true
sudo k8s kubectl get service cilium-gateway-my-gateway
```

Get the external IP of the Gateway from the output.

```
NAME                        TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
cilium-gateway-my-gateway   LoadBalancer   10.152.183.189   10.0.1.0      80:30230/TCP   6m
```

Verify access from the external IP with the target port.

```
curl 10.0.1.0:80
```

The output should display a welcome to Nginx message.

## Disable gateway

You can `disable` the built-in Gateway:

``` {warning}
If you have an active cluster, disabling Gateway may impact external access to services within your cluster. Ensure that you have alternative configurations in place before disabling Gateway.
```

```
sudo k8s disable gateway
```
<!-- LINKS -->
[gateway API]:https://gateway-api.sigs.k8s.io/
[getting-started-guide]: ../../tutorial/getting-started
[kubectl-guide]: ../../tutorial/kubectl
[sample workload]: https://raw.githubusercontent.com/canonical/k8s-snap/refs/heads/main/tests/integration/templates/gateway-test.yaml
