## Etcd Node Configuration

### Etcd Node Configuration

#### Control 2.1

Description: Ensure that the --cert-file and --key-file arguments are set as
appropriate (Automated)

Audit:

```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Expected output:

```
ETCD_CERT_FILE and ETCD_KEY_FILE are set
```

Remediation:

Follow the etcd service documentation and configure TLS
encryption.
Then, edit the etcd pod specification file
/etc/kubernetes/manifests/etcd.yaml
on the master node and set the below parameters.
--cert-file=</path/to/ca-file>
--key-file=</path/to/key-file>

#### Control 2.2

Description: Ensure that the --client-cert-auth argument is set to true
(Automated)

Audit:

```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Expected output:

```
ETCD_CLIENT_CERT_AUTH is set to true
```

Remediation:

Edit the etcd pod specification file /etc/default/etcd on the master
node and set the below parameter.
--client-cert-auth="true"

#### Control 2.3

Description: Ensure that the --auto-tls argument is not set to true
(Automated)

Audit:

```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Expected output:

```
ETCD_AUTO_TLS is not set
```

Remediation:

Edit the etcd pod specification file /etc/default/etcd on the master
node and either remove the --auto-tls parameter or set it to
false.
  --auto-tls=false

#### Control 2.4

Description: Ensure that the --peer-cert-file and --peer-key-file arguments
are set as appropriate (Automated)

Audit:

```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Expected output:

```
ETCD_PEER_CERT_FILE and ETCD_PEER_KEY_FILE are set
```

Remediation:

Follow the etcd service documentation and configure peer TLS
encryption as appropriate
for your etcd cluster.
Then, edit the etcd pod specification file /etc/default/etcd on the
master node and set the below parameters.
--peer-client-file=</path/to/peer-cert-file>
--peer-key-file=</path/to/peer-key-file>

#### Control 2.5

Description: Ensure that the --peer-client-cert-auth argument is set to true
(Automated)

Audit:

```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Expected output:

```
ETCD_PEER_CLIENT_CERT_AUTH is set to true
```

Remediation:

Edit the etcd pod specification file /etc/default/etcd on the master
node and set the below parameter.
--peer-client-cert-auth=true

#### Control 2.6

Description: Ensure that the --peer-auto-tls argument is not set to true
(Automated)

Audit:

```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Expected output:

```
ETCD_PEER_AUTO_TLS is not set
```

Remediation:

Edit the etcd pod specification file /etc/default/etcd on the master
node and either remove the --peer-auto-tls parameter or set it
to false.
--peer-auto-tls=false

#### Control 2.7

Description: Ensure that a unique Certificate Authority is used for etcd
(Manual)

Audit:

```
/bin/ps -ef | /bin/grep etcd | /bin/grep -v grep
```

Expected output:

```
ETCD_TRUSTED_CA_FILE is set
```

Remediation:

[Manual test]
Follow the etcd documentation and create a dedicated certificate
authority setup for the
etcd service.
Then, edit the etcd pod specification file /etc/default/etcd on the
master node and set the below parameter.
--trusted-ca-file=</path/to/ca-file>

