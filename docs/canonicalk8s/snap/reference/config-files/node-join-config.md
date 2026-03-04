--- 
tocdepth: 2
---

(node-join)=
# Node join configuration file

A YAML file can be supplied to the `k8s join-cluster ` command to configure and
customize new worker and control plane nodes.


(control-plane-node-join-config)=
## Control plane configuration options 

```{include} /_parts/control_plane_join_config.md
```

---

(worker-node-join-config)=
## Worker configuration options

```{include} /_parts/worker_join_config.md
```