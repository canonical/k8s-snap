## k8s completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(k8s completion zsh)

To load completions for every new session, execute once:

#### Linux:

	k8s completion zsh > "${fpath[1]}/_k8s"

#### macOS:

	k8s completion zsh > $(brew --prefix)/share/zsh/site-functions/_k8s

You will need to start a new shell for this setup to take effect.


```
k8s completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
```

### SEE ALSO

* [k8s completion](k8s_completion.md)	 - Generate the autocompletion script for the specified shell

