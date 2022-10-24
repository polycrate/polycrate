/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"errors"
	goErrors "errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/sync/errgroup"

	"github.com/jeremywohl/flatten"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage the workspace",
	Long:  `Manage the workspace`,
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load().Flush()

		workspace.Inspect()
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}

type ImageConfig struct {
	Reference string `yaml:"reference" mapstructure:"reference" json:"reference" validate:"required"`
	Version   string `yaml:"version" mapstructure:"version" json:"version" validate:"required"`
}

type Metadata struct {
	Name        string            `mapstructure:"name" json:"name" validate:"required,metadata_name"`
	Description string            `mapstructure:"description" json:"description"`
	Labels      map[string]string `mapstructure:"labels" json:"labels"`
	Alias       []string          `mapstructure:"alias" json:"alias"`
}

type WorkspaceConfig struct {
	Image      ImageConfig `yaml:"image" mapstructure:"image" json:"image" validate:"required"`
	BlocksRoot string      `yaml:"blocksroot" mapstructure:"blocksroot" json:"blocksroot" validate:"required"`
	// The block configuration file (default: block.poly)
	BlocksConfig    string                 `yaml:"blocksconfig" mapstructure:"blocksconfig" json:"blocksconfig" validate:"required"`
	WorkspaceConfig string                 `yaml:"workspaceconfig" mapstructure:"workspaceconfig" json:"workspaceconfig" validate:"required"`
	WorkflowsRoot   string                 `yaml:"workflowsroot" mapstructure:"workflowsroot" json:"workflowsroot" validate:"required"`
	ArtifactsRoot   string                 `yaml:"artifactsroot" mapstructure:"artifactsroot" json:"artifactsroot" validate:"required"`
	ContainerRoot   string                 `yaml:"containerroot" mapstructure:"containerroot" json:"containerroot" validate:"required"`
	SshPrivateKey   string                 `yaml:"sshprivatekey" mapstructure:"sshprivatekey" json:"sshprivatekey" validate:"required"`
	SshPublicKey    string                 `yaml:"sshpublickey" mapstructure:"sshpublickey" json:"sshpublickey" validate:"required"`
	RemoteRoot      string                 `yaml:"remoteroot" mapstructure:"remoteroot" json:"remoteroot" validate:"required"`
	Dockerfile      string                 `yaml:"dockerfile" mapstructure:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	Globals         map[string]interface{} `yaml:"globals" mapstructure:"globals" json:"globals"`
}

type Workspace struct {
	Name        string            `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description string            `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias       []string          `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	// Format: block:version
	Dependencies    []string        `yaml:"dependencies,omitempty" mapstructure:"dependencies,omitempty" json:"dependencies,omitempty"`
	Config          WorkspaceConfig `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	Blocks          []Block         `yaml:"blocks,omitempty" mapstructure:"blocks,omitempty" json:"blocks,omitempty" validate:"dive,required"`
	Workflows       []Workflow      `yaml:"workflows,omitempty" mapstructure:"workflows,omitempty" json:"workflows,omitempty"`
	currentBlock    *Block
	currentAction   *Action
	currentWorkflow *Workflow
	currentStep     *Step
	index           *WorkspaceIndex
	env             map[string]string
	mounts          map[string]string
	err             error
	runtimeDir      string
	installedBlocks []Block
	Path            string      `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	SyncOptions     SyncOptions `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	LocalPath       string      `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath   string      `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
	//overrides       []string
	ExtraEnv    []string `yaml:"extraenv,omitempty" mapstructure:"extraenv,omitempty" json:"extraenv,omitempty"`
	ExtraMounts []string `yaml:"extramounts,omitempty" mapstructure:"extramounts,omitempty" json:"extramounts,omitempty"`
	containerID string
	loaded      bool
	Version     string                 `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
	Identifier  string                 `yaml:"identifier,omitempty" mapstructure:"identifier,omitempty" json:"identifier,omitempty"`
	Meta        map[string]interface{} `yaml:"meta,omitempty" mapstructure:"meta,omitempty" json:"meta,omitempty"`
}

type WorkspaceIndex struct {
	Actions   map[string]*Action   `yaml:"actions" mapstructure:"actions" json:"actions"`
	Steps     map[string]*Step     `yaml:"steps" mapstructure:"steps" json:"steps"`
	Blocks    map[string]*Block    `yaml:"blocks" mapstructure:"blocks" json:"blocks"`
	Workflows map[string]*Workflow `yaml:"workflows" mapstructure:"workflows" json:"workflows"`
}

type WorkspaceSnapshot struct {
	Workspace *Workspace        `yaml:"workspace,omitempty" mapstructure:"workspace,omitempty" json:"workspace,omitempty"`
	Action    *Action           `yaml:"action,omitempty" mapstructure:"action,omitempty" json:"action,omitempty"`
	Block     *Block            `yaml:"block,omitempty" mapstructure:"block,omitempty" json:"block,omitempty"`
	Workflow  *Workflow         `yaml:"workflow,omitempty" mapstructure:"workflow,omitempty" json:"workflow,omitempty"`
	Step      *Step             `yaml:"step,omitempty" mapstructure:"step,omitempty" json:"step,omitempty"`
	Env       map[string]string `yaml:"env,omitempty" mapstructure:"env,omitempty" json:"env,omitempty"`
	Mounts    map[string]string `yaml:"mounts,omitempty" mapstructure:"mounts,omitempty" json:"mounts,omitempty"`
}

func (c *Workspace) CreateSshKeys() *Workspace {
	err := CreateSSHKeys()
	if err != nil {
		c.err = err
		return c
	}
	return c
}

