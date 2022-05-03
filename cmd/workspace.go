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
	"context"
	goErrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage the workspace",
	Long:  `Manage the workspace`,
	Run: func(cmd *cobra.Command, args []string) {
		workspace.load()
		if workspace.Flush() != nil {
			log.Fatal(workspace.Flush)
		}

		workspace.Inspect()
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}

type ImageConfig struct {
	Reference string `mapstructure:"reference" json:"reference" validate:"required"`
	Version   string `mapstructure:"version" json:"version" validate:"required"`
}

type Metadata struct {
	Name        string            `mapstructure:"name" json:"name" validate:"required,metadata_name"`
	Description string            `mapstructure:"description" json:"description"`
	Labels      map[string]string `mapstructure:"labels" json:"labels"`
	Alias       []string          `mapstructure:"alias" json:"alias"`
}

type WorkspaceConfig struct {
	Image           ImageConfig            `mapstructure:"image" json:"image" validate:"required"`
	BlocksRoot      string                 `mapstructure:"blocksroot" json:"blocksroot" validate:"required"`
	BlocksConfig    string                 `mapstructure:"blocksconfig" json:"blocksconfig" validate:"required"`
	WorkspaceConfig string                 `mapstructure:"workspaceconfig" json:"workspaceconfig" validate:"required"`
	WorkflowsRoot   string                 `mapstructure:"workflowsroot" json:"workflowsroot" validate:"required"`
	ArtifactsRoot   string                 `mapstructure:"artifactsroot" json:"artifactsroot" validate:"required"`
	ContainerRoot   string                 `mapstructure:"containerroot" json:"containerroot" validate:"required"`
	SshPrivateKey   string                 `mapstructure:"sshprivatekey" json:"sshprivatekey" validate:"required"`
	SshPublicKey    string                 `mapstructure:"sshpublickey" json:"sshpublickey" validate:"required"`
	RemoteRoot      string                 `mapstructure:"remoteroot" json:"remoteroot" validate:"required"`
	Dockerfile      string                 `mapstructure:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	Globals         map[string]interface{} `mapstructure:"globals,remain" json:"globals"`
}

type Workspace struct {
	//Metadata        Metadata          `mapstructure:"metadata,squash" json:"metadata" validate:"required"`
	// alphanum,unique,startsnotwith='/',startsnotwith='-',startsnotwith='.',excludesall=!@#?
	Name            string            `yaml:"name,omitempty" mapstructure:"name,omitempty" json:"name,omitempty" validate:"required,metadata_name"`
	Description     string            `yaml:"description,omitempty" mapstructure:"description,omitempty" json:"description,omitempty"`
	Labels          map[string]string `yaml:"labels,omitempty" mapstructure:"labels,omitempty" json:"labels,omitempty"`
	Alias           []string          `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	Config          WorkspaceConfig   `yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`
	Blocks          []Block           `yaml:"blocks,omitempty" mapstructure:"blocks,omitempty" json:"blocks,omitempty" validate:"dive,required"`
	Workflows       []Workflow        `yaml:"workflows,omitempty" mapstructure:"workflows,omitempty" json:"workflows,omitempty"`
	currentBlock    *Block
	currentAction   *Action
	currentWorkflow *Workflow
	currentStep     *Step
	index           *WorkspaceIndex
	env             map[string]string
	mounts          map[string]string
	err             error
	Path            string
	//overrides       []string
	ExtraEnv      []string `yaml:"extraenv,omitempty" mapstructure:"extraenv,omitempty" json:"extraenv,omitempty"`
	ExtraMounts   []string `yaml:"extramounts,omitempty" mapstructure:"extramounts,omitempty" json:"extramounts,omitempty"`
	ContainerPath string   `yaml:"containerpath,omitempty" mapstructure:"containerpath,omitempty" json:"containerpath,omitempty"`
	containerID   string
}

type WorkspaceIndex struct {
	Actions   map[string]*Action   `mapstructure:"actions" json:"actions"`
	Steps     map[string]*Step     `mapstructure:"steps" json:"steps"`
	Blocks    map[string]*Block    `mapstructure:"blocks" json:"blocks"`
	Workflows map[string]*Workflow `mapstructure:"workflows" json:"workflows"`
}

