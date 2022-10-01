# Create a Docker node on HETZNER Cloud

```yaml
# workspace.poly
name: docker-workspace
dependencies:
  - ayedo/hcloud/inventory
  - ayedo/hcloud/vm
  - ayedo/docker/traefik
config:
  globals:
    hcloud:
      token: YOUR_HCLOUD_TOKEN
blocks:
  - name: inventory
    from: ayedo/hcloud/inventory
  - name: docker-vm
    from: ayedo/hcloud/vm
    config:
      node_group: docker
      nodes:
        - name: docker-1
          location: fsn1
      network:
        enabled: true
    inventory:
      from: inventory
  - name: traefik
    from: ayedo/docker/traefik
    inventory:
      from: inventory
    config:
      tls:  
        resolver:
          email: info@example.com
```

To install the Docker node and Traefik, run the following commands:

- Create the dynamic HETZNER inventory: `polycrate run inventory install`
- Create the Docker node: `polycrate run docker-vm install`
- Create the Traefik deployment: `polycrate run traefik install`