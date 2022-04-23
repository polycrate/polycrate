# Plugin Development

- create plugin dir in `ansible/roles/paas/PLUGIN_NAME`
  - create `tasks` dir
  - create `install.yml`, `main.yml`, `uninstall.yml`
  - create `templates` dir
- add this to `main.yml`:
  ```yaml
  - name: executing plugin_command
    include_tasks: "{{ plugin_command }}.yml"
  ```
- add plugin invocation hook in `ansible/PLUGIN_NAME.yml`
  ```yaml
  - name: "loading plugin PLUGIN_NAME"
    hosts: localhost
    connection: local
    gather_facts: True
    pre_tasks:
      - name: "preflight"
        include_role:
          name: "helpers/prepare"
      - name: "k8s preflight"
        include_role:
          name: "helpers/k8s_preflight"
    tasks:
      - name: "loading plugin PLUGIN_NAME"
        include_role:
          name: "paas/PLUGIN_NAME"
  ```
- add plugin config to `cli/cmd/helpers`- function `loadDefaults()`

## Helm based plugin

- create `templates/values.yml.j2` to store the helm values
- add this to `tasks/install.yml`:
  ```yaml
  - name: installing plugin
    include_role:
      name: helpers/install_helm
    vars:
      release_values: "{{ lookup('template', 'values.yml.j2') | from_yaml }}"
  ```
- add this to `tasks/uninstall.yml`:
  ```yaml
  - name: uninstalling
    community.kubernetes.helm:
      release_name: "{{ plugin }}"
      release_state: absent
      release_namespace: "{{ cloudstack['plugins'][plugin]['namespace'] }}"
      purge: True
      wait: True
      wait_timeout: 10m0s

  - name: deleting namespace
    community.kubernetes.k8s:
      name: "{{ cloudstack['plugins'][plugin]['namespace'] }}"
      api_version: v1
      kind: Namespace
      state: absent
  ```

## Manifest based plugin

- create a file for each manifest inside the `templates` dir
  - suffix for each file must be `.yml.j2`
- add this to `tasks/install.yml`
  ```yaml
  - name: "installing {{ plugin }}"
    block:
      - name: installing PLUGIN_NAME
        community.kubernetes.k8s:
          state: present
          definition: "{{ lookup('template', item+'.yml.j2') | from_yaml }}"
        with_items:
          # Adjust this to the templates you need
          - ClusterRoleBinding
          - ConfigMap
          - ServiceAccount
          - RBAC
          - Deployment
  ```

## Test the installation

Run `cloudstack --dev-dir PATH_WHERE_YOU_CLONED_CLOUDSTACK plugins install PLUGIN_NAME
