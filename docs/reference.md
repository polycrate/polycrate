# Reference

## Container

All the Cloudstack magic happens inside of a Docker container running on the system that invokes the `cloudstack` command.

The container will be started whenever you call a plugin command. It is based on public image provided by [ayedo](https://www.ayedo.de) (see the [Dockerfile]([Cloudstack image](https://gitlab.com/ayedocloudsolutions/cloudstack/cloudstack/ansible/Dockerfile)) here) and contains most of the best-practice tooling of cloud-native development and operations.

## Stack

Whatever you build with Cloudstack is called a **stack**. It's a named instantiation of the configuration in your [context](#context) directory.

## Context

The **context** directory contains the configuration and lifecycle artifacts of a [stack](#stack). It's a directory on your filesystem that can be synced and collaborated on via **git**.

## Inventory

The **inventory** is a regular, yaml-formated [Ansible inventory](https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html) inside the **context** directory. The **inventory** must be named `inventory.yml`.

The **inventory** can be created automatically through plugins (such as [hcloud_vms](plugins/hcloud_vms.md)) or provided manually.

## Kubeconfig

Cloudstack is tightly integrated with Kubernetes and can connect to a cluster using a [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file. By default, the CLI looks for a kubeconfig file in `$HOME/.kube/config` and in the `KUBECONFIG` environment variable. 

When working with a [stack](#stack) (e.g. by invoking `cloudstack` from within a [context](#context) directory or using the `--context` flag), the CLI looks for a `kubeconfig.yml` file within the [context](#context) directory. If it finds one, this has precedence over a kubeconfig file discovered from the default path or the environment.

Additionally, you can specify a kubeconfig file using the `--kubeconfig PATH/TO/YOUR/kubeconfig` flag.

## Kubernetes

Kubernetes is the container orchestrator of choice in Cloudstack. Most of the [plugins](#plugins) (like [prometheus](plugins/prometheus.md)) are directly depending on Kubernetes  - they have a `need` for `kubernetes`. Some plugins (like [k3s](plugins/k3s.md)) can be used to install Kubernetes - they `provide` `kubernetes`.

!!! note
    Plugins `need` and `provide` certain functionalities. This way a dependency system can be established even for [custom plugins](plugins/custom_plugins.md).

## Plugins

Cloudstack is a modular system built of plugins that provide the actual functionality. A [stack](#stack) is composed of one or more plugins and their configuration. Plugins may have dependencies (implemented by `needs` and `provides`) that imply constraints to the order the plugins should be installed in.

The installation order for plugins is defined in `stack.plugins`:

```yaml
stack:
  name: my-cloudstack
  hostname: my.cloudstack.one
  exposed: true
  tls: false
  plugins:
    - prometheus_crds
    - eventrouter
    - prometheus
    - loki
    - promtail
    - nginx_ingress
    - portainer
```

This is mainly relevant when using `cloudstack launch` as it iterates through the list of plugins and invokes the `install` command.

!!! note
    The order in which plugins are installed is important. For example, the [prometheus](plugins/prometheus.md) and [nginx_ingress](plugins/nginx_ingress.md) plugin needs the Prometheus [CRDs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/) to be available in the cluster, so they can install a `ServiceMonitor` that enabled automated monitoring. Thus, the plugin [prometheus_crds](plugins/prometheus_crds.md) must be installed before the other plugins.

### Configuration

A plugin can be configured in the `plugins` configuration section of a [Stackfile](#stackfile) or in its `Pluginfile`.

```yaml
...

plugins:
  hcloud_vms:
    config:
      token: YOUR_TOKEN_HERE
      master:
        count: 3
        type: cx31
      node:
        count: 3
        type: cpx31

...
```

Each plugin has a shared set of configuration:

- `version`: the version of the plugin to be installed. For most plugins, this equals to the application version of the plugin
- `subdomain`: the subdomain under `stack.hostname` that the plugin will be exposed to (if `exposed` is true)
- `namespace` (only applies to plugins that will be installed into Kubernetes): the namespace the plugin will be installed to
- `chart`: this section holds configuration for Helm-based plugins. For [Core plugins](#core-plugins) this should be left untouched. It can be useful to customize this for [Custom plugins](#custom-plugins)
- `config`: plugin-specific configuration that is not common to all plugins belongs here
- `commands`: an array of commands that the plugin implements. Defaults to `install` and `uninstall`
- `needs`: a list of arbitrary strings that can be matched with a `provides` list
- `provides`: a list of arbitrary strings that can be matched with a `needs` list
- `rejects`: a list of plugins that can't be used together with this plugin

!!! note
    To see the default configuration, run `cloudstack show defaults`. The default output format is yaml. If you prefer JSON, add `--output-format json` to the command. To see the rendered stack configuration, use `cloudstack show config`.

### Needs and provides

A plugin `needs` and `provides` certain functionality. By matching the needs and provides of the active plugins in a stack, a loose dependency system can be established. For example, the plugin [k3s](plugins/k3s.md) provides functionality `kubernetes` which plugin [cert_manager](plugins/cert_manager.md) needs. Plugin [k3s](plugins/k3s.md) also needs functionality `vms` which [hcloud_vms](plugins/hcloud_vms.md) provides.

### Core plugins

Core plugins are maintained by [Ayedo Cloud Solutions GmbH](https://www.ayedo.de) and part of the [Cloudstack code base](https://gitlab.com/ayedocloudsolutions/cloudstack/cloudstack).

### Custom plugins

Custom plugins are living in a `plugins` directory inside the [context](#context) directory. Each plugin must have a dedicated directory having the plugin's name. Inside the plugin directory, any files are allowed. Cloudstack is looking for a file called [Pluginfile](#pluginfile) that must contain the plugin's configuration (see [plugin configuration](#configuration)). Usually the `Pluginfile` is used for a plugin's default configuration (like the [commands](#commands) it supports) while instance-specific configuration belongs into the [Stackfile](#stackfile).

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

### Commands

A plugin can expose an arbitrary amount of commands. Commands can be used to implement the actual functionality of a plugin. Examples would be `install` or `uninstall`, but also `status` or `debug`.

The `script` section of a plugin command is a list of commands that will be concatenated to a Bash script and executed inside the Cloudstck container when you call `cloudstack plugins my_custom_plugin install` (or any of the commands you specified).

While `install` and `uninstall` should be defined by convention, you're free to implement anything here. Command names are limited to certain characters though: `a-zA-Z_`.

The script will be executed within a container created from the [Cloudstack image](https://gitlab.com/ayedocloudsolutions/cloudstack/cloudstack/ansible/Dockerfile) and has access to all the tools available in the container (namely kubectl, helm, ansible, cloudstack, etc). The [context](#context) directory is available at `/context`, the active [kubeconfig](#kubeconfig) file is mounted to `/root/.kube/config` and set via `KUBECONFIG` environment variable so all the tooling should work out of the box.

The Cloudstack image is based on debian buster so you're free to install any additional software as part of a plugin command.

!!! note
    Cloudstack does not persist data between invocations apart from changes made to the [context](#context) directory (mounted at `/context` inside the execution container) or the [kubeconfig](#kubeconfig) file (mounted at `/root/.kube/config` inside the execution container).

!!! note
    It's fine to write data to the [context](#context) directory. It's advised to not touch the [kubeconfig](#kubeconfig) file unless you're absolutely sure what you do as it might affect your whole stack.

## Dependencies

Plugins can be declared as external dependencies for a [stack](#stack). They can be downloaded from an external source (currently only `git` repositories are supported) and saved to the [plugins](#plugins) directory. Dependencies of a stack must be defined in `stack.dependencies`:

```yaml
stack:
  name: my-cloudstack
  hostname: my-cloudstack.one
  dependencies:
    - name: hcloud_vms
      provider: git
      git:
        repository: https://gitlab.com/ayedocloudsolutions/cloudstack/plugins/hcloud_vms
```

Running `cloudstack deps download` will download all configured dependencies.

!!! note
    A plugin will only be downloaded if it doesn't exist already. If you want to update or override a plugin, just use `--force` when downloading dependencies like so: `cloudstack deps download --force`

!!! note
    Currently the Cloudstack [container](#container) ships with the [Core Plugins](#core-plugins). This will change with the release of Cloudstack 2.0.0. Form then on, all plugins (core or custom) must be defined as a dependency and downloaded if they're not part of the stack's local plugins. The Core Plugins will no longer be shipped with the Cloudstack container to improve customizability of plugins per [stack](#stack).

## Stackfile

The **Stackfile** holds the configuration for a [stack](#stack). Its default location is the [context](#context) directory, but it can be anywhere on your system.

!!! note
    You can specify a custom **Stackfile** with the CLI by using `--stackfile YOUR/CUSTOM/Stackfile`. This can be especially helpful when using Cloudstack in CI.

## Pluginfile

The **Pluginfile** contains configuration for [Custom plugins](#custom-plugins). Its default location is the plugin's directory inside the `plugins` folder of the [context](#context) directory.