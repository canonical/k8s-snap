<!-- snap start -->
sudo snap install k8s --classic --channel=1.33-classic/stable
<!-- snap end -->
<!-- lxd start -->
lxc exec k8s -- sudo snap install k8s --classic --channel=1.33-classic/stable
<!-- lxd end -->
<!-- offline start -->
sudo snap download k8s --channel 1.33-classic/stable --basename k8s
<!-- offline end -->
<!-- juju control start -->
juju deploy k8s --channel=1.33/stable
<!-- juju control end -->
<!-- juju worker start -->
juju deploy k8s-worker --channel=1.33/stable -n 2
<!-- juju worker end -->
<!-- juju control constraints start -->
juju deploy k8s --channel=1.33/stable --constraints='cores=2 mem=16G root-disk=40G'
<!-- juju control constraints end -->
<!-- juju worker constraints start -->
juju deploy k8s-worker --channel=1.33/stable --constraints='cores=2 mem=16G root-disk=40G'
<!-- juju worker constraints end -->
<!-- juju vm start -->
juju deploy k8s --channel=latest/stable \
    --base "ubuntu@22.04" \
    --constraints "cores=2 mem=8G root-disk=16G virt-type=virtual-machine"
<!-- juju vm end -->
<!-- juju controlplane custom config start -->
juju deploy k8s --config ./k8s-config.yaml --channel=1.33/stable
<!-- juju controlplane custom config end -->
