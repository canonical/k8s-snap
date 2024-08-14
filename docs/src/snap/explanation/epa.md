# Enhanced Platform Awareness

Enhanced Platform Awareness (EPA) is a methodology and a set of
enhancements across various layers of the orchestration stack, aimed at
optimising platform capability, configuration and capacity usage. For
telecommunications service providers using virtual network functions (VNFs) to
deliver services, EPA offers enhanced application performance, increased
input/output throughput and improved determinism.

EPA focuses on three key objectives: discovering, scheduling and isolating
server hardware capabilities. This document provides a detailed description of
EPA capabilities and its integration into {{product}}. These capabilities are:

-  **HugePage support**: In GA from Kubernetes v1.14, this feature enables the
   discovery, scheduling and allocation of HugePages as a first-class
   resource.  
-  **Real-time kernel**: Ensures that high-priority tasks are run within a
   predictable time frame, providing the low latency and high determinism
   essential for time-sensitive applications.  
-  **CPU pinning** (CPU Manager for Kubernetes (CMK)): In GA from Kubernetes
   v1.26, provides mechanisms for CPU pinning and isolation of containerised
   workloads.  
-  **NUMA topology awareness**: Ensures that CPU and memory allocation are
   aligned according to the NUMA architecture, reducing memory latency and
   increasing performance for memory-intensive applications.  
-  **Single Root I/O Virtualization (SR-IOV)**: Enhances networking by enabling
   virtualisation of a single physical network device into multiple virtual
   devices.  
-  **DPDK (Data Plane Development Kit)**: A set of libraries and drivers for
   fast packet processing, designed to run in user space, optimising network
   performance and reducing latency.

This document provides detailed instructions for setting up and installing the
aforementioned technologies. It is designed for developers and architects
looking to integrate these new technologies into their {{product}}-based
networking solutions. The goal is to achieve enhanced network I/O,
deterministic compute performance and optimised server platform sharing.

## HugePages 

HugePages are a feature in the Linux kernel which enables the allocation of
larger memory pages. This reduces the overhead of managing large amounts of
memory and can improve performance for applications that require significant
memory access.

### Key features

-  **Larger memory pages**: HugePages provide larger memory pages (e.g., 2MB or
   1GB) compared to the standard 4KB pages, reducing the number of pages the
   system must manage.  
-  **Reduced overhead**: By using fewer, larger pages, the system reduces the
   overhead associated with page table entries, leading to improved memory
   management efficiency.  
-  **Improved TLB performance**: The Translation Lookaside Buffer (TLB) stores
   recent translations of virtual memory to physical memory addresses. Using
   HugePages increases TLB hit rates, reducing the frequency of memory
   translation lookups.  
-  **Enhanced application performance**: Applications that access large amounts
   of memory can benefit from HugePages by experiencing lower latency and
   higher throughput due to reduced page faults and better memory access
   patterns.  
-  **Support for high-performance workloads**: Ideal for high-performance
   computing (HPC) applications, databases and other memory-intensive
   workloads that demand efficient and fast memory access.  
-  **Native Kubernetes integration**: Starting from Kubernetes v1.14, HugePages
   are supported as a native, first-class resource, enabling their
   discovery, scheduling and allocation within Kubernetes environments.

### Application to Kubernetes

The architecture for HugePages on Kubernetes integrates the management and
allocation of large memory pages into the Kubernetes orchestration system. Here
are the key architectural components and their roles:

-  **Node configuration**: Each Kubernetes node must be configured to reserve
   HugePages. This involves setting the number of HugePages in the node's
   kernel boot parameters.  
-  **Kubelet configuration**: The kubelet on each node must be configured to
   recognise and manage HugePages. This is typically done through the kubelet
   configuration file, specifying the size and number of HugePages.  
-  **Pod specification**: HugePages are requested and allocated at the pod
   level through resource requests and limits in the pod specification. Pods
   can request specific sizes of HugePages (e.g., 2MB or 1GB).  
