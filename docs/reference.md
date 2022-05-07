# Reference

## Container

Most of the Polycrate magic happens inside of a Docker container running on the system that invokes the `polycrate` command.

The container will be started whenever you call a Block Action. It is based on public image (`ghcr.io/polycrate/polycrate`) provided by [ayedo](https://www.ayedo.de) ([Dockerfile](https://github.com/polycrate/polycrate/Dockerfile.goreleaser)) and contains most of the best-practice tooling of cloud-native development and operations.

## Workspace

Whatever you build with Polycrate happens in your **workspace**. The worpspace contains configuration and lifecycle artifacts. It's a directory on your filesystem that can be synced and collaborated on via **git** or other tooling.

## Ansible Inventory

Polycrate can consume yaml-formated [Ansible inventory](https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html) files inside the artifacts directory of a block. The inventory file **must** be named `inventory.yml`.

An inventory file can be created automatically by a block or provided manually (useful for existing infrastructure).

The inventories can be consumed by the owning block itself or by other blocks using the `inventory` stanza in the block configuration:

```yaml
# block.poly
name: plugin-a
  inventory:
    from: plugin-b
```

This will add an environment variable (`ANSIBLE_INVENTORY=path/to/inventory/of/plugin-b`) to the container that points Ansible to the right inventory to work with.

## Kubeconfig

Polycrate is integrated with Kubernetes and can connect to a cluster using a [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file. By default, Polycrate looks for kubeconfig files inside the artifacts directory of a block. 

A kubeconfig file can be created automatically by a block or provided manually (useful for existing infrastructure).

The kubeconfig file can be consumed by the owning block itself or by other blocks using the `kubeconfig` stanza in the block configuration:

```yaml
# block.poly
name: plugin-a
  kubeconfig:
    from: plugin-b
```

This will add an environment variable (`KUBECONFIG=path/to/kubeconfig/of/plugin-b`) to the container that points kubectl, etc to the right kubeconfig to work with.

## Blocks

A Polycrate workspace is a modular system built out of so called **blocks** and their configuration. Blocks can be inherited `from` other blocks which merges configuration of the parent block into its child and changes the workdir to the parent block's workdir whenever an [action](#actions) runs.

!!! note
    When merging a parent block's config into a child block, existing config in the child block will not be overriden. The most common scenario where this is relevant is when you define defaults inside a `block.poly` file and overwrite them in your `workspace.poly` file.

Blocks can be created on the fly by defining their configuration in the workspace configuration directly. If the block is composed of custom code, simply create a directory (the so-called **block directory**) in the **blocks root** (`mkdir blocks/custom-block`) and place your code there.

!!! note
    The name of the directory shoult be the same you defined in the `name` stanza of that Block.

When a custom workdir for a block exists, Polycrate will change to this workdir when executing an action of that block.

A block can be configured through a block configuration file (`block.poly`) in its block dir (this is a good way to store defaults) or by adding the block configuration directly to the workspace configuration. Workspace-level configuration will always have precedence over block-level configuration.

```yaml
# block.poly
name: custom-block
config:
  foo: bar
  baz:
    foo: bar
actions:
  install:
    script:
      - echo "Install"
  uninstall:
    script:
      - echo "Uninstall"
```

## Block configuration

The **block configuration** (default: `block.poly`) holds the configuration for a [block](#block) and must be located in the block directory.

### Actions

A block can expose an arbitrary amount of actions. Actions are used to implement the actual functionality of a block. Examples would be `install` or `uninstall`, but also `status` or `init`.

The `script` section of an action is a list of commands that will be merged into a Bash script and executed inside the Polycrate container (or locally if you specifiy `--local`) when you run the action.

!!! note
    You can use [Go Templates](https://learn.hashicorp.com/tutorials/nomad/go-template-syntax) in your action scripts. 

    ```yaml
    [...]
    actions:
      - name: template
          script:
            - echo This is Action {{ .Action.Name }} in Block {{ .Block.Name }}
            - echo Running in Workspace {{ .Workspace.Name }}
    [...]
    ```

!!! note
    Action names are limited to certain characters: `^[a-zA-Z]+([-/_]?[a-zA-Z0-9_]+)+$`.

    This constraint applies to **ALL** `name` stanzas in Polycrate.

!!! note
    Polycrate does not persist data between runs apart from changes made to the [workspace](#workspace) directory (mounted at `/workspace`) inside the execution container).

!!! note
    It's fine to write data to the [workspace](#workspace) directory. However it's best-practice to use [Artifacts](#artifacts) to persist custom data.

## Artifacts

Artifacts can be stored in the **artifacts root** inside your workspace (which is configurable using `--artifacts-root` and defaults to `artifacts`).

By default, Polycrate looks for Ansible Inventories and Kubeconfigs in the Artifacts Directory of a Block.

## Workspace configuration

The **workspace configuration** (default: `workspace.poly`) holds the configuration for a [workspace](#workspace) and must be located inside the workspace directory.

!!! note
    You can specify a custom workspace configuration by using `--workspace-config YOUR/CUSTOM/workspace-config.yml`. This can be especially helpful when using Polycrate in CI.
