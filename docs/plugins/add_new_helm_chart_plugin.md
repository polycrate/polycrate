# Adding a Helm Chart to cloudstack

*Prerequists:*

- install go
  - mac `brew install go`
- add the following to your .bashrc or .zshrc  or wathever_rc
  - EXPORTS HERE
- vscode
  - add the go extension.
- clone the repository

## Botkube example

In this example we will add a new `botkube` plugin to the `cloudstack` cli. After completion of this example, we can add the new `botkube` plugin to the `Stackfile` and install them.

## which files and folder we need to create/change?

- `ansible/botkube.yaml`
- `cli/cmd/vars.go`
- `ansible/paas/botkube/tasks/*`
  - `ansible/paas/botkube/main.yml`
  - `ansible/paas/botkube/install.yml`
  - `ansible/paas/botkube/uninstall.yml`
- `ansible/paas/botkube/templates/values.yml.j2`

### `ansible/botkube.yaml`

Create a new file under `ansible` called `botkube.yaml` and add the following:

```yaml
- name: "loading plugin botkube"
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
    - name: "loading plugin botkube"
      include_role:
        name: "paas/botkube"
```

### `cli/cmd/vars.go`

to get the `plugins.botkube.version`, `plugins.botkube.chart.*`, values you can check the values after `helm repo add` command.

```bash
helm repo add infracloudio https://infracloudio.github.io/charts
```

