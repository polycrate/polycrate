# Create a Kubernetes cluster on HETZNER Cloud

```yaml
# workspace.poly
name: hcloud-workspace
config:
  globals:
    letsencrypt:
      email: info@example.com
    hcloud:
      network:
        cidr: 10.11.3.0/24
        enabled: true
      token: YOUR_HCLOUD_TOKEN
dependencies:
  - ayedo/hcloud/inventory
  - ayedo/hcloud/k8s-infra
  - ayedo/hcloud/k8s
blocks:
  - name: inventory
    from: ayedo/hcloud/inventory
  - name: controlplane
    config:
      node_group: controlplane
      nodes:
      - location: fsn1
        name: controlplane-1
        type: cx21
      - location: nbg1
        name: controlplane-2
        type: cx21
      - location: hel1
        name: controlplane-3
        type: cx21
    from: ayedo/hcloud/k8s
    inventory:
      from: inventory
  - name: worker
    config:
      node_group: worker
      nodes:
      - location: fsn1
        name: worker-1
        type: cx41
      - location: fsn1
        name: worker-2
        type: cx41
      - location: nbg1
        name: worker-3
        type: cx41
    from: ayedo/hcloud/k8s
    inventory:
      from: inventory
    kubeconfig:
      from: controlplane
  - name: infrastructure
    from: ayedo/hcloud/k8s-infra
```

To install the Kubernetes cluster, run the following commands:

- Create the dynamic HETZNER inventory: `polycrate run inventory install`
- Create the private network, ssh key and placement group at HETZNER: `polycrate run infrastructure install`
- Create the controlplane nodes: `polycrate run controlplane install`
- Create the worker nodes: `polycrate run worker install`

After successful creation of the cluster, the Kubeconfig will be saved to `artifacts/blocks/controlplane/kubeconfig.yml`.