-  **Scheduler awareness**: The Kubernetes scheduler is aware of HugePages as a
   resource and schedules pods onto nodes that have sufficient HugePages
   available. This ensures that pods with HugePages requirements are placed
   appropriately. Scheduler configurations and policies can be adjusted to
   optimise HugePages allocation and utilisation.  
-  **Node Feature Discovery (NFD)**: Node Feature Discovery can be used to
   label nodes with their HugePages capabilities. This enables scheduling
   decisions to be based on the available HugePages resources.  
-  **Resource quotas and limits**: Kubernetes enables the definition of resource
   quotas and limits to control the allocation of HugePages across namespaces.
   This helps in managing and isolating resource usage effectively.  
-  **Monitoring and metrics**: Kubernetes provides tools and integrations
   (e.g., Prometheus, Grafana) to monitor and visualise HugePages usage across
   the cluster. This helps in tracking resource utilisation and performance.
   Metrics can include HugePages allocation, usage and availability on each
   node, aiding in capacity planning and optimization.

## Real-time kernel 

A real-time kernel ensures that high-priority tasks are executed within a
predictable timeframe, crucial for applications requiring low latency and high
determinism.

### Key features

-  **Predictable task execution**: A real-time kernel ensures that
   high-priority tasks are executed within a predictable and bounded timeframe,
   reducing the variability in task execution time.  
-  **Low latency**: The kernel is optimised to minimise the time it takes to
   respond to high-priority tasks, which is crucial for applications that
   require immediate processing.  
-  **Priority-based scheduling**: Tasks are scheduled based on their priority
   levels, with real-time tasks being given precedence over other types of
   tasks to ensure they are processed promptly.  
-  **Deterministic behaviour**: The kernel guarantees deterministic behaviour,
   meaning the same task will have the same response time every time it is
   run, essential for time-sensitive applications.  
-  **Pre-emption:** The real-time kernel supports preemptive multitasking,
   allowing high-priority tasks to interrupt lower-priority tasks to ensure
   critical tasks are run without delay.  
-  **Resource reservation**: System resources (such as CPU and memory) can be
   reserved by the kernel for real-time tasks, ensuring that these resources
   are available when needed.  
-  **Enhanced interrupt handling**: Interrupt handling is optimised to ensure
   minimal latency and jitter, which is critical for maintaining the
   performance of real-time applications.  
-  **Real-time scheduling policies**: The kernel includes specific scheduling
   policies (e.g., SCHED\_FIFO, SCHED\_RR) designed to manage real-time tasks
   effectively and ensure they meet their deadlines.

These features make a real-time kernel ideal for applications requiring precise
timing and high reliability.

### Application to Kubernetes

The architecture for integrating a real-time kernel into Kubernetes involves
several components and configurations to ensure that high-priority, low-latency
tasks can be managed effectively within a Kubernetes environment. Here are the
key architectural components and their roles:

-  **Real-Time Kernel installation**: Each Kubernetes node must run a real-time
   kernel. This involves installing a real-time kernel package and configuring
   the system to use it.  
-  **Kernel boot parameters**: The kernel boot parameters must be configured to
   optimise for real-time performance. This includes isolating CPU cores and
   configuring other kernel parameters for real-time behaviour.  
-  **Kubelet configuration**: The `kubelet` on each node must be configured to
   recognise and manage real-time workloads. This can involve setting specific
   `kubelet` flags and configurations.  
-  **Pod specification**: Real-time workloads are specified at the pod level
   through resource requests and limits. Pods can request dedicated CPU cores
   and other resources to ensure they meet real-time requirements.  
-  **CPU Manager**: Kubernetes’ CPU Manager is a critical component for
   real-time workloads. It allows for the static allocation of CPUs to
   containers, ensuring that specific CPU cores are dedicated to particular
   workloads.  
-  **Scheduler awareness**: The Kubernetes scheduler must be aware of real-time
   requirements and prioritise scheduling pods onto nodes with available
   real-time resources.  
