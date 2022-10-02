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
	goErrors "errors"
	"fmt"
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
	Path            string      `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty"`
	SyncOptions     SyncOptions `yaml:"sync,omitempty" mapstructure:"sync,omitempty" json:"sync,omitempty"`
	LocalPath       string      `yaml:"localpath,omitempty" mapstructure:"localpath,omitempty" json:"localpath,omitempty"`
	ContainerPath   string      `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
	//overrides       []string
	ExtraEnv    []string `yaml:"extraenv,omitempty" mapstructure:"extraenv,omitempty" json:"extraenv,omitempty"`
	ExtraMounts []string `yaml:"extramounts,omitempty" mapstructure:"extramounts,omitempty" json:"extramounts,omitempty"`
	containerID string
	loaded      bool
	Version     string `yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`
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

func (c *Workspace) loadWorkspaceConfig() *Workspace {
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
	workspaceConfigFilePath := filepath.Join(c.LocalPath, workspace.Config.WorkspaceConfig)

	if _, err := os.Stat(workspaceConfigFilePath); os.IsNotExist(err) {
		// The config file does not exist
		// We try to look in the list of workspaces in $HOME/.polycrate/workspaces

		// Assuming the "path" given is actually the name of a workspace
		workspaceName := c.LocalPath

		log.WithFields(log.Fields{
			"path":      workspaceConfigFilePath,
			"workspace": workspaceName,
		}).Debugf("Workspace config not found. Looking in the local workspace index", workspaceName)

		// Check if workspaceName exists in the local workspace index (see discoverWorkspaces())
		if localWorkspaceIndex[workspaceName] != "" {
			// We found a workspace with that name in the index
			path := localWorkspaceIndex[workspaceName]
			log.WithFields(log.Fields{
				"workspace": workspaceName,
				"path":      path,
			}).Debugf("Found workspace in the local workspace index")

			// Update the workspace config file path with the local workspace path from the index
			c.LocalPath = path
			workspaceConfigFilePath = filepath.Join(c.LocalPath, workspace.Config.WorkspaceConfig)
		} else {
			c.err = fmt.Errorf("couldn't find workspace config at %s", workspaceConfigFilePath)
			return c
		}
	}

	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(workspaceConfigFilePath)

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		// log.Warnf("Couldn't find workspace config at %s: %s", workspaceConfigFilePath, err.Error())
		// log.Warnf("Creating ad-hoc workspace with name %s", filepath.Base(cwd))

		// workspaceConfig.SetDefault("name", filepath.Base(cwd))
		// workspaceConfig.SetDefault("description", "Ad-hoc Workspace in "+cwd)
		c.err = err
		return c
	}

	if err := workspaceConfig.Unmarshal(&c); err != nil {
		c.err = err
		return c
	}

	if err := c.validate(); err != nil {
		c.err = err
		return c
	}

	// set runtime dir
	c.runtimeDir = filepath.Join(polycrateRuntimeDir, c.Name)

	return c
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

func (c *Workspace) load() *Workspace {
	if c.loaded {
		return c
	}

	if local {
		c.Path = c.LocalPath
	} else {
		c.Path = c.ContainerPath
	}

	log.WithFields(log.Fields{
		"path": workspace.LocalPath,
	}).Debugf("Loading workspace")

	// Load Workspace config (e.g. workspace.poly)
	c.loadWorkspaceConfig().Flush()

	// Cleanup workspace runtime
	c.cleanup().Flush()

	// Bootstrap the workspace index
	c.bootstrapIndex().Flush()

	// Find all blocks in the workspace
	c.DiscoverInstalledBlocks().Flush()

	// Load installed blocks
	c.LoadInstalledBlocks().Flush()

	// Pull dependencies
	c.PullDependencies().Flush()

	// Load all discovered blocks in the workspace
	c.ImportInstalledBlocks().Flush()

	// Resolve block dependencies
	c.ResolveBlockDependencies().Flush()

	// Update workflow and step addresses
	c.resolveWorkflows().Flush()

	// Bootstrap env vars
	c.bootstrapEnvVars().Flush()

	// Bootstrap container mounts
	c.bootstrapMounts().Flush()

	// Template action scripts
	c.templateActionScripts().Flush()

	log.WithFields(log.Fields{
		"workspace": c.Name,
		"blocks":    len(workspace.Blocks),
		"workflows": len(workspace.Workflows),
	}).Debugf("Workspace ready")

	// Mark workspace as loaded
	c.loaded = true

	// Load sync
	sync.Load().Flush()
	if sync.Options.Enabled && sync.Options.Auto {
		sync.Sync().Flush()
	}

	return c
}

func (c *Workspace) cleanup() *Workspace {
	// Create runtime dir
	log.WithFields(log.Fields{
		"workspace": c.Name,
		"path":      c.runtimeDir,
	}).Debugf("Creating runtime directory")

	err := os.MkdirAll(c.runtimeDir, os.ModePerm)
	if err != nil {
		c.err = err
		return c
	}

	// Purge all contents of runtime dir
	dir, err := ioutil.ReadDir(c.runtimeDir)
	if err != nil {
		c.err = err
		return c
	}
	for _, d := range dir {
		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      c.runtimeDir,
			"file":      d.Name(),
		}).Debugf("Cleaning runtime directory")
		err := os.RemoveAll(filepath.Join([]string{c.runtimeDir, d.Name()}...))
		if err != nil {
			c.err = err
			return c
		}
	}
	return c
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
	}).Debugf("Resolving block dependencies")

	// Iterate over all Blocks in the Workspace
	// Until nothing is "missing" anymore
	for missing != 0 {
		for i := 0; i < len(w.Blocks); i++ {
			loadedBlock := &w.Blocks[i]

			// Block has not been resolved yet
			if !loadedBlock.resolved {
				log.WithFields(log.Fields{
					"block":     loadedBlock.Name,
					"workspace": w.Name,
					"missing":   missing,
				}).Debugf("Resolving block %d/%d", i+1, len(w.Blocks))

				// Check if a "from:" stanza is given and not empty
				// This means that the loadedBlock should inherit from another Block
				if loadedBlock.From != "" {
					// a "from:" stanza is given
					log.WithFields(log.Fields{
						"block":      loadedBlock.Name,
						"dependency": loadedBlock.From,
						"workspace":  w.Name,
						"missing":    missing,
					}).Debugf("Dependency detected")

					// Try to load the referenced Block
					dependency := w.getBlockByName(loadedBlock.From)

					if dependency == nil {
						log.WithFields(log.Fields{
							"block":      loadedBlock.Name,
							"dependency": loadedBlock.From,
							"workspace":  w.Name,
							"missing":    missing,
						}).Errorf("Dependency not found in the workspace")
						// There's no Block to load from

						// Let's check if there's one in the registry
						if _, err := registry.GetBlock(loadedBlock.From); err == nil {
							w.err = goErrors.New("Block '" + loadedBlock.From + "' not found in the Workspace. Please check the 'from' stanza of Block " + loadedBlock.Name + " or run 'polycrate block install " + loadedBlock.From + "'")
							return w
						} else {
							w.err = goErrors.New("Block '" + loadedBlock.From + "' not found in the Workspace. Please check the 'from' stanza of Block " + loadedBlock.Name)
							return w
						}
					}

					log.WithFields(log.Fields{
						"block":      loadedBlock.Name,
						"dependency": loadedBlock.From,
						"workspace":  w.Name,
						"missing":    missing,
					}).Debugf("Dependency loaded")

					dep := *dependency

					// Check if the dependency Block has already been resolved
					// If not, re-queue the loaded Block so it can be resolved in another iteration
					if !dep.resolved {
						// Needed Block from 'from' stanza is not yet resolved
						log.WithFields(log.Fields{
							"block":               loadedBlock.Name,
							"dependency":          loadedBlock.From,
							"workspace":           w.Name,
							"missing":             missing,
							"dependency_resolved": dep.resolved,
						}).Debugf("Postponing block")
						loadedBlock.resolved = false
						continue
					}

					// Merge the dependency Block into the loaded Block
					// We do NOT OVERWRITE existing values in the loaded Block because we must assume
					// That this is configuration that has been explicitly set by the user
					// The merge works like "loading defaults" for the loaded Block
					err := loadedBlock.MergeIn(dep)
					if err != nil {
						w.err = err
						return w
					}

					// Handle Workdir
					loadedBlock.Workdir.LocalPath = dep.Workdir.LocalPath
					loadedBlock.Workdir.ContainerPath = dep.Workdir.ContainerPath

					// Handle Actions
					// Iterate over the dependency Block's Actions
					// and check if the Action exists in the loaded Block
					// If not, add it
					// If yes, merge it
					// for _, blockAction := range dep.Actions {
					// 	//blockAction := dep.getActionByName(action.Name)
					// 	log.Debugf("Analyzing action '%s' (%p) of dependency '%s' (%p)", blockAction.Name, &blockAction, dep.Name, &dep)
					// 	log.Debugf("Dependency action before merge (%p):", &blockAction)
					// 	printObject(blockAction)

					// 	existingAction := loadedBlock.getActionByName(blockAction.Name)

					// 	if existingAction != nil {
					// 		log.Debugf("Existing action before merge (%p):", existingAction)
					// 		printObject(existingAction)
					// 		log.Debugf("Found existing action '%s' in the block. Merging with loaded action", existingAction.Name)
					// 		// An Action with the same name exists
					// 		// We merge!
					// 		err := existingAction.MergeIn(blockAction)
					// 		if err != nil {
					// 			c.err = err
					// 			return c
					// 		}
					// 		existingAction.address = strings.Join([]string{loadedBlock.Name, existingAction.Name}, ".")
					// 		existingAction.Block = loadedBlock.Name
					// 		log.Debugf("Existing action after merge (%p):", existingAction)
					// 		printObject(existingAction)

					// 	} else {
					// 		log.Debugf("No user-provided action with that name found. Adding '%s' to the workspace.", blockAction.Name)
					// 		blockAction.address = strings.Join([]string{loadedBlock.Name, blockAction.Name}, ".")
					// 		blockAction.Block = loadedBlock.Name
					// 		loadedBlock.Actions = append(loadedBlock.Actions, blockAction)
					// 		log.Debugf("Dependency action without merge (%p):", &blockAction)
					// 		printObject(blockAction)
					// 	}
					// 	log.Debugf("Dependency action after merge (%p):", &blockAction)
					// 	printObject(blockAction)

					// }
					// opts := conjungo.NewOptions()
					// opts.Overwrite = false // do not overwrite existing values in workspaceConfig
					// if err := conjungo.Merge(loadedBlock, dependency, opts); err != nil {
					// 	return err
					// }
					log.WithFields(log.Fields{
						"block":      loadedBlock.Name,
						"dependency": loadedBlock.From,
						"workspace":  w.Name,
						"missing":    missing,
					}).Debugf("Dependency resolved")

				} else {
					log.WithFields(log.Fields{
						"block":     loadedBlock.Name,
						"workspace": w.Name,
						"missing":   missing,
					}).Debugf("Block has no dependencies")

				}

				// Register the Block to the Index
				loadedBlock.address = loadedBlock.Name
				w.registerBlock(loadedBlock)

				// Update Artifacts, Kubeconfig & Inventory
				err := loadedBlock.LoadArtifacts()
				if err != nil {
					w.err = err
					return w
				}
				loadedBlock.LoadInventory()
				loadedBlock.LoadKubeconfig()

				// Update Action addresses for all blocks not covered by dependencies
				if len(loadedBlock.Actions) > 0 {
					log.WithFields(log.Fields{
						"block":     loadedBlock.Name,
						"workspace": w.Name,
						"missing":   missing,
					}).Debugf("Updating action addresses")

					for _, action := range loadedBlock.Actions {

						existingAction := loadedBlock.getActionByName(action.Name)

						actionAddress := strings.Join([]string{loadedBlock.Name, existingAction.Name}, ".")
						if existingAction.address != actionAddress {
							existingAction.address = actionAddress
							log.WithFields(log.Fields{
								"block":     loadedBlock.Name,
								"action":    existingAction.Name,
								"workspace": w.Name,
								"address":   actionAddress,
								"missing":   missing,
							}).Debugf("Updated action address")
						}

						if existingAction.Block != loadedBlock.Name {
							existingAction.Block = loadedBlock.Name
							log.WithFields(log.Fields{
								"block":     loadedBlock.Name,
								"action":    existingAction.Name,
								"workspace": w.Name,
								"address":   actionAddress,
								"missing":   missing,
							}).Debugf("Updated action block")
						}

						// Register the Action to the Index
						w.registerAction(existingAction)
					}
				}

				loadedBlock.resolved = true
				missing--

				log.WithFields(log.Fields{
					"block":     loadedBlock.Name,
					"workspace": w.Name,
					"missing":   missing,
				}).Debugf("Block resolved")

				// log.WithFields(log.Fields{
				// 	"workspace": c.Name,
				// 	"missing":   missing,
				// }).Debugf("%d blocks left", missing)
			}

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

func (c *Workspace) bootstrapIndex() *Workspace {
	c.index = &WorkspaceIndex{}
	c.index.Actions = make(map[string]*Action)
	c.index.Blocks = make(map[string]*Block)
	c.index.Workflows = make(map[string]*Workflow)
	c.index.Steps = make(map[string]*Step)
	return c
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
	for _, installedBlock := range installedBlocks {
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

			// create block runtime dir
			// blockRuntimeDir := filepath.Join(c.runtimeDir, blockName)
			// log.WithFields(log.Fields{
			// 	"block":       blockName,
			// 	"workspace":   c.Name,
			// 	"runtime-dir": blockRuntimeDir,
			// }).Debugf("Creating block runtime dir")
			// err = os.MkdirAll(blockRuntimeDir, os.ModePerm)
			// if err != nil {
			// 	return err
			// }

			// Download blocks from registry
			err = w.PullBlock(blockName, blockVersion)
			if err != nil {
				return err
			}
			return nil
		})

		// NOT NEEDED ANYMORE

		// block := c.getBlockByName(blockName)
		// if block != nil {
		// 	// Check if block already has the desired version
		// 	if block.Version == blockVersion {
		// 		log.WithFields(log.Fields{
		// 			"workspace":       c.Name,
		// 			"block":           block.Name,
		// 			"path":            block.Workdir.LocalPath,
		// 			"current_version": block.Version,
		// 			"desired_version": blockVersion,
		// 		}).Debugf("Block is already installed")
		// 		return nil
		// 	}

		// 	log.WithFields(log.Fields{
		// 		"workspace":       c.Name,
		// 		"block":           block.Name,
		// 		"path":            block.Workdir.LocalPath,
		// 		"current_version": block.Version,
		// 		"desired_version": blockVersion,
		// 	}).Infof("Updating block")

		// 	// Download blocks from registry
		// 	err = c.PullBlock(blockName, blockVersion)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	return nil

		// 	// NOT NEEDED ANYMORE

		// 	// Search block in registry
		// 	registryBlock, err := registry.GetBlock(blockName)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	// Check if release exists
		// 	registryRelease, err := registryBlock.GetRelease(blockVersion)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	// Check again if we're already on the desired version
		// 	// At this point "latest" has been resolved to a specific release version
		// 	if block.Version == registryRelease.Version {
		// 		log.WithFields(log.Fields{
		// 			"workspace":       c.Name,
		// 			"block":           block.Name,
		// 			"path":            block.Workdir.LocalPath,
		// 			"current_version": block.Version,
		// 			"desired_version": registryRelease.Version,
		// 		}).Infof("Block already has desired version")
		// 		return nil
		// 	}

		// 	// Uninstall block
		// 	// pruneBlock constains a boolean triggered by --prune
		// 	err = block.Uninstall(pruneBlock)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	// Now install the wanted version
		// 	registryBlockDir := filepath.Join(workspace.LocalPath, workspace.Config.BlocksRoot, blockName)
		// 	registryBlockVersion := blockVersion

		// 	err = registryBlock.Install(registryBlockDir, registryBlockVersion)
		// 	if err != nil {
		// 		return err
		// 	}

		// } else {
		// 	if pull {
		// 		// If we pull the container image automatically
		// 		// We certainly want to install dependencies automatically

		// 		// Search block in registry
		// 		registryBlock, err := registry.GetBlock(blockName)
		// 		if err != nil {
		// 			return err
		// 		}

		// 		// Check if release exists
		// 		_, err = registryBlock.GetRelease(blockVersion)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		// Now install the wanted version
		// 		registryBlockDir := filepath.Join(workspace.LocalPath, workspace.Config.BlocksRoot, blockName)
		// 		registryBlockVersion := blockVersion

		// 		err = registryBlock.Install(registryBlockDir, registryBlockVersion)
		// 		if err != nil {
		// 			return err
		// 		}

		// 	} else {
		// 		err := fmt.Errorf("Block not found: %s. Run 'polycrate block install %s' to install it.", blockName, blockName)
		// 		return err
		// 	}
		// }
	}
	//return nil

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
func (c *Workspace) InstallBlocks(args []string) error {
	for _, arg := range args {
		blockName, blockVersion, err := registry.resolveArg(arg)
		if err != nil {
			return err
		}

		block := c.getBlockByName(blockName)
		if block != nil {
			// A block exists already
			// Let's check if it has a workdir
			if block.Workdir.LocalPath != "" {
				// The block has a workdir
				// Let's check if it exists
				if _, err := os.Stat(block.Workdir.LocalPath); !os.IsNotExist(err) {
					// The workdir exists
					// We're done here
					log.WithFields(log.Fields{
						"workspace": c.Name,
						"block":     block.Name,
						"path":      block.Workdir.LocalPath,
					}).Debugf("Block is already installed. Use 'polycrate block update %s' to update it", block.Name)
				} else {
					// The workdir does not exist
					// We can download the block
					download = true
				}
			} else {
				download = true
			}

		} else {
			download = true
		}

		// Search block in registry
		if download {
			registryBlock, err := registry.GetBlock(blockName)
			if err != nil {
				log.WithFields(log.Fields{
					"workspace": c.Name,
					"block":     blockName,
				}).Debug(err)
				return err
			}

			registryBlockDir := filepath.Join(workspace.LocalPath, workspace.Config.BlocksRoot, blockName)
			registryBlockVersion := blockVersion

			err = registryBlock.Install(registryBlockDir, registryBlockVersion)
			if err != nil {
				return err
			}

		}

	}
	return nil
}

func (c *Workspace) PullBlock(blockName string, blockVersion string) error {
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

	// // NOT NEEDED ANYMORE

	// // Search block in registry
	// registryBlock, err := config.Registry.GetBlock(blockName)
	// if err != nil {
	// 	// Block not found should lead to the creation of the block
	// 	// WIP
	// 	log.WithFields(log.Fields{
	// 		"workspace": c.Name,
	// 		"block":     blockName,
	// 	}).Debug(err)

	// 	// Create block
	// 	registryBlock, err = registry.AddBlock(blockName)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// printObject(registryBlock)

	// _, err = registryBlock.GetRelease(block.Version)
	// if err == nil {
	// 	err := fmt.Errorf("release with version %s of block %s already exists in the registry", block.Version, block.Name)
	// 	// Release already exists
	// 	log.WithFields(log.Fields{
	// 		"workspace": c.Name,
	// 		"block":     block.Name,
	// 		"version":   block.Version,
	// 	}).Debug(err)
	// 	return err
	// }

	// // No release exists with that version
	// // Now let's see if there's a latest release
	// latestRelease, err := registryBlock.GetRelease("latest")
	// if err != nil {
	// 	// Latest release does not exist
	// 	// This means that the block has just been created or there haven't been any releases yet
	// 	log.WithFields(log.Fields{
	// 		"workspace": c.Name,
	// 		"block":     block.Name,
	// 		"version":   block.Version,
	// 	}).Debug(err)
	// }

	// // Compare version if there's a release
	// if latestRelease != nil {
	// 	// Check if our version is bigger than the current version
	// 	blockSemVer, err := semver.NewVersion(block.Version)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.WithFields(log.Fields{
	// 		"workspace": c.Name,
	// 		"block":     block.Name,
	// 		"version":   block.Version,
	// 	}).Debugf("Current block version: %s", blockSemVer.String())

	// 	latestReleaseSemVer, err := semver.NewVersion(latestRelease.Version)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.WithFields(log.Fields{
	// 		"workspace": c.Name,
	// 		"block":     block.Name,
	// 		"version":   block.Version,
	// 	}).Debugf("Latest registry release version: %s", latestReleaseSemVer.String())

	// 	// Compare versions
	// 	comparison := fmt.Sprintf("> %s", latestReleaseSemVer)
	// 	constraint, err := semver.NewConstraint(comparison)
	// 	if err != nil {
	// 		log.WithFields(log.Fields{
	// 			"workspace": c.Name,
	// 			"block":     block.Name,
	// 			"version":   block.Version,
	// 		}).Debugf(err.Error())
	// 		return err
	// 	}
	// 	isGreater := constraint.Check(blockSemVer)

	// 	if !isGreater {

	// 		err := fmt.Errorf("block version (%s) is lower than registry version (%s)", blockSemVer.String(), latestReleaseSemVer.String())
	// 		return err
	// 	}
	// }

	// // Zip the block workdir
	// log.WithFields(log.Fields{
	// 	"workspace": c.Name,
	// 	"block":     block.Name,
	// 	"version":   block.Version,
	// }).Debugf("Creating release bundle")

	// blockFileName := slugify([]string{block.Name, block.Version})
	// zipFilePath, err := createZipFile(block.Workdir.LocalPath, blockFileName)
	// if err != nil {
	// 	return err
	// }

	// log.WithFields(log.Fields{
	// 	"workspace": c.Name,
	// 	"block":     block.Name,
	// 	"version":   block.Version,
	// 	"path":      zipFilePath,
	// }).Debugf("Saved release bundle to %s", zipFilePath)

	// release, err := registryBlock.AddRelease(block.Version, zipFilePath, blockFileName)
	// if err != nil {
	// 	return err
	// }

	// printObject(release)

	// log.WithFields(log.Fields{
	// 	"workspace": c.Name,
	// 	"block":     block.Name,
	// 	"version":   block.Version,
	// 	//"id":        release.Id,
	// }).Debugf("Successfully pushed release to registry")

	// // Create / Upload attachment, save id (registry.CreateAttachment(path string) string)
	// // Create Release, link to attachment and Block ID

	// return nil
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

	for _, installedBlock := range installedBlocks {
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

		// var loadedBlock Block
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

			// // Handle Actions
			// // Iterate over the loaded Block's Actions
			// // and check if it exists in the existing Block
			// // If not, add it
			// log.Debugf("Merging actions")
			// for _, loadedAction := range loadedBlock.Actions {
			// 	existingAction := existingBlock.getActionByName(loadedAction.Name)

			// 	if existingAction != nil {
			// 		log.Debugf("Analyzing Action '%s' (%p) of block '%s' (%p)", existingAction.Name, existingAction, loadedBlock.Name, &loadedBlock)

			// 		// An Action with the same name exists
			// 		// We merge!
			// 		log.Debugf("Existing Action (%p) before merge:", &existingAction)
			// 		printObject(existingAction)
			// 		err := existingAction.MergeIn(loadedAction)
			// 		if err != nil {
			// 			c.err = err
			// 			return c
			// 		}
			// 		log.Debugf("Existing Action (%p) after merge:", &existingAction)
			// 		printObject(existingAction)
			// 	} else {
			// 		log.Debugf("No existing Action found. Adding '%s'", loadedAction.Name)
			// 		existingBlock.Actions = append(existingBlock.Actions, loadedAction)
			// 	}
			// }

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
		blockConfigFilePath := filepath.Join(blockPath, c.Config.BlocksConfig)

		blockConfigObject := viper.New()
		blockConfigObject.SetConfigType("yaml")
		blockConfigObject.SetConfigFile(blockConfigFilePath)

		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      blockPath,
		}).Debugf("Loading installed block")
		if err := blockConfigObject.MergeInConfig(); err != nil {
			c.err = err
			return c
		}

		var loadedBlock Block
		if err := blockConfigObject.UnmarshalExact(&loadedBlock); err != nil {
			c.err = err
			return c
		}
		if err := loadedBlock.validate(); err != nil {
			c.err = err
			return c
		}
		log.WithFields(log.Fields{
			"block":     loadedBlock.Name,
			"workspace": c.Name,
		}).Debugf("Loaded block")

		// Set Block vars
		loadedBlock.Workdir.LocalPath = blockPath
		loadedBlock.Workdir.ContainerPath = filepath.Join(c.ContainerPath, c.Config.BlocksRoot, loadedBlock.Name)

		if local {
			loadedBlock.Workdir.Path = loadedBlock.Workdir.LocalPath
		} else {
			loadedBlock.Workdir.Path = loadedBlock.Workdir.ContainerPath
		}

		// Add block to installedBlocks
		installedBlocks = append(installedBlocks, loadedBlock)
	}
	return c

}

func (c *Workspace) DiscoverInstalledBlocks() *Workspace {
	blocksDir := filepath.Join(workspace.LocalPath, workspace.Config.BlocksRoot)

	if _, err := os.Stat(blocksDir); !os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      blocksDir,
		}).Debugf("Discovering blocks")

		// This function adds all valid Blocks to the list of
		err := filepath.WalkDir(blocksDir, walkBlocksDir)
		if err != nil {
			c.err = err
		}
	} else {
		log.WithFields(log.Fields{
			"workspace": c.Name,
			"path":      blocksDir,
		}).Debugf("Skipping block discovery. Blocks directory not found")
	}

	return c
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