type WorkspaceSnapshot struct {
	Workspace *Workspace
	Action    *Action
	Block     *Block
	Workflow  *Workflow
	Step      *Step
	Env       map[string]string
	Mounts    map[string]string
}

func (c *Workspace) Snapshot() {
	snapshot := WorkspaceSnapshot{
		Workspace: c,
		Action:    c.currentAction,
		Block:     c.currentBlock,
		Workflow:  c.currentWorkflow,
		Step:      c.currentStep,
		Env:       c.env,
		Mounts:    c.mounts,
	}
	printObject(snapshot)
	//convertToEnv(&snapshot)
}

func (c *Workspace) Inspect() {
	printObject(c)
}

func (c *Workspace) Flush() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *Workspace) RunAction(address string) error {
	// Find action in index and report
	action := c.LookupAction(address)

	if action != nil {
		workspace.registerCurrentAction(action)
		workspace.registerCurrentBlock(&action.Block)

		if snapshot {
			c.Snapshot()
		} else {
			err := action.Run()
			if err != nil {
				return err
			}
		}
	} else {
		return goErrors.New("cannot find Action with address " + address)
	}
	return nil
}

func (c *Workspace) RunWorkflow(name string) error {
	// Find action in index and report
	workflow := c.LookupWorkflow(name)

	if workflow != nil {
		workspace.registerCurrentWorkflow(workflow)

		if snapshot {
			c.Snapshot()
		} else {
			err := workflow.Run()
			if err != nil {
				return err
			}
		}
	} else {
		return goErrors.New("cannot find Workflow " + name)
	}
	return nil
}

func (c *Workspace) RunStep(name string) error {
	// Find action in index and report
	step := c.LookupStep(name)

	if step != nil {
		if snapshot {
			c.Snapshot()
		} else {
			err := step.Run()
			if err != nil {
				return err
			}
		}
	} else {
		return goErrors.New("cannot find step " + name)
	}
	return nil
}