-  **Priority and preemption**: Kubernetes supports priority and preemption to
   ensure that critical real-time pods are scheduled and run as needed. This
   involves defining pod priorities and enabling preemption to ensure
   high-priority pods can displace lower-priority ones if necessary.  
-  **Resource quotas and limits**: Kubernetes can define resource quotas
   and limits to control the allocation of resources for real-time workloads
   across namespaces. This helps manage and isolate resource usage effectively.
-  **Monitoring and metrics**: Monitoring tools such as Prometheus and Grafana
   can be used to track the performance and resource utilisation of real-time
   workloads. Metrics include CPU usage, latency and task scheduling times,
   which help in optimising and troubleshooting real-time applications.  
-  **Security and isolation**: Security contexts and isolation mechanisms
   ensure that real-time workloads are protected and run in a controlled
   environment. This includes setting privileged containers and configuring
   namespaces.

## CPU pinning 

CPU pinning enables specific CPU cores to be dedicated to a particular process
or container, ensuring that the process runs on the same CPU core(s) every
time, which reduces context switching and cache invalidation.

### Key features 

-  **Dedicated CPU Cores**: CPU pinning allocates specific CPU cores to a
   process or container, ensuring consistent and predictable CPU usage.  
-  **Reduced context switching**: By running a process or container on the same
   CPU core(s), CPU pinning minimises the overhead associated with context
   switching, leading to better performance.  
-  **Improved cache utilisation**: When a process runs on a dedicated CPU core,
   it can take full advantage of the CPU cache, reducing the need to fetch data
   from main memory and improving overall performance.  
-  **Enhanced application performance**: Applications that require low latency
   and high performance benefit from CPU pinning as it ensures they have
   dedicated processing power without interference from other processes.  
-  **Consistent performance**: CPU pinning ensures that a process or container
   receives consistent CPU performance, which is crucial for real-time and
   performance-sensitive applications.  
-  **Isolation of workloads**: CPU pinning isolates workloads on specific CPU
   cores, preventing them from being affected by other workloads running on
   different cores. This is especially useful in multi-tenant environments.  
-  **Improved predictability**: By eliminating the variability introduced by
   sharing CPU cores, CPU pinning provides more predictable performance
   characteristics for critical applications.  
-  **Integration with Kubernetes**: Kubernetes supports CPU pinning through the
   CPU Manager (in GA since v1.26), which allows for the static allocation of
   CPUs to containers. This ensures that containers with high CPU demands have
   the necessary resources.

### Application to Kubernetes

The architecture for CPU pinning in Kubernetes involves several components and
configurations to ensure that specific CPU cores can be dedicated to particular
processes or containers, thereby enhancing performance and predictability. Here
are the key architectural components and their roles:

-  **Kubelet configuration**: The kubelet on each node must be configured to
   enable CPU pinning. This involves setting specific kubelet flags to activate
   the CPU Manager.  
-  **CPU manager**: Kubernetes’ CPU Manager is a critical component for CPU
   pinning. It allows for the static allocation of CPUs to containers, ensuring
   that specific CPU cores are dedicated to particular workloads. The CPU
   Manager can be configured to either static or none. Static policy allows for
   exclusive CPU core allocation to Guaranteed QoS (Quality of Service) pods.  
-  **Pod specification**: Pods must be specified to request dedicated CPU
   resources. This is done through resource requests and limits in the pod
   specification.  
-  **Scheduler awareness**: The Kubernetes scheduler must be aware of the CPU
   pinning requirements. It schedules pods onto nodes with available CPU
   resources as requested by the pod specification. The scheduler ensures that
   pods with specific CPU pinning requests are placed on nodes with sufficient
   free dedicated CPUs.  
-  **NUMA Topology Awareness**: For optimal performance, CPU pinning should be
   aligned with NUMA (Non-Uniform Memory Access) topology. This ensures that
   memory accesses are local to the CPU, reducing latency. Kubernetes can be
   configured to be NUMA-aware, using the Topology Manager to align CPU
   and memory allocation with NUMA nodes.  
