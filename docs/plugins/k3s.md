# k3s

`k3s` installs [k3s](https://k3s.io) on VMs/Bare-Metal hosts. When a plugin like `hcloud_vms` is used, a file called `inventory.yml` will be available for `k3s` to determine the connection data for the machines to provision Kubernetes to.

- **needs**: `vms`
- **provides**: `kubernetes`

## Directories used

- /usr/local/bin/ (k3s binary)
- /etc/systemd/system (k3s systemd units)
- /var/lib/kubelet (kubelet data)
- /var/log/ (logs)
- /var/lib/rancher/k3s (k3s data, containers, images)