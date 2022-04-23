# external_dns

`external_dns` installs [external-dns](https://github.com/kubernetes-sigs/external-dns) on a Kubernetes cluster. external-dns picks up hostnames from ingress and service objects and adjusts the respective DNS zone (here's a list of [supported providers](https://github.com/kubernetes-sigs/external-dns#status-of-providers))

!!! note
    The default provider for external-dns in Cloudstack is `cloudflare` which can be configured through the [cloudflare](cloudflare.md) plugin.

- **needs**: `kubernetes`