func (c *Workspace) RegisterSnapshotEnv(snapshot WorkspaceSnapshot) *Workspace {
	// create empty map to hold the flattened keys
	var jsonMap map[string]interface{}
	// err := mapstructure.Decode(snapshot, &jsonMap)
	// if err != nil {
	// 	panic(err)
	// }
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	// marshal the snapshot into json
	jsonData, err := json.Marshal(snapshot)
	if err != nil {
		log.Errorf("Error marshalling: %s", err)
		c.err = err
		return c
	}

	// unmarshal the json into the previously created json map
	// flatten needs this input format: map[string]interface
	// which we achieve by first marshalling the struct (snapshot)
	// and then unmarshalling the resulting bytes into our json structure
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		log.Errorf("Error unmarshalling: %s", err)
		c.err = err
		return c
	}

	// flatten to key_key_0_key=value
	flat, err := flatten.Flatten(jsonMap, "", flatten.UnderscoreStyle)
	if err != nil {
		c.err = err
		return c
	}

	for key := range flat {
		keyString := fmt.Sprintf("%v", flat[key])
		//fmt.Printf("%s=%s\n", strings.ToUpper(key), keyString)
		workspace.registerEnvVar(strings.ToUpper(key), keyString)
	}

	return c
}

func (c *Workspace) Snapshot() {
	snapshot := c.GetSnapshot()
	printObject(snapshot)
	//convertToEnv(&snapshot)
}

func (c *Workspace) Inspect() {
	printObject(c)
}

func (c *Workspace) Flush() *Workspace {
	if c.err != nil {
		log.Fatal(c.err)
	}
	return c
}

func (c *Workspace) RunAction(address string) *Workspace {
	// Find action in index and report
	action := c.LookupAction(address)

	if action != nil {
		block := c.GetBlockFromIndex(action.Block)

		workspace.registerCurrentAction(action)
		workspace.registerCurrentBlock(block)

		log.WithFields(log.Fields{
			"block":     block.Name,
			"workspace": c.Name,
		}).Debugf("Registering current block")

		log.WithFields(log.Fields{
			"action":    action.Name,
			"block":     block.Name,
			"workspace": c.Name,
		}).Debugf("Registering current action")

		if block.Template {
			log.WithFields(log.Fields{
				"block":     block.Name,
				"action":    action.Name,
				"workspace": c.Name,
				"template":  block.Template,
			}).Errorf("This is a template block")
			c.err = goErrors.New("this is a template block. Not running action")
			return c
		}

		// Write log here
		if snapshot {
			c.Snapshot()
		} else {
			sync.Log(fmt.Sprintf("Running action %s of block %s", action.Name, block.Name)).Flush()
			err := action.Run()
			if err != nil {
				c.err = err

				sync.Log(fmt.Sprintf("Action %s of block %s failed", action.Name, block.Name)).Flush()
				return c
			}
			sync.Log(fmt.Sprintf("Action %s of block %s was successful", action.Name, block.Name)).Flush()
		}

		// Reload Block after action execution to update artifacts, inventory and kubeconfig
		block.Reload()
	} else {
		c.err = goErrors.New("cannot find Action with address " + address)
		return c
	}
	if sync.Options.Enabled && sync.Options.Auto {
		sync.Sync().Flush()
	}
	return c
}

func (c *Workspace) RunWorkflow(name string) *Workspace {
	// Find action in index and report
	workflow := c.LookupWorkflow(name)

	if workflow != nil {
		workspace.registerCurrentWorkflow(workflow)

		if snapshot {
			c.Snapshot()
		} else {
			err := workflow.run()
			if err != nil {
				c.err = err
				return c
			}
		}
	} else {
		c.err = goErrors.New("cannot find Workflow " + name)
		return c
	}
	return c
}

func (c *Workspace) RunStep(name string) *Workspace {
	// Find action in index and report
	step := c.LookupStep(name)

	if step != nil {
		if snapshot {
			c.Snapshot()
		} else {
			err := step.run()
			if err != nil {
				c.err = err
				return c
			}
		}
	} else {
		c.err = goErrors.New("cannot find step " + name)
		return c
	}
	return c
}

func (c *Workspace) registerBlock(block *Block) {
	c.index.Blocks[block.address] = block
}
func (c *Workspace) registerWorkflow(workflow *Workflow) {
	c.index.Workflows[workflow.address] = workflow
}
func (c *Workspace) registerAction(action *Action) {
	c.index.Actions[action.address] = action
}
func (c *Workspace) registerStep(step *Step) {
	c.index.Steps[step.address] = step
}

func (c *Workspace) LookupBlock(address string) *Block {
	return c.index.Blocks[address]
}
func (c *Workspace) LookupAction(address string) *Action {
	return c.index.Actions[address]
}
func (c *Workspace) ActivateAction(address string) *Action {
	return c.index.Actions[address]
}
func (c *Workspace) LookupWorkflow(address string) *Workflow {
	return c.index.Workflows[address]
}
func (c *Workspace) LookupStep(address string) *Step {
	return c.index.Steps[address]
}

func (w *Workspace) SoftloadWorkspaceConfig() *Workspace {
	var workspaceConfig = viper.New()
	workspaceConfigFilePath := filepath.Join(w.Path, WorkspaceConfigFile)

	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		w.err = err
		return w
	}

	if err := workspaceConfig.Unmarshal(&w); err != nil {
		w.err = err
		return w
	}

	return w
}

