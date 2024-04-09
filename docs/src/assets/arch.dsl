!constant  c4 "c4.dsl"

workspace "Canonical K8s Workspace" {
    model {

        admin = person "K8s Admin"
        user = person "K8s User"
        charm = softwareSystem "Charm K8s" "Orchestrating the lifecycle management of K8s"

        external_lb = softwareSystem "Load Balancer" "External LB, offered by the substrate (cloud). Could be replaced by any alternative solution." "Extern"
        storage = softwareSystem "Storage" "External storage, offered by the substrate (cloud). Could be replaced by any storage solution." "Extern"
        iam = softwareSystem "Identity Management System" "The external identity system, offered by the substrate (cloud). Could be replaced by any alternative system." "Extern"
        external_datastore = softwareSystem "External datastore" "etcd" "Extern"
  
       k8s_snap = softwareSystem "K8s Snap Distribution" "The Kubernetes distribution in a snap" {

            kubectl = container "Kubectl" "kubectl client for accessing the cluster"

            kubernetes = container "Kubernetes Services" "API server, kubelet, kube-proxy, scheduler, kube-controller" {
                systemd = component "systemd daemons" "Daemons holding the k8s services" 
                apiserver = component "API server"
                kubelet = component "kubelet"
                kube_proxy = component "kube-proxy"
                scheduler = component "scheduler"
                kube_controller = component "kube-controller"
                network = component "Network CNI" "The network implementation of K8s (from Cilium)"
                storage_provider = component "Local storage provider" "Simple storage for workloads"
                ingress = component "Ingress" "Ingress for workloads (from Cilium)"
                gw = component "Gateway" "Gateway API for workloads (from Cilium)"
                dns = component "DNS" "Internal DNS"
                metrics_server = component "Metrics server" "Keep track of cluster metrics"
                loadbalancer = component "Load-balancer" "The load balancer (from Cilium)"
            }

            rt = container "Runtime" "Containerd and runc"

            k8sd = container "K8sd" "Daemon implementing the features available in the k8s snap" {
                cli = component "CLI" "The CLI the offered" "CLI"
                api = component "API via HTTP" "The API interface offered" "REST"
                cluster_manager = component "CLuster management" "Management of the cluster with the help of MicroCluster"
            }

            state = container "State" "Datastores holding the cluster state" {
                k8sd_db = component "k8sd-dqlite" "MicroCluster DB"
                k8s_dqlite = component "k8s-dqlite" "Datastore holding the K8s cluster state"
            }
        }

        admin -> cli "Administers the cluster"
        admin -> charm "Manages cluster's lifecycle"
        admin -> kubectl "Uses to manage the cluster"
        user -> loadbalancer "Interact with workloads hosted in K8s"
        charm -> api "Orchestrates the lifecycle management of K8s"

        k8s_snap -> storage "Hosted workloads use storage"
        k8s_snap -> iam "Users identity is retrieved"

        k8s_dqlite -> external_datastore "Stores cluster data" "" "Runtime"
        loadbalancer -> external_lb "Routes client requests" "" "Runtime"

        cluster_manager -> systemd "Configures"

        systemd -> apiserver "Is a service"
        systemd -> kubelet "Is a service"
        systemd -> kube_proxy "Is a service"
        systemd -> kube_controller "Is a service"
        systemd -> scheduler "Is a service"

        network -> apiserver "Keeps state in"
        dns -> apiserver "Keeps state in"
        apiserver -> k8s_dqlite "Uses by default"

        network -> ingress "May provide" "HTTP/HTTPS" "Runtime"
        network -> gw "May provide" "HTTP/HTTPS" "Runtime"
        network -> loadbalancer "May provide" "HTTP/HTTPS" "Runtime"

        cluster_manager -> k8sd_db "Keeps state in"

        kubectl -> apiserver "Interacts"
        api -> systemd "Configures"
        api -> rt "Configures"
        api -> cluster_manager "Uses"

        cli -> api "CLI is based on the API primitives"

    }
    views {

        systemLandscape Overview "K8s Snap Overview" {
          include * 
          autoLayout
        }

        container k8s_snap {
            include *
            autoLayout
            title "K8s Snap Context View"
        }

        component state {
            include *
            autoLayout
            title "Datastores"
        }

        component k8sd {
            include *
            autoLayout
            title "k8sd"
        }

        component kubernetes {
            include *
            autoLayout
            title "Kubernetes services"
        }

        styles {
            element "Person" {
                background #08427b
                color #ffffff
                fontSize 22
                shape Person
            }

            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Structurizr" {
                background #77FF44
                color #000000
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
            element "BuiltIn" {
                background #1988f6
                color #FFFFFF
            }
            element "Extern" {
                background #dddddd
                color #000000
            }

            element "Extension" {
                background #FFdd88
                color #000000
            }

            element "File" {
                shape Folder
                background #448704
                color #ffffff
            }

            relationship "Relationship" {
                dashed false
            }

            relationship "Runtime" {
                dashed true
                color #0000FF
            }

        }
    }

}
