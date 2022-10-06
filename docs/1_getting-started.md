---
title: Getting Started
description: Create your first workspace
---

# Getting Started

## Install Polycrate

See [Installation](6_installation.md).

## Create your first workspace

- Create a workspace directory: `mkdir -p $HOME/.polycrate/workspaces/my-workspace`
- Enter the directory: `cd $HOME/.polycrate/workspaces/my-workspace`
- Create the workspace configuration and a first block:

```yaml
cat <<EOF > workspace.poly
name: my-workspace
extraenv:
  - "YOUR_NAME=Max Mustermann"
blocks:
  - name: hello-world
    config:
      your_name: "Max Mustermann"
    actions:
      - name: greet
        script:
          - echo "HELLO WORLD"
      - name: greet-me
        script:
          - echo "HELLO $YOUR_NAME"
      - name: greet-me2
        script:
          - echo "HELLO $BLOCK_CONFIG_YOUR_NAME"
      - name: show-workspace
        script:
          - echo "This is workspace '$WORKSPACE_NAME'"
EOF
```

!!! note
    You can create a directory anywhere you like and run Polycrate from inside it. However, if you choose to create your workspace directory in `$HOME/.polycrate/workspaces`, which is Polycrate's home directory, you can simply reference a workspace by name when interfacing with it: `polycrate -w my-workspace run block action`.

Now run the following commands:

- `polycrate run hello-world greet`
- `polycrate run hello-world greet-me`
- `polycrate run hello-world greet-me2`
- `polycrate run hello-world show-workspace`

### What happened?

- The first action simply echoes `Hello World` because you defined it like that.
- The second action took an additional environment variable that you defined at workspace level and echoed what you set as a value. With `polycrate --env "ANOTHER_NAME=Santa Claus" --env "YET_ANOTHER_NAME=John Doe"` you can inject further environment variables at runtime.
- The third action echoed the content of another environment variable that Polycrate automatically created for you by converting the block's config path to env vars: `block.config.your_name` -> `BLOCK_CONFIG_YOUR_NAME`. 
- The fourth action shows that the magic from action number 3 works with every stanza in a workspace configuration, even with workspace-level settings like `workspace.name` which converts to `WORKSPACE_NAME`.

## Play with Python

Add the following block to your workspace configuration:

```
blocks:
  - name: python
    actions:
      - name: hello-world
        script:
          - python -c "print('hello world')"
      - name: version
        script:
          - python --version
```

Now run the following commands:

- `polycrate run python hello-world`
- `polycrate run python version`

While it's fun to do things ad-hoc, it makes much more sense to run code from files, right? Let's do this:

