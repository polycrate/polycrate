# Reference

## Workspace

Whatever you build with Polycrate happens in your **workspace**. The workspace contains configuration and lifecycle artifacts. It's a directory on your filesystem (the so called **workspace directory**; can be specified using `--workspace`) that can be synced and collaborated on via **git** or other tooling.

The workspace can be assembled using the [workspace configuration](#workspace-configuration).

## Container

Most of the Polycrate magic happens inside of a Docker container running on the system that invokes the `polycrate` command.

The container will be started whenever you run an action. It is based on a public image (`ghcr.io/polycrate/polycrate`) provided by [ayedo](https://www.ayedo.de) ([Dockerfile](https://github.com/polycrate/polycrate/Dockerfile.goreleaser)) and contains most of the best-practice tooling of cloud-native development and operations.

The container gives you access to a state-of-the-art DevOps runtime. Polycrate exports a [snapshot](#workspace-snapshot) of the workspace in various formats (yaml, json, environment vars, hcl, ...) and makes it available to the tooling inside the container so you can start building right away.

!!! note
    You can run actions locally instead of in a container using the `--local` flag. 

    The main purpose for this would be:

    - Use the Polycrate CLI inside its own container in CI
    - Run local actions like setting up developer environments

    This is currently **EXPERIMENTAL** and not well tested.

## Dockerfile

Polycrate looks for a Dockerfile in your workspace (defaults to `Dockerfile.poly`, can be configured using `--dockerfile`). If it finds one, it will build it and run the [container](#container) based on this image instead of the default image.

This can be used to persist changes to the workspace, like installing additional tools or libraries.

By convention, the Dockerfile should be built upon the official Polycrate image:

```
FROM ghcr.io/polycrate/polycrate:latest

RUN pip install hcloud==1.16.0
```

Using [loglevel](#loglevel) 2 you can see and debug the build process:

```
DEBU[0000] Found Dockerfile.poly in Workspace           
DEBU[0000] Building image 'polycrate-demo:latest', --build=true 
WARN[0000] Building custom image polycrate-demo:latest  
DEBU[0000] Assembling docker context                    
DEBU[0000] Building image                               
DEBU[0001] Step 1/2 : FROM ghcr.io/polycrate/polycrate  
DEBU[0001]  ---> 67237198f4a5                           
DEBU[0001] Step 2/2 : RUN pip install hcloud==1.16.0    
DEBU[0001]  ---> Using cache                            
DEBU[0001]  ---> 92a78743b4f4                           
DEBU[0001] Successfully built 92a78743b4f4              
DEBU[0001] Successfully tagged polycrate-demo:latest
```

## Workspace configuration

The **workspace configuration** (default: `workspace.poly`) holds the configuration for a [workspace](#workspace) and must be located inside the workspace directory.

!!! note
    You can specify a custom workspace configuration by using `--workspace-config YOUR/CUSTOM/workspace-config.yml`. This can be especially helpful when using Polycrate in CI.

```yaml
# workspace.poly
name: polycrate-demo
blocks:
  - name: custom-block
    config:
      foo: bar
```

## Blocks

A Polycrate workspace is a modular system built out of so called **blocks** and their configuration. Blocks can be inherited `from` other blocks which merges configuration of the parent block into its child and changes the workdir to the parent block's workdir whenever an [action](#actions) runs.

!!! note
    When merging a parent block's config into a child block, existing config in the child block will not be overriden. The most common scenario where this is relevant is when you define defaults inside a `block.poly` file and overwrite them in your `workspace.poly` file.

Blocks can be created on the fly by defining their configuration in the workspace configuration directly. If the block is composed of custom code, simply create a directory (the so-called **block directory**) in the **blocks root** (`mkdir blocks/custom-block`) and place your code there.

!!! note
    The name of the directory shoult be the same you defined in the `name` stanza of that Block.

When a custom workdir for a block exists, Polycrate will change to this workdir when executing an action of that block.

A block can be configured through a [block configuration](#block-configuration) file (`block.poly`) in its block dir (this is a good way to store defaults) or by adding the block configuration directly to the workspace configuration. Workspace-level configuration will always have precedence over block-level configuration.

## Block configuration

The **block configuration** (default: `block.poly`) holds the configuration for a single [block](#block) and must be located in the block directory.

```yaml
# block.poly
name: custom-block
config:
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

By default, Polycrate looks for [Ansible Inventories](#ansible-inventory) and [Kubeconfigs](#kubeconfig) in the Artifacts Directory of a Block.

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

## Workspace snapshot

Whenever you run an action with Polycrate, a **workspace snapshot** will be captured. This snapshot contains the computed workspace configuration and will be exported in the following formats:

### yaml

Polycrate exports the workspace snapshot to a yaml file that will be saved locally and mounted into the container. The path of the file will be exported to the `POLYCRATE_WORKSPACE_SNAPSHOT_YAML` environment variable.

### environment vars

Polycrate converts the workspace snapshot to a flattened map and exports it to the environment.

```bash
[...]
ACTION_BLOCK: custom-block
ACTION_NAME: hello
ACTION_SCRIPT_0: echo Hello
ACTION_SCRIPT_1: echo World
BLOCK_ACTIONS_0_BLOCK: custom-block
BLOCK_ACTIONS_0_NAME: hello
BLOCK_ACTIONS_0_SCRIPT_0: echo Hello
BLOCK_ACTIONS_0_SCRIPT_1: echo World
BLOCK_ARTIFACTS_CONTAINERPATH: /workspace/artifacts/blocks/custom-block
BLOCK_ARTIFACTS_LOCALPATH: $HOME/.polycrate/workspaces/polycrate-demo/artifacts/blocks/custom-block
BLOCK_CONFIG_FOO: bar
BLOCK_NAME: custom-block
BLOCK_VERSION: ""
BLOCK_WORKDIR_CONTAINERPATH: /workspace/blocks/custom-block
BLOCK_WORKDIR_LOCALPATH: $HOME/.polycrate/workspaces/polycrate-demo/blocks/custom-block
[...]
```

You can use the `--snapshot` flag when invoking `polycrate run`. This will prevent any action from running and instead dumps the workspace snapshot. Also, `polycrate workspace snapshot` has the same effect, but doesn't contain data about the current action and block.

## Loglevel

Polycrate supports 4 loglevel:

- Loglevel 0: The default. Will only print logs of type **WARN** or above
- Loglevel 1: Info-Level. Will be more verbose
- Loglevel 2: Debug-Level. You know ...
- Loglevel 3: Trace level. 

The loglevel will be mapped to the respective **Ansible verbosity** meaning `--loglevel 3` will result in Ansible executing as if you used `-vvv`.