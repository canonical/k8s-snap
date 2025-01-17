# Install with Multipass (Ubuntu/macOS/Windows)

[Multipass] is a simple way to run Ubuntu in a virtual machine, no matter
what your underlying OS. It is the recommended way to run {{product}} on
Windows and macOS systems, and is equally useful for running multiple instances
of the `k8s` snap on Ubuntu too.

## Install Multipass

Choose your OS for the install procedure

````{tabs}

```{group-tab} Ubuntu/Linux

Multipass is shipped as a snap for Ubuntu and other Linux distributions which
support the [snap package system][snap-support].

    sudo snap install multipass


```

```{group-tab} Windows

Windows users should download and install the Multipass installer from the
website.

The [latest Windows version][] is available to download, though you may wish
to visit the [Multipass website][] for more details.


```

```{group-tab} macOS

Users running macOS should download and install the Multipass installer from the
website.

The [latest macOS version] is available to download,
though you may wish to visit the [Multipass website][] for more details, including
an alternate install method using `brew`.

```

````

## Create an instance

The `k8s` snap will require a certain amount of resources, so the default
settings for a Multipass VM aren't going to be suitable. Exactly what resources
will be required depends on your use case. We recommend at least 4G of memory
and 20G of disk space for each instance.

Open a terminal (or Shell on Windows) and enter the following command:

```
multipass launch 24.04 --name k8s-node --memory 4G --disk 20G --cpus 2
```

This command specifies:

- `24.04`: The Ubuntu image used as the base for the instance
- `--name`: The name by which you will refer to the instance
- `--memory`: The memory to allocate
- `--disk`: The disk space to allocate
- `--cpus`: The number of CPU cores to reserve for this instance

For more details of creating instances with Multipass, please see the
[Multipass documentation][Multipass-options] about instance creation.

## Access the created instance

To access the image you just created, run:

```
multipass shell k8s-node
```

This will immediately open a shell on the instance, so further commands you
enter will be executed on the Ubuntu instance you created.

You can now use this terminal to install the `k8s` snap, following the standard
[install instructions][], or following along with the [Getting started][]
tutorial if you are new to {{product}}.

To end the shell session on the instance, enter:

```
exit
```

...and you will be returned to the original terminal session.

## Stop/Remove the instance

The instance you created will keep running in the background until it is either
stopped or the host computer is shut down. You can stop the running instance at
any time by running:

```
multipass stop k8s-node
```

And it can be permanently removed with:

```
multipass delete k8s-node -p
```

<!-- LINKS -->
<!-- markdownlint-disable MD053 -->
[Multipass]:https://multipass.run/
[snap-support]: https://snapcraft.io/docs/installing-snapd
[Multipass-options]: https://canonical.com/multipass/docs/tutorial#p-71169-create-a-customised-instance
[install instructions]: ./snap
[Getting started]: ../../tutorial/getting-started
[Multipass website]: https://multipass.run/docs
[latest Windows version]:https://canonical.com/multipass/download/windows
[latest macOS version]:https://canonical.com/multipass/download/macos