func (w *Workspace) LoadWorkspaceConfig() *Workspace {
	// This variable holds the configuration loaded from the workspace config file (e.g. workspace.poly)
	var workspaceConfig = viper.New()

	// Match CLI Flags with Config options
	// CLI Flags have precedence
	workspaceConfig.BindPFlag("config.image.version", rootCmd.Flags().Lookup("image-version"))
	workspaceConfig.BindPFlag("config.image.reference", rootCmd.Flags().Lookup("image-ref"))
	workspaceConfig.BindPFlag("config.blocksroot", rootCmd.Flags().Lookup("blocks-root"))
	workspaceConfig.BindPFlag("config.blocksconfig", rootCmd.Flags().Lookup("blocks-config"))
	workspaceConfig.BindPFlag("config.workflowsroot", rootCmd.Flags().Lookup("workflows-root"))
	workspaceConfig.BindPFlag("config.workspaceconfig", rootCmd.Flags().Lookup("workspace-config"))
	workspaceConfig.BindPFlag("config.artifactsroot", rootCmd.Flags().Lookup("artifacts-root"))
	workspaceConfig.BindPFlag("config.containerroot", rootCmd.Flags().Lookup("container-root"))
	workspaceConfig.BindPFlag("config.remoteroot", rootCmd.Flags().Lookup("remote-root"))
	workspaceConfig.BindPFlag("config.sshprivatekey", rootCmd.Flags().Lookup("ssh-private-key"))
	workspaceConfig.BindPFlag("config.sshpublickey", rootCmd.Flags().Lookup("ssh-public-key"))
	workspaceConfig.BindPFlag("config.dockerfile", rootCmd.Flags().Lookup("dockerfile"))

	workspaceConfig.SetEnvPrefix(EnvPrefix)
	workspaceConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	workspaceConfig.AutomaticEnv()

	// Check if a full path has been given
	workspaceConfigFilePath := filepath.Join(w.LocalPath, workspace.Config.WorkspaceConfig)

	if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
		// The config file does not exist
		// We try to look in the list of workspaces in $HOME/.polycrate/workspaces

		// Assuming the "path" given is actually the name of a workspace
		workspaceName := w.LocalPath

		log.WithFields(log.Fields{
			"path":      workspaceConfigFilePath,
			"workspace": workspaceName,
		}).Debugf("Workspace config not found. Looking in the local workspace index")

		// Check if workspaceName exists in the local workspace index (see discoverWorkspaces())
		if localWorkspaceIndex[workspaceName] != "" {
			// We found a workspace with that name in the index
			path := localWorkspaceIndex[workspaceName]
			log.WithFields(log.Fields{
				"workspace": workspaceName,
				"path":      path,
			}).Debugf("Found workspace in the local workspace index")

			// Update the workspace config file path with the local workspace path from the index
			w.LocalPath = path
			workspaceConfigFilePath = filepath.Join(w.LocalPath, WorkspaceConfigFile)
		} else {
			w.err = fmt.Errorf("couldn't find workspace config at %s", workspaceConfigFilePath)
			return w
		}
	}

	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		w.err = err
		return w
	}

	if err := workspaceConfig.Unmarshal(&w); err != nil {
		w.err = err
		return w
	}

	if err := w.validate(); err != nil {
		w.err = err
		return w
	}

	// set runtime dir
	w.runtimeDir = filepath.Join(polycrateRuntimeDir, w.Name)

	return w
}

func (c *Workspace) saveWorkspace(path string) *Workspace {
	workspaceConfigFilePath := filepath.Join(path, workspace.Config.WorkspaceConfig)

	if _, err := os.Stat(workspaceConfigFilePath); !os.IsNotExist(err) {
		c.err = fmt.Errorf("config file already exists: %s", workspaceConfigFilePath)
		return c
	}

	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		c.err = err
		return c
	}

	err = ioutil.WriteFile(workspaceConfigFilePath, yamlBytes, 0644)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

func (c *Workspace) updateConfig(path string, value string) *Workspace {
	var sideloadConfig = viper.New()
	//var sideloadStruct Workspace

	// Check if a full path has been given
	workspaceConfigFilePath := filepath.Join(c.LocalPath, workspace.Config.WorkspaceConfig)

	log.WithFields(log.Fields{
		"workspace": c.Name,
		"key":       path,
		"value":     value,
		"path":      workspaceConfigFilePath,
	}).Debugf("Updating workspace config")

	if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
		c.err = fmt.Errorf("couldn't find workspace config at %s: %s", workspaceConfigFilePath, err)
		return c
	}

	yamlFile, err := ioutil.ReadFile(workspaceConfigFilePath)
	if err != nil {
		c.err = err
		return c
	}

	sideloadConfig.SetConfigType("yaml")
	sideloadConfig.SetConfigName("workspace")
	sideloadConfig.SetConfigFile(workspaceConfigFilePath)
	//sideloadConfig.ReadInConfig()

	err = sideloadConfig.ReadConfig(bytes.NewBuffer(yamlFile))
	if err != nil {
		c.err = err
		return c
	}

	// Update here
	sideloadConfig.Set(path, value)

	// if err := sideloadConfig.Unmarshal(&sideloadStruct); err != nil {
	// 	c.err = err
	// 	return c
	// }

	// if err := sideloadStruct.validate(); err != nil {
	// 	c.err = err
	// 	return c
	// }

	// Write back
	s := sideloadConfig.AllSettings()
	bs, err := yaml.Marshal(s)
	if err != nil {
		c.err = err
		return c
	}

	err = ioutil.WriteFile(workspaceConfigFilePath, bs, 0)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

