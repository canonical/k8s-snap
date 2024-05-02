# Ingress Configuration

## Format Specification

### ingress.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled. If omitted defaults to `false`

### ingress.default-tls-secret
**Type:** `string`<br>
**Required:** `No` <br>

Sets the name of the secret to be used for providing default encryption to ingresses.

Ingresses can specify another TLS secret in their resource definitions, in which case the default secret won't be used.

### ingress.enable-proxy-protocol
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the proxy protocol should be enabled for ingresses. If omitted defaults to `false`
