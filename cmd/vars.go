package cmd

import (
	_ "embed"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var validate *validator.Validate

var defaultsYml []byte

// Globals
var local bool = false
var imageRef string
var imageVersion string
var logLevel string
var pull bool
var force bool
var interactive bool
var envPrefix string = "polycrate"

const defaultFailedCode = 1

//var devDir string
var cwd, _ = os.Getwd()
var callUUID = getCallUUID()
var timeFormat string = "2006-01-02T15:04:05-0700"
var utcNow time.Time = time.Now().UTC()
var now string = time.Now().Format(timeFormat)
var home, _ = homedir.Dir()
var overrides []string
var workdir string = workspaceDir
var workdirContainer string = "/workdir"
var outputFormat string
var polycrateVersion string
var sshPrivateKey string
var sshPublicKey string
var remoteRoot string

// Build meta
var version string = "latest"
var commit string
var date string

var envVars []string
var mounts []string
var ports []string

// Git
var gitConfigObject = viper.New()

// Workspace
var workspace Workspace // Struct
var workspaceInitConfig = viper.New()
var workspaceConfig = viper.New()               // Viper
var workspaceDir string                         // local
var workspaceContainerDir string = "/workspace" // Container
var workspaceConfigFile string = "workspace.yml"

var workspaceConfigFilePath string
var workspaceContainerConfigFilePath string

// Blocks
var blocksRoot string = "blocks"
var blockConfigFile string = "block.yml"
var blockName string
var blockPaths []string
var blocks []Block

//var block Block
var blockDir string
var blocksDir string
var blocksContainerDir string
var blockContainerDir string
var blockContainerConfigFilePath string

// Actions
var action Action
var actionName string

// Config Objects
var defaultConfigObject = viper.New()

// Inventory
var inventoryLocalPath string = "/tmp/inventory.yml"
var inventoryContainerPath string = "/tmp/inventory.yml"
var inventory string
var inventoryConfigObject = viper.New()

// Kubernetes Kubeconfig
var kubeconfig string
var kubeconfigPath string = kubeconfig
var kubeconfigContainerPath string = "/root/.kube/config"

// State
var stateRoot string = "state"
var state Statefile
var stateConfigObject = viper.New()
var statefile string

// Workflows
var workflowsRoot string = "workflows"

//var stackfile string
//var runtimeStackfile string

//var k8sDistro string
//var k8sOnline bool = false

//var command string
//var plugins string
var pipeline string

//var version = BuildVersion(cwd)
var currentHistoryItem StateHistoryItem

//var cloudstack Stackfile

// var cloudstackDefaults = &Stackfile{
// 	Namespace:        "cloudstack",
// 	Legacy_namespace: "acs",
// 	Image: ImageConfig{
// 		Reference: "registry.gitlab.com/ayedocloudsolutions/cloudstack/cloudstack",
// 		Version:   version,
// 	},
// 	Stack: StackConfig{
// 		Hostname: "my.cloudstack.one",
// 		Exposed:  true,
// 		Tls:      false,
// 		Sso:      false,
// 		Plugins: []string{
// 			"prometheus_crds",
// 			"prometheus",
// 			"loki",
// 			"promtail",
// 			"nginx_ingress",
// 			"portainer",
// 		},
// 	},
// 	Plugins: map[string]PluginConfig{
// 		"bastion": {
// 			Enabled: false,
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook install-bastion.yml"},
// 				},
// 			},
// 		},
// 		"hcloud": {
// 			Enabled: false,
// 			Config: map[string]interface{}{
// 				"token": "",
// 				"master": map[string]interface{}{
// 					"count":    1,
// 					"type":     "cx31",
// 					"location": "fsn1",
// 				},
// 				"node": map[string]interface{}{
// 					"count":    3,
// 					"type":     "cx31",
// 					"location": "fsn1",
// 				},
// 				"os_image": "ubuntu-20.04",
// 			},
// 		},
// 		"hcloud_csi": {
// 			Enabled: false,
// 			Version: version, // TODO
// 			Config: map[string]interface{}{
// 				"token": "",
// 			},
// 			Needs:    []string{"kubernetes"},
// 			Provides: []string{"csi"},
// 			Rejects:  []string{"longhorn", "azure_aks"},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook hcloud_csi.yml"},
// 				},
// 			},
// 		},
// 		"hcloud_vms": {
// 			Enabled: false,
// 			Version: version, // TODO
// 			Config: map[string]interface{}{
// 				"token": "",
// 				"master": map[string]interface{}{
// 					"count":    1,
// 					"type":     "cx31",
// 					"location": "fsn1",
// 				},
// 				"node": map[string]interface{}{
// 					"count":    3,
// 					"type":     "cx31",
// 					"location": "fsn1",
// 				},
// 				"inventory": map[string]string{
// 					"path": "inventory.yml",
// 				},
// 				"os_image": "ubuntu-20.04",
// 			},
// 			Needs:    []string{},
// 			Provides: []string{"inventory", "teardown"},
// 			Rejects:  []string{"azure_aks", "proxmox_vms"},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook hcloud_vms.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook hcloud_vms.yml"},
// 				},
// 			},
// 		},
// 		"proxmox_vms": {
// 			Enabled: false,
// 			Version: version, // TODO
// 			Config: map[string]interface{}{
// 				"target_node": "",
// 				"clone":       "",
// 				"ssh_user":    "cloudstack",
// 				"user":        "cloudstack",
// 				"password":    "",
// 				"api": map[string]interface{}{
// 					"url":      "",
// 					"user":     "root@pam",
// 					"password": "",
// 				},
// 				"master": map[string]interface{}{
// 					"count":          1,
// 					"cores":          4,
// 					"sockets":        1,
// 					"memory":         4096,
// 					"disk":           "150G",
// 					"storage":        "local-zfs",
// 					"network_model":  "virtio",
// 					"network_bridge": "vmbr0",
// 					"ipconfig":       []string{},
// 				},
// 				"node": map[string]interface{}{
// 					"count":          1,
// 					"cores":          4,
// 					"sockets":        1,
// 					"memory":         4096,
// 					"disk":           "150G",
// 					"storage":        "local-zfs",
// 					"network_model":  "virtio",
// 					"network_bridge": "vmbr0",
// 					"ipconfig": []string{
// 						"dhcp",
// 					},
// 				},
// 			},
// 			Needs:    []string{},
// 			Provides: []string{"inventory", "teardown"},
// 			Rejects:  []string{"azure_aks", "hcloud_vms"},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook proxmox_vms.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook proxmox_vms.yml"},
// 				},
// 			},
// 		},
// 		"inventory": {
// 			Enabled: false,
// 			Version: version, // TODO
// 			Config: map[string]interface{}{
// 				"inventory": map[string]string{
// 					"path": "inventory.yml",
// 				},
// 			},
// 			Needs:    []string{},
// 			Provides: []string{},
// 			Rejects:  []string{"hcloud_vms"},
// 		},
// 		"azure": {
// 			Enabled: false,
// 			Version: version, // TODO
// 			Config: map[string]interface{}{
// 				"serviceprincipal": map[string]interface{}{
// 					"client": map[string]interface{}{
// 						"id":     "",
// 						"secret": "",
// 					},
// 				},
// 			},
// 		},
// 		"azure_aks": {
// 			Enabled: false,
// 			Version: version, // TODO
// 			Config: map[string]interface{}{
// 				"serviceprincipal": map[string]interface{}{
// 					"client": map[string]interface{}{
// 						"id":     "",
// 						"secret": "",
// 					},
// 				},
// 				"node": map[string]interface{}{
// 					"count":    3,
// 					"type":     "Standard_D2_v2",
// 					"location": "Germany West Central",
// 				},
// 				"resource_group": "",
// 				"os_image":       "ubuntu-20.04",
// 				"tags":           map[string]string{},
// 			},
// 			Needs:    []string{},
// 			Provides: []string{"kubernetes", "teardown"},
// 			Rejects:  []string{"hcloud_csi", "longhorn", "proxmox_vms"},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook azure_aks.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook azure_aks.yml"},
// 				},
// 			},
// 		},
// 		"ssh": {
// 			Enabled: false,
// 			Version: version, // TODO
// 			Config: map[string]interface{}{
// 				"keys": map[string]interface{}{
// 					"public":  "id_rsa.pub",
// 					"private": "id_rsa",
// 				},
// 				"fallback_port": 3519,
// 				"port":          22,
// 			},
// 			Needs:    []string{"inventory"},
// 			Provides: []string{},
// 			Rejects:  []string{},
// 		},
// 		"sshd": {
// 			Enabled:  false,
// 			Version:  version, // TODO
// 			Needs:    []string{},
// 			Provides: []string{},
// 			Rejects:  []string{},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook sshd.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook sshd.yml"},
// 				},
// 			},
// 		},
// 		"k3s": {
// 			Enabled:  false,
// 			Version:  "v1.21.1+k3s1", // TODO
// 			Needs:    []string{"inventory"},
// 			Provides: []string{"kubernetes"},
// 			Rejects:  []string{"azure_aks"},
// 			Config: map[string]interface{}{
// 				"server": map[string]interface{}{
// 					"args": "--kube-scheduler-arg 'bind-address=0.0.0.0' --kube-scheduler-arg 'address=0.0.0.0' --kube-proxy-arg 'metrics-bind-address=0.0.0.0' --kube-controller-manager-arg 'bind-address=0.0.0.0' --kube-controller-manager-arg 'address=0.0.0.0' --kube-controller-manager-arg 'allocate-node-cidrs' --etcd-expose-metrics --disable traefik,local-storage --disable-cloud-controller",
// 				},
// 				"agent": map[string]interface{}{
// 					"args": "",
// 				},
// 				"api": map[string]interface{}{
// 					"endpoint": "",
// 					"token":    "",
// 				},
// 				"systemd_dir": "/etc/systemd/system",
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook k3s.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook k3s.yml"},
// 				},
// 			},
// 			Artifacts: []ArtifactConfig{
// 				{
// 					Name:     "k3s-amd64",
// 					Filename: "k3s-amd64",
// 					Url:      "https://github.com/k3s-io/k3s/releases/download/v1.21.1+k3s1/k3s",
// 				},
// 				{
// 					Name:     "k3s-arm64",
// 					Filename: "k3s-arm64",
// 					Url:      "https://github.com/k3s-io/k3s/releases/download/v1.21.1+k3s1/k3s-arm64",
// 				},
// 				{
// 					Name:     "k3s-armhf",
// 					Filename: "k3s-armhf",
// 					Url:      "https://github.com/k3s-io/k3s/releases/download/v1.21.1+k3s1/k3s-arm64",
// 				},
// 			},
// 		},
// 		// 	wget -O "k3s/amd64/k3s-${K3S_VERSION}" -q "https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s" && \
// 		// wget -O "k3s/arm64/k3s-${K3S_VERSION}" -q "https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-arm64" && \
// 		// wget -O "k3s/armhf/k3s-${K3S_VERSION}" -q "https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-arm64"

// 		"longhorn": {
// 			Enabled:   false,
// 			Version:   "1.2.2", // TODO
// 			Namespace: "longhorn-system",
// 			Needs:     []string{"kubernetes"},
// 			Provides:  []string{"csi"},
// 			Rejects:   []string{"hcloud_csi"},
// 			Chart: ChartConfig{
// 				Name:    "longhorn",
// 				Version: "1.2.2",
// 				Url:     "https://github.com/longhorn/charts/releases/download/longhorn-1.2.2/longhorn-1.2.2.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://charts.longhorn.io",
// 					Name: "longhorn",
// 				},
// 			},
// 			Config: map[string]interface{}{
// 				"server": map[string]interface{}{
// 					"args": "--kube-scheduler-arg 'bind-address=0.0.0.0' --kube-scheduler-arg 'address=0.0.0.0' --kube-proxy-arg 'metrics-bind-address=0.0.0.0' --kube-controller-manager-arg 'bind-address=0.0.0.0' --kube-controller-manager-arg 'address=0.0.0.0' --kube-controller-manager-arg 'allocate-node-cidrs' --etcd-expose-metrics --disable traefik,local-storage --disable-cloud-controller",
// 				},
// 				"agent": map[string]interface{}{
// 					"args": "",
// 				},
// 				"api": map[string]interface{}{
// 					"endpoint": "",
// 					"token":    "",
// 				},
// 				"datapath":    "/var/lib/longhorn/",
// 				"systemd_dir": "/etc/systemd/system",
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook longhorn.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook longhorn.yml"},
// 				},
// 			},
// 		},
// 		"letsencrypt": {
// 			Enabled:   false,
// 			Version:   version, // TODO
// 			Namespace: "cloudstack",
// 			Needs:     []string{"kubernetes", "cert_manager", "cert_manager_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Config: map[string]interface{}{
// 				"environments": []map[string]string{
// 					{
// 						"name":   "staging",
// 						"server": "https://acme-staging-v02.api.letsencrypt.org/directory",
// 					},
// 					{
// 						"name":   "prod",
// 						"server": "https://acme-v02.api.letsencrypt.org/directory",
// 					},
// 				},
// 				"issuers": map[string]interface{}{
// 					"dns01": map[string]interface{}{
// 						"enabled":  false,
// 						"provider": "letsencrypt",
// 						"solver":   "cloudflare",
// 						"zone":     "",
// 					},
// 					"http01": map[string]interface{}{
// 						"enabled":  true,
// 						"provider": "letsencrypt",
// 					},
// 				},
// 				"environment": "prod",
// 				"mail":        "",
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook letsencrypt.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook letsencrypt.yml"},
// 				},
// 			},
// 		},
// 		"argocd": {
// 			Enabled:   false,
// 			Version:   "2.0.4", // TODO
// 			Namespace: "argocd",
// 			Subdomain: "argocd",
// 			Needs:     []string{"kubernetes"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Backup: BackupConfig{
// 				Enabled: false,
// 			},
// 			Ingress: IngressConfig{
// 				Enabled:  false,
// 				Hostname: "",
// 				Tls: TlsConfig{
// 					Enabled:  false,
// 					Provider: "letsencrypt",
// 				},
// 			},
// 			Chart: ChartConfig{
// 				Name:    "argo-cd",
// 				Version: "2.0.4",
// 				Url:     "https://charts.bitnami.com/bitnami/argo-cd-2.0.4.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://charts.bitnami.com/bitnami",
// 					Name: "bitnami",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook argocd.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook argocd.yml"},
// 				},
// 			},
// 		},
// 		"cert_manager_crds": {
// 			Enabled:   false,
// 			Version:   "1.6.1", // TODO
// 			Namespace: "kube-system",
// 			Needs:     []string{"kubernetes"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook cert_manager_crds.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook cert_manager_crds.yml"},
// 				},
// 			},
// 		},
// 		"cert_manager": {
// 			Enabled:   false,
// 			Version:   "1.6.1", // TODO
// 			Namespace: "cert-manager",
// 			Needs:     []string{"kubernetes", "prometheus_crds", "cert_manager_crds"},
// 			Provides:  []string{"tls"},
// 			Rejects:   []string{},
// 			Config: map[string]interface{}{
// 				"ca_issuer": map[string]interface{}{
// 					"enabled":          false,
// 					"certificate_path": "ca.crt",
// 					"key_path":         "ca.key",
// 				},
// 			},
// 			Chart: ChartConfig{
// 				Name:    "cert-manager",
// 				Version: "1.6.1",
// 				Url:     "https://charts.jetstack.io/charts/cert-manager-v1.6.1.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://charts.jetstack.io",
// 					Name: "jetstack",
// 				},
// 			},
// 			Artifacts: []ArtifactConfig{
// 				{
// 					Name:     "cert-manager-crds",
// 					Filename: "cert-manager-crds.yml",
// 					Url:      "https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.crds.yaml",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook cert_manager.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook cert_manager.yml"},
// 				},
// 			},
// 		},
// 		"route53": {
// 			Enabled:   false,
// 			Version:   version, // TODO
// 			Namespace: "cloudstack",
// 			Needs:     []string{},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Config: map[string]interface{}{
// 				"access_key_id":     "",
// 				"secret_access_key": "",
// 			},
// 		},
// 		"cloudflare": {
// 			Enabled:   false,
// 			Version:   version, // TODO
// 			Namespace: "cloudstack",
// 			Needs:     []string{},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Config: map[string]interface{}{
// 				"api": map[string]interface{}{
// 					"key":   "",
// 					"token": "",
// 				},
// 				"proxied": false,
// 			},
// 		},
// 		"debug": {
// 			Enabled:   true,
// 			Version:   "latest", // TODO
// 			Namespace: "",
// 			Needs:     []string{},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Commands: map[string]Command{
// 				"run": {
// 					Script: []string{"bash"},
// 				},
// 			},
// 		},
// 		"eventrouter": {
// 			Enabled:   true,
// 			Version:   "latest", // TODO
// 			Namespace: "kube-system",
// 			Needs:     []string{"kubernetes"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook eventrouter.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook eventrouter.yml"},
// 				},
// 			},
// 		},
// 		"external_dns": {
// 			Enabled:   false,
// 			Version:   "5.4.4", // TODO
// 			Namespace: "external-dns",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "external-dns",
// 				Version: "5.4.4",
// 				Url:     "https://charts.bitnami.com/bitnami/external-dns-5.4.4.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://charts.bitnami.com/bitnami",
// 					Name: "bitnami",
// 				},
// 			},
// 			Config: map[string]interface{}{
// 				"provider": "cloudflare",
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook external_dns.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook external_dns.yml"},
// 				},
// 			},
// 		},
// 		"cilium_cni": {
// 			Enabled:   false,
// 			Version:   "1.10.5", // TODO
// 			Namespace: "kube-system",
// 			Subdomain: "cilium",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{"cni"},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "cilium",
// 				Version: "1.10.5",
// 				Url:     "https://helm.cilium.io/cilium-1.10.5.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://helm.cilium.io/",
// 					Name: "cilium",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook cilium_cni.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook cilium_cni.yml"},
// 				},
// 			},
// 		},
// 		"loki": {
// 			Enabled:   true,
// 			Version:   "2.6.0", // TODO
// 			Namespace: "loki",
// 			Subdomain: "loki",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{"logs"},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "loki",
// 				Version: "2.6.0",
// 				Url:     "https://github.com/grafana/helm-charts/releases/download/loki-2.6.0/loki-2.6.0.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://grafana.github.io/helm-charts",
// 					Name: "grafana",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook loki.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook loki.yml"},
// 				},
// 			},
// 		},
// 		"promtail": {
// 			Enabled:   true,
// 			Version:   "3.8.2", // TODO
// 			Namespace: "promtail",
// 			Subdomain: "promtail",
// 			Needs:     []string{"kubernetes", "prometheus_crds", "loki"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "promtail",
// 				Version: "3.8.2",
// 				Url:     "https://github.com/grafana/helm-charts/releases/download/promtail-3.8.2/promtail-3.8.2.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://grafana.github.io/helm-charts",
// 					Name: "grafana",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook promtail.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook promtail.yml"},
// 				},
// 			},
// 		},
// 		"nginx_ingress": {
// 			Enabled:   true,
// 			Version:   "3.36.0", // TODO
// 			Namespace: "nginx-ingress",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{"ingress"},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "ingress-nginx",
// 				Version: "3.36.0",
// 				Url:     "https://github.com/kubernetes/ingress-nginx/releases/download/helm-chart-3.36.0/ingress-nginx-3.36.0.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://kubernetes.github.io/ingress-nginx",
// 					Name: "ingress-nginx",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook nginx_ingress.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook nginx_ingress.yml"},
// 				},
// 			},
// 		},
// 		"slack": {
// 			Enabled:   false,
// 			Version:   version, // TODO
// 			Namespace: "cloudstack",
// 			Needs:     []string{},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Config: map[string]interface{}{
// 				"api": map[string]interface{}{
// 					"url": "",
// 				},
// 				"channel": "#cloudstack",
// 			},
// 		},
// 		"alertmanager": {
// 			Enabled:   false,
// 			Version:   version, // TODO
// 			Namespace: "alertmanager",
// 			Subdomain: "alertmanager",
// 			Needs:     []string{"prometheus", "prometheus_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Config: map[string]interface{}{
// 				"webhook": map[string]interface{}{
// 					"url":   "",
// 					"token": "",
// 				},
// 				"receiver": "",
// 			},
// 		},
// 		"grafana": {
// 			Enabled:   true,
// 			Version:   version, // TODO
// 			Namespace: "grafana",
// 			Subdomain: "grafana",
// 			Needs:     []string{"prometheus", "prometheus_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 		},
// 		"prometheus": {
// 			Enabled:   true,
// 			Version:   "19.2.2", // TODO
// 			Namespace: "monitoring",
// 			Subdomain: "prometheus",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{"metrics"},
// 			Rejects:   []string{},
// 			Backup: BackupConfig{
// 				Enabled: false,
// 			},
// 			Config: map[string]interface{}{
// 				"remote_write": map[string]interface{}{
// 					"enabled": false,
// 					"targets": []map[string]interface{}{
// 						{
// 							"url":           "http://prometheus.example.com",
// 							"name":          "default",
// 							"remoteTimeout": "30s",
// 							"sendExemplars": false,
// 							"headers": map[string]string{
// 								"X-Scope-OrgID": "",
// 							},
// 							"writeRelabelConfigs": "",
// 							"oauth2":              map[string]string{},
// 							"basicAuth": map[string]string{
// 								"username": "",
// 								"password": "",
// 							},
// 							"bearerToken":     "",
// 							"bearerTokenFile": "",
// 							"authorization": map[string]interface{}{
// 								"type":        "Bearer",
// 								"credentials": map[string]string{},
// 							},
// 						},
// 					},
// 				},
// 				"receiver": "",
// 			},
// 			Chart: ChartConfig{
// 				Name:    "kube-prometheus-stack",
// 				Version: "19.2.2",
// 				Url:     "https://github.com/prometheus-community/helm-charts/releases/download/kube-prometheus-stack-19.2.2/kube-prometheus-stack-19.2.2.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://prometheus-community.github.io/helm-charts",
// 					Name: "prometheus-community",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook prometheus.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook prometheus.yml"},
// 				},
// 			},
// 		},
// 		"prometheus_crds": {
// 			Enabled:   true,
// 			Version:   "19.2.2", // TODO
// 			Namespace: "kube-system",
// 			Needs:     []string{"kubernetes"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook prometheus_crds.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook prometheus_crds.yml"},
// 				},
// 			},
// 			Artifacts: []ArtifactConfig{
// 				{
// 					Name:     "crd-alertmanagerconfigs.yaml",
// 					Filename: "crd-alertmanagerconfigs.yaml",
// 					Url:      "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-19.2.2/charts/kube-prometheus-stack/crds/crd-alertmanagerconfigs.yaml",
// 				},
// 				{
// 					Name:     "crd-alertmanagers.yaml",
// 					Filename: "crd-alertmanagers.yaml",
// 					Url:      "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-19.2.2/charts/kube-prometheus-stack/crds/crd-alertmanagers.yaml",
// 				},
// 				{
// 					Name:     "crd-podmonitors.yaml",
// 					Filename: "crd-podmonitors.yaml",
// 					Url:      "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-19.2.2/charts/kube-prometheus-stack/crds/crd-podmonitors.yaml",
// 				},
// 				{
// 					Name:     "crd-prometheuses.yaml",
// 					Filename: "crd-prometheuses.yaml",
// 					Url:      "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-19.2.2/charts/kube-prometheus-stack/crds/crd-prometheuses.yaml",
// 				},
// 				{
// 					Name:     "crd-prometheusrules.yaml",
// 					Filename: "crd-prometheusrules.yaml",
// 					Url:      "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-19.2.2/charts/kube-prometheus-stack/crds/crd-prometheusrules.yaml",
// 				},
// 				{
// 					Name:     "crd-servicemonitors.yaml",
// 					Filename: "crd-servicemonitors.yaml",
// 					Url:      "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-19.2.2/charts/kube-prometheus-stack/crds/crd-servicemonitors.yaml",
// 				},
// 				{
// 					Name:     "crd-thanosrulers.yaml",
// 					Filename: "crd-thanosrulers.yaml",
// 					Url:      "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-19.2.2/charts/kube-prometheus-stack/crds/crd-thanosrulers.yaml",
// 				},
// 			},
// 		},
// 		"tempo": {
// 			Enabled:   false,
// 			Version:   "0.2.6", // TODO
// 			Namespace: "tempo",
// 			Subdomain: "tempo",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{"tracing"},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "grafana-tempo",
// 				Version: "0.2.6",
// 				Url:     "https://charts.bitnami.com/bitnami/grafana-tempo-0.2.6.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://charts.bitnami.com/bitnami",
// 					Name: "bitnami",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook tempo.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook tempo.yml"},
// 				},
// 			},
// 		},
// 		"portainer": {
// 			Enabled:   true,
// 			Version:   "1.0.18", // TODO
// 			Namespace: "portainer",
// 			Subdomain: "portainer",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{},
// 			Backup: BackupConfig{
// 				Enabled: false,
// 			},
// 			Rejects: []string{},
// 			Config: map[string]interface{}{
// 				"image": map[string]interface{}{
// 					"version": "2.9.2",
// 				},
// 			},
// 			Chart: ChartConfig{
// 				Name:    "portainer",
// 				Version: "1.0.18",
// 				Url:     "https://github.com/portainer/k8s/releases/download/portainer-1.0.18/portainer-1.0.18.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://portainer.github.io/k8s/",
// 					Name: "portainer",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook portainer.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook portainer.yml"},
// 				},
// 				"status": {
// 					Script: []string{"ansible-playbook portainer.yml"},
// 				},
// 			},
// 		},
// 		"velero": {
// 			Enabled:   true,
// 			Version:   "1.0.18", // TODO
// 			Namespace: "velero",
// 			Subdomain: "velero",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{"backups"},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "velero",
// 				Version: "1.0.18",
// 				Url:     "https://github.com/portainer/k8s/releases/download/portainer-1.0.18/portainer-1.0.18.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://portainer.github.io/k8s/",
// 					Name: "velero",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook velero.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook velero.yml"},
// 				},
// 				"status": {
// 					Script: []string{"ansible-playbook velero.yml"},
// 				},
// 				"backup": {
// 					Script: []string{"ansible-playbook velero.yml"},
// 				},
// 			},
// 		},
// 		"portainer_agent": {
// 			Enabled:   false,
// 			Version:   "2.9.0", // TODO
// 			Namespace: "portainer-agent",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook portainer_agent.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook portainer_agent.yml"},
// 				},
// 			},
// 		},
// 		"weave_scope": {
// 			Enabled:   false,
// 			Version:   "1.13.2", // TODO
// 			Namespace: "weave",
// 			Subdomain: "scope",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook weave_scope.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook weave_scope.yml"},
// 				},
// 			},
// 		},
// 		"keycloak": {
// 			Enabled:   false,
// 			Version:   "14.0.1", // TODO
// 			Namespace: "keycloak",
// 			Subdomain: "login",
// 			Needs:     []string{"kubernetes", "prometheus_crds", "ingress"},
// 			Provides:  []string{"sso"},
// 			Rejects:   []string{},
// 			Config: map[string]interface{}{
// 				"admin": map[string]interface{}{
// 					"user":     "",
// 					"password": "",
// 				},
// 			},
// 			Chart: ChartConfig{
// 				Name:    "keycloak",
// 				Version: "14.0.1",
// 				Url:     "https://github.com/codecentric/helm-charts/releases/download/keycloak-14.0.1/keycloak-14.0.1.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://codecentric.github.io/helm-charts",
// 					Name: "codecentric",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook keycloak.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook keycloak.yml"},
// 				},
// 			},
// 		},
// 		"kubeapps": {
// 			Enabled:   false,
// 			Version:   "7.3.2", // TODO
// 			Namespace: "kubeapps",
// 			Subdomain: "kubeapps",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "kubeapps",
// 				Version: "7.3.2",
// 				Url:     "https://charts.bitnami.com/bitnami/kubeapps-7.3.2.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://charts.bitnami.com/bitnami",
// 					Name: "bitnami",
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook kubeapps.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook kubeapps.yml"},
// 				},
// 			},
// 		},
// 		"sonarqube": {
// 			Enabled: false,
// 		},
// 		"metallb": {
// 			Enabled: false,
// 		},
// 		"fission": {
// 			Enabled: false,
// 		},
// 		"gitlab": {
// 			Enabled: false,
// 		},
// 		"harbor": {
// 			Enabled: false,
// 		},
// 		"botkube": {
// 			Enabled:   false,
// 			Version:   "v0.12.3", // TODO
// 			Namespace: "botkube",
// 			Needs:     []string{"kubernetes", "prometheus_crds"},
// 			Provides:  []string{},
// 			Rejects:   []string{},
// 			Chart: ChartConfig{
// 				Name:    "botkube",
// 				Version: "v0.12.3",
// 				Url:     "https://infracloudio.github.io/charts/botkube-v0.12.3.tgz",
// 				Repo: ChartRepoConfig{
// 					Url:  "https://infracloudio.github.io/charts",
// 					Name: "infracloudio",
// 				},
// 			},
// 			Config: map[string]interface{}{
// 				"communications": map[string]interface{}{
// 					"mattermost": map[string]interface{}{
// 						"enabled": false,
// 						"botname": "botkube",
// 						"url":     "",
// 						"token":   "",
// 						"team":    "",
// 						"channel": "",
// 					},
// 				},
// 				"config": map[string]interface{}{
// 					"settings": map[string]interface{}{
// 						"clustername": "",
// 						"kubectl": map[string]bool{
// 							"enabled": true,
// 						},
// 					},
// 					"resources": []map[string]interface{}{
// 						// map[string]interface{}{
// 						// 	"name": "velero.io/v1/backups",
// 						// 	"namespaces": map[string]interface{}{
// 						// 		"include": []string{
// 						// 			"all",
// 						// 		},
// 						// 	},
// 						// 	"events": []string{
// 						// 		"all",
// 						// 	},
// 						// 	"updateSetting": map[string]interface{}{
// 						// 		"includeDiff": true,
// 						// 		"fields": []string{
// 						// 			"status.phase",
// 						// 		},
// 						// 	},
// 						// },
// 						// map[string]interface{}{
// 						// 	"name": "velero.io/v1/restores",
// 						// 	"namespaces": map[string]interface{}{
// 						// 		"include": []string{
// 						// 			"all",
// 						// 		},
// 						// 	},
// 						// 	"events": []string{
// 						// 		"all",
// 						// 	},
// 						// 	"updateSetting": map[string]interface{}{
// 						// 		"includeDiff": true,
// 						// 		"fields": []string{
// 						// 			"status.phase",
// 						// 		},
// 						// 	},
// 						// },
// 					},
// 				},
// 			},
// 			Commands: map[string]Command{
// 				"install": {
// 					Script: []string{"ansible-playbook botkube.yml"},
// 				},
// 				"uninstall": {
// 					Script: []string{"ansible-playbook botkube.yml"},
// 				},
// 			},
// 		},
// 	},
// }