-  **Node Feature Discovery (NFD)**: Node Feature Discovery can be used to
   label nodes with their CPU capabilities, including the availability of
   isolated and reserved CPU cores.  
-  **Resource quotas and limits**: Kubernetes allows defining resource quotas
   and limits to control the allocation of CPU resources across namespaces.
   This helps in managing and isolating resource usage effectively.  
-  **Monitoring and metrics**: Monitoring tools such as Prometheus and Grafana
   can be used to track the performance and resource utilisation of CPU-pinned
   workloads. Metrics include CPU usage, core allocation and task scheduling
   times, which help in optimising and troubleshooting performance-sensitive
   applications.  
-  **Isolation and security**: Security contexts and isolation mechanisms
   ensure that CPU-pinned workloads are protected and run in a controlled
   environment. This includes setting privileged containers and configuring
   namespaces to avoid resource contention.  
-  **Performance Tuning**: Additional performance tuning can be achieved by
   isolating CPU cores at the OS level and configuring kernel parameters to
   minimise interference from other processes. This includes setting CPU
   isolation and nohz_full parameters (reduces the number of scheduling-clock
   interrupts, improving energy efficiency and [reducing OS jitter][no_hz]).

## NUMA topology awareness 

NUMA (Non-Uniform Memory Access) topology awareness ensures that the CPU and
memory allocation are aligned according to the NUMA architecture, which can
reduce memory latency and increase performance for memory-intensive
applications.

The Kubernetes Memory Manager enables the feature of guaranteed memory (and
hugepages) allocation for pods in the Guaranteed QoS (Quality of Service)
class.

The Memory Manager employs hint generation protocol to yield the most suitable
NUMA affinity for a pod. The Memory Manager feeds the central manager (Topology
Manager) with these affinity hints. Based on both the hints and Topology
Manager policy, the pod is rejected or admitted to the node.

Moreover, the Memory Manager ensures that the memory which a pod requests is
allocated from a minimum number of NUMA nodes.

### Key features

-  **Aligned CPU and memory allocation**: NUMA topology awareness ensures that
   CPUs and memory are allocated in alignment with the NUMA architecture,
   minimising cross-node memory access latency.  
-  **Reduced memory latency**: By ensuring that memory is accessed from the
   same NUMA node as the CPU, NUMA topology awareness reduces memory latency,
   leading to improved performance for memory-intensive applications.  
-  **Increased performance**: Applications benefit from increased performance
   due to optimised memory access patterns, which is especially critical for
   high-performance computing and data-intensive tasks.  
-  **Kubernetes Memory Manager**: The Kubernetes Memory Manager supports
   guaranteed memory allocation for pods in the Guaranteed QoS (Quality of
   Service) class, ensuring predictable performance.  
-  **Hint generation protocol**: The Memory Manager uses a hint generation
   protocol to determine the most suitable NUMA affinity for a pod, helping to
   optimise resource allocation based on NUMA topology.  
-  **Integration with Topology Manager**: The Memory Manager provides NUMA
   affinity hints to the Topology Manager. The Topology Manager then decides
   whether to admit or reject the pod based on these hints and the configured
   policy.  
-  **Optimised resource allocation**: The Memory Manager ensures that the
   memory requested by a pod is allocated from the minimum number of NUMA
   nodes, thereby optimising resource usage and performance.  
-  **Enhanced scheduling decisions**: The Kubernetes scheduler, in conjunction
   with the Topology Manager, makes informed decisions about pod placement to
   ensure optimal NUMA alignment, improving overall cluster efficiency.  
-  **Support for HugePages**: The Memory Manager also supports the allocation
   of HugePages, ensuring that large memory pages are allocated in a NUMA-aware
   manner, further enhancing performance for applications that require large
   memory pages.  
-  **Improved application predictability**: By aligning CPU and memory
   allocation with NUMA topology, applications experience more predictable
   performance characteristics, crucial for real-time and latency-sensitive
   workloads.  
