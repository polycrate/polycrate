# Reference

## Workspace

Whatever you build with Polycrate happens in your **workspace**. The workspace contains configuration and lifecycle artifacts. It's a directory on your filesystem (the so called **workspace directory**; can be specified using `--workspace/-w`) that can be synced and collaborated on via **git** or other tooling.

The workspace can be assembled using the [workspace configuration](#workspace-configuration).

## Polycrate container

Most of the Polycrate magic happens inside of a Docker container running on the system that invokes the `polycrate` command.

The container will be started whenever you run an action. It is based on a public image (`cargo.ayedo.cloud/library/polycrate`) provided by [ayedo](https://www.ayedo.de) ([Dockerfile](https://gitlab.ayedo.de/polycrate/polycrate/Dockerfile.goreleaser)) and contains most of the best-practice tooling of cloud-native development and operations.

The container gives you access to a state-of-the-art DevOps runtime. Polycrate exports a [snapshot](#workspace-snapshot) of the workspace in various formats (yaml, json, environment vars, hcl, ...) and makes it available to the tooling inside the container so you can start building right away.

!!! note
    You can run actions locally instead of in a container using the `--local` flag. 

    The main purpose for this would be:

    - Use the Polycrate CLI inside its own container in CI
    - Run local actions like setting up developer environments

    This is currently **EXPERIMENTAL** and not well tested.

## Dockerfile

Polycrate looks for a Dockerfile in your workspace (defaults to `Dockerfile.poly`, can be configured using `--dockerfile`). If it finds one, it will build it and run the [container](#polycrate-container) based on this image instead of the default image.

This can be used to persist changes to the workspace, like installing additional tools or libraries.

By convention, the Dockerfile should be built upon the official Polycrate image:

```
FROM cargo.ayedo.cloud/library/polycrate:latest

RUN pip install hcloud==1.16.0
```

Using [loglevel](#loglevel) 2 you can see and debug the build process:

```
DEBU[0000] Found Dockerfile.poly in Workspace           
DEBU[0000] Building image 'polycrate-demo:latest', --build=true 
WARN[0000] Building custom image polycrate-demo:latest  
DEBU[0000] Assembling docker context                    
DEBU[0000] Building image                               
DEBU[0001] Step 1/2 : FROM cargo.ayedo.cloud/library/polycrate:latest
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
    You can specify a custom workspace configuration file by using `--workspace-config YOUR/CUSTOM/workspace-config.yml`. This can be especially helpful when using Polycrate in CI.

```yaml
# workspace.poly
name: polycrate-demo
blocks:
  - name: custom-block
    config:
      foo: bar
```

!!! note
    The workspace name is limited to certain characters: `^[a-zA-Z]+([-/_]?[a-zA-Z0-9_]+)+$`.

    This constraint applies to **ALL** `name` stanzas in Polycrate.

## Blocks

A Polycrate workspace is a modular system built out of so called **blocks**. Blocks are dedicated pieces of code/functionality that can be configured using the `config` stanza in the block configuration (default: `block.poly`) or the workspace configuration (default: `workspace.poly`). Blocks expose actions that can be executed using `polycrate run $BLOCK_NAME $ACTION_NAME`. 

Polycrate looks for blocks inside the **blocks root** (defaults to `blocks`). Nested directories (e.g. `blocks/foo/bar/baz`) are allowed. 

!!! note
    If a block's name contains one or multiple slashes (`/`) and is installed through the registry, it will be saved to a nested directory structure: the block `ayedo/k8s/harbor` will be saved to `blocks/ayedo/k8s/harbor`. This also applies to the block's [artifact directory](#artifacts).

### Inheritance

Blocks can be inherited `from` other blocks which merges configuration of the parent block into its child and changes the workdir to the parent block's workdir whenever an [action](#actions) runs. Such blocks that are based on other blocks are called [dynamic blocks](#dynamic-blocks).

!!! note
    When merging a parent block's config into a child block, existing config in the child block will not be overriden. The most common scenario where this is relevant is when you define defaults inside a `block.poly` file and overwrite them in your `workspace.poly` file.

```yaml
...
blocks:
  - name: harbor
    from: ayedo/k8s/harbor
...
```

### Dynamic blocks

Blocks can be created dynamically by defining their configuration in the workspace configuration directly or creating a directory (the so-called **block directory**) in the **blocks root** (`mkdir blocks/custom-block`). This directory can contain custom code and a block configuration file.

!!! note
    By convention, the name of the directory shoult be the same you defined in the `name` stanza of that Block.

When a custom workdir for a block exists, Polycrate will change to this workdir when executing an action of that block. It's also possible to build blocks that do not use custom code but only rely on the available tooling inside the [Polycrate container](#polycrate-container).

Workspace-level configuration (made in `workspace.poly`) will always have precedence over block-level configuration (made in `block.poly`).

### Dependencies

Polycrate supports workspace-level dependencies by using the `dependencies` stanza:

```yaml
...
dependencies:
  - ayedo/hcloud/inventory:0.0.1
  - ayedo/hcloud/k8s-infra:0.0.2
  - ayedo/hcloud/k8s:0.0.9
  - ayedo/k8s/nginx:0.0.2
  - ayedo/k8s/portainer:0.0.7
  - ayedo/k8s/cert-manager:0.0.3
  - ayedo/k8s/external-dns:0.0.21
  - ayedo/k8s/harbor:0.0.1
...
```

Polycrate checks the configured dependencies against the installed blocks in the workspace with every invocation of the `polycrate` command. If a dependency is missing, it will be downloaded.

### Pull blocks

To dynamically add blocks to the workspace from the [registry](#registry), you can run `polycrate pull $BLOCK_NAME:$BLOCK_VERSION`. If `$BLOCK_VERSION` is not defined, `latest` will be assumed.

Blocks can be uninstalled from the workspace by running `polycrate block uninstall BLOCK1 BLOCK2:0.0.1` or simply deleting the block's directory.

## Block configuration

The **block configuration** (defaults to `block.poly`) holds the configuration for a single [block](#block) and must be located in the block directory.

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

The `config` stanza of the block configuration is free form and not typed. You can use it to define the configuration structure of your block according to your needs.

!!! note
    Block names are limited to certain characters: `^[a-zA-Z]+([-/_]?[a-zA-Z0-9_]+)+$`.

    This constraint applies to **ALL** `name` stanzas in Polycrate.

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
    Polycrate does not persist data between runs apart from changes made to the [workspace](#workspace) directory (mounted at `/workspace` inside the execution container).

!!! note
    It's fine to write data to the [workspace](#workspace) directory. However it's best-practice to use [Artifacts](#artifacts) to persist custom data.

## Artifacts

Artifacts can be stored in the **artifacts root** inside your workspace (which is configurable using `--artifacts-root` and defaults to `artifacts`).

By default, Polycrate looks for [Ansible Inventories](#ansible-inventory) and [Kubeconfigs](#kubeconfig) in the Artifacts Directory of a Block.

For each block in the workspace a directory will automatically be created underneath the artifacts root (e.g. `artifacts/block/`).

## Ansible Inventory

Polycrate can consume yaml-formated [Ansible inventory](https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html) files inside the artifacts directory of a block. Polycrate looks for a file named `inventory.yml` by default - this can be overridden using the `inventory.filename` stanza in the block configuration.

An inventory file can be created automatically by a block or provided manually (useful for existing infrastructure).

The inventories can be consumed by the owning block itself or by other blocks using the `inventory` stanza in the block configuration:

```yaml
# block.poly
name: block-a
  inventory:
    from: block-b
    filename: inventory.yml
```

This will add an environment variable (`ANSIBLE_INVENTORY=path/to/inventory/of/block-b`) to the container that points Ansible to the right inventory to work with.

## Kubeconfig

Polycrate is integrated with Kubernetes and can connect to a cluster using a [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file. By default, Polycrate looks for kubeconfig files named `kubeconfig.yml` inside the artifacts directory of a block. This can be overridden using the `kubeconfig.filename` stanza in the block configuration.

A kubeconfig file can be created automatically by a block or provided manually (useful for existing infrastructure).

The kubeconfig file can be consumed by the owning block itself or by other blocks using the `kubeconfig` stanza in the block configuration:

```yaml
# block.poly
name: block-a
  kubeconfig:
    from: block-b
    filename: kubeconfig.yml
```

This will add an environment variable (`KUBECONFIG=path/to/kubeconfig/of/block-b`) to the container that points kubectl, etc to the right kubeconfig to work with.

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

Polycrate supports 3 loglevel:

- Loglevel 1: The default. Will only print logs of type **INFO** or above
- Loglevel 2: Debug-Level. You know ...
- Loglevel 3: Trace level. 

The loglevel will be mapped to the respective **Ansible verbosity** meaning `--loglevel 3` will result in Ansible executing as if you used `-vvv`.

## Sync

Polycrate can be configured to automatically sync with a git repository whenever it is invoked. Polycrate then writes a history log to `history.log` in the workspace directory and commits and pushes before and after invoking an action, saving the command that has been issued as well as the exit code the command resulted in.

Sync must be enabled in the workspace configuration:

```yaml
sync:
  enabled: true
  auto: true
  remote:
    url: ssh://git@gitlab.ayedo.de:10022/devops/workspaces/starkiller.git
```

If `sync.auto` is false, you need to manually push to the connected git repository. 

With `polycrate log $MESSAGE` you can write and sync arbitrary history logs.

## Registry

Polycrate blocks can be pushed and pulled to and from a OCI-compatible registry. By default, Polycrate uses `cargo.ayedo.cloud` as its registry targets to obtain blocks from. 

You can define your own registry by using `--registry-url` and pointing it to your OCI-compatible registry.

Polycrate does not implement authentication with the registry but instead makes use of your local Docker credential helper. To authenticate against a registry, run `docker login $REGISTRY_URL` before pushing or pulling blocks.

## Ansible

Polycrate provides a special integration with [Ansible](https://www.ansible.com/). The [workspace snapshot](#workspace-snapshot) that is being exported to yaml format and mounted to the Polycrate container will be consumed by Ansible automagically. As a result, the snapshot is available directly as top-level variables in Ansible which can be used in playbooks and templates.

The following example shows:

- the default configuration (`block.poly`) of a block called `traefik`
- the user-provided configuration for the block in `workspace.poly`
- the Ansible playbook using the exposed variables ( `block.config...`)
- an Ansible template using the exposed variables (`templates/docker-compose.yml.j2`)
- the resulting file `/polycrate/docker-compose.yml` that is templated to a remote host

!!! note
    Note how in `block.poly` the configured image is `traefik:2.6` but in `workspace.poly` it's `traefik:2.7`. In the resulting `docker-compose.yml`, the image is `traefik:2.7` as defaults in `block.poly` will be overridden by user-provided configuration in `workspace.poly`.

The `block` variable contains the configuration of the current block invoked by `polycrate run traefik install`. Additionally, there's a variable `workspace` available, that contains the fully compiled workspace including additional blocks that are available in the workspace.

Polycrate makes use of a special [Ansible Vars Plugin](https://docs.ansible.com/ansible/latest/plugins/vars.html) to read in the Yaml-Snapshot and expose it as top-level variables to the Ansible facts.


=== "block.poly"

    ```yaml
    name: traefik
    config:
      image: "traefik:v2.6"
      letsencrypt:
        email: ""
        resolver: letsencrypt
    actions:
      - name: install
        script:
          - ansible-playbook install.yml
      - name: uninstall
        script:
          - ansible-playbook uninstall.yml
      - name: prune
        script:
          - ansible-playbook prune.yml
    ```

=== "workspace.poly"

    ```yaml
    name: ansible-traefik-demo
    blocks:
      - name: traefik
        inventory:
          from: inventory-block
        config:
          letsencrypt:
            email: info@example.com
          image: traefik:2.7
    ```

=== "install.yml"

    ```yaml
    - name: "install"
      hosts: all
      gather_facts: yes
      tasks:
        - name: Create remote block directory
          ansible.builtin.file:
            path: "{{ item }}"
            state: directory
            mode: '0755'
          with_items:
            - "/polycrate/{{ block.name }}"

        - name: Copy compose file
          ansible.builtin.template:
            src: docker-compose.yml.j2
            dest: "/polycrate/{{ block.name }}/docker-compose.yml"

        - name: Deploy compose stack
          docker_compose:
            project_src: "/polycrate/{{ block.name }}"
            remove_orphans: true
            files:
              - docker-compose.yml
    ```

=== "templates/docker-compose.yml.j2"

    ```yaml
    version: "3.9"

    services:
      traefik:
        image: "{{ block.config.image }}"
        container_name: "traefik"
        command:
          - "--providers.docker=true"
          - "--providers.docker.exposedbydefault=false"
          - "--entrypoints.web.address=:80"
          - "--entrypoints.websecure.address=:443"
          - "--certificatesresolvers.{{ block.config.letsencrypt.resolver }}.acme.email={{ block.config.letsencrypt.email }}"
          - "--certificatesresolvers.{{ block.config.letsencrypt.resolver }}.acme.storage=/letsencrypt/acme.json"
          - "--certificatesresolvers.{{ block.config.letsencrypt.resolver }}.acme.tlschallenge=true"
        ports:
          - "80:80"
          - "443:443"
        volumes:
          - "/var/run/docker.sock:/var/run/docker.sock:ro"
          - "traefik-letsencrypt:/letsencrypt"
        networks:
          - traefik

    networks:
      traefik:
        name: traefik
      
    volumes:
      traefik-letsencrypt:
    ```

=== "/polycrate/traefik/docker-compose.yml"

    ```yaml
    version: "3.9"

    services:
      traefik:
        image: "traefik:2.7"
        container_name: "traefik"
        command:
          - "--providers.docker=true"
          - "--providers.docker.exposedbydefault=false"
          - "--entrypoints.web.address=:80"
          - "--entrypoints.websecure.address=:443"
          - "--certificatesresolvers.letsencrypt.acme.email=info@example.com"
          - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
          - "--certificatesresolvers.letsencrypt.acme.tlschallenge=true"
        ports:
          - "80:80"
          - "443:443"
        volumes:
          - "/var/run/docker.sock:/var/run/docker.sock:ro"
          - "traefik-letsencrypt:/letsencrypt"
        networks:
          - traefik

    networks:
      traefik:
        name: traefik
      
    volumes:
      traefik-letsencrypt:
    ```


