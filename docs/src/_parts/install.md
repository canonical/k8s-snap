<!-- snap start -->
sudo snap install k8s --classic --channel=1.32-classic/stable
<!-- snap end -->
<!-- lxd start -->
lxc exec k8s -- sudo snap install k8s --classic --channel=1.32-classic/stable
<!-- lxd end -->
<!-- offline start -->
sudo snap download k8s --channel 1.32-classic/stable --basename k8s
<!-- offline end -->
<!-- juju control start -->
juju deploy k8s --channel=1.32/edge
<!-- juju control end -->
<!-- juju worker start -->
juju deploy k8s-worker --channel=1.32/edge -n 2
<!-- juju worker end -->
<!-- juju control constraints start -->
juju deploy k8s --channel=1.32/edge --constraints='cores=2 mem=16G root-disk=40G'
<!-- juju control constraints end -->
<!-- juju worker constraints start -->
juju deploy k8s-worker --channel=1.32/edge --constraints='cores=2 mem=16G root-disk=40G'
<!-- juju worker constraints end -->
<!-- juju vm start -->
juju deploy k8s --channel=latest/edge \
    --base "ubuntu@22.04" \
    --constraints "cores=2 mem=8G root-disk=16G virt-type=virtual-machine"
<!-- juju vm end -->
