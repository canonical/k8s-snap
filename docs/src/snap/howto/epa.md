# How to set up Enhanced Platform Awareness

This section explains how to set up the Enhanced Platform Awareness (EPA) features in a {{product}} cluster. 

The content starts with the setup of the environment (including steps for using
[MAAS][MAAS]). Then the setup of {{product}}, including the Multus & SR-IOV/DPDK
networking components. Finally, the steps needed to test every EPA feature:
HugePages, Real-time Kernel, CPU Pinning / Numa Topology Awareness and
SR-IOV/DPDK. 

## What you'll need

- An Ubuntu Pro subscription (required for real-time kernel)
- Ubuntu instances **or** a MAAS environment to run {{product}} on 


## Prepare the Environment 


`````{tabs}
````{group-tab} Ubuntu

First, run the `numactl` command to get the number of CPUs available for NUMA: 

```
numactl -s
```

This example output shows that there are 32 CPUs available for NUMA:

```
policy: default
preferred node: current
physcpubind: 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31
cpubind: 0 1
nodebind: 0 1
membind: 0 1
```

```{dropdown} Detailed explanation of output

- `policy: default`: indicates that the system is using the default NUMA policy. The default policy typically tries to allocate memory on the same node as the processor executing a task, but it can fall back to other nodes if necessary.  
- `preferred node: current`: processes will prefer to allocate memory from the current node (the node where the process is running). However, if memory is not available on the current node, it can be allocated from other nodes.  
- `physcpubind: 0 1 2 3 ... 31 `: shows the physical CPUs that processes are allowed to run on. In this case, the system has 32 physical CPUs enabled for NUMA, and processes can use any of them.  
- `cpubind: 0 1 `: indicates the specific CPUs that the current process (meaning the process “numactl \-s”) is bound to. It's currently using CPUs 0 and 1.  
- `nodebind: 0 1 `: shows the NUMA nodes that the current process (meaning the process “numactl \-s”) is allowed to use for memory allocation. It has access to both node 0 and node 1.  
- `membind`: 0 1 `: confirms that the current process (meaning the process “numactl \-s”) can allocate memory from both node 0 and node 1.
```

### Enable the real-time kernel

The real-time kernel enablement requires an ubuntu pro subscription and some additional tools to be available.

```
sudo pro attach
sudo apt update && sudo apt install ubuntu-advantage-tools
sudo pro enable realtime-kernel
```

This should produce output similar to:

```
One moment, checking your subscription first
Real-time kernel cannot be enabled with Livepatch.
Disable Livepatch and proceed to enable Real-time kernel? (y/N) y
Disabling incompatible service: Livepatch
The Real-time kernel is an Ubuntu kernel with PREEMPT_RT patches integrated.

This will change your kernel. To revert to your original kernel, you will need
to make the change manually.

Do you want to continue? [ default = Yes ]: (Y/n) Y
Updating Real-time kernel package lists
Updating standard Ubuntu package lists
Installing Real-time kernel packages
Real-time kernel enabled
A reboot is required to complete install.
```

First the Ubuntu system is attached to an Ubuntu Pro subscription 
(needed to use the real-time kernel), requiring you to enter a token 
associated with the subscription. After successful attachment, your 
system gains access to the Ubuntu Pro repositories, including the one 
containing the real-time kernel packages. Once the tools and 
real-time kernel are installed, a reboot is required to start using 
the new kernel.

### Create a configuration file to enable HugePages and CPU isolation

The bootloader will need a configuration file to enable the recommended 
boot options (explained below) to enable HugePages and CPU isolation.

In this example, the host has 128 CPUs, and 2M / 1G HugePages are enabled. 
This is the command to update the boot options and reboot the system:

```
cat <<EOF > /etc/default/grub.d/epa_kernel_options.cfg
GRUB_CMDLINE_LINUX_DEFAULT="${GRUB_CMDLINE_LINUX_DEFAULT} intel_iommu=on iommu=pt usbcore.autosuspend=-1 selinux=0 enforcing=0 nmi_watchdog=0 crashkernel=auto softlockup_panic=0 audit=0 tsc=nowatchdog intel_pstate=disable mce=off hugepagesz=1G hugepages=1000 hugepagesz=2M hugepages=0 default_hugepagesz=1G kthread_cpus=0-31 irqaffinity=0-31 nohz=on nosoftlockup nohz_full=32-127 rcu_nocbs=32-127 rcu_nocb_poll skew_tick=1 isolcpus=managed_irq,32-127 console=tty0 console=ttyS0,115200n8"
EOF
sudo chmod 0644 /etc/netplan/99-sriov_vfs.yaml
update-grub
reboot
```

