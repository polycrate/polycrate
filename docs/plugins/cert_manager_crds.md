# cert_manager_crds

Installs cert-manager CRDs into a Kubernetes cluster.

```yaml
plugins:
  cert_manager_crds:
    namespace: kube-system
    version: 1.5.3
    needs:
      - kubernetes
```