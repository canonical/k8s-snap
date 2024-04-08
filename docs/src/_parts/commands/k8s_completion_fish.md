## k8s completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	k8s completion fish | source

To load completions for every new session, execute once:

	k8s completion fish > ~/.config/fish/completions/k8s.fish

You will need to start a new shell for this setup to take effect.


```
k8s completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
      --timeout duration       the max time to wait for the command to execute (default 1m30s)
```

### SEE ALSO

* [k8s completion](k8s_completion.md)	 - Generate the autocompletion script for the specified shell

