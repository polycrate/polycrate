package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var home, _ = os.UserHomeDir()

// Constants
// These are mainly used for setting defaults to the CLI flags
// As such they can be overriden by the user

// default image to use for the Polycrate container
const WorkspaceConfigImageRef string = "cargo.ayedo.cloud/library/polycrate"

// default blocks directory inside the workspace
const WorkspaceConfigBlocksRoot string = "blocks"

// default artifacts directory inside the workspace
const WorkspaceConfigArtifactsRoot string = "artifacts"

// default workflows directory inside the workspace
const WorkspaceConfigWorkflowsRoot string = "workflows"

// default remote root (can be used when running commands on remote machines)
const WorkspaceConfigRemoteRoot string = "/polycrate"

// default Dockerfile inside the workspace
const WorkspaceConfigDockerfile string = "Dockerfile.poly"

// default root directory inside the Polycrate container
const WorkspaceContainerRoot string = "/workspace"

// default filename for the ssh public key inside the workspace
const WorkspaceConfigSshPublicKey string = "id_rsa.pub"

// default filename for the ssh private key inside the workspace
const WorkspaceConfigSshPrivateKey string = "id_rsa"

// default workspace config file
const WorkspaceConfigFile string = "workspace.poly"

// default workspace logs path
const WorkspaceConfigLogsRoot string = ".logs"

// default block config file
const BlocksConfigFile string = "block.poly"

// default block changelog file
const BlocksChangelogFile string = "CHANGELOG.poly"

// default env prefix
const EnvPrefix string = "polycrate"

// default registry url
const HubUrl string = "https://hub.polycrate.io"

// default registry url
const RegistryUrl string = "cargo.ayedo.cloud"

// default registry namespace
const RegistryBlockNamespace string = "polycrate-blocks"

// default registry api base
//const RegistryApiBase string = "wp-json/wp/v2"

// default registry base image that blocks are packaged with
const RegistryBaseImage string = "cargo.ayedo.cloud/library/scratch:latest"

// default local editor to open when using `workspace edit`
const DefaultEditor string = "code"

// default exit code for failed commands
const defaultFailedCode int = 1

const GitLabDefaultUrl string = "https://gitlab.com"
const GitLabDefaultTransport string = "ssh"

const GitDefaultBranch string = "main"
const GitDefaultRemote string = "origin"

const ValidateMetaDataNameRegex string = "^[a-zA-Z]+([-/_]?[a-zA-Z0-9_]+)+$"
const ValidateBlockNameRegex string = `([^\/]+\.[^\/.]+)?\/?([^:]+):?(.+)?`

const WorkspaceEventHandler string = "workspace"

const WorkspaceLogDateFormat string = "2006/01/02"
const WorkspaceLogDateOutputFormat string = "02.01.2006 15:04:05"

// Variable to decide if polycrate is run during block development
var dev bool

// Global variable to decide if an action runs local or in the container
// Can be overriden with the --local flag
var local bool

// Global variable to decide if the container image should be pulled before running the container
// Can be overriden with the --pull flag
var pull bool

// Global variable to decide if blocks that are missing from the workspace should automatically be pulled
// Can be overriden with the --blocks-auto-pull flag
var blocksAutoPull bool

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
//var history HistoryLog

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
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Build meta
// These variables will be set at build time
var version string = "latest"
var commit string
var date string

// Global sync variable
// This variable holds the sync struct
//var sync Sync

// Global workspace variable
// This variable holds the allmighty workspace struct
var workspace Workspace
var defaultWorkspace Workspace

// Global registry variable
// This variable holds the registry struct
// var registry Registry

// Global variable that holds the block paths discovered at block discovery in workspace.load()
//var blockPaths []string

// var installedBlocks []Block

// Global variable that holds the workspace paths discovered at workspace discovery
//var workspacePaths []string

var localWorkspaceIndex map[string]string = make(map[string]string)

// Inventory
// var inventory string
// var inventoryConfigObject = viper.New()

var polycrate Polycrate
var polycrateConfigDir = filepath.Join(home, ".polycrate")
var polycrateWorkspaceDir = filepath.Join(polycrateConfigDir, "workspaces")
var polycrateRuntimeDir = filepath.Join(polycrateConfigDir, "run")
var polycrateConfigFilePath = filepath.Join(polycrateConfigDir, "polycrate.yml")

var KubeconfigPath string = filepath.Join(home, ".kube", "config")

// Errors
var ErrDependencyNotResolved = errors.New("block dependency not resolved")
var ErrWorkspaceConfigNotFound = errors.New("workspace config not found")

//var signals = make(chan os.Signal, 1)

var globalCmd *cobra.Command

var backplane Backplane
