# Deploy to an existing EKS cluster

Polycrate includes the [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator). By adding your AWS credentials to the workspace configuration and supplying an `aws-iam-authenticator`-scoped `kubeconfig.yml` in a block, you can interface with EKS clusters in a safe manner.

```yaml
# workspace.poly
name: eks-workspace
extraenv:
  - AWS_ACCESS_KEY_ID=YOU_ACCESS_KEY
  - AWS_SECRET_ACCESS_KEY=YOUR_SECRET_ACCESS_KEY
blocks:
  - name: eks # put the kubeconfig to artifacts/blocks/eks/kubeconfig.yml
  - name: nginx
    kubeconfig:
      from: eks # use the eks block's kubeconfig
    actions:
      - name: deploy
        script:
          - kubectl run nginx --image=nginx --replicas=1 --port=80

```