```{dropdown} Explanation of boot options

-  `intel_iommu=on`: Enables Intel's Input-Output Memory Management Unit (IOMMU), which is used for device virtualization and Direct Memory Access (DMA) remapping.  
-  `iommu=pt`: Sets the IOMMU to passthrough mode, allowing devices to directly access physical memory without translation.  
-  `usbcore.autosuspend=-1`: Disables USB autosuspend, preventing USB devices from being automatically suspended to save power.  
-  `selinux=0`: Disables Security-Enhanced Linux (SELinux), a security module that provides mandatory access control.  
-  `enforcing=0`: If SELinux is enabled, this option sets it to permissive mode, where policies are not enforced but violations are logged.  
-  `nmi_watchdog=0`: Disables the Non-Maskable Interrupt (NMI) watchdog, which is used to detect and respond to system hangs.  
-  `crashkernel=auto`: Reserves a portion of memory for capturing a crash dump in the event of a kernel crash.  
-  `softlockup_panic=0`: Prevents the kernel from panicking (crashing) on detecting a soft lockup, where a CPU appears to be stuck.  
-  `audit=0`: Disables the kernel auditing system, which logs security-relevant events.  
-  `tsc=nowatchdog`: Disables the Time Stamp Counter (TSC) watchdog, which checks for issues with the TSC.  
-  `intel_pstate=disable`: Disables the Intel P-state driver, which controls CPU frequency scaling.  
-  `mce=off`: Disables Machine Check Exception (MCE) handling, which detects and reports hardware errors.  
-  `hugepagesz=1G hugepages=1000`: Allocates 1000 huge pages of 1GB each.  
-  `hugepagesz=2M hugepages=0`: Configures huge pages of 2MB size but sets their count to 0\.  
-  `default_hugepagesz=1G`: Sets the default size for huge pages to 1GB.  
-  `kthread_cpus=0-31`: Restricts kernel threads to run on CPUs 0-31.  
-  `irqaffinity=0-31`: Restricts interrupt handling to CPUs 0-31.  
-  `nohz=on`: Enables the nohz (no timer tick) mode, reducing timer interrupts on idle CPUs.  
-  `nosoftlockup`: Disables the detection of soft lockups.  
-  `nohz_full=32-127`: Enables nohz\_full (full tickless) mode on CPUs 32-127, reducing timer interrupts during application processing.  
-  `rcu_nocbs=32-127`: Offloads RCU (Read-Copy-Update) callbacks to CPUs 32-127, preventing them from running on these CPUs.  
-  `rcu_nocb_poll`: Enables polling for RCU callbacks instead of using interrupts.  
-  `skew_tick=1`: Skews the timer tick across CPUs to reduce contention.  
-  `isolcpus=managed_irq,32-127`: Isolates CPUs 32-127 and assigns managed IRQs to them, reducing their involvement in system processes and dedicating them to specific workloads.  
-  `console=tty0`: Sets the console output to the first virtual terminal.  
-  `console=ttyS0,115200n8`: Sets the console output to the serial port ttyS0 with a baud rate of 115200, 8 data bits, no parity, and 1 stop bit.  
```

Once the reboot has taken place, ensure the HugePages configuration has been applied:

```
grep HugePages /proc/meminfo
```

This should  generate output indicating the number of pages allocated

```
HugePages_Total:    1000
HugePages_Free:     1000
HugePages_Rsvd:        0
HugePages_Surp:        0
```


Next, create a configuration file to configure the network interface 
to use SR-IOV (so it can create virtual functions afterwards) using 
Netplan. In the example below the file is created first, then the configuration is
applied, making  128 virtual functions available for use in the environment:

```
cat <<EOF > /etc/netplan/99-sriov_vfs.yaml
  network:
    ethernets:
      enp152s0f1:
        virtual-function-count: 128
EOF
sudo chmod 0600 /etc/netplan/99-sriov_vfs.yaml
sudo netplan apply
ip link show enp152s0f1
```

The output of the last command should indicate the device is working and has generated the expected
virtual functions.

```
5: enp152s0f1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 9000 qdisc mq state UP mode DEFAULT group default qlen 1000
    link/ether 40:a6:b7:96:d8:89 brd ff:ff:ff:ff:ff:ff
    vf 0     link/ether ae:31:7f:91:09:97 brd ff:ff:ff:ff:ff:ff, spoof checking on, link-state auto, trust off
    vf 1     link/ether 32:09:8b:f7:07:4b brd ff:ff:ff:ff:ff:ff, spoof checking on, link-state auto, trust off
    vf 2     link/ether 12:b9:c6:08:fc:36 brd ff:ff:ff:ff:ff:ff, spoof checking on, link-state auto, trust off
    ..........
    vf 125     link/ether 92:10:ff:8a:e5:0c brd ff:ff:ff:ff:ff:ff, spoof checking on, link-state auto, trust off
    vf 126     link/ether 66:fe:ad:f2:d3:05 brd ff:ff:ff:ff:ff:ff, spoof checking on, link-state auto, trust off
    vf 127     link/ether ca:20:00:c6:83:dd brd ff:ff:ff:ff:ff:ff, spoof checking on, link-state auto, trust off
```

```{dropdown} Explanation of steps
  * Breakdown of the content of the file /etc/netplan/99-sriov\_vfs.yaml :  
    * path: /etc/netplan/99-sriov\_vfs.yaml: This specifies the location of the configuration file. The "99" prefix in the filename usually indicates that it will be processed last, potentially overriding other configurations.  
    * enp152s0f1:  This is the name of the physical network interface you want to create VFs on. This name may vary depending on your system.  
    * virtual-function-count: 128: This is the key line that instructs Netplan to create 128 virtual functions on the specified physical interface. Each of these VFs can be assigned to a different virtual machine or container, effectively allowing them to share the physical adapter's bandwidth.  
    * permissions: "0600": This is an optional line that sets the file permissions to 600 (read and write access only for the owner).  
  * Breakdown of the output of ip link show enp152s0f1 command:  
    * Main interface:  
      * 5: The index number of the network interface in the system.  
      * enp152s0f1: The name of the physical network interface.  
      * \<BROADCAST,MULTICAST,UP,LOWER\_UP\>: The interface's flags indicating its capabilities (e.g., broadcast, multicast) and current status (UP).  
      * mtu 9000: The maximum transmission unit (MTU) is set to 9000 bytes, larger than the typical 1500 bytes, likely for jumbo frames.  
      * qdisc mq: The queuing discipline (qdisc) is set to "mq" (multi-queue), designed for multi-core systems.  
      * state UP: The interface is currently active and operational.  
      * mode DEFAULT: The interface is in the default mode of operation.  
      * qlen 1000: The maximum number of packets allowed in the transmit queue.  
      * link/ether 40:a6:b7:96:d8:89: The interface's MAC address (a unique hardware identifier).  
    * Virtual functions:  
      * vf \<number\>: The index number of the virtual function.  
      * link/ether \<MAC address\>: The MAC address assigned to the virtual function.  
      * spoof checking on: A security feature to prevent MAC address spoofing (pretending to be another device).  
      * link-state auto: The link state (up/down) is determined automatically based on the physical connection.  
      * trust off: The interface doesn't trust the incoming VLAN (Virtual LAN) tags.  
    * Results:  
      * Successful VF Creation: The output confirms a success creation of 128 VFs (numbered 0 through 127\) on the enp152s0f1 interface.  
      * VF Availability: Each VF is ready for use, and they can be assigned i.e. to {{product}} containers to give them direct access to the network through this physical network interface.  
      * MAC Addresses: Each VF has its own unique MAC address, which is essential for network communication.
```


