## Etcd Node Configuration

### Etcd Node Configuration

#### Control 2.1

Description: Ensure that the ETCD_CERT_FILE and ETCD_KEY_FILE environment
variables are set as appropriate (Automated)

Audit:

```
cat /proc/$(pidof /usr/bin/etcd)/environ
```

Expected output:

```
ETCD_CERT_FILE and ETCD_KEY_FILE are set
```

Remediation:

Follow the etcd service documentation and configure TLS
encryption.
Then, edit the etcd daemon configuration file /etc/default/etcd
on the master node and set the below variables.
ETCD_CERT_FILE=</path/to/ca-file>
ETCD_KEY_FILE=</path/to/key-file>

#### Control 2.2

Description: Ensure that the ETCD_CLIENT_CERT_AUTH variable is set to true
(Automated)

Audit:

```
cat /proc/$(pidof /usr/bin/etcd)/environ
```

Expected output:

```
ETCD_CLIENT_CERT_AUTH is set to true
```

Remediation:

Edit the etcd daemon configuration file /etc/default/etcd on the master
node and set the below variable.
ETCD_CLIENT_CERT_AUTH=true

#### Control 2.3

Description: Ensure that the ETCD_AUTO_TLS argument is not set to true
(Automated)

Audit:

```
cat /proc/$(pidof /usr/bin/etcd)/environ
```

Expected output:

```
ETCD_AUTO_TLS is not set
```

Remediation:

Edit the etcd daemon configuration file /etc/default/etcd on the master
node and either remove the ETCD_AUTO_TLS variable or set it to
false.
  ETCD_AUTO_TLS=false

#### Control 2.4

Description: Ensure that the ETCD_PEER_CERT_FILE and ETCD_PEER_KEY_FILE
variables are set as appropriate (Automated)

Audit:

```
cat /proc/$(pidof /usr/bin/etcd)/environ
```

Expected output:

```
ETCD_PEER_CERT_FILE and ETCD_PEER_KEY_FILE are set
```

Remediation:

Follow the etcd service documentation and configure peer TLS
encryption as appropriate
for your etcd cluster.
Then, edit the etcd daemon configuration file /etc/default/etcd on the
master node and set the below variables.
ETCD_PEER_CERT_FILE=</path/to/peer-cert-file>
ETCD_PEER_KEY_FILE=</path/to/peer-key-file>

#### Control 2.5

Description: Ensure that the ETCD_PEER_CLIENT_CERT_AUTH variable is set to
true (Automated)

Audit:

```
cat /proc/$(pidof /usr/bin/etcd)/environ
```

Expected output:

```
ETCD_PEER_CLIENT_CERT_AUTH is set to true
```

Remediation:

Edit the etcd daemon configuration file /etc/default/etcd on the master
node and set the below parameter.
ETCD_PEER_CLIENT_CERT_AUTH=true

#### Control 2.6

Description: Ensure that the ETCD_PEER_AUTO_TLS argument is not set to true
(Automated)

Audit:

```
cat /proc/$(pidof /usr/bin/etcd)/environ
```

Expected output:

```
ETCD_PEER_AUTO_TLS is not set
```

Remediation:

Edit the etcd daemon configuration file /etc/default/etcd on the master
node and either remove the ETCD_PEER_AUTO_TLS parameter or set
it to false.
ETCD_PEER_AUTO_TLS=false

#### Control 2.7

Description: Ensure that a unique Certificate Authority is used for etcd
(Manual)

Audit:

```
cat /proc/$(pidof /usr/bin/etcd)/environ
```

Expected output:

```
ETCD_TRUSTED_CA_FILE is set
```

Remediation:

Follow the etcd documentation and create a dedicated certificate
authority setup for the
etcd service.
Then, edit the etcd daemon configuration file /etc/default/etcd on the
master node and set the below parameter.
ETCD_TRUSTED_CA_FILE=</path/to/ca-file>

