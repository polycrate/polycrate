package cmd

import (
	_ "embed"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var validate *validator.Validate

var defaultsYml []byte

// Globals
var local bool = false

var logLevel string
var pull bool
var force bool
var interactive bool
var envPrefix string = "polycrate"
var build bool
var logrusLevel log.Level

const defaultFailedCode = 1

//var devDir string
var cwd, _ = os.Getwd()
var callUUID = getCallUUID()
var timeFormat string = "2006-01-02T15:04:05-0700"
var utcNow time.Time = time.Now().UTC()
var now string = time.Now().Format(timeFormat)
var home, _ = homedir.Dir()
var overrides []string
var workdir string = workspace.path
var workdirContainer string = "/workdir"
var outputFormat string
var extraEnv []string
var extraMounts []string
var snapshot bool

// Build meta
var version string = "latest"
var commit string
var date string

// Container stuff
var mounts []string
var ports []string

// current
var currentBlock *Block
var currentAction *Action
var currentWorkflow *Workflow

// Git
var gitConfigObject = viper.New()

// Workspace
var workspace Workspace           // Struct
var workspaceConfig = viper.New() // Viper
//var workspaceDir string           // local

var workspaceConfigFilePath string
var workspaceContainerConfigFilePath string

// Blocks

var blockName string
var blockPaths []string

//var block Block
var blockDir string
var blocksDir string
var blocksContainerDir string
var blockContainerDir string

// Actions
var action Action
var actionName string

// Inventory
var inventoryLocalPath string = "/tmp/inventory.yml"
var inventoryContainerPath string = "/tmp/inventory.yml"
var inventory string
var inventoryConfigObject = viper.New()

// Kubernetes Kubeconfig
var kubeconfig string

// Globals
var blocksRoot string = "blocks"
var artifactsRoot string = "artifacts"
var blockConfigFile string = "block.poly"
var workspaceConfigFile string = "workspace.poly"
var workspaceContainerDir string = "/workspace" // Container
