# Blocks

Polycrate blocks are dedicated pieces of code/functionality that can be configured using the `config` stanza in the block configuration file (default: `block.poly`). Blocks expose actions that can be executed using `polycrate run $BLOCK_NAME $ACTION_NAME`.

To add blocks to a workspace, you can use the `dependencies` stanza in the workspace configuration file or run `polycrate pull $BLOCK_NAME:$BLOCK_VERSION`. If `$BLOCK_VERSION` is not defined, `latest` will be assumed.

## Install blocks

Run `polycrate block install BLOCK1 BLOCK2:0.0.1`

## Uninstall blocks

Run `polycrate block uninstall BLOCK1 BLOCK2:0.0.1`

## Workspace dependencies

You can specify a list of blocks in the `dependencies` stanza of the workspace configuration file.

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
