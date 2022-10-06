# Create a Kubernetes cluster on Azure Cloud

```yaml
# workspace.poly
name: aks-workspace
dependencies:
  - ayedo/azure/aks
blocks:
  - name: aks
    from: ayedo/azure/aks
    config:
      tenant_id: "YOUR_TENANT_ID"
      subscription_id: "YOUR_SUBSCRIPTION_ID"
      serviceprincipal:
        id: "YOUR_SP_CLIENT_ID"
        secret: "YOUR_SP_CLIET_SECRET"
      resource_group: "rg-name-of-your-choice"
      tags:
        label: "value"
      pools:
        default:
          count: 1
          size: standard_e2bds_v5
          mode: System
        user:
          count: 3
          size: standard_e2bds_v5
          mode: User
```

To install the Kubernetes cluster, run the following command: `polycrate run aks install`. This will setup an AKS cluster with 1 node in the default pool and 3 nodes in the user pool. The resulting Kubeconfig will be saved to `artifacts/blocks/aks/kubeconfig.yml`.