Now enable DPDK, first by cloning the DPDK repo, and then placing the script which
will bind the VFs to the VFIO-PCI driver in the location that will run
automatically each time the system boots up, so the VFIO
(Virtual Function I/O) bindings are applied consistently:

```
git clone https://github.com/DPDK/dpdk.git /home/ubuntu/dpdk
cat <<EOF > /var/lib/cloud/scripts/per-boot/dpdk_bind.sh 
      #!/bin/bash
      if [ -d /home/ubuntu/dpdk ]; then
        modprobe vfio-pci
        vfs=$(python3 /home/ubuntu/dpdk/usertools/dpdk-devbind.py -s | grep drv=iavf | awk '{print $1}' | tail -n +11)
        python3 /home/ubuntu/dpdk/usertools/dpdk-devbind.py --bind=vfio-pci $vfs
      fi
sudo chmod 0755 /var/lib/cloud/scripts/per-boot/dpdk_bind.sh
```

```{dropdown} Explanation 
  * Load VFIO Module (modprobe vfio-pci): If the DPDK directory exists, the script loads the VFIO-PCI kernel module. This module is necessary for the VFIO driver to function.  
  * The script uses the dpdk-devbind.py tool (included with DPDK) to list the available network devices and their drivers.  
    * It filters this output using grep drv=iavf to find devices that are currently using the iavf driver (a common driver for Intel network adapters), excluding the physical network interface itself and just focusing on the virtual functions (VFs).  
  * Bind VFs to VFIO: The script uses dpdk-devbind.py again, this time with the \--bind=vfio-pci option, to bind the identified VFs to the VFIO-PCI driver. This step essentially tells the kernel to relinquish control of these devices to DPDK.  
```

To test that the VFIO Kernel Module and DPDK are enabled:

```
lsmod | grep -E 'vfio'
```

...should indicate the kernel module is loaded

```
vfio_pci               16384  0
vfio_pci_core          94208  1 vfio_pci
vfio_iommu_type1       53248  0
vfio                   73728  3 vfio_pci_core,vfio_iommu_type1,vfio_pci
iommufd                98304  1 vfio
irqbypass              12288  2 vfio_pci_core,kvm

```

Running the helper script:

```
python3 /home/ubuntu/dpdk/usertools/dpdk-devbind.py -s
```

...should return a list of network devices using DPDK:

```
Network devices using DPDK-compatible driver
============================================
0000:98:12.2 'Ethernet Adaptive Virtual Function 1889' drv=vfio-pci unused=iavf
0000:98:12.3 'Ethernet Adaptive Virtual Function 1889' drv=vfio-pci unused=iavf
0000:98:12.4 'Ethernet Adaptive Virtual Function 1889' drv=vfio-pci unused=iavf
....
```

With these preparation steps we have enabled the features of EPA:

- NUMA and CPU Pinning are available to the first 32 CPUs  
- Real-Time Kernel is enabled  
- HugePages are enabled and 1000 1G huge pages are available   
- SRIOV is enabled in the enp152s0f1 interface, with 128 virtual 
  function interfaces bound to the vfio-pci driver (that could also use the iavf driver)  
- DPDK is enabled in all the 128 virtual function interfaces

````

````{group-tab} MAAS

To prepare a machine for CPU isolation, Hugepages, real-time kernel, 
SRIOV and DPDK we leverage cloud-init through MAAS.

