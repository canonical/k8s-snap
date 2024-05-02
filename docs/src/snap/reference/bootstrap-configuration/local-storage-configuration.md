# Local-storage Configuration

## Format Specification

### local-storage.enabled
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the feature should be enabled. If omitted defaults to `false`

### local-storage.local-path
**Type:** `string`<br>
**Required:** `No` <br>

Sets the path to be used for storing volume data. If omitted defaults to `/var/snap/k8s/common/rawfile-storage`

### local-storage.reclaim-policy
**Type:** `string`<br>
**Required:** `No` <br>
**Possible Values:** `Retain | Recycle | Delete`

Sets the reclaim policy of the storage class. If omitted defaults to `Delete`

### local-storage.default
**Type:** `bool`<br>
**Required:** `No` <br>

Determines if the storage class should be set as default. If omitted defaults to `true`
