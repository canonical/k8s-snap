workspace {

    model {
        user = person "Clients" {
        }


        cluster = softwareSystem "Cluster" {




            ingress = container "Ingress" {
                user -> this "External Traffic (HTTP/HTTPS)"
            }
            
            service = container "Service 1" {
                technology "Kubernetes Service"
                ingress -> this
            }

            service2 = container "Service 2" {
                technology "Kubernetes Service"
                ingress -> this
            }
            
            pod = container "Pod 1"{
                technology "Kubernetes Pod"
                service -> this
            }

            pod2 = container "Pod 2"{
                technology "Kubernetes Pod"
                service -> this
            }
        
            pod3 = container "Pod 3"{
                technology "Kubernetes Pod"
                service2 -> this
            }        
        
        }

    }

    views {

        container cluster {
            include *
            autolayout tb
        }

        theme default
    }

}
