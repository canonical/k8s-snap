## Datastore Node Configuration

### Datastore Node Configuration

#### Control 2.1

##### Description:

Ensure that the --cert-file and --key-file arguments are set as
appropriate (Automated)


##### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


#### Control 2.2

##### Description:

Ensure that the --client-cert-auth argument is set to true
(Automated)


##### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


#### Control 2.3

##### Description:

Ensure that the --auto-tls argument is not set to true
(Automated)


##### Remediation:

Not applicable. Canonical K8s uses dqlite and the communication
to this service is done through a
local socket
(/var/snap/k8s/common/var/lib/k8s-dqlite/k8s-dqlite.sock)
accessible to users with root permissions.


#### Control 2.4

##### Description:

Ensure that the --peer-cert-file and --peer-key-file arguments
are set as appropriate (Automated)


##### Remediation:

The certificate pair for dqlite and tls peer communication is
/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt and
/var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key.


##### Audit (as root):

```
if test -e /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.crt && test -e /var/snap/k8s/common/var/lib/k8s-dqlite/cluster.key; then echo 'certs-found'; fi
```

##### Expected output:

```
certs-found
```

#### Control 2.5

##### Description:

Ensure that the --peer-client-cert-auth argument is set to true
(Automated)


##### Remediation:

Dqlite peer communication uses TLS unless the --enable-tls is
set to false in
/var/snap/k8s/common/args/k8s-dqlite.


##### Audit (as root):

```
/bin/cat /var/snap/k8s/common/args/k8s-dqlite | /bin/grep enable-tls || true; echo $?
```

##### Expected output:

```
0
```

#### Control 2.6

##### Description:

Ensure that the --peer-auto-tls argument is not set to true
(Automated)


##### Remediation:

Not applicable. Canonical K8s uses dqlite and tls peer
communication uses the certificates
created upon the snap creation.


#### Control 2.7

##### Description:

Ensure that a unique Certificate Authority is used for the
datastore (Manual)


##### Remediation:

Not applicable. Canonical K8s uses dqlite and tls peer
communication uses certificates
created upon cluster setup.