func (w *Workspace) load() *Workspace {
	// Return the workspace if it has been already loaded
	if w.loaded {
		return w
	}

	// Set workspace.Path depending on --local
	if local {
		w.Path = w.LocalPath
	} else {
		w.Path = w.ContainerPath
	}

	log.WithFields(log.Fields{
		"path": w.LocalPath,
	}).Debugf("Loading workspace")

	// Load Workspace config (e.g. workspace.poly)
	w.LoadWorkspaceConfig().Flush()

	// Cleanup workspace runtime dir
	w.Cleanup().Flush()

	// Bootstrap the workspace index
	w.BootstrapIndex().Flush()

	// Find all blocks in the workspace
	w.DiscoverInstalledBlocks().Flush()

	// Pull dependencies
	w.PullDependencies().Flush()

	// Load all discovered blocks in the workspace
	w.ImportInstalledBlocks().Flush()

	// Resolve block dependencies
	w.ResolveBlockDependencies().Flush()

	// Update workflow and step addresses
	w.resolveWorkflows().Flush()

	// Bootstrap env vars
	w.bootstrapEnvVars().Flush()

	// Bootstrap container mounts
	w.bootstrapMounts().Flush()

	// Template action scripts
	w.templateActionScripts().Flush()

	log.WithFields(log.Fields{
		"workspace": w.Name,
		"blocks":    len(workspace.Blocks),
		"workflows": len(workspace.Workflows),
	}).Debugf("Workspace ready")

	// Mark workspace as loaded
	w.loaded = true

	// Load sync
	sync.Load().Flush()
	if sync.Options.Enabled && sync.Options.Auto {
		sync.Sync().Flush()
	}

	return w
}

func (w *Workspace) Cleanup() *Workspace {
	// Create runtime dir
	log.WithFields(log.Fields{
		"workspace": w.Name,
		"path":      w.runtimeDir,
	}).Debugf("Creating runtime directory")

	err := os.MkdirAll(w.runtimeDir, os.ModePerm)
	if err != nil {
		w.err = err
		return w
	}

	// Purge all contents of runtime dir
	dir, err := ioutil.ReadDir(w.runtimeDir)
	if err != nil {
		w.err = err
		return w
	}
	for _, d := range dir {
		log.WithFields(log.Fields{
			"workspace": w.Name,
			"path":      w.runtimeDir,
			"file":      d.Name(),
		}).Debugf("Cleaning runtime directory")
		err := os.RemoveAll(filepath.Join([]string{w.runtimeDir, d.Name()}...))
		if err != nil {
			w.err = err
			return w
		}
	}
	return w
}

func (c *Workspace) Uninstall() error {
	// e.g. $HOME/.polycrate/workspaces/workspace-1/artifacts/blocks/block-1
	if _, err := os.Stat(c.LocalPath); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      c.LocalPath,
		}).Debugf("Workspace directory does not exist")
	} else {
		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      c.LocalPath,
		}).Debugf("Removing workspace directory")
		err := os.RemoveAll(c.LocalPath)
		if err != nil {
			return err
		}
	}
	log.WithFields(log.Fields{
		"workspace": c.Name,
		"path":      c.LocalPath,
	}).Debugf("Successfully uninstalled workspace")
	return nil

}

// Resolves the 'from:' stanza of all blocks
func (w *Workspace) ResolveBlockDependencies() *Workspace {
	missing := len(w.Blocks)

	log.WithFields(log.Fields{
		"workspace": w.Name,
		"missing":   missing,
	}).Debugf("Resolving block dependencies")

	// Iterate over all Blocks in the Workspace
	// Until nothing is "missing" anymore
	for missing > 0 {
		for i := 0; i < len(w.Blocks); i++ {
			loadedBlock := &w.Blocks[i]

			if !loadedBlock.resolved {
				loadedBlock.Resolve()
				if loadedBlock.err != nil {
					switch {
					case errors.Is(loadedBlock.err, DependencyNotResolved):
						loadedBlock.resolved = false
						continue
					default:
						w.err = loadedBlock.err
						return w
					}
				}

				log.WithFields(log.Fields{
					"block":     loadedBlock.Name,
					"workspace": w.Name,
					"missing":   missing,
				}).Debugf("Block resolved")
			}
			missing--
		}
	}
	return w
}
func (w *Workspace) resolveWorkflows() *Workspace {
	log.WithFields(log.Fields{
		"workspace": w.Name,
	}).Debugf("Resolving workflows")

	for i := 0; i < len(w.Workflows); i++ {
		loadedWorkflow := &w.Workflows[i]

		loadedWorkflow.address = loadedWorkflow.Name
		// Register the workflow to the Index
		w.registerWorkflow(loadedWorkflow)

		// Loop over the steps
		for _, step := range loadedWorkflow.Steps {
			loadedStep := loadedWorkflow.getStepByName(step.Name)

			// Set step address
			loadedStep.address = strings.Join([]string{loadedWorkflow.Name, loadedStep.Name}, ".")
			loadedStep.Workflow = loadedWorkflow.Name

			log.WithFields(log.Fields{
				"workspace": w.Name,
				"workflow":  loadedWorkflow.Name,
				"step":      loadedStep.Name,
			}).Debugf("Validating step")
			if err := loadedStep.validate(); err != nil {
				w.err = err
				return w
			}

			log.WithFields(log.Fields{
				"workspace": w.Name,
				"workflow":  loadedWorkflow.Name,
				"step":      loadedStep.Name,
			}).Debugf("Registering step")
			w.registerStep(loadedStep)
		}

		log.WithFields(log.Fields{
			"workspace": w.Name,
			"workflow":  loadedWorkflow.Name,
		}).Debugf("Validating workflow")
		if err := loadedWorkflow.validate(); err != nil {
			w.err = err
			return w
		}

	}
	return w
}