func (c *Workspace) pullContainerImage(image string) error {
	ctx := context.Background()
	cli, err := getDockerCLI()
	if err != nil {
		return err
	}

	_, err = cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	log.Debugf("Successfully pulled image %s", image)
	return nil
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

func (c *Workspace) loadWorkspaceConfig() error {
	workspaceConfig.SetConfigType("yaml")
	workspaceConfig.SetConfigFile(filepath.Join(workspace.Path, workspace.Config.WorkspaceConfig))

	err := workspaceConfig.MergeInConfig()
	if err != nil {
		log.Warn("Error loading workspaceConfig: " + err.Error())

		workspaceConfig.SetDefault("name", filepath.Base(cwd))
		workspaceConfig.SetDefault("description", "Ad-hoc Workspace in "+cwd)
	}

	return nil
}

func (c *Workspace) unmarshalWorkspaceConfig() error {
	log.Debugf("Unmarshalling loaded config to internal struct")
	err := workspaceConfig.Unmarshal(&c)

	if err != nil {
		return err
	}

	err = c.Validate()
	return err
}

func (c *Workspace) load() {
	log.Infof("Loading Workspace")

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

	// Load Workspace config (e.g. workspace.poly)
	if err := c.loadWorkspaceConfig(); err != nil {
		log.Fatal(err)
	}

	// Unmarshal + Validate Workspace config
	if err := c.unmarshalWorkspaceConfig(); err != nil {
		log.Fatal(err)
	}

	// Bootstrap Index
	c.bootstrapIndex()

	// Find all Blocks in the Workspace
	if err := c.discoverBlocks(); err != nil {
		log.Fatal(err)
	}

	// Load all discovered Blocks in the Workspace
	if err := c.loadBlockConfigs(); err != nil {
		log.Fatal(err)
	}

	// Resolve Block dependencies
	if err := c.resolveBlockDependencies(); err != nil {
		log.Fatal(err)
	}

	// Update workflow and step addresses
	for i := 0; i < len(c.Workflows); i++ {
		loadedWorkflow := &c.Workflows[i]

		loadedWorkflow.address = loadedWorkflow.Name
		// Register the workflow to the Index
		c.registerWorkflow(loadedWorkflow)

		// Loop over the steps
		for _, step := range loadedWorkflow.Steps {
			loadedStep := loadedWorkflow.getStepByName(step.Name)

			// Set step address
			loadedStep.address = strings.Join([]string{loadedWorkflow.Name, loadedStep.Name}, ".")

			workflowCopy := *loadedWorkflow
			loadedStep.Workflow = workflowCopy

			// Remove nested steps
			loadedStep.Workflow.Steps = nil

			log.Debugf("Validating step %s", loadedStep.Name)
			if err := loadedStep.Validate(); err != nil {
				log.Fatal(err)
			}

			log.Debugf("Registered step %s of workflow %s", loadedStep.Name, loadedWorkflow.Name)
			c.registerStep(loadedStep)
		}

		log.Debugf("Validating workflow %s", loadedWorkflow.Name)
		if err := loadedWorkflow.Validate(); err != nil {
			log.Fatal(err)
		}

	}

	// Bootstrap env vars
	c.bootstrapEnvVars()

	// Bootstrap container mounts
	c.bootstrapMounts()

	log.Infof("Workspace ready")

	log.Debugf("Blocks: %d", len(workspace.Blocks))
	log.Debugf("Workflows: %d", len(workspace.Workflows))
}

func (c *Workspace) resolveBlockDependencies() error {
	missing := len(c.Blocks)

	log.Debug("Resolving Block dependencies")

	// Iterate over all Blocks in the Workspace
	// Until nothing is "missing" anymore
	for missing != 0 {
		for i := 0; i < len(c.Blocks); i++ {
			loadedBlock := &c.Blocks[i]

			// Block has not been resolved yet
			if !loadedBlock.resolved {

				log.Debugf("Trying to resolve Block '%s'", loadedBlock.Name)

				// Check if a "from:" stanza is given and not empty
				// This means that the loadedBlock should inherit from another Block
				if loadedBlock.From != "" {
					// a "from:" stanza is given
					log.Debugf("Block has dependency: '%s'", loadedBlock.From)

					// Try to load the referenced Block
					dependency := c.getBlockByName(loadedBlock.From)

					if dependency == nil {
						// There's no Block to load from
						return goErrors.New("Block '" + loadedBlock.From + "' not found in the Workspace. Please check the 'from' stanza of Block " + loadedBlock.Name)
					}

					// Check if the dependency Block has already been resolved
					// If not, re-queue the loaded Block so it can be resolved in another iteration
					if !dependency.resolved {
						// Needed Block from 'from' stanza is not yet resolved
						log.Debugf("Postponing Block '%s' because its dependency '%s' is not yet resolved", loadedBlock.Name, dependency.Name)
						loadedBlock.resolved = false
						continue
					}

					// Merge the dependency Block into the loaded Block
					// We do NOT OVERWRITE existing values in the loaded Block because we must assume
					// That this is configuration that has been explicitly set by the user
					// The merge works like "loading defaults" for the loaded Block
					err := loadedBlock.MergeIn(dependency)
					if err != nil {
						return err
					}

					// Handle Workdir
					// If the dependency Block has a workdir, apply this workdir to the loaded block
					if loadedBlock.Artifacts.LocalPath == "" {
						loadedBlock.Workdir.LocalPath = dependency.Workdir.LocalPath
						loadedBlock.Workdir.ContainerPath = dependency.Workdir.ContainerPath
					}

					// Handle Actions
					// Iterate over the loaded Block's Actions
					// and check if the Action exists in the dependency Block
					// If not, add it
					for _, loadedAction := range dependency.Actions {
						log.Debugf("Analyzing action '%s'", loadedAction.Name)
						existingAction := loadedBlock.getActionByName(loadedAction.Name)

						if existingAction != nil {
							log.Debugf("Found user-provided action '%s' in the workspace. Merging with loaded action", existingAction.Name)
							// An Action with the same name exists
							// We merge!
							err := existingAction.MergeIn(&loadedAction)
							if err != nil {
								return err
							}
						} else {
							log.Debugf("No user-provided action with that name found. Adding '%s' to the workspace.", loadedAction.Name)
							loadedBlock.Actions = append(loadedBlock.Actions, loadedAction)
						}
					}
					// opts := conjungo.NewOptions()
					// opts.Overwrite = false // do not overwrite existing values in workspaceConfig
					// if err := conjungo.Merge(loadedBlock, dependency, opts); err != nil {
					// 	return err
					// }
					loadedBlock.resolved = true
					loadedBlock.Parent = dependency
					missing--
					log.Debugf("Resolved Block '%s' from dependency '%s'", loadedBlock.Name, loadedBlock.From)
				} else {
					loadedBlock.resolved = true
					missing--
					log.Debugf("Resolved Block '%s'", loadedBlock.Name)
				}

				// Register the Block to the Index
				loadedBlock.address = strings.Join([]string{loadedBlock.Name}, ".")
				c.registerBlock(loadedBlock)

				// Update Artifacts, Kubeconfig & Inventory
				err := loadedBlock.LoadArtifacts()
				if err != nil {
					return err
				}
				loadedBlock.LoadInventory()
				loadedBlock.LoadKubeconfig()

				// Update Actions
				log.Debugf("Updating actions")
				for _, action := range loadedBlock.Actions {
					log.Debugf("Updating action %s of block %s", action.Name, loadedBlock.Name)

					existingAction := loadedBlock.getActionByName(action.Name)
					existingAction.address = strings.Join([]string{loadedBlock.Name, action.Name}, ".")

					blockCopy := *loadedBlock

					existingAction.Block = blockCopy

					// Remove nested actions
					existingAction.Block.Actions = nil

					// Register the Action to the Index
					c.registerAction(existingAction)
				}

				// Update Workflow addresses
				log.Debugf("%d Blocks left", missing)
			}

		}
	}
	return nil
}

func (c *Workspace) Validate() error {
	log.Debugf("Validating workspace struct")
	err := validate.Struct(c)

	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Error(err)
			return nil
		}

		for _, err := range err.(validator.ValidationErrors) {
			log.Error("Configuration option `" + strings.ToLower(err.Namespace()) + "` failed to validate: " + err.Tag())
		}

		// from here you can create your own error messages in whatever language you wish
		return goErrors.New("error validating Workspace")
	}
	return nil
}