```
#cloud-config

apt:
  sources:
    rtk.list:
      source: "deb https://<launchpad_id>:<ppa_subscription_password>@private-ppa.launchpadcontent.net/canonical-kernel-rt/ppa/ubuntu jammy main"

write_files:
  # set kernel option with hugepages and cpu isolation
  - path: /etc/default/grub.d/100-telco_kernel_options.cfg
    content: |
      GRUB_CMDLINE_LINUX_DEFAULT="${GRUB_CMDLINE_LINUX_DEFAULT} intel_iommu=on iommu=pt usbcore.autosuspend=-1 selinux=0 enforcing=0 nmi_watchdog=0 crashkernel=auto softlockup_panic=0 audit=0 tsc=nowatchdog intel_pstate=disable mce=off hugepagesz=1G hugepages=1000 hugepagesz=2M hugepages=0 default_hugepagesz=1G kthread_cpus=0-31 irqaffinity=0-31 nohz=on nosoftlockup nohz_full=32-127 rcu_nocbs=32-127 rcu_nocb_poll skew_tick=1 isolcpus=managed_irq,32-127 console=tty0 console=ttyS0,115200n8"
    permissions: "0644"

  # create sriov VFs
  - path: /etc/netplan/99-sriov_vfs.yaml
    content: |
      network:
          ethernets:
              enp152s0f1:
                  virtual-function-count: 128
    permissions: "0600"

  # ensure VFs are bound to vfio-pci driver (so they can be consumed by pods)
  - path: /var/lib/cloud/scripts/per-boot/dpdk_bind.sh
    content: |
      #!/bin/bash
      if [ -d /home/ubuntu/dpdk ]; then
        modprobe vfio-pci
        vfs=$(python3 /home/ubuntu/dpdk/usertools/dpdk-devbind.py -s | grep drv=iavf | awk '{print $1}' | tail -n +11)
        python3 /home/ubuntu/dpdk/usertools/dpdk-devbind.py --bind=vfio-pci $vfs
      fi
    permissions: "0755"

  # set proxy variables
  - path: /etc/environment
    content: |
      HTTPS_PROXY=http://10.18.2.1:3128
      HTTP_PROXY=http://10.18.2.1:3128
      NO_PROXY=10.0.0.0/8,192.168.0.0/16,127.0.0.1,172.16.0.0/16,.svc,localhost
      https_proxy=http://10.18.2.1:3128
      http_proxy=http://10.18.2.1:3128
      no_proxy=10.0.0.0/8,192.168.0.0/16,127.0.0.1,172.16.0.0/16,.svc,localhost
    append: true

  # add rtk ppa key
  - path: /etc/apt/trusted.gpg.d/rtk.asc
    content: |
      -----BEGIN PGP PUBLIC KEY BLOCK-----
      Comment: Hostname:
      Version: Hockeypuck 2.2

      xsFNBGAervwBEADHCeEuR7WKRiEII+uFOu8J+W47MZOcVhfNpu4rdcveL4qe4gj4
      nsROMHaINeUPCmv7/4EXdXtTm1VksXeh4xTeqH6ZaQre8YZ9Hf4OYNRcnFOn0KR+
      aCk0OWe9xkoDbrSYd3wmx8NG/Eau2C7URzYzYWwdHgZv6elUKk6RDbDh6XzIaChm
      kLsErCP1SiYhKQvD3Q0qfXdRG908lycCxgejcJIdYxgxOYFFPcyC+kJy2OynnvQr
      4Yw6LJ2LhwsA7bJ5hhQDCYZ4foKCXX9I59G71dO1fFit5O/0/oq0xe7yUYCejf7Z
      OqD+TzEK4lxLr1u8j8lXoQyUXzkKIL0SWEFT4tzOFpWQ2IBs/sT4X2oVA18dPDoZ
      H2SGxCUcABfne5zrEDgkUkbnQRihBtTyR7QRiE3GpU19RNVs6yAu+wA/hti8Pg9O
      U/5hqifQrhJXiuEoSmmgNb9QfbR3tc0ZhKevz4y+J3vcnka6qlrP1lAirOVm2HA7
      STGRnaEJcTama85MSIzJ6aCx4omCgUIfDmsi9nAZRkmeomERVlIAvcUYxtqprLfu
      6plDs+aeff/MAmHbak7yF+Txj8+8F4k6FcfNBT51oVSZuqFwyLswjGVzWol6aEY7
      akVIrn3OdN2u6VWlU4ZO5+sjP4QYsf5K2oVnzFVIpYvqtO2fGbxq/8dRJQARAQAB
      zSVMYXVuY2hwYWQgUFBBIGZvciBDYW5vbmljYWwgS2VybmVsIFJUwsGOBBMBCgA4
      FiEEc4Tsv+pcopCX6lNfLz1Vl/FsjCEFAmAervwCGwMFCwkIBwIGFQoJCAsCBBYC
      AwECHgECF4AACgkQLz1Vl/FsjCF9WhAAnwfx9njs1M3rfsMMuhvPxx0WS65HDlq8
      SRgl9K2EHtZIcS7lHmcjiTR5RD1w+4rlKZuE5J3EuMnNX1PdCYLSyMQed+7UAtX6
      TNyuiuVZVxuzJ5iS7L2ZoX05ASgyoh/Loipc+an6HzHqQnNC16ZdrBL4AkkGhDgP
      ZbYjM3FbBQkL2T/08NcwTrKuVz8DIxgH7yPAOpBzm91n/pV248eK0a46sKauR2DB
      zPKjcc180qmaVWyv9C60roSslvnkZsqe/jYyDFuSsRWqGgE5jNyIb8EY7K7KraPv
      3AkusgCh4fqlBxOvF6FJkiYeZZs5YXvGQ296HTfVhPLOqctSFX2kuUKGIq2Z+H/9
      qfJFGS1iaUsoDEUOaU27lQg5wsYa8EsCm9otroH2P3g7435JYRbeiwlwfHMS9EfK
      dwD38d8UzZj7TnxGG4T1aLb3Lj5tNG6DSko69+zqHhuknjkRuAxRAZfHeuRbACgE
      nIa7Chit8EGhC2GB12pr5XFWzTvNFdxFhbG+ed7EiGn/v0pVQc0ZfE73FXltg7et
      bkoC26o5Ksk1wK2SEs/f8aDZFtG01Ys0ASFICDGW2tusFvDs6LpPUUggMjf41s7j
      4tKotEE1Hzr38EdY+8faRaAS9teQdH5yob5a5Bp5F5wgmpqZom/gjle4JBVaV5dI
      N5rcnHzcvXw=
      =asqr
      -----END PGP PUBLIC KEY BLOCK-----
    permissions: "0644"

# install the snap
snap:
  commands:
    00: 'snap install k8s --classic --channel=1.30-moonray/beta'

runcmd:
# fetch dpdk driver binding script
- su ubuntu -c "git config --global http.proxy http://10.18.2.1:3128"
- su ubuntu -c "git clone https://github.com/DPDK/dpdk.git /home/ubuntu/dpdk"
- apt update
- DEBIAN_FRONTEND=noninteractive apt install -y linux-headers-6.8.1-1004-realtime linux-image-6.8.1-1004-realtime linux-modules-6.8.1-1004-realtime linux-modules-extra-6.8.1-1004-realtime

# enable kernel options
- update-grub

# reboot to activate realtime-kernel and kernel options
power_state:
  mode: reboot
```

