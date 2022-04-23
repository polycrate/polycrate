# Creating a stack

After you've installed the CLI, you can bootstrap a new [stack](reference.md#stack). Let's call it `my-cloudstack` and create the corresponding [context](reference.md#context) directory.

!!! note
    The **context** directory is where the configuration (the so called **Stackfile**) and lifecycle artifacts (e.g. TLS certifcates and SSH keys) of a stack are saved. You can set a custom context directory with `--context`. By default, the context directory is set to your current working directory.

Running `cloudstack init` will get you started with an example config that deploys a single-node Kubernetes cluster on [HETZNER Cloud](https://console.hetzner.cloud/).

```bash
mkdir my-cloudstack
cd my-cloudstack
cloudstack init
```

Edit the `Stackfile`: add your HETZNER Cloud API key to `plugins.hcloud_vms.config.token` and configure a hostname for your cluster in `stack.hostname`. Now you're good to go.

!!! note
     If you're not using the [external_dns](plugins/external_dns.md) plugin, make sure you grab you server's IP address from the [inventory](reference.md#inventory) and update the DNS zone of your hostname accordingly so traffic can reach the cluster.

!!! note
    Instead of creating and entering the context directory manually, you can initialize a new stack from anywhere and have the directory created automatically by using the following command: `cloudstack --context $PWD/my-cloudstack init`

You can launch your new stack running the following command:

```bash
cloudstack launch --pull
```

!!! note
    `--pull` downloads the latest Cloudstack image from GitLab's [container registry](https://gitlab.com/ayedocloudsolutions/cloudstack/cloudstack/container_registry/2239893). You can use `--image-ref` to specify your own image (e.g. if you forked Cloudstack). By default, the CLI will use the `latest` tag of the image. Use `--image-version` to specify a different tag.

!!! note
    In case you're not using the CLI from within the context directory, you can specify the context directory manually: `cloudstack --context $PWD/my-cloudstack install --pull`.

!!! note
    The CLI tries to automatically discover the Kubernetes cluster to work with by detecting a [kubeconfig](reference.md#kubeconfig) file. By default, it looks for a `kubeconfig.yml` in the context directory. This file is automatically created for you if you use a [plugin](reference.md#plugins) that provides [kubernetes](reference.md#kubernetes). If it doesn't exist, the CLI tries to fall back to `$HOME/.kube/config` and the `KUBECONFIG` environment variable in this order. If no kubeconfig can be found, the parts of the stack that interface with your Kubernetes cluster won't work.

    You can manually specify a kubeconfig file using `--kubeconfig` which has precedence over the automatic discovery process.

After installation, you can access your stack under the following URLs (replace `my.cloudstack.one` with your configured hostname):

- [grafana.my.cloudstack.one](https://grafana.my.cloudstack.one)

Additionally, a [kubeconfig](reference.md#kubeconfig) is now located inside your [context](reference.md#context) directory which you can load with [Lens](https://k8slens.dev/) or any other tool to connect to your cluster.
