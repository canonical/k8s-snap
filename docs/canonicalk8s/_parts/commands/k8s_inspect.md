## k8s inspect

Generate inspection report

### Synopsis

Generate an inspection report tarball containing diagnostics and relevant information from a Kubernetes node.

This command collects diagnostics from either a control-plane or worker node and compiles them into
a tarball report. The collected data includes service arguments, Kubernetes cluster info, SBOM, system
diagnostics, network diagnostics, and more. The command needs to be run with elevated permissions (sudo).

Arguments:
  output-file             (Optional) The full path and filename for the generated tarball.
                          If not provided, a default filename based on the current date
                          and time will be used.
  --all-namespaces        (Optional) Acquire detailed debugging information, including logs
                          from all Kubernetes namespaces.
  --num-snap-log-entries  (Optional) The maximum number of log entries to collect
                          from snap services. Default: 100000.
  --timeout               (Optional) The maximum time in seconds to wait for a command.
                          Default: 180s.
  --core-dump-dir         (Optional) Core dump location. Default: /var/crash.


```
k8s inspect <output-file> [flags]
```

```{warning}
This command removes sensitive information, such as secrets and certificates,
from the generated report, but it is recommended to review the contents of
the tarball before sharing it with support or other parties.
```

### Options

```
  -h, --help   help for inspect
```

### SEE ALSO

* [k8s](k8s.md)	 - Canonical Kubernetes CLI