Notes:

* In the above, realtime kernel 6.8 is installed from a private ppa. It was recently backported from 24.04 to 22.04 and is still going through some validation stages. Once it is officially released, it will be installable via the Ubuntu Pro cli.


````
`````





## {{product}} setup 

{{product}} is delivered as a 
[snap](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/).

This section explains how to set up a dual node {{product}} cluster for testing
EPA capabilities.

### Control plane and worker node 

1. [Install the
   snap](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/howto/install/snap/)
   from the relevant track, currently `{{track}}`. The [beta
   channel](https://documentation.ubuntu.com/canonical-kubernetes/latest/snap/explanation/channels/)
   is used at this point as the end configuration of the k8s snap is not
   finalised yet.

```
sudo snap install k8s --classic --channel=1.30-moonray/beta
```

2. Create a file called *configuration.yaml*. In this configuration file we let
   the snap start with its default CNI (calico), with CoreDNS deployed and we
   also point k8s to the external etcd. 

```yaml
cluster-config:
  network:
    enabled: true
  dns:
    enabled: true
  local-storage:
    enabled: true
extra-node-kubelet-args:
  --reserved-cpus: "0-31"
  --cpu-manager-policy: "static"
  --topology-manager-policy: "best-effort"
```

3. Bootstrap {{product}} using the above configuration file.

```
sudo k8s bootstrap --file configuration.yaml
```

#### Verify control plane node is up

After a few seconds you can query the API server with:

```
sudo k8s kubectl get all -A
```

### Add second k8s node as worker 

1. Install the k8s snap on the second node

```
sudo snap install k8s --classic --channel=1.30-moonray/beta
```

2. On the control plane node generate a join token to be used for joining the
   second node

```
sudo k8s get-join-token --worker
```

3. On the worker node create the configuration.yaml file

```
extra-node-kubelet-args:
  --reserved-cpus: "0-31"
  --cpu-manager-policy: "static"
  --topology-manager-policy: "best-effort"
```

4. On the worker node use the token to join the cluster

```
sudo k8s join-cluster --file configuration.yaml <token-generated-on-the-control-plane-node>
```


#### Verify the two node cluster is ready 

After a few seconds the second worker node will register with the control
plane. You can query the available workers from the first node:

```
sudo k8s kubectl get nodes
```

The output should list the connected nodes:

```
NAME          STATUS   ROLES                  AGE   VERSION
pc6b-rb4-n1   Ready    control-plane,worker   22h   v1.30.2
pc6b-rb4-n3   Ready    worker                 22h   v1.30.2
```

### Multus and SRIOV setup 

Get the thick plugin (in case of resource scarcity we can consider deploying
the thin flavor)

```
sudo k8s kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset-thick.yml
```

```{note} 
The memory limits for the multus pod spec in the DaemonSet should be
increased (i.e. to 500Mi instead 50Mi) to avoid OOM issues when deploying
multiple workload pods in parallel.
```

#### SRIOV Network Device Plugin 

Create sriov-dp.yaml configMap:

```
cat <<EOF | tee sriov-dp.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sriovdp-config
  namespace: kube-system
data:
  config.json: |
    {
      "resourceList": [{
        "resourceName": "intel_sriov_netdevice",
        "selectors": [{
          "vendors": ["8086"],
          "devices": ["1889"],
          "drivers": ["iavf"]
          }]
        },
        {
        "resourceName": "intel_sriov_dpdk",
        "resourcePrefix": "intel.com",
        "selectors": [{
            "vendors": ["8086"],
            "devices": ["1889"],
            "drivers": ["vfio-pci"],
            "pfNames": ["enp152s0f1"],
            "needVhostNet": true
            }]
        }
      ]
    }
EOF
```

Apply the configMap definition:

```
sudo k8s kubectl create -f ./sriov-dp.yaml
```

Install the SRIOV network device plugin:

```
sudo k8s kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/sriov-network-device-plugin/master/deployments/sriovdp-daemonset.yaml
```

#### SRIOV CNI 

Install the SRIOV CNI daemonset:

```
sudo k8s kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/sriov-cni/master/images/sriov-cni-daemonset.yaml
```

#### Multus NetworkAttachmentDefinition 

Create the sriov-nad.yaml NetworkAttachmentDefinition:

```
cat <<EOF | tee sriov-nad.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
 name: sriov-net1
 annotations:
   k8s.v1.cni.cncf.io/resourceName: intel.com/intel_sriov_netdevice
spec:
 config: '{
 "type": "sriov",
 "cniVersion": "0.3.1",
 "name": "sriov-network",
 "ipam": {
   "type": "host-local",
   "subnet": "10.18.2.153/24", # The subnet for SR-IOV VFs
   "routes": [{
     "dst": "0.0.0.0/0"
   }],
   "gateway": "10.18.2.1" 
 }
}'
EOF

cat <<EOF | tee dpdk-nad.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
 name: dpdk-net1
 annotations:
   k8s.v1.cni.cncf.io/resourceName: intel.com/intel_sriov_dpdk