func (c *Workspace) validate() error {
	log.WithFields(log.Fields{
		"workspace": c.Name,
	}).Debugf("Validating workspace")
	err := validate.Struct(c)

	if err != nil {
		log.WithFields(log.Fields{
			"workspace": c.Name,
		}).Errorf("Error validating workspace")

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error(err)
			return nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			log.WithFields(log.Fields{
				"workspace": c.Name,
				"option":    strings.ToLower(err.Namespace()),
				"error":     err.Tag(),
			}).Errorf("Validation error")
		}

		// from here you can create your own error messages in whatever language you wish
		return goErrors.New("error validating Workspace")
	}
	return nil
}

func (c *Workspace) ListBlocks() *Workspace {
	for _, block := range c.Blocks {
		str := block.Name

		if block.From != "" {
			str = str + " (from: " + block.From + ")"
		}
		fmt.Println(str)
	}
	return c
}

func (c *Workspace) ListWorkflows() *Workspace {
	for workflow := range c.index.Workflows {
		fmt.Println(workflow)
	}
	return c
}

func (c *Workspace) ListActions() *Workspace {
	for action := range c.index.Actions {
		fmt.Println(action)
	}
	return c
}

func (w *Workspace) BootstrapIndex() *Workspace {
	w.index = &WorkspaceIndex{}
	w.index.Actions = make(map[string]*Action)
	w.index.Blocks = make(map[string]*Block)
	w.index.Workflows = make(map[string]*Workflow)
	w.index.Steps = make(map[string]*Step)
	return w
}

func (c *Workspace) bootstrapMounts() *Workspace {
	// Prepare mounts map
	c.mounts = map[string]string{}

	// Mount the Workspace
	c.registerMount(c.LocalPath, c.ContainerPath)

	// Mount the Docker Socket if possible
	dockerSocket := "/var/run/docker.sock"
	if _, err := os.Stat(dockerSocket); !os.IsNotExist(err) {
		c.registerMount(dockerSocket, dockerSocket)
	}

	for _, extraMount := range c.ExtraMounts {
		// Split by :
		p := strings.Split(extraMount, ":")

		if len(p) == 2 {
			c.registerMount(p[0], p[1])
		} else {
			c.err = goErrors.New("Illegal value for mount found: " + extraMount)
			return c
		}
	}
	return c
}

// Resolve the templates of all actions
// If we don't do this, Ansible reads in Go's template tags
// which results in a jinja error
func (c *Workspace) templateActionScripts() *Workspace {
	// Loop over all blocks
	for _, block := range c.index.Blocks {
		// Generate intermediate snapshot
		snapshot := WorkspaceSnapshot{
			Workspace: c,
			Block:     block,
		}

		for index := range block.Actions {
			action := &block.Actions[index]
			// Loop over script of action
			var newScript = []string{}
			//printObject(snapshot)
			if len(action.Script) > 0 {
				// Loop over script slice and substitute the template
				for _, scriptLine := range action.Script {
					//log.Debug("Templating at snapshot generation, line: " + scriptLine)
					t, err := template.New("script-line").Parse(scriptLine)
					if err != nil {
						c.err = err
						return c
					}

					// Execute the template and save the substituted content to a var
					var substitutedScriptLine bytes.Buffer
					err = t.Execute(&substitutedScriptLine, snapshot)
					if err != nil {
						c.err = err
						return c
					}
					//log.Debug("Substituted the following content: " + substitutedScriptLine.String())

					//scriptLineStr := fmt.Sprintf("%v", scriptLine)
					//scriptStrings = append(scriptStrings, scriptLineStr)
					//log.Debugf("scriptLine as String: %s", scriptLineStr)
					newScript = append(newScript, substitutedScriptLine.String())
				}
			}

			// Update the action script with the substituted one
			action.Script = newScript
		}
	}
	return c
}

func (c *Workspace) DumpEnv() []string {
	envVars := []string{}
	for envVar := range workspace.env {
		value := workspace.env[envVar]
		s := strings.Join([]string{envVar, value}, "=")
		envVars = append(envVars, s)
	}
	return envVars
}

func (c *Workspace) DumpMounts() []string {
	mounts := []string{}
	for mount := range workspace.mounts {
		value := workspace.mounts[mount]
		s := strings.Join([]string{mount, value}, "=")
		mounts = append(mounts, s)
	}
	return mounts
}