func (c *Workspace) ListBlocks() {
	for _, block := range c.Blocks {
		fmt.Println(block.Name)
	}
}

func (c *Workspace) ListWorkflows() {
	for workflow := range c.index.Workflows {
		fmt.Println(workflow)
	}
}

func (c *Workspace) ListActions() {
	// for _, block := range c.Blocks {
	// 	for _, action := range block.Actions {
	// 		fmt.Printf("%s.%s\n", block.Name, action.Name)
	// 	}
	// }
	for action := range c.index.Actions {
		fmt.Println(action)
	}
}

func (c *Workspace) bootstrapIndex() {
	c.index = &WorkspaceIndex{}
	c.index.Actions = make(map[string]*Action)
	c.index.Blocks = make(map[string]*Block)
	c.index.Workflows = make(map[string]*Workflow)
	c.index.Steps = make(map[string]*Step)
}

func (c *Workspace) bootstrapMounts() {
	// Prepare mounts map
	c.mounts = map[string]string{}

	// Mount the Workspace
	c.registerMount(c.Path, c.ContainerPath)

	// Mount the Docker Socket if possible
	dockerSocket := "/var/run/docker.sock"
	if _, err := os.Stat(dockerSocket); !os.IsNotExist(err) {
		c.registerMount(dockerSocket, dockerSocket)
	}

	// if extraMounts != nil {

	// 	mounts = append(mounts, extraMounts...)
	// }
	// if devDir != "" {
	// 	mounts = append(mounts, strings.Join([]string{devDir, "/polycrate"}, ":"))
	// }
	for _, extraMount := range c.ExtraMounts {
		// Split by =
		p := strings.Split(extraMount, ":")

		if len(p) == 2 {
			c.registerMount(p[0], p[1])
		} else {
			c.err = goErrors.New("Illegal value for mount found: " + extraMount)
			break
		}
	}
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

func (c *Workspace) bootstrapEnvVars() {
	// Prepare env map
	c.env = map[string]string{}

	var _force string
	if force {
		_force = "1"
	} else {
		_force = "0"
	}
	c.registerEnvVar("ANSIBLE_DISPLAY_SKIPPED_HOSTS", "no")
	c.registerEnvVar("ANSIBLE_DISPLAY_OK_HOSTS", "yes")
	c.registerEnvVar("ANSIBLE_HOST_KEY_CHECKING", "no")
	c.registerEnvVar("ANSIBLE_ROLES_PATH", "/root/.ansible/roles:/usr/share/ansible/roles:/etc/ansible/roles")
	c.registerEnvVar("ANSIBLE_COLLECTIONS_PATH", "/root/.ansible/collections:/usr/share/ansible/collections:/etc/ansible/collections")
	c.registerEnvVar("ANSIBLE_VERBOSITY", logLevel)
	c.registerEnvVar("ANSIBLE_CALLBACKS_ENABLED", "timer,profile_tasks,profile_roles")
	c.registerEnvVar("POLYCRATE_CLI_VERSION", version)
	c.registerEnvVar("POLYCRATE_IMAGE_REFERENCE", workspace.Config.Image.Reference)
	c.registerEnvVar("POLYCRATE_IMAGE_VERSION", workspace.Config.Image.Version)
	c.registerEnvVar("POLYCRATE_FORCE", _force)
	c.registerEnvVar("POLYCRATE_VERSION", version)
	c.registerEnvVar("IN_CI", "true")
	c.registerEnvVar("TERM", "xterm-256color")

	if local {
		// Not in container
		c.registerEnvVar("POLYCRATE_WORKSPACE", workspace.Path)
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
			break
		}
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

func (c *Workspace) loadBlockConfigs() error {
	log.Debugf("Loading Blocks")

	for _, blockPath := range blockPaths {
		blockConfigFilePath := filepath.Join(blockPath, c.Config.BlocksConfig)

		blockConfigObject := viper.New()
		blockConfigObject.SetConfigType("yaml")
		blockConfigObject.SetConfigFile(blockConfigFilePath)

		log.Debug("Loading ", c.Config.BlocksConfig, " from "+blockPath)
		if err := blockConfigObject.MergeInConfig(); err != nil {
			return err
		}

		var loadedBlock Block
		if err := blockConfigObject.UnmarshalExact(&loadedBlock); err != nil {
			return err
		}
		if err := loadedBlock.Validate(); err != nil {
			return err
		}
		log.Debugf("Loaded Block '%s'", loadedBlock.Name)

		// Set Block vars
		loadedBlock.Workdir.LocalPath = blockPath
		loadedBlock.Workdir.ContainerPath = filepath.Join(c.ContainerPath, c.Config.BlocksRoot, loadedBlock.Name)

		// Check if Block exists
		existingBlock := c.getBlockByName(loadedBlock.Name)

		if existingBlock != nil {
			// Block exists
			log.Debugf("Found existing Block '%s' in the Workspace. Merging with loaded Block.", existingBlock.Name)

			err := existingBlock.MergeIn(&loadedBlock)
			if err != nil {
				return err
			}

			// Handle Actions
			// Iterate over the loaded Block's Actions
			// and check if it exists in the existing Block
			// If not, add it
			for _, loadedAction := range loadedBlock.Actions {
				log.Debugf("Analyzing Action '%s'", loadedAction.Name)
				existingAction := existingBlock.getActionByName(loadedAction.Name)

				if existingAction != nil {
					log.Debugf("Found existing Action '%s'", existingAction.Name)
					// An Action with the same name exists
					// We merge!
					err := existingAction.MergeIn(&loadedAction)
					if err != nil {
						return err
					}
				} else {
					log.Debugf("No existing Action found. Adding '%s'", loadedAction.Name)
					existingBlock.Actions = append(existingBlock.Actions, loadedAction)
				}
			}

		} else {
			c.Blocks = append(c.Blocks, loadedBlock)
		}
	}
	return nil

}

func (c *Workspace) discoverBlocks() error {
	blocksDir := filepath.Join(workspace.Path, workspace.Config.BlocksRoot)
	log.Debugf("Starting Block Discovery at %s", blocksDir)

	// This function adds all valid Blocks to the list of
	err := filepath.WalkDir(blocksDir, walkBlocksDir)
	if err != nil {
		return err
	}

	return nil
}
