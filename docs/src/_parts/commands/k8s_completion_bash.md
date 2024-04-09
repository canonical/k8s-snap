## k8s completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(k8s completion bash)

To load completions for every new session, execute once:

#### Linux:

	k8s completion bash > /etc/bash_completion.d/k8s

#### macOS:

	k8s completion bash > $(brew --prefix)/etc/bash_completion.d/k8s

You will need to start a new shell for this setup to take effect.


```
k8s completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --output-format string   set the output format to one of plain, json or yaml (default "plain")
```

### SEE ALSO

* [k8s completion](k8s_completion.md)	 - Generate the autocompletion script for the specified shell

