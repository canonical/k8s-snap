## k8s completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	k8s completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
k8s completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
      --timeout duration       the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s completion](k8s_completion.md)	 - Generate the autocompletion script for the specified shell