- Create a directory called `python` inside the blocks directory of your workspace: `mkdir -p blocks/python`
- Create a block configuration file inside that directory (this is needed for Polycrate to consider this directory a [block directory](2_reference.md#block-configuration)):

```python
cat <<EOF > blocks/python/block.poly
name: python
config:
  greeter: "world"
actions:
  - name: hello-code
    script:
      - python app.py
EOF
```

- Create the following script inside the block directory:

```python
cat <<EOF > blocks/python/app.py
import os
print('Hello, '+os.getenv('BLOCK_CONFIG_GREETER')+'!')
EOF
```

Now run the following commands:

- `polycrate run python hello-code`
- `polycrate block inspect python`

### What happened?

In the first two examples we simply used Polycrate's built-in Python interpreter to fiddle around around with a so called [dynamic block](#2_reference-md#dynamic-blocks), i.e. a block that has been defined on the fly in the workspace configuration file. The last example shows what happens when you create a directory for your block that contains configuration and actual code:

- Polycrate will read in the [block configuration](2_reference.md#block-configuration) from its block directory
- Polycrate will merge the configuration defined in the workspace configuration file with that of the block configuration file (with the workspace configuration **always** overwriting existing block configuration)
- Polycrate will change its workdir to the block directory (`blocks/python`)

This means: once you have a valid block configuration, you can start writing your own code inside the block directory and Polycrate will happily execute it from there without you specifying any path.

## Inheritance

Once a block has been defined (in the workspace or a specific block configuration), it can be [inherited](2_reference.md#inheritance) and thus modified by other blocks, just like classes in typical programming languages work.

- Create a new dynamic block in your workspace configuration:

```yaml
blocks:
  - name: greet-me
    from: python
    config:
      greeter: Max Mustermann
```

Now run the following commands:

- `polycrate run greet-me hello-code`
- `polycrate block inspect greet-me`

### What happened?

As mentioned, the `greet-me` block inherited the base configuration from the `python` block the to the `from: python` stanza. It also inherited its workdir, meaning the code that will be executed is the one from the `python` block we created earlier.

Due to the fact that **workspace configuration overwrites block configuration**, we can simply change the inherited block's `config.greeter` stanza to run the same code but with different outcome. As mentioned, this basically works like classes in regular programming languages.

### But what is it good for?

Well, the Python example might be a bit dull, I agree. Here's a few examples to outline how this can be used:

- create a base block that contains code to deploy S3 buckets, then create 3 dynamic blocks for `dev`, `qa` and `prod` with different configuration to create those buckets for each environment without replicating code
- use Ansible to write a Playbook that deploys a Pod on Kubernetes, then create dynamic blocks with different configuration for namespace and Pod-name to deploy that Pod for different teams or customers (see [how Polycrate integrates with Ansible](2_reference.md#ansible) to see how that works)
- use kubectl to manage resources in different Kubernetes clusters from the same set of actions (see how Polycrate integrates with [Kubernetes](2_reference.md#kubeconfig) to learn how that works)

There are certainly many more use-cases. If you find one, [join our Discord](https://discord.gg/8cQZfXWeXP) and let us know about it so we can add it here.

## Workspace snapshot

The most important thing to know when integrating your own tooling with Polycrate is the concept of the [workspace snapshot](2_reference.md#workspace-snapshot). The workspace snapshot is compiled at runtime and exposed to the Polycrate environment in various ways:

- as environment variables 
- as yaml file (you can get the location of this file from the `POLYCRATE_WORKSPACE_SNAPSHOT_YAML` environment variable)

That snapshot is automatically read-in by Ansible and can be used by your own code to access the workspace snapshot and thus the whole configuration (manual and compiled one) available.

!!! note
    Want to know which environment variables are available? Create an action like this and run it:

    ```yaml
    actions:
      - name: env
        script:
          - env
    ```

You can get a representation of the workspace snapshot by running these commands:

- `polycrate workspace snapshot` (doesn't contain current block/action info)
- `polycrate run block action --snapshot` (contains info on the current block and action; DOES NOT EXECUTE THE ACTION!)

## Play with Ansible

Polycrate has a very special integration with Ansible in that the workspace snapshot is available as Ansible facts automatically. Polycrate provides an [Ansible Vars Plugin](https://docs.ansible.com/ansible/latest/plugins/vars.html) to make that happen. Let's see what that can do for us.

```yaml
# blocks/ansible/playbook.yml
- name: "Debug workspace"
  hosts: localhost
  tasks:
    - name: Show current block
      ansible.builtin.debug:
        var: block
    
    - name: Show current block name
      ansible.builtin.debug:
        var: block.name
    
    - name: Show current action name
      ansible.builtin.debug:
        var: action.name
    
    - name: Show current block config
      ansible.builtin.debug:
        var: block.config

    - name: Show workspace
      ansible.builtin.debug:
        var: workspace

    - name: Get config from block with name 'custom-block'
      ansible.builtin.debug:
        var: (workspace.blocks | selectattr('name', 'match', 'custom-block') | first).config
    
    - name: Show 'greeter' stanza of all blocks that have it
      ansible.builtin.debug:
        var: item
      loop: "{{ workspace | community.general.json_query('blocks[*].config.greeter') }}"
```

If you run `polycrate run block action` you will see 4 top-level keys that will also be available directly from within Ansible:

- `workspace`
- `block`
- `action`
- `env`

### What is it good for?

This allows you to access the configuration of one block from another one by simply querying the workspace for the requested block's config. One way we use this is for example to have a block that provides configuration for LetsEncrypt and then reference this in each block that makes use of an ingress or Traefik labels to get the e-mail address or cert-manager issuer without the need to redefine it in each block. We can simply reference the block that holds the information and take it from there.

Also it's very convenient to access the configuration of the current block (from `polycrate run current-block action`) in a template by simply using `{{ block.config.greeter }}`. The same goes for workspace-level configuration, e.g. `{{ workspace.name }}` which we often use in Kubernetes labels or resource groups at cloud providers to reference the created resources back to the workspace they originate from.

Either way, this allows you to feed a crazy amount of variables to Ansible without using `-e extra-vars-file.yml` or any other mechanism. The whole workspace is directly at your hands when you use Ansible inside a block.


!!! note
    Polycrate provides additional integrations with Ansible, especially for using an inventory or a kubeconfig file. You can learn more about that [here](2_reference.md#ansible).

## Wrapup

You should now have a basic understanding of what Polycrate is capable of and it's now up to you to build stuff with it. You can check more [examples](examples/README.md) or join our Discord for feedback.

[Examples](examples/README.md){ .md-button .md-button--primary }
[:simple-discord: Discord](https://discord.gg/8cQZfXWeXP){ .md-button }