-  **Policy-Based Management**: NUMA topology awareness can be managed through
   policies, allowing administrators to configure how resources should be
   allocated based on the NUMA architecture, providing flexibility and control.

### Application to Kubernetes

The architecture for NUMA topology awareness in Kubernetes involves several
components and configurations to ensure that CPU and memory allocations are
optimised according to the NUMA architecture. This setup reduces memory latency
and enhances performance for memory intensive applications. Here are the key
architectural components and their roles:

-  **Node configuration**: Each Kubernetes node must have NUMA-aware hardware.
   The system's NUMA topology can be inspected using tools such as `lscpu` or
   `numactl`.  
-  **Kubelet configuration**: The kubelet on each node must be configured to
   enable NUMA topology awareness. This involves setting specific kubelet flags
   to activate the Topology Manager.  
-  **Topology Manager**: The Topology Manager is a critical component that
   coordinates resource allocation based on NUMA topology. It receives NUMA
   affinity hints from other managers (e.g., CPU Manager, Device Manager) and
   makes informed scheduling decisions.  
-  **Memory Manager**: The Kubernetes Memory Manager is responsible for
   managing memory allocation, including HugePages, in a NUMA-aware manner. It
   ensures that memory is allocated from the minimum number of NUMA nodes
   required. The Memory Manager uses a hint generation protocol to provide NUMA
   affinity hints to the Topology Manager.  
-  **Pod specification**: Pods can be specified to request NUMA-aware resource
   allocation through resource requests and limits, ensuring that they get
   allocated in alignment with the NUMA topology.  
-  **Scheduler awareness**: The Kubernetes scheduler works in conjunction with
   the Topology Manager to place pods on nodes that meet their NUMA affinity
   requirements. The scheduler considers NUMA topology during the scheduling
   process to optimise performance.  
-  **Node Feature Discovery (NFD)**: Node Feature Discovery can be used to
   label nodes with their NUMA capabilities, providing the scheduler with
   information to make more informed placement decisions.  
-  **Resource quotas and limits**: Kubernetes allows defining resource quotas
   and limits to control the allocation of NUMA-aware resources across
   namespaces. This helps in managing and isolating resource usage effectively.
-  **Monitoring and metrics**: Monitoring tools such as Prometheus and Grafana
   can be used to track the performance and resource utilisation of NUMA-aware
   workloads. Metrics include CPU and memory usage per NUMA node, helping in
   optimising and troubleshooting performance-sensitive applications.  
-  **Isolation and security**: Security contexts and isolation mechanisms
   ensure that NUMA-aware workloads are protected and run in a controlled
   environment. This includes setting privileged containers and configuring
   namespaces to avoid resource contention.  
-  **Performance tuning**: Additional performance tuning can be achieved by
   configuring kernel parameters and using tools like numactl to bind processes
   to specific NUMA nodes.

## SR-IOV (Single Root I/O Virtualization)

SR-IOV enables a single physical network device to appear as multiple separate
virtual devices. This can be beneficial for network-intensive applications that
require direct access to the network hardware.

### Key features

-  **Multiple Virtual Functions (VFs)**: SR-IOV enables a single physical
   network device to be partitioned into multiple virtual functions (VFs), each
   of which can be assigned to a virtual machine or container as a separate
   network interface.  
-  **Direct hardware access**: By providing direct access to the physical
   network device, SR-IOV bypasses the software-based network stack, reducing
   overhead and improving network performance and latency.  
-  **Improved network throughput**: Applications can achieve higher network
   throughput as SR-IOV enables high-speed data transfer directly
   between the network device and the application.  
-  **Reduced CPU utilisation**: Offloading network processing to the hardware
   reduces the CPU load on the host system, freeing up CPU resources for other
   tasks and improving overall system performance.  
-  **Isolation and security**: Each virtual function (VF) is isolated from
   others, providing security and stability. This isolation ensures that issues
   in one VF do not affect other VFs or the physical function (PF).  
