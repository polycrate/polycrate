# cert_manager

Installs cert-manager into a Kubernetes cluster.

```yaml
plugins:
  cert_manager:
    namespace: cert-manager
    version: 1.5.3
    needs:
      - kubernetes
      - prometheus_crds
      - cert_manager_crds
    provides:
      - tls
```