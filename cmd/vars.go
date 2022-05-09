package cmd

import (
	"os"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Constants
// These are mainly used for setting defaults to the CLI flags
// As such they can be overriden by the user
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
const EnvPrefix string = "polycrate"
const defaultFailedCode int = 1

// Global variable to decide if an action runs local or in the container
// Can be overriden with the --local flag
var local bool

// Global variable to set the loglevel
// Can be overriden with the --loglevel flag
var logLevel string

// Global variable to decide if the container image should be pulled before running the container
// Can be overriden with the --pull flag
var pull bool

// Global variable to decide if some actions in the workspace should be forced
// Can be overriden with the --force flag
var force bool

// Global variable to decide if the container should run in interactive mode (i.e. `-it`)
// Can be overriden with the --interactive flag
var interactive bool

// Global variable to decide if the container should be built from the workspace Dockerfile (should it exist)
// Can be overriden with the --build flag
var build bool

// Global Logrus loglevel variable
// This one is used by the main logger as well as the special
// logger of the buildContainerImage() function
var logrusLevel log.Level

// Global variable for the current working directory
var cwd, _ = os.Getwd()

// Global variable to decide the output format (json/yaml)
// Can be overriden with the --output-format flag
var outputFormat string

// Global variable to decide if the workspace snapshot should be printed
// This usually means that no other action/command will be executed
// Can be overriden with the --snapshot flag
var snapshot bool

// Global validator variable
// We're hooking up special validators in root.go->init()
var validate = validator.New()

// Build meta
// These variables will be set at build time
var version string = "latest"
var commit string
var date string

// Global workspace variable
// This variable holds the allmighty workspace struct
var workspace Workspace

// Global workspace config variable
// This variable holds the configuration loaded from the workspace config file (e.g. workspace.poly)
var workspaceConfig = viper.New()

// Global variable that holds the block paths discovered at block discovery in workspace.load()
var blockPaths []string

// Inventory
var inventory string
var inventoryConfigObject = viper.New()
