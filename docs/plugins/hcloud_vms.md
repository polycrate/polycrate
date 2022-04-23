# hcloud_vms

Provisions virtual machines on [HETZNER Cloud](https://www.hetzner.com/de/cloud). Generates a file called `inventory.yml` in the [context](Glossary.md#context) directory - this is an Ansible-compatible inventory that can be consumed by plugins like [k3s](#k3s).

- **needs**: `-`
- **provides**: `vms`