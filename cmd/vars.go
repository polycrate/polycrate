package cmd

import (
	_ "embed"
	"os"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Globals
const WorkspaceConfigImageRef string = "ghcr.io/polycrate/polycrate"
const WorkspaceConfigBlocksRoot string = "blocks"
const WorkspaceConfigArtifactsRoot string = "artifacts"
const WorkspaceConfigWorkflowsRoot string = "workflows"
const WorkspaceConfigRemoteRoot string = "/polycrate"
const WorkspaceConfigDockerfile string = "Dockerfile.poly"
const WorkspaceContainerRoot string = "/workspace"
const WorkspaceConfigSshPublicKey string = "id_rsa.pub"
const WorkspaceConfigSshPrivateKey string = "id_rsa"
const WorkspaceConfigFile string = "workspace.poly"
const BlocksConfigFile string = "block.poly"

var blockConfigFile string = "block.poly"
var workspaceConfigFilePath string
var workspaceContainerConfigFilePath string

var local bool = false

var logLevel string
var pull bool
var force bool
var interactive bool
var envPrefix string = "polycrate"
var build bool
var logrusLevel log.Level

const defaultFailedCode = 1

var cwd, _ = os.Getwd()

var outputFormat string
var snapshot bool

var validate = validator.New()

// Build meta
var version string = "latest"
var commit string
var date string

// Workspace
var workspace Workspace           // Struct
var workspaceConfig = viper.New() // Viper
var blockPaths []string

//var block Block
var blocksDir string

// Inventory
var inventory string
var inventoryConfigObject = viper.New()

// Globals