func (c *Workspace) bootstrapEnvVars() *Workspace {
	// Prepare env map
	c.env = map[string]string{}

	var _force string
	if force {
		_force = "1"
	} else {
		_force = "0"
	}
	var _in_container string
	if local {
		_in_container = "0"
	} else {
		_in_container = "1"
	}
	c.registerEnvVar("ANSIBLE_DISPLAY_SKIPPED_HOSTS", "False")
	c.registerEnvVar("ANSIBLE_DISPLAY_OK_HOSTS", "True")
	c.registerEnvVar("ANSIBLE_HOST_KEY_CHECKING", "False")
	c.registerEnvVar("ANSIBLE_ACTION_WARNINGS", "False")
	c.registerEnvVar("ANSIBLE_COMMAND_WARNINGS", "False")
	c.registerEnvVar("ANSIBLE_LOCALHOST_WARNING", "False")
	c.registerEnvVar("ANSIBLE_DEPRECATION_WARNINGS", "False")
	c.registerEnvVar("ANSIBLE_ROLES_PATH", "/root/.ansible/roles:/usr/share/ansible/roles:/etc/ansible/roles")
	c.registerEnvVar("ANSIBLE_COLLECTIONS_PATH", "/root/.ansible/collections:/usr/share/ansible/collections:/etc/ansible/collections")
	c.registerEnvVar("ANSIBLE_VERBOSITY", logLevel)
	c.registerEnvVar("ANSIBLE_SSH_PRIVATE_KEY_FILE", filepath.Join(c.ContainerPath, c.Config.SshPrivateKey))
	c.registerEnvVar("ANSIBLE_PRIVATE_KEY_FILE", filepath.Join(c.ContainerPath, c.Config.SshPrivateKey))
	c.registerEnvVar("ANSIBLE_VARS_ENABLED", "polycrate_vars")
	c.registerEnvVar("ANSIBLE_RUN_VARS_PLUGINS", "start")
	c.registerEnvVar("ANSIBLE_VARS_PLUGINS", "/root/.ansible/plugins/vars:/usr/share/ansible/plugins/vars")
	c.registerEnvVar("DEFAULT_VARS_PLUGIN_PATH", "/root/.ansible/plugins/vars:/usr/share/ansible/plugins/vars")
	//c.registerEnvVar("ANSIBLE_CALLBACKS_ENABLED", "timer,profile_tasks,profile_roles")
	c.registerEnvVar("POLYCRATE_CLI_VERSION", version)
	c.registerEnvVar("POLYCRATE_IMAGE_REFERENCE", workspace.Config.Image.Reference)
	c.registerEnvVar("POLYCRATE_IMAGE_VERSION", workspace.Config.Image.Version)
	c.registerEnvVar("POLYCRATE_FORCE", _force)
	c.registerEnvVar("POLYCRATE_VERSION", version)
	c.registerEnvVar("IN_CI", "true")
	c.registerEnvVar("IN_CONTAINER", _in_container)
	c.registerEnvVar("TERM", "xterm-256color")

	if local {
		// Not in container
		c.registerEnvVar("POLYCRATE_WORKSPACE", workspace.LocalPath)
	} else {
		// In container
		c.registerEnvVar("POLYCRATE_WORKSPACE", workspace.ContainerPath)
	}

	for _, envVar := range c.ExtraEnv {
		// Split by =
		p := strings.Split(envVar, "=")

		if len(p) == 2 {
			c.registerEnvVar(p[0], p[1])
		} else {
			c.err = goErrors.New("Illegal value for env var found: " + envVar)
			return c
		}
	}

	return c
}

func (c *Workspace) GetSnapshot() WorkspaceSnapshot {
	log.Debug("Generating snapshot")
	snapshot := WorkspaceSnapshot{
		Workspace: c,
		Action:    c.currentAction,
		Block:     c.currentBlock,
		Workflow:  c.currentWorkflow,
		Step:      c.currentStep,
		Env:       c.env,
		Mounts:    c.mounts,
	}

	c.RegisterSnapshotEnv(snapshot).Flush()
	return snapshot
}

func (c *Workspace) SaveSnapshot() error {
	snapshot := c.GetSnapshot()

	//c.RegisterSnapshotEnv(snapshot).Flush()

	// Marshal the snapshot to yaml
	data, err := yaml.Marshal(snapshot)
	if err != nil {
		return err
	}

	if data != nil {
		snapshotSlug := slugify([]string{workspace.Name, workspace.currentBlock.Name, c.Name})
		snapshotFilename := strings.Join([]string{snapshotSlug, "yml"}, ".")

		f, err := workspace.getTempFile(snapshotFilename)
		if err != nil {
			return err
		}

		// Create a viper config object
		snapshotConfig := viper.New()
		snapshotConfig.SetConfigType("yaml")
		snapshotConfig.ReadConfig(bytes.NewBuffer(data))
		err = snapshotConfig.WriteConfigAs(f.Name())
		if err != nil {
			return err
		}
		log.Debugf("Saved snapshot to %s", f.Name())

		// Closing file descriptor
		// Getting fatal errors on windows WSL2 when accessing
		// the mounted script file from inside the container if
		// the file descriptor is still open
		// Works flawlessly with open file descriptor on M1 Mac though
		// It's probably safer to close the fd anyways
		f.Close()

		c.registerEnvVar("POLYCRATE_WORKSPACE_SNAPSHOT_YAML", f.Name())
		c.registerMount(f.Name(), f.Name())

		return nil
	} else {
		return fmt.Errorf("cannot save snapshot")
	}

}

func (c *Workspace) registerEnvVar(key string, value string) {
	c.env[key] = value
}

func (c *Workspace) registerMount(host string, container string) {
	c.mounts[host] = container
}

func (c *Workspace) registerCurrentBlock(block *Block) {

	c.registerEnvVar("POLYCRATE_BLOCK", block.Name)

	if block.Workdir.exists {
		c.registerEnvVar("POLYCRATE_BLOCK_WORKDIR", block.Workdir.ContainerPath)
	}
	c.currentBlock = block
}
func (c *Workspace) registerCurrentAction(action *Action) {

	c.registerEnvVar("POLYCRATE_ACTION", action.Name)
	c.currentAction = action
}
func (c *Workspace) registerCurrentWorkflow(workflow *Workflow) {

	c.registerEnvVar("POLYCRATE_WORKFLOW", workflow.Name)
	c.currentWorkflow = workflow
}
func (c *Workspace) registerCurrentStep(step *Step) {

	c.registerEnvVar("POLYCRATE_STEP", step.Name)
	c.currentStep = step
}

func (c *Workspace) GetActionFromIndex(name string) *Action {
	if action, ok := c.index.Actions[name]; ok {
		return action
	}
	return nil
}

// func (c *Workspace) getActionByPath(actionPath string) *Action {
// 	// Validate actionPath
// 	s := strings.Split(actionPath, ".")
// 	blockName := s[0]
// 	actionName := s[1]

// 	block := workspace.getBlockByName(blockName)

// 	if block != nil {
// 		action := block.getActionByName(actionName)