spec:
 config: '{
 "type": "sriov",
 "cniVersion": "0.3.1",
 "name": "dpdk-network"
}'
EOF
```

Apply the two NetworkAttachmentDefinition definition files

```
sudo k8s kubectl create -f ./sriov-nad.yaml
sudo k8s kubectl create -f ./dpdk-nad.yaml
```

## Testing

It is important to verify that all of these enabled features are working as
expected before relying on them. This section confirms that
everything is working as expected.

### Test HugePages 

Verify that HugePages are allocated on your Kubernetes nodes. You can do this
by checking the node's capacity and allocatable resources:

```
sudo k8s kubectl get nodes
```

This should return the available nodes

```
NAME          STATUS   ROLES                  AGE   VERSION
pc6b-rb4-n1   Ready    control-plane,worker   22h   v1.30.2
pc6b-rb4-n3   Ready    worker                 22h   v1.30.2
```


```
  hugepages-1Gi:      1000Gi
  hugepages-2Mi:      0
  hugepages-1Gi:      1000Gi
  hugepages-2Mi:      0
  hugepages-1Gi      0 (0%)      0 (0%)
  hugepages-2Mi      0 (0%)      0 (0%)
```

So this example has 1000 huge pages of 1Gi size each and we have a worker node
labelled properly. Then you can create a Pod that explicitly requests one 1G
Huge Page in its resource limits:

```
cat <<EOF | sudo k8s kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: hugepage-test-ubuntu
spec:
  containers:
  - name: ubuntu-container
    image: ubuntu:latest
    command: ["sleep", "infinity"]
    resources:
      limits:
        hugepages-1Gi: 1Gi
        memory: 128Mi
      requests:
        memory: 128Mi
EOF
```

```{note} 
To ensure proper resource management and prevent conflicts, Kubernetes
enforces that a pod requesting HugePages also explicitly requests a minimum
Now, ensure that the `1Gi` HugePage is allocated in the pod:
```

Now ensure that the 1Gi HugePage is allocated in the pod:

```
sudo k8s kubectl describe pod hugepage-test-ubuntu
``` 

The output should reflect the HugePage request:


```
....
    Limits:
      hugepages-1Gi:  1Gi
      memory:         128Mi
    Requests:
      hugepages-1Gi:  1Gi
      memory:         128Mi
....

```

### Test the real-time kernel

First, verify that real-time kernel is enabled in the worker node by checking
if “PREEMPT RT” appears after running the `uname -a` command:

```
uname -a
```

The output should show the “PREEMPT RT” identifier:

```
Linux pc6b-rb4-n3 6.8.1-1004-realtime #4~22.04.1-Ubuntu SMP PREEMPT_RT Mon Jun 24 16:45:51 UTC 2024 x86_64 x86_64 x86_64 GNU/Linux
```

The test will use cyclictest, commonly used to assess the real-time performance
of a system, especially when running a real-time kernel. It measures the time
it takes for a thread to cycle between high and low priority states, giving you
an indication of the system's responsiveness to real-time events.  Lower
latencies typically indicate better real-time performance.

The output of cyclictest will provide statistics including:

- Average latency: The average time taken for a cycle.  
- Minimum latency: The shortest observed cycle time.  
- Maximum latency: The longest observed cycle time.

Create a pod that will run the cyclictest tool with specific options:

- `-l 1000000`: Sets the number of test iterations to 1 million.  
- `-m`: Measures the maximum latency.  
- `-p 80`: Sets the real-time scheduling priority to 80 (a high priority, typically used for real-time tasks).  
- `-t 1`: Specifies CPU core 1 to be used for the test.

```
cat <<EOF | sudo k8s kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: realtime-kernel-test
spec:
  containers:
  - name: realtime-test
    image: ubuntu:latest
    command: ["/bin/bash"]
    args: ["-c", "apt-get update && apt-get install rt-tests -y && cyclictest -l 1000000 -m -p 80 -t 1"]
    securityContext:
      privileged: true     
EOF
```

Confirm that the test is running by checking the pod's logs:

```
sudo k8s kubectl logs realtime-kernel-test -f

```

```
...
# /dev/cpu_dma_latency set to 0us
policy: fifo: loadavg: 7.92 8.34 9.32 1/3698 2965

T: 0 ( 2965) P:80 I:1000 C: 241486 Min:      3 Act:    4 Avg:    3 Max:      18
```

```{dropdown} Explanation of output

- `/dev/cpu_dma\_latency set to 0us`: This line indicates that the CPU DMA (Direct Memory Access) latency has been set to 0 microseconds. This setting is relevant for real-time systems as it controls how long a device can hold the CPU bus during a DMA transfer.  
- `policy: fifo`: This means the scheduling policy for the cyclictest thread is set to FIFO (First In, First Out). In FIFO scheduling, the highest priority task that is ready to run gets the CPU first and continues running until it is blocked or voluntarily yields the CPU.  
- `loadavg: 7.92 8.34 9.32 1/3698 2965:` This shows the load average of your system over the last 1, 5, and 15 minutes. The numbers are quite high, indicating that your system is under significant load. This can potentially affect the latency measurements.  
- `T: 0 ( 2965) P:80 I:1000 C: 241486`:  
  - `T: 0`: The number of the CPU core the test was run on (CPU 0 in this case).  
  - `(2965)`: The PID (Process ID) of the cyclictest process.  
  - `P:80`: The priority of the cyclictest thread.  
  - `I:1000`: The number of iterations (loops) the test ran for (1000 in this case).  
  - `C: 241486`: The number of cycles per second that the test has aimed for.  
- `Min: 3 Act: 4 Avg: 3 Max: 18`: These are the key latency statistics in microseconds (us):  
  - `Min`: The minimum latency observed during the test (3 us).  
  - `Act`: The actual average latency (4 us).  
  - `Avg`: The expected average latency (3us).  
  - `Max`: The maximum latency observed during the test (18 us).  