Add `index.yaml` to the URL and [open this](https://infracloudio.github.io/charts/index.yaml). Search the *.tgz file with the version you need.

```yaml
plugins:
  //some code
  botkube:
    enabled: false
    commands:
      install:
        script:
          - ansible-playbook botkube.yml
      uninstall:
        script:
          - ansible-playbook botkube.yml
    namespace: botkube
    subdomain: ""
    version: v0.12.4
    backup:
      enabled: false
    ingress:
      enabled: false
      hostname: ""
      tls:
        enabled: false
        provider: ""
        secret:
          cert:
            content: ""
            path: ""
          key:
            content: ""
            path: ""
          name: ""
    chart:
      name: botkube
      version: v0.12.4
      url: https://infracloudio.github.io/charts/botkube-v0.12.4.tgz
      repo:
        url: https://infracloudio.github.io/charts
        name: infracloudio
    artifacts: []
    needs:
      - kubernetes
      - prometheus_crds
    provides: []
    rejects: []
    config:
      communications:
        mattermost:
          botname: botkube
          channel: ""
          enabled: false
          team: ""
          token: ""
          url: ""
      config:
        resources: []
        settings:
          clustername: ""
          kubectl:
            enabled: true
```

### `ansible/paas/botkube/tasks/*`

now, we can copy `ansible/paas/external_dns/` to `ansible/paas/botkube/`.

```bash
cp -r /ansible/paas/external_dns/ /ansible/paas/botkube/
```

After copying we need to modify the following files:

### `ansible/paas/botkube/main.yml`

Nothing to change:

```yaml
- name: executing plugin_command
  include_tasks: "{{ plugin_command }}.yml"
```

### `ansible/paas/botkube/install.yml`

Nothing to change:

```yaml
- name: installing module
  include_role:
    name: helpers/install_helm
  vars:
    release_values: "{{ lookup('template', 'values.yml.j2') | from_yaml }}"
```

### `ansible/paas/botkube/uninstall.yml`

replace `external_dns` and `external-dns` with `botkube`:

```yaml
- name: uninstalling
  community.kubernetes.helm:
    release_name: "botkube"
    release_state: absent
    release_namespace: "{{ cloudstack.plugins.botkube.namespace }}"
    purge: True
    wait: True
    wait_timeout: 10m0s

- name: deleting namespace
  community.kubernetes.k8s:
    name: "{{ cloudstack.plugins.botkube.namespace }}"
    api_version: v1
    kind: Namespace
    state: absent

```

### `ansible/paas/botkube/templates/values.yml.j2`

Now, add your costum HelmChart values. Add the path from `vars.go` to the file in mustache style.
For bool values, you dont need to add the quotation marks.

```yaml
communications:
  mattermost:
    botName: "{{ cloudstack.plugins.botkube.config.communications.mattermost.botname }}"
    enabled: {{ cloudstack.plugins.botkube.config.communications.mattermost.enabled }}
    url: "{{ cloudstack.plugins.botkube.config.communications.mattermost.url }}"
    token: "{{ cloudstack.plugins.botkube.config.communications.mattermost.token }}"
    team: "{{ cloudstack.plugins.botkube.config.communications.mattermost.team }}"
    channel: "{{ cloudstack.plugins.botkube.config.communications.mattermost.channel }}"
config:
  settings:
    clustername: "{{ cloudstack.plugins.botkube.config.config.settings.clustername }}"
    kubectl:
      enabled: {{ cloudstack.plugins.botkube.config.config.settings.kubectl.enabled }}
  resources:
    - name: v1/pods             # Name of the resource. Resource name must be in group/version/resource (G/V/R) format
                                # resource name should be plural (e.g apps/v1/deployments, v1/pods)
      namespaces:               # List of namespaces, "all" will watch all the namespaces
        include:
          - all
        ignore:                 # List of namespaces to be ignored (omitempty), used only with include: all, can contain a wildcard (*)
          -                     # example : include [all], ignore [x,y,secret-ns-*]
      events:                   # List of lifecycle events you want to receive, e.g create, update, delete, error OR all
        - create
        - delete
        - error
    - name: v1/services
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: apps/v1/deployments
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - update
        - delete
        - error
      updateSetting:
        includeDiff: true
        fields:
          - spec.template.spec.containers[*].image
          - status.availableReplicas
    - name: apps/v1/statefulsets
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - update
        - delete
        - error
      updateSetting:
        includeDiff: true
        fields:
          - spec.template.spec.containers[*].image
          - status.readyReplicas
    - name: networking.k8s.io/v1beta1/ingresses
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: v1/nodes
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: v1/namespaces
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: v1/persistentvolumes
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: v1/persistentvolumeclaims
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: v1/configmaps
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: apps/v1/daemonsets
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - update
        - delete
        - error
      updateSetting:
        includeDiff: true
        fields:
          - spec.template.spec.containers[*].image
          - status.numberReady
    - name: batch/v1/jobs
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - update
        - delete
        - error
      updateSetting:
        includeDiff: true
        fields:
          - spec.template.spec.containers[*].image
          - status.conditions[*].type
    - name: rbac.authorization.k8s.io/v1/roles
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: rbac.authorization.k8s.io/v1/rolebindings
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: rbac.authorization.k8s.io/v1/clusterrolebindings
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: rbac.authorization.k8s.io/v1/clusterroles
      namespaces:
        include:
          - all
        ignore:
          -
      events:
        - create
        - delete
        - error
    - name: velero.io/v1/backups
      namespaces:
        include:
          - all
      events:
        - all
      updateSetting:
        includeDiff: true
        fields:
          - status.phase
    - name: velero.io/v1/restores
      namespaces:
        include:
          - all
      events:
        - all
      updateSetting:
        includeDiff: true
        fields:
          - status.phase
```

## Test your plugin

Now we can test our new plugin.

at first, we need to modify our `Stackfile`. Add `botkube` to `stack.plugins`.

```yaml
stack:
  # some values
  plugins:
    # other plugins
    - botkube
```

Now add `botkube` to `plugins`

```yaml
plugins:
  # other plugins
  botkube:
    config:
      communications:
        mattermost:
          enabled: true
          url: https://your.mattermost.url
          token: yourMattermostTokenHere
          team: yourMattermostTeamName
          channel: yourMattermostChannelName
      config:
        settings:
          clustername: "yourClusterName"
```

- `cd cli/`
- `go run main.go --context /path/to/Stackfilelocation/ --pull --image-version 1.15.0-rc-cloudstack-v2.0.0.c6fa0e59 plugins botkube install --dev-dir $PWD/../ansible` if an error occours you can add `--loglevel 2` for more informations.