// 		if action != nil {
// 			return action
// 		}
// 	}
// 	return nil
// }

func (c *Workspace) GetBlockFromIndex(name string) *Block {
	if block, ok := c.index.Blocks[name]; ok {
		return block
	}
	return nil
}

func (c *Workspace) getBlockByName(blockName string) *Block {
	for i := 0; i < len(c.Blocks); i++ {
		block := &c.Blocks[i]
		if block.Name == blockName {
			return block
		}
	}
	return nil
}

func (c *Workspace) GetWorkflowFromIndex(name string) *Workflow {
	if workflow, ok := c.index.Workflows[name]; ok {
		return workflow
	}
	return nil
}

func (c *Workspace) Create() *Workspace {
	workspacePath := filepath.Join(polycrateWorkspaceDir, c.Name)

	// Check if a workspace with this name already exists
	if localWorkspaceIndex[c.Name] != "" {
		// We found a workspace with that name in the index

		c.err = fmt.Errorf("workspace already exists: %s", localWorkspaceIndex[c.Name])
		return c
	}

	// Check if the directory for this workspace already exists in polycrateWorkspaceDir
	if _, err := os.Stat(workspacePath); !os.IsNotExist(err) {
		// Directory already exists
		c.err = fmt.Errorf("workspace directory already exists: %s", workspacePath)
		return c
	} else {
		err := CreateDir(workspacePath)
		if err != nil {
			c.err = err
			return c
		}

		// Set the new directory as workspace.Path
		c.LocalPath = workspacePath
	}
	// Save a new workspace config to workspace.poly in the given directory
	var n Workspace

	// Update workspace vars with minimum defaults
	n.Name = c.Name

	if c.SyncOptions.Enabled {
		n.SyncOptions.Enabled = true
	}

	if c.SyncOptions.Auto {
		n.SyncOptions.Auto = true
	}

	if c.SyncOptions.Remote.Url != "" {
		n.SyncOptions.Remote.Url = c.SyncOptions.Remote.Url
		n.SyncOptions.Remote.Branch.Name = c.SyncOptions.Remote.Branch.Name
	}

	n.saveWorkspace(c.LocalPath).Flush()

	return c
}

func (c *Workspace) Sync() *Workspace {
	sync.Sync().Flush()
	// Check if a remote is configured
	// if not, fail
	// Check if git has already been initialized
	// if no, run git init and configure remote
	// if yes: Check if the configured remote matches the current remote
	// 				 if no, fail
	// Check if the remote is ahead, even or behind
	// Add all files of the worktree
	// Commit with the current command (as seen from bash)
	// Push
	return c
}

func (w *Workspace) IsBlockInstalled(blockName string, blockVersion string) bool {
	for _, installedBlock := range w.installedBlocks {
		if installedBlock.Name == blockName {
			if installedBlock.Version == blockVersion {
				return true
			}
		}
	}

	return false
}