- In this case, the results suggest the following:  
  - Low Latencies: The minimum, average, and maximum latencies are all very low (3-18 us), which is a good sign for real-time performance. It indicates that your real-time kernel is responding promptly to events.  
  - High Load: The high load average indicates that your system is busy, but even under this load, the real-time kernel is maintaining low latencies for the high-priority cyclictest thread.

```

### Test CPU Pinning and NUMA 

First check if CPU Manager and NUMA Topology Manager is set up in the worker node:

```
ps -ef | grep /snap/k8s/678/bin/kubelet
``` 

```
root        9139       1  1 Jul17 ?        00:20:03 /snap/k8s/678/bin/kubelet --anonymous-auth=false --authentication-token-webhook=true --authorization-mode=Webhook --client-ca-file=/etc/kubernetes/pki/client-ca.crt --cluster-dns=10.152.183.97 --cluster-domain=cluster.local --container-runtime-endpoint=/var/snap/k8s/common/run/containerd.sock --containerd=/var/snap/k8s/common/run/containerd.sock --cpu-manager-policy=static --eviction-hard=memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi --fail-swap-on=false --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=10.18.2.153 --node-labels=node-role.kubernetes.io/worker=,k8sd.io/role=worker --read-only-port=0 --register-with-taints= --reserved-cpus=0-31 --root-dir=/var/lib/kubelet --serialize-image-pulls=false --tls-cert-file=/etc/kubernetes/pki/kubelet.crt --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_GCM_SHA384 --tls-private-key-file=/etc/kubernetes/pki/kubelet.key --topology-manager-policy=best-effort
```

```{dropdown} Explanation of output

* \--cpu-manager-policy=static : This flag within the Kubelet command line arguments explicitly tells us that the CPU Manager is active and using the static policy. Here's what this means:  
  * CPU Manager:  This is a component of Kubelet that manages how CPU resources are allocated to pods running on a node.  
  * Static Policy:  This policy is designed to provide stricter control over CPU allocation. With the static policy, you can request integer CPUs for your containers (e.g., 1, 2, etc.), and {{product}} will try to assign them to dedicated CPU cores on the node, providing a greater degree of isolation and predictability.  
* \--reserved-cpus=0-31: This line indicates that no CPUs are reserved for the Kubelet or system processes. This implies that all CPUs might be available for pod scheduling, depending on the cluster's overall resource allocation strategy.  
* \--topology-manager-policy=best-effort: This flag sets the topology manager policy to "best-effort." The topology manager helps optimise pod placement on nodes by considering factors like NUMA nodes, CPU cores, and devices. The "best-effort" policy tries to place pods optimally, but it doesn't enforce strict requirements.
```

You can also confirm the total number of NUMA CPUs available in the worker node:

```
lscpu
```

```
....
NUMA:
  NUMA node(s):           2
  NUMA node0 CPU(s):      0,2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32,34,36,38,40,42,44,46,48,50,52,54,56,58,60,62,64,66,68,70,72,74,76,78,80,82,84,86,88,90,92,94,96,98,100,102,104,106,108,110,112,114,116,118,120,122,124,126
  NUMA node1 CPU(s):      1,3,5,7,9,11,13,15,17,19,21,23,25,27,29,31,33,35,37,39,41,43,45,47,49,51,53,55,57,59,61,63,65,67,69,71,73,75,77,79,81,83,85,87,89,91,93,95,97,99,101,103,105,107,109,111,113,115,117,119,121,123,125,127
...
```

Now let’s label the node with information about the available CPU/NUMA nodes, and then create a pod selecting that label:

```
sudo k8s kubectl label node pc6b-rb4-n3 topology.kubernetes.io/zone=NUMA

```

```
node/pc6b-rb4-n3 labeled
```

```
$ cat <<EOF | sudo k8s kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: cpu-pinning-test
spec:
  containers:
  - name: test-container
    image: ubuntu:latest
    command: ["sleep", "infinity"]
    resources:
      requests:
        cpu: "2"
        memory: "512Mi"
      limits:
        cpu: "2"
        memory: "512Mi"
  nodeSelector:
    topology.kubernetes.io/zone: NUMA
EOF
```

Finally, describing the node and the pod will confirm that the pod is running
on the intended node and that its CPU requests are being met; also running
taskset inside the pod will identify the pod pinned to the process running
inside the pod:

```
sudo k8s kubectl describe node pc6b-rb4-n3
```

...which should produce the output:

```
  Namespace                   Name                                 CPU Requests  CPU Limits  Memory Requests  Memory Limits  Age
  ---------                   ----                                 ------------  ----------  ---------------  -------------  ---
  ....
  default                     cpu-pinning-test                     2 (2%)        2 (2%)      512Mi (0%)       512Mi (0%)     72s
  ....
```

```
sudo k8s kubectl describe pod cpu-pinning-test
```

```
...
    Limits:
      cpu:     2
      memory:  512Mi
    Requests:
      cpu:        2
      memory:     512Mi
