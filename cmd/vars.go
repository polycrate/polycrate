package cmd

import (
	"os"
	"path/filepath"

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
const RegistryUrl string = "cargo.ayedo.cloud"
const RegistryBlockNamespace string = "polycrate-blocks"
const RegistryApiBase string = "wp-json/wp/v2"
const RegistryBaseImage string = "cargo.ayedo.dev/library/scratch:latest"
const DefaultEditor string = "code"
const defaultFailedCode int = 1

const GitLabDefaultUrl string = "https://gitlab.com"
const GitLabDefaultTransport string = "ssh"

const GitDefaultBranch string = "main"
const GitDefaultRemote string = "origin"

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

// Global variable to configure the editor used to open files
var editor string

// Global variable to decide if the container should be built from the workspace Dockerfile (should it exist)
// Can be overriden with the --build flag
var build bool

// Global Logrus loglevel variable
// This one is used by the main logger as well as the special
// logger of the buildContainerImage() function
var logrusLevel log.Level

// Global history variable
var history HistoryLog

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

// Global sync variable
// This variable holds the sync struct
var sync Sync

// Global workspace variable
// This variable holds the allmighty workspace struct
var workspace Workspace

// Global registry variable
// This variable holds the registry struct
var registry Registry

// Global variable that holds the block paths discovered at block discovery in workspace.load()
var blockPaths []string
var installedBlocks []Block

// Global variable that holds the workspace paths discovered at workspace discovery
var workspacePaths []string

var localWorkspaceIndex map[string]string = make(map[string]string)

// Inventory
var inventory string
var inventoryConfigObject = viper.New()

var home, _ = os.UserHomeDir()
var polycrateHome = filepath.Join(home, ".polycrate")
var polycrateWorkspaceDir = filepath.Join(polycrateHome, "workspaces")
var polycrateRuntimeDir = filepath.Join(polycrateHome, "run")
var polycrateConfigFilePath = filepath.Join(polycrateHome, "polycrate.yml")
var config PolycrateConfig
