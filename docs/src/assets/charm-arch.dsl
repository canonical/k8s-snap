workspace {

    model {
        user = person "Administrator"


        jujuSystem = softwareSystem "Juju" {

            k8sCharm = container "K8s" "K8s Charm" {
                technology "Charmed Operator"
                
            }

            container "K8s Relation Data" {
                k8sCharm -> this "Reads from and writes to"
                this -> k8sCharm "Retrieves Peer Data"
            }

            charmWorker = container "K8s Worker" "K8s Worker Charm" {
                technology "Charmed Operator"
            }

            container "K8s Worker Relation Data" {
                technology "Juju Relation Databag"
                k8sCharm -> this "Share Cluster Data"
                charmWorker -> this "Reads from and writes to"
                this -> charmWorker "Retrieves Peer Data"
            }


            jujuController = container "Juju Controller" {
                technology "Snap Package"
                this -> k8sCharm "Manages"
                this -> charmWorker "Manages"
            }

            jujuCLI = container "Juju Client" {
                technology "Snap Package"
                user -> this "Uses"
                this -> jujuController "Manages"
            }
        
        
            externalCharms = container "Compatible Charms" "Other Compatible Canonical Charms" {
                k8sCharm -> this "Integrates with"
                charmWorker -> this "Integrates with
            }
        
        }

    }

    views {

        container jujuSystem {
            include *
            autolayout lr
        }

        theme default
    }

}
