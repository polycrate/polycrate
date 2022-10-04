# Why Polycrate

## Human-friendly interface

Create simple [actions](2_reference.md#actions) to execute complex, pre-configured deployment or installation logic:

```yaml
# block.poly
name: nginx
actions:
  - name: install
    script:
      - kubectl run nginx --image=nginx --port=80
  - name: uninstall
    script:
      - kubectl delete deployment nginx
  - name: status
    script:
      - kubectl get deployment nginx
```

Use these actions by running the following commands: 

- `polycrate run nginx install`
- `polycrate run nginx status`
- `polycrate run nginx uninstall`

## Included cloud-native toolchain

There's no need to locally install abstract toolchains and conflicting dependencies when working with tools like Ansible, Kubernetes or Terraform.

Learn more about the included tooling [here](5_included-tools.md).

## Abstract away complexity

No knowledge of the logic of single [blocks](2_reference.md#blocks) is necessary to use the exposed actions. A polycrate [workspace](#2_reference.md#workspace) can be designed as a very user-friendly interface to very complex logic - it's all in the hands of the workspace creator to make it easy for users to execute pre-defined tasks in a safe way.

## Build well-integrated systems based on a single configuration file

Gone are the times of scattered GitOps repositories containing only parts of the logic to deploy a systems. A Polycrate workspace can contain blocks that handle the complete lifecycle of a system, even multiple systems (like dev, qa and prod) at once.

## No custom DSL or complex configuration structure to learn

Polycrate lets you build on your own terms with minimum constraints, giving you the ability to configure your workspace and blocks the way YOU need to. All you need to know is **how to YAML**. How you implement a specific block is totally up to you.

## Share the operational load of managing complex systems with your team

By building blocks with meaningful actions, you can have even non-technical people run errands on your systems without them knowing how Terraform, Ansible or Kubernetes works.

```yaml
# block.poly
name: myapp
kubeconfig:
  from: aks
actions:
  - name: dns-logs
    script:
      - kubectl --namespace kube-system logs -l k8s-app=kube-dns
  - name: ingress-logs
    script:
      - kubectl --namespace ingress-nginx logs -l app.kubernetes.io/name=ingress-nginx
  - name: restart-myapp
    script:
      - kubectl rollout restart deployment myapp
```

## Works with and improves existing tools

No need to rewrite existing code to work with Polycrate. Just make a block out of it and expose it as an action

```yaml
# block.poly
name: hello-world
actions:
  - name: hello-world
    script:
      - python hello-world.py
```

You can use this to write infrastructure as code with any of the [supported tools and languages](5_included-tools.md).

## Use tools like Ansible to achieve idempotent deployments inside your blocks

```yaml
# block.poly
name: docker
inventory:
  from: servers
actions:
  - name: install
    script:
      - ansible-playbook install.yml
```

You can either place an inventory file in the `docker` block's artifact directory directly or create a virtual block that only holds the inventory and then reference it in this block.

Polycrate makes sure that Ansible will find the configured inventory so you don't have to use `-i inventory-file` in the call to `ansible-playbook`.

## Share common blocks through an OCI-comptabible registry

Polycrate uses OCI-compatible registries as a transport to push and pull blocks and share them between workspaces. You can define them as workspace [dependencies](2_reference.md#dependencies) and Polycrate will try to pull them on every invokation if they're not already in the workspace (just like Docker does with its images).

- Pull a block from a registry into the workspace: `polycrate block pull ayedo/hcloud/k8s:0.0.1`
- Push a block in the workspace to a registry: `polycrate --registry-url my-registry.io block push my/custom/block`

## Share workspaces through git-repositories

A workspace is like a blackbox, containing everything necessary to "function". The easiest way to collaborate on them is to share them through git-repositories. Simply do `git add . && git commit -am "feat: created the workspace" && git push` and you'll be good.

## Always be in sync

You can configure a workspace to automatically sync with a git-repository on every invocation by adding the following config to `workspace.poly`:

```yaml
name: your-workspace
sync:
  enabled: true
  auto: true
  remote:
    url: ssh://git@github.com/your/namespace/your-workspace.git
```

!!! note
    Polycrate will shell out to your local git installation. Make sure the repository exists and you have appropriate access to it.