func (w *Workspace) UpdateBlocks(args []string) error {
	eg := new(errgroup.Group)
	for _, arg := range args {
		arg := arg // https://go.dev/doc/faq#closures_and_goroutines
		eg.Go(func() error {
			blockName, blockVersion, err := registry.resolveArg(arg)
			if err != nil {
				return err
			}

			if w.IsBlockInstalled(blockName, blockVersion) {
				log.WithFields(log.Fields{
					"workspace":       w.Name,
					"block":           blockName,
					"desired_version": blockVersion,
				}).Debugf("Block is already installed")
				return nil
			}

			log.WithFields(log.Fields{
				"workspace":       w.Name,
				"block":           blockName,
				"desired_version": blockVersion,
			}).Debugf("Updating block")

			// Download blocks from registry
			err = w.PullBlock(blockName, blockVersion)
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

// - Accepts 1 arg: the block name/slug as it is in the registry
// - Accepts 1 flag: version; defaults to latest
// - Checks if a block with that name exists already AND has a block dir
// - If the block exists, the command fails with a warning and shows a hint to the update command
// - If no block exists, looks up the name of the block via Wordpress API at polycrate.io
// - If a block is found, gets the list of releases
// - Marks the youngest release as "latest"
// - Downloads the release bundle
// - If download succeeds, creates a block dir for the block
// - unpacks the release bundle to the block dir
// func (c *Workspace) InstallBlocks(args []string) error {
// 	for _, arg := range args {
// 		blockName, blockVersion, err := registry.resolveArg(arg)
// 		if err != nil {
// 			return err
// 		}

func (w *Workspace) PullBlock(blockName string, blockVersion string) error {
	targetDir := filepath.Join(workspace.LocalPath, workspace.Config.BlocksRoot, blockName)

	//log.Debugf("Pulling block %s:%s", blockName, blockVersion)
	log.WithFields(log.Fields{
		"block":   blockName,
		"version": blockVersion,
	}).Debugf("Pulling block")
	err := UnwrapOCIImage(targetDir, blockName, blockVersion)
	if err != nil {
		return err
	}

	// Load Block
	var block Block
	block.Workdir.LocalPath = targetDir
	block.Load().Flush()

	// Register installed block
	w.installedBlocks = append(w.installedBlocks, block)

	return nil
}

func (c *Workspace) PushBlock(blockName string) error {
	// Get the block
	// Check that it has a version
	// Check if release with new version exists in registry
	// Get the latest version from the registry
	// Check if new version > current version
	//

	block := c.GetBlockFromIndex(blockName)
	if block != nil {
		// The block exists
		// Check it has a version
		if block.Version == "" {
			return fmt.Errorf("block has no version")
		}
	} else {
		return fmt.Errorf("block not found in workspace: %s", blockName)
	}

	log.WithFields(log.Fields{
		"block":   block.Name,
		"version": block.Version,
		"path":    block.Workdir.LocalPath,
	}).Debugf("Pushing block")

	err := WrapOCIImage(block.Workdir.LocalPath, block.Name, block.Version, block.Labels)
	if err != nil {
		return err
	}
	return nil
}

func (c *Workspace) UninstallBlocks(args []string) error {
	for _, blockName := range args {
		block := c.GetBlockFromIndex(blockName)
		if block != nil {
			err := block.Uninstall(pruneBlock)
			if err != nil {
				return err
			}
		} else {
			err := fmt.Errorf("Block not found: %s", blockName)
			return err
		}
	}
	return nil
}

// func (c *Workspace) getWorkflowByName(workflowName string) *Workflow {

// 	for i := 0; i < len(c.Workflows); i++ {
// 		workflow := &c.Workflows[i]
// 		if workflow.Name == workflowName {
// 			return workflow
// 		}
// 	}
// 	return nil
// }

func (c *Workspace) print() {
	//fmt.Printf("%#v\n", c)
	printObject(c)
}

func (w *Workspace) ImportInstalledBlocks() *Workspace {
	log.WithFields(log.Fields{
		"workspace": w.Name,
	}).Debugf("Importing installed blocks")

	for _, installedBlock := range w.installedBlocks {

		// Check if Block exists
		existingBlock := w.getBlockByName(installedBlock.Name)

		if existingBlock != nil {
			// Block exists
			log.WithFields(log.Fields{
				"block":     installedBlock.Name,
				"workspace": w.Name,
			}).Debugf("Found existing block. Merging.")

			err := existingBlock.MergeIn(installedBlock)
			if err != nil {
				w.err = err
				return w
			}
		} else {
			w.Blocks = append(w.Blocks, installedBlock)
		}
	}
	return w

}
func (c *Workspace) LoadInstalledBlocks() *Workspace {
	log.WithFields(log.Fields{
		"workspace": c.Name,
	}).Debugf("Loading installed blocks")

	for _, blockPath := range blockPaths {
		var loadedBlock Block
		loadedBlock.Workdir.LocalPath = filepath.Join(blockPath, c.Config.BlocksConfig)
		loadedBlock.Load().Flush()

		// blockConfigFilePath := filepath.Join(blockPath, c.Config.BlocksConfig)

		// blockConfigObject := viper.New()
		// blockConfigObject.SetConfigType("yaml")
		// blockConfigObject.SetConfigFile(blockConfigFilePath)

		// log.WithFields(log.Fields{
		// 	"workspace": c.Name,
		// 	"path":      blockPath,
		// }).Debugf("Loading installed block")
		// if err := blockConfigObject.MergeInConfig(); err != nil {
		// 	c.err = err
		// 	return c
		// }

		// if err := blockConfigObject.UnmarshalExact(&loadedBlock); err != nil {
		// 	c.err = err
		// 	return c
		// }
		// if err := loadedBlock.validate(); err != nil {
		// 	c.err = err
		// 	return c
		// }
		// log.WithFields(log.Fields{
		// 	"block":     loadedBlock.Name,
		// 	"workspace": c.Name,
		// }).Debugf("Loaded block")

		// // Set Block vars
		// loadedBlock.Workdir.LocalPath = blockPath
		// loadedBlock.Workdir.ContainerPath = filepath.Join(c.ContainerPath, c.Config.BlocksRoot, loadedBlock.Name)

		// if local {
		// 	loadedBlock.Workdir.Path = loadedBlock.Workdir.LocalPath
		// } else {
		// 	loadedBlock.Workdir.Path = loadedBlock.Workdir.ContainerPath
		// }

		// // Add block to installedBlocks
		// c.installedBlocks = append(installedBlocks, loadedBlock)
	}
	return c

}

func (w *Workspace) DiscoverInstalledBlocks() *Workspace {
	blocksDir := filepath.Join(workspace.LocalPath, workspace.Config.BlocksRoot)

	if _, err := os.Stat(blocksDir); !os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"workspace": w.Name,
			"path":      blocksDir,
		}).Debugf("Discovering blocks")

		// This function adds all valid Blocks to the list of
		err := filepath.WalkDir(blocksDir, w.WalkBlocksDir)
		if err != nil {
			w.err = err
		}
	} else {
		log.WithFields(log.Fields{
			"workspace": w.Name,
			"path":      blocksDir,
		}).Debugf("Skipping block discovery. Blocks directory not found")
	}

	return w
}

func (w *Workspace) WalkBlocksDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() {
		fileinfo, _ := d.Info()

		if fileinfo.Name() == w.Config.BlocksConfig {
			blockConfigFileDir := filepath.Dir(path)
			log.WithFields(log.Fields{
				"path": blockConfigFileDir,
			}).Debugf("Block detected")

			var block Block
			block.Workdir.LocalPath = blockConfigFileDir
			block.Load().Flush()

			// Add block to installedBlocks
			w.installedBlocks = append(w.installedBlocks, block)
		}
	}
	return nil
}

func (c *Workspace) PullDependencies() *Workspace {
	// Loop through dependencies
	// Resolve each dep into block/version
	// Check if block already exists
	// If yes, check if version matches
	// If version matches, continue
	// If version doesn't match, update
	// If no, install

	log.WithFields(log.Fields{
		"workspace": c.Name,
	}).Debugf("Resolving workspace dependencies")

	err := c.UpdateBlocks(c.Dependencies)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

func (c *Workspace) getTempFile(filename string) (*os.File, error) {
	fp := filepath.Join(workspace.runtimeDir, filename)
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	return f, nil
}