...
```

<!-- this needs to be explained -->

```
sudo k8s kubectl exec -ti cpu-pinning-test -- /bin/bash
root@cpu-pinning-test:/# ps -ef
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 08:51 ?        00:00:00 sleep infinity
root          17       0  0 08:58 pts/0    00:00:00 /bin/bash
root          25      17  0 08:58 pts/0    00:00:00 ps -ef
root@cpu-pinning-test:/# taskset -p 1
pid 1's current affinity mask: 1000000000000000100000000
```

This hexadecimal mask (1000000000000000100000000) might seem unusual, but it
represents the binary equivalent: 0b1000000000000000100000000  
In this binary representation, each '1' bit indicates a CPU core that the
process is allowed to run on, while a '0' bit indicates a core the process
cannot use. Counting from right to left (starting at 0), the '1' bits in this
mask correspond to CPU cores 0 and 32\.

Based on the output, the sleep infinity process (PID 1\) is indeed being pinned
to specific CPU cores (0 and 32). This indicates that the CPU pinning is
working correctly. 

### Test SR-IOV & DPDK

First check if SR-IOV Device Plugin pod is running and healthy in the cluster,
if SR-IOV is allocatable in the worker node and the PCI IDs of the VFs
available in the node (describing one of them to get further details):

```
sudo k8s kubectl get pods -n kube-system | grep sriov-device-plugin
```

This should indicate some running pods:

```
kube-sriov-device-plugin-7mxz5        1/1     Running   0          7m31s
kube-sriov-device-plugin-fjzgt        1/1     Running   0          7m31s
```

Now check the VFs:

```
sudo k8s kubectl describe node pc6b-rb4-n3
```

This should indicate the presence of the SRIOV device:

```
...
Allocatable:
  cpu:                              96
  ephemeral-storage:                478444208Ki
  hugepages-1Gi:                    1000Gi
  hugepages-2Mi:                    0
  intel.com/intel_sriov_dpdk:       118
  intel.com/intel_sriov_netdevice:  10
  memory:                           1064530020Ki
  pods:                             110
....
```

The virtual functions should also appear on th 

```
lspci | grep Virtual
```

```
98:11.0 Ethernet controller: Intel Corporation Ethernet Adaptive Virtual Function (rev 02)
98:11.1 Ethernet controller: Intel Corporation Ethernet Adaptive Virtual Function (rev 02)
98:11.2 Ethernet controller: Intel Corporation Ethernet Adaptive Virtual Function (rev 02)
...
99:00.5 Ethernet controller: Intel Corporation Ethernet Adaptive Virtual Function (rev 02)
99:00.6 Ethernet controller: Intel Corporation Ethernet Adaptive Virtual Function (rev 02)
99:00.7 Ethernet controller: Intel Corporation Ethernet Adaptive Virtual Function (rev 02)
```

```
lspci -s 98:1f.2 -vv
98:1f.2 Ethernet controller: Intel Corporation Ethernet Adaptive Virtual Function (rev 02)
	Subsystem: Intel Corporation Ethernet Adaptive Virtual Function
	Control: I/O- Mem- BusMaster- SpecCycle- MemWINV- VGASnoop- ParErr- Stepping- SERR- FastB2B- DisINTx-
	Status: Cap+ 66MHz- UDF- FastB2B- ParErr- DEVSEL=fast >TAbort- <TAbort- <MAbort- >SERR- <PERR- INTx-
	NUMA node: 1
	IOMMU group: 391
	Region 0: Memory at d6e40000 (64-bit, prefetchable) [virtual] [size=128K]
	Region 3: Memory at d81e8000 (64-bit, prefetchable) [virtual] [size=16K]
	Capabilities: <access denied>
	Kernel driver in use: vfio-pci
	Kernel modules: iavf
```

Now, create a test pod that will claim a network interface from the DPDK network:

```
cat <<EOF | sudo k8s kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: sriov-test-pod
  annotations:
    k8s.v1.cni.cncf.io/networks: dpdk-net1 
spec:
  containers:
    - name: sriov-test-container
      image: ubuntu:latest
      resources:
        limits:
          cpu: "10"
          memory: "4000Mi"
          intel.com/intel_sriov_dpdk: '1'
          hugepages-1Gi: "16Gi"
        requests:
          cpu: "10"
          memory: "4000Mi"
          hugepages-1Gi: "16Gi"
          intel.com/intel_sriov_dpdk: '1'
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
            - CAP_SYS_NICE
            - SYS_NICE
            - IPC_LOCK
            - NET_ADMIN
            - SYS_TIME
            - CAP_NET_RAW
            - CAP_BPF
            - CAP_SYS_ADMIN
            - SYS_ADMIN
        privileged: true
      command: ["bash", "-c", "sleep infinity"]
EOF
```

Finally, describe the pod to confirm the DPDK network assignment and also the
virtual function PCI ID (in this case, `0000:98:1f.2`) that was assigned
automatically to the `net1` interface:

```
sudo k8s kubectl describe pod sriov-test-pod
```


```
...
                  k8s.v1.cni.cncf.io/network-status:
                    [{
                        "name": "k8s-pod-network",
                        "ips": [
                            "10.1.17.141"
                        ],
                        "default": true,
                        "dns": {}
                    },{
                        "name": "default/dpdk-net1",
                        "interface": "net1",
                        "mac": "26:e4:aa:f4:ce:ba",
                        "dns": {},
                        "device-info": {
                            "type": "pci",
                            "version": "1.1.0",
                            "pci": {
                                "pci-address": "0000:98:1f.2"
                            }
                        }
                    }]
                  k8s.v1.cni.cncf.io/networks: dpdk-net1
...

```



## Further reading

* [How to enable Real-time Ubuntu](https://canonical-ubuntu-pro-client.readthedocs-hosted.com/en/latest/howtoguides/enable\_realtime\_kernel/\#how-to-enable-real-time-ubuntu)  
* [Manage HugePages](https://kubernetes.io/docs/tasks/manage-hugepages/scheduling-hugepages/)  
* [Utilising the NUMA-aware Memory Manager](https://kubernetes.io/docs/tasks/administer-cluster/memory-manager/)  
* [SR-IOV Network Device Plugin for Kubernetes](https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin)  
* [VMware Telco Cloud Automation \- EPA](https://docs.vmware.com/en/VMware-Telco-Cloud-Automation/3.1.1/com-vmware-tca-userguide/GUID-3F4BA111-D344-4022-A635-7D5774385EF8.html)


<!-- LINKS -->

[MAAS]: https://maas.io