-  **Dynamic resource allocation**: SR-IOV supports dynamic allocation of
   virtual functions, enabling resources to be adjusted based on application
   demands without requiring changes to the physical hardware setup.  
-  **Enhanced virtualisation support**: SR-IOV is particularly beneficial in
   virtualised environments, enabling better network performance for virtual
   machines and containers by providing them with dedicated network interfaces.
-  **Kubernetes integration**: Kubernetes supports SR-IOV through the use of
   network device plugins, enabling the automatic discovery, allocation,
   and management of virtual functions.  
-  **Compatibility with Network Functions Virtualization (NFV)**: SR-IOV is
   widely used in NFV deployments to meet the high-performance networking
   requirements of virtual network functions (VNFs), such as firewalls,
   routers and load balancers.  
-  **Reduced network latency**: As network packets can bypass the
   hypervisor's virtual switch, SR-IOV significantly reduces network latency,
   making it ideal for latency-sensitive applications.

### Application to Kubernetes

The architecture for SR-IOV (Single Root I/O Virtualization) in Kubernetes
involves several components and configurations to ensure that virtual functions
(VFs) from a single physical network device can be managed and allocated
efficiently. This setup enhances network performance and provides direct access
to network hardware for applications requiring high throughput and low latency.
Here are the key architectural components and their roles:

-  **Node configuration**: Each Kubernetes node with SR-IOV capable hardware
   must have the SR-IOV drivers and tools installed. This includes the SR-IOV
   network device plugin and associated drivers.  
-  **SR-IOV enabled network interface**: The physical network interface card
   (NIC) must be configured to support SR-IOV. This involves enabling SR-IOV in
   the system BIOS and configuring the NIC to create virtual functions (VFs).  
-  **SR-IOV network device plugin**: The SR-IOV network device plugin is
   deployed as a DaemonSet in Kubernetes. It discovers SR-IOV capable network
   interfaces and manages the allocation of virtual functions (VFs) to pods.  
-  **Device Plugin Configuration**: The SR-IOV device plugin requires a
   configuration file that specifies the network devices and the number of
   virtual functions (VFs) to be managed.  
-  **Pod specification**: Pods can request SR-IOV virtual functions by
   specifying resource requests and limits in the pod specification. The SR-IOV
   device plugin allocates the requested VFs to the pod.  
-  **Scheduler awareness**: The Kubernetes scheduler must be aware of the
   SR-IOV resources available on each node. The device plugin advertises the
   available VFs as extended resources, which the scheduler uses to place pods
   accordingly. Scheduler configuration ensures pods with SR-IOV requests are
   scheduled on nodes with available VFs.  
-  **Resource quotas and limits**: Kubernetes enables the definition of
   resource quotas and limits to control the allocation of SR-IOV resources
   across namespaces. This helps manage and isolate resource usage effectively.
-  **Monitoring and metrics**: Monitoring tools such as Prometheus and Grafana
   can be used to track the performance and resource utilisation of
   SR-IOV-enabled workloads. Metrics include VF allocation, network throughput,
   and latency, helping optimise and troubleshoot performance-sensitive
   applications.  
-  **Isolation and security**: SR-IOV provides isolation between VFs, ensuring
   that each VF operates independently and securely. This isolation is critical
   for multi-tenant environments where different workloads share the same
   physical network device.  
-  **Dynamic resource allocation**: SR-IOV supports dynamic allocation and
   deallocation of VFs, enabling Kubernetes to adjust resources based on
   application demands without requiring changes to the physical hardware
   setup.

## DPDK (Data Plane Development Kit) 

The Data Plane Development Kit (DPDK) is a set of libraries and drivers for
fast packet processing. It is designed to run in user space, so that
applications can achieve high-speed packet processing by bypassing the kernel.
DPDK is used to optimise network performance and reduce latency, making it
ideal for applications that require high-throughput and low-latency networking,
such as telecommunications, cloud data centres and network functions
virtualisation (NFV).

### Key features

-  **High performance**: DPDK can process millions of packets per second per
   core, using multi-core CPUs to scale performance.  
