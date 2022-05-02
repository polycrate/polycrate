# Reference

## Container

Most of the Polycrate magic happens inside of a Docker container running on the system that invokes the `polycrate` command.

The container will be started whenever you call a Block Action. It is based on public image (`ghcr.io/polycrate/polycrate`) provided by [ayedo](https://www.ayedo.de) ([Dockerfile](https://github.com/polycrate/polycrate/Dockerfile.goreleaser)) and contains most of the best-practice tooling of cloud-native development and operations.

## Workspace

Whatever you build with Polycrate happens in your **Workspace**. The **Worpspace** contains configuration and lifecycle artifacts. It's a directory on your filesystem that can be synced and collaborated on via **git** or other tooling.

## Inventory

Polycrate can consume json-formated [Ansible inventory](https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html) files inside the **Artificats** directory of a **Block** or the **Workspace** root. The inventory file **must** be named `inventory.json`.

An inventory file can be created automatically by a Block or provided manually (useful for existing infrastructure).

## Kubeconfig

Polycrate is integrated with Kubernetes and can connect to a cluster using a [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file. By default, Polycrate looks for `kubeconfig.yml` in the **Workspace** directory. 

Additionally, you can specify a kubeconfig file using the `--kubeconfig PATH/TO/YOUR/kubeconfig` flag.

## Kubernetes

Kubernetes is the container orchestrator of choice in Cloudstack. Most of the [plugins](#plugins) (like [prometheus](plugins/prometheus.md)) are directly depending on Kubernetes  - they have a `need` for `kubernetes`. Some plugins (like [k3s](plugins/k3s.md)) can be used to install Kubernetes - they `provide` `kubernetes`.

!!! note
    Plugins `need` and `provide` certain functionalities. This way a dependency system can be established even for [custom plugins](plugins/custom_plugins.md).

## Block

A Polycrate Workspace is a modular system built out of so called **Blocks** that provide the actual functionality. A **Workspace** is composed of one or more **Blocks** and their configuration. Blocks can be inherited `from` other blocks which merges the configuration of the parent block into its child. Also, the workdir changes to the parent block's workdir whenever an **Action** runs.

!!! note
    Child-Block config overwrites Parent-Block config

Blocks can be created on the fly by defining their **Block Configuration** and **Actions** in the **Workspace Configuration** directly. If the Block is composed of custom code, simply create a directory (the so-called **Block Dir**) with the Block's name in the **Blocks Root** (`mkdir blocks/custom-plugin`) and place your code there.

!!! note
    The name of the directory must be the same you defined in `name` for that Block.

When a custom Workdir for a Block exists, Polycrate will change to this Workdir when executing an Action of that Block.

A Block can be configured though a **Block Configuration** file (`block.poly`) in its **Block Dir** (this is a good way to store defaults) or by adding the Block Configuration directly to the **Workspace Configuration**. Workspace Configuration will always overwrite Block Configuration.

```yaml
# Pluginfile
version: 1.2.3
subdomain: customdomain
config:
  foo: bar
  baz:
    foo: bar
commands:
  install:
    script:
      - echo "Install"
  uninstall:
    script:
      - echo "Uninstall"
```

!!! note
    The `version` key of the plugin config is required.

### Actions

A Block can expose an arbitrary amount of Actions. Actions can be used to implement the actual functionality of a plugin. Examples would be `install` or `uninstall`, but also `status` or `init`.

The `script` section of an Action is a list of commands that will be concatenated to a Bash script and executed inside the Polycrate container (or locally if you specifiy `--local`) when you run the Action.

While `init` and `destroy` should be defined by convention, you're free to implement anything here. Command names are limited to certain characters though: `a-zA-Z_-0-9`.

The Polycrate image is based on debian buster so you're free to install any additional software as part of a plugin command.

!!! note
    Polycrate does not persist data between invocations apart from changes made to the [Workspace](#workspace) directory (mounted at `/workspace` inside the execution container).

!!! note
    It's fine to write data to the [Workspace](#workspace) directory. However it's best-practice to use the **Artidacts Root** to persist custom data.


## Workspace Configuration

The **Workspace Configuration** (default: `workspace.poly`) holds the configuration for a [Workspace](#workspace) and is typically located inside the **Workspace**.

!!! note
    You can specify a custom **Workspace Config** with the CLI by using `--workspace-config YOUR/CUSTOM/Stackfile`. This can be especially helpful when using Cloudstack in CI.

## Block Configuration

The **Block Configuration** (default: `block.poly`) holds the configuration for a [Block](#block) and must be located in the Blocks's **Block Dir**.