-  **User-space processing**: By running in user space, DPDK avoids the
   overhead of kernel context switches and uses HugePages for better
   memory performance.  
-  **Poll Mode Drivers (PMD)**: DPDK uses PMDs that poll for packets instead of
   relying on interrupts, which reduces latency.

### DPDK architecture

The main goal of the DPDK is to provide a simple, complete framework for fast
packet processing in data plane applications. Anyone can use the code to
understand some of the techniques employed, to build upon for prototyping or to
add their own protocol stacks. 

The framework creates a set of libraries for specific environments through the
creation of an Environment Abstraction Layer (EAL), which may be specific to a
mode of the Intel® architecture (32-bit or 64-bit), user space
compilers or a specific platform. These environments are created through the
use of Meson files (needed by Meson, the software tool for automating the
building of software that DPDK uses) and configuration files. Once the EAL
library is created, the user may link with the library to create their own
applications. Other libraries, outside of EAL, including the Hash, Longest
Prefix Match (LPM) and rings libraries are also provided. Sample applications
are provided to help show the user how to use various features of the DPDK.

The DPDK implements a run to completion model for packet processing, where all
resources must be allocated prior to calling data plane applications, running
as execution units on logical processing cores. The model does not support a
scheduler and all devices are accessed by polling. The primary reason for not
using interrupts is the performance overhead imposed by interrupt processing.

In addition to the run-to-completion model, a pipeline model may also be used
by passing packets or messages between cores via the rings. This enables work
to be performed in stages and is a potentially more efficient use of code on
cores. This is suitable for scenarios where each pipeline must be mapped to a
specific application thread or when multiple pipelines must be mapped to the
same thread. 

### Application to Kubernetes

The architecture for integrating the Data Plane Development Kit (DPDK) into
Kubernetes involves several components and configurations to ensure high-speed
packet processing and low-latency networking. DPDK enables applications to
bypass the kernel network stack, providing direct access to network hardware
and significantly enhancing network performance. Here are the key architectural
components and their roles:

-  **Node configuration**: Each Kubernetes node must have the DPDK libraries
   and drivers installed. This includes setting up HugePages and binding
   network interfaces to DPDK-compatible drivers.  
-  **HugePages configuration**: DPDK requires HugePages for efficient memory
   management. Configure the system to reserve HugePages.  
-  **Network interface binding**: Network interfaces must be bound to
   DPDK-compatible drivers (e.g., vfio-pci) to be used by DPDK applications.  
-  **DPDK application container**: Create a Docker container image with the
   DPDK application and necessary libraries. Ensure that the container runs
   with appropriate privileges and mounts HugePages.  
-  **Pod specification**: Deploy the DPDK application in Kubernetes by
   specifying the necessary resources, including CPU pinning and HugePages, in
   the pod specification.  
-  **CPU pinning**: For optimal performance, DPDK applications should use
   dedicated CPU cores. Configure CPU pinning in the pod specification.  
-  **SR-IOV for network interfaces**: Combine DPDK with SR-IOV to provide
   high-performance network interfaces. Allocate SR-IOV virtual functions (VFs)
   to DPDK pods.  
-  **Scheduler awareness**: The Kubernetes scheduler must be aware of the
   resources required by DPDK applications, including HugePages and CPU
   pinning, to place pods appropriately on nodes with sufficient resources.  
-  **Monitoring and metrics**: Use monitoring tools like Prometheus and Grafana
   to track the performance of DPDK applications, including network throughput,
   latency and CPU usage.  
-  **Resource quotas and limits**: Define resource quotas and limits to control
   the allocation of resources for DPDK applications across namespaces,
   ensuring fair resource distribution and preventing resource contention.  
-  **Isolation and security**: Ensure that DPDK applications run in isolated
   and secure environments. Use security contexts to provide the necessary
   privileges while maintaining security best practices.


<!-- LINKS -->

[no_hz]: https://www.kernel.org/doc/Documentation/timers/NO_